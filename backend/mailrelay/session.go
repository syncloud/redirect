package mailrelay

import (
	"errors"
	"io"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type Session struct {
	relay       *Relay
	sender      Sender
	scanner     Scanner
	limiter     *Limiter
	connections *Connections
	inFlight    *InFlight
	peer        string
	logger      *zap.Logger
	domain      *model.Domain
	from        string
	recipients  []string
}

func (s *Session) AuthMechanisms() []string {
	return []string{sasl.Plain}
}

func (s *Session) Auth(_ string) (sasl.Server, error) {
	return sasl.NewPlainServer(func(_, login, password string) error {
		domain, err := s.relay.Authorize(login, password)
		if err != nil {
			s.logger.Info("mail relay auth rejected", zap.String("login", login), zap.Error(err))
			return authError(err)
		}
		s.domain = domain
		return nil
	}), nil
}

func authError(err error) error {
	switch {
	case errors.Is(err, ErrUnknownToken), errors.Is(err, ErrNotOwned),
		errors.Is(err, ErrBlocked), errors.Is(err, ErrNotAllowed):
		return authRejected(err)
	case errors.Is(err, ErrOverLimit):
		return authTryAgain(err)
	default:
		return authTryAgain(errors.New("temporary authentication failure"))
	}
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
		return permanent(err, smtp.EnhancedCode{5, 5, 3})
	}
	if err := s.limiter.Allow(s.domain.Name, int64(len(s.recipients))); err != nil {
		s.logger.Info("mail relay rate limited",
			zap.String("domain", s.domain.Name), zap.Error(err))
		return tryAgain(err, smtp.EnhancedCode{4, 7, 1})
	}
	message, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err := s.scanner.Scan(s.from, s.recipients, s.domain.Name, message); err != nil {
		if errors.Is(err, ErrRejectedAsSpam) {
			return permanent(err, smtp.EnhancedCode{5, 7, 1})
		}
		return tryAgain(err, smtp.EnhancedCode{4, 7, 0})
	}
	if !s.inFlight.Acquire() {
		s.logger.Info("mail relay at capacity", zap.String("domain", s.domain.Name))
		return tryAgain(ErrBusy, smtp.EnhancedCode{4, 3, 2})
	}
	defer s.inFlight.Release()
	if err := s.sender.Send(s.from, s.recipients, message); err != nil {
		s.logger.Error("mail relay send failed", zap.String("domain", s.domain.Name), zap.Error(err))
		return tryAgain(err, smtp.EnhancedCode{4, 4, 0})
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
	s.connections.Release(s.peer)
	return nil
}
