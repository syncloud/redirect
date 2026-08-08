package inbound

import (
	"errors"
	"net"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/pires/go-proxyproto"
	"github.com/syncloud/redirect/mail"
	"go.uber.org/zap"
)

const DialTimeout = 30 * time.Second

type Server struct {
	server        *smtp.Server
	proxyProtocol bool
	logger        *zap.Logger
}

func NewServer(address string, hostname string, router *Router, dialer DeviceDialer,
	connections *mail.Connections, inFlight *mail.InFlight, maxMessageBytes int64,
	proxyProtocol bool, logger *zap.Logger) *Server {
	s := &Server{
		proxyProtocol: proxyProtocol,
		logger:        logger,
	}
	server := smtp.NewServer(smtp.BackendFunc(func(c *smtp.Conn) (smtp.Session, error) {
		peer := peerOf(c)
		if !connections.Acquire(peer) {
			logger.Info("inbound refused a connection", zap.String("peer", peer))
			return nil, mail.ErrTooManyConnections
		}
		return &Session{
			router: router, dialer: dialer, connections: connections, inFlight: inFlight,
			hostname: hostname, peer: peer, logger: logger,
		}, nil
	}))
	server.Addr = address
	server.Domain = hostname
	server.ReadTimeout = 5 * time.Minute
	server.WriteTimeout = 5 * time.Minute
	server.MaxMessageBytes = maxMessageBytes
	server.AllowInsecureAuth = false
	s.server = server
	return s
}

func (s *Server) Start() error {
	listener, err := s.listen()
	if err != nil {
		return err
	}

	s.logger.Info("inbound mail listening",
		zap.String("address", s.server.Addr), zap.Bool("proxy protocol", s.proxyProtocol))
	go func() {
		if err := s.server.Serve(listener); err != nil && !errors.Is(err, smtp.ErrServerClosed) {
			s.logger.Error("inbound mail stopped", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) listen() (net.Listener, error) {
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return nil, err
	}
	if s.proxyProtocol {
		return &proxyproto.Listener{Listener: listener, Policy: fromLoopbackOnly}, nil
	}
	return listener, nil
}

func fromLoopbackOnly(upstream net.Addr) (proxyproto.Policy, error) {
	host, _, err := net.SplitHostPort(upstream.String())
	if err != nil {
		return proxyproto.REJECT, err
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return proxyproto.REQUIRE, nil
	}
	return proxyproto.REJECT, nil
}

func (s *Server) Close() error {
	return s.server.Close()
}

func peerOf(c *smtp.Conn) string {
	if c == nil || c.Conn() == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(c.Conn().RemoteAddr().String())
	if err != nil {
		return c.Conn().RemoteAddr().String()
	}
	return host
}
