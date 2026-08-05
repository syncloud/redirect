package mailin

import (
	"errors"
	"net"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/syncloud/redirect/mailnet"
	"go.uber.org/zap"
)

const dialTimeout = 30 * time.Second

type Server struct {
	server *smtp.Server
	logger *zap.Logger
}

func NewServer(address string, hostname string, router *Router, connections *mailnet.Connections,
	inFlight *mailnet.InFlight, certificate *mailnet.CertificateLoader, maxMessageBytes int64,
	logger *zap.Logger) *Server {
	s := &Server{logger: logger}
	server := smtp.NewServer(smtp.BackendFunc(func(c *smtp.Conn) (smtp.Session, error) {
		peer := peerOf(c)
		if !connections.Acquire(peer) {
			logger.Info("inbound refused a connection", zap.String("peer", peer))
			return nil, mailnet.ErrTooManyConnections
		}
		return &Session{
			router: router, connections: connections, inFlight: inFlight,
			hostname: hostname, dialTimeout: dialTimeout, peer: peer, logger: logger,
		}, nil
	}))
	server.Addr = address
	server.Domain = hostname
	server.ReadTimeout = 5 * time.Minute
	server.WriteTimeout = 5 * time.Minute
	server.MaxMessageBytes = maxMessageBytes
	server.AllowInsecureAuth = false
	s.certificate(server, certificate)
	s.server = server
	return s
}

func (s *Server) certificate(server *smtp.Server, certificate *mailnet.CertificateLoader) {
	config, err := certificate.Load()
	if err != nil {
		s.logger.Error("inbound certificate load failed, starting without starttls", zap.Error(err))
		return
	}
	server.TLSConfig = config
}

func (s *Server) Start() error {
	s.logger.Info("inbound mail listening", zap.String("address", s.server.Addr))
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, smtp.ErrServerClosed) {
			s.logger.Error("inbound mail stopped", zap.Error(err))
		}
	}()
	return nil
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
