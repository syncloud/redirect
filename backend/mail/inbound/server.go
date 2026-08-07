package inbound

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/pires/go-proxyproto"
	"github.com/syncloud/redirect/mail"
	"go.uber.org/zap"
)

const (
	DialTimeout     = 30 * time.Second
	certificateWait = 30 * time.Second
)

type Server struct {
	server          *smtp.Server
	certificate     *mail.CertificateLoader
	certificateWait time.Duration
	proxyProtocol   bool
	stopped         chan struct{}
	logger          *zap.Logger
}

func NewServer(address string, hostname string, router *Router, dialer DeviceDialer,
	connections *mail.Connections, inFlight *mail.InFlight,
	certificate *mail.CertificateLoader, maxMessageBytes int64,
	proxyProtocol bool, logger *zap.Logger) *Server {
	s := &Server{
		certificateWait: certificateWait,
		proxyProtocol:   proxyProtocol,
		stopped:         make(chan struct{}),
		logger:          logger,
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
	s.certificate = certificate
	return s
}

func (s *Server) Start() error {
	tlsConfig, err := s.certificate.Load()
	if errors.Is(err, mail.ErrCertificateMissing) {
		s.logger.Warn("waiting for the inbound mail certificate before accepting mail",
			zap.String("address", s.server.Addr), zap.Duration("retry", s.certificateWait))
		go s.serveOnceCertified()
		return nil
	}
	if err != nil {
		return fmt.Errorf("inbound mail certificate: %w", err)
	}
	return s.serve(tlsConfig)
}

func (s *Server) serveOnceCertified() {
	for {
		select {
		case <-s.stopped:
			return
		case <-time.After(s.certificateWait):
		}
		tlsConfig, err := s.certificate.Load()
		if errors.Is(err, mail.ErrCertificateMissing) {
			s.logger.Warn("still no inbound mail certificate, not accepting mail yet",
				zap.String("address", s.server.Addr))
			continue
		}
		if err != nil {
			s.logger.Error("inbound mail certificate cannot be read, not accepting mail",
				zap.Error(err))
			continue
		}
		if err := s.serve(tlsConfig); err != nil {
			s.logger.Error("inbound mail cannot listen", zap.Error(err))
		}
		return
	}
}

func (s *Server) serve(tlsConfig *tls.Config) error {
	s.server.TLSConfig = tlsConfig

	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	if s.proxyProtocol {
		listener = &proxyproto.Listener{Listener: listener, Policy: fromLoopbackOnly}
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
	close(s.stopped)
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
