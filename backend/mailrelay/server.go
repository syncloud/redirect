package mailrelay

import (
	"errors"
	"io"
	"time"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type Server struct {
	relay       *Relay
	sender      Sender
	certificate *Certificate
	server      *smtp.Server
	logger      *zap.Logger
}

func NewServer(address string, domain string, relay *Relay, sender Sender, scanner Scanner,
	limiter *Limiter, certificate *Certificate, maxMessageBytes int64, logger *zap.Logger) *Server {
	s := &Server{relay: relay, sender: sender, certificate: certificate, logger: logger}
	server := smtp.NewServer(smtp.BackendFunc(func(_ *smtp.Conn) (smtp.Session, error) {
		return &Session{relay: relay, sender: sender, scanner: scanner, limiter: limiter, logger: logger}, nil
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
	// credentials travel in plain text under AUTH PLAIN, so without a
	// certificate of our own this listener must stay on loopback behind caddy
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

type Session struct {
	relay      *Relay
	sender     Sender
	scanner    Scanner
	limiter    *Limiter
	logger     *zap.Logger
	domain     *model.Domain
	from       string
	recipients []string
}

func (s *Session) AuthMechanisms() []string {
	return []string{sasl.Plain}
}

func (s *Session) Auth(_ string) (sasl.Server, error) {
	return sasl.NewPlainServer(func(_, login, password string) error {
		domain, err := s.relay.Authorize(login, password)
		if err != nil {
			s.logger.Info("mail relay auth rejected", zap.String("login", login), zap.Error(err))
			return err
		}
		s.domain = domain
		return nil
	}), nil
}

func (s *Session) Mail(from string, _ *smtp.MailOptions) error {
	if s.domain == nil {
		return smtp.ErrAuthRequired
	}
	if !s.relay.Allowed(s.domain, from) {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 7, 1},
			Message:      "sender does not belong to this device domain",
		}
	}
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string, _ *smtp.RcptOptions) error {
	if s.domain == nil {
		return smtp.ErrAuthRequired
	}
	s.recipients = append(s.recipients, to)
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if s.domain == nil {
		return smtp.ErrAuthRequired
	}
	if err := s.limiter.AllowRecipients(len(s.recipients)); err != nil {
		return err
	}
	if err := s.limiter.Allow(s.domain.Name, int64(len(s.recipients))); err != nil {
		s.logger.Info("mail relay rate limited",
			zap.String("domain", s.domain.Name), zap.Error(err))
		return err
	}
	message, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err := s.scanner.Scan(s.from, s.recipients, s.domain.Name, message); err != nil {
		return err
	}
	if err := s.sender.Send(s.from, s.recipients, message); err != nil {
		s.logger.Error("mail relay send failed", zap.String("domain", s.domain.Name), zap.Error(err))
		return err
	}
	s.logger.Info("mail relay sent",
		zap.String("domain", s.domain.Name), zap.Int("recipients", len(s.recipients)))
	return s.relay.Sent(s.domain, len(s.recipients))
}

func (s *Session) Reset() {
	s.from = ""
	s.recipients = nil
}

func (s *Session) Logout() error {
	return nil
}
