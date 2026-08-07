package outbound

import (
	"errors"
	"net"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/syncloud/redirect/mail"
	"go.uber.org/zap"
)

type Server struct {
	relay       *Relay
	sender      Sender
	certificate *mail.CertificateLoader
	server      *smtp.Server
	logger      *zap.Logger
}

func NewServer(address string, domain string, relay *Relay, sender Sender, scanner Scanner,
	limiter *Limiter, connections *mail.Connections, inFlight *mail.InFlight, certificate *mail.CertificateLoader,
	maxMessageBytes int64, logger *zap.Logger) *Server {
	s := &Server{relay: relay, sender: sender, certificate: certificate, logger: logger}
	server := smtp.NewServer(smtp.BackendFunc(func(c *smtp.Conn) (smtp.Session, error) {
		peer := peerOf(c)
		if !connections.Acquire(peer) {
			logger.Info("mail relay refused a connection", zap.String("peer", peer))
			return nil, mail.ErrTooManyConnections
		}
		return &Session{
			relay: relay, sender: sender, scanner: scanner, limiter: limiter,
			connections: connections, inFlight: inFlight, peer: peer, logger: logger,
		}, nil
	}))
	server.Addr = address
	server.Domain = domain
	server.ReadTimeout = time.Minute
	server.WriteTimeout = time.Minute
	server.MaxMessageBytes = maxMessageBytes
	s.server = server
	return s
}

func (s *Server) Start() error {
	tlsConfig, err := s.certificate.Load()
	if err != nil {
		return err
	}
	s.server.TLSConfig = tlsConfig
	s.server.AllowInsecureAuth = tlsConfig == nil
	s.logger.Info("mail relay listening", zap.String("address", s.server.Addr))
	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, smtp.ErrServerClosed) {
			s.logger.Error("mail relay stopped", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) Close() error {
	return s.server.Close()
}

func tryAgain(err error, code smtp.EnhancedCode) error {
	return &smtp.SMTPError{Code: 451, EnhancedCode: code, Message: err.Error()}
}

func permanent(err error, code smtp.EnhancedCode) error {
	return &smtp.SMTPError{Code: 550, EnhancedCode: code, Message: err.Error()}
}

func authRejected(err error) error {
	return &smtp.SMTPError{Code: 535, EnhancedCode: smtp.EnhancedCode{5, 7, 8}, Message: err.Error()}
}

func authTryAgain(err error) error {
	return &smtp.SMTPError{Code: 454, EnhancedCode: smtp.EnhancedCode{4, 7, 0}, Message: err.Error()}
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
