package mailin

import (
	"errors"
	"io"
	"strings"

	"github.com/emersion/go-smtp"
	"github.com/syncloud/redirect/mailnet"
	"go.uber.org/zap"
)

var (
	ErrUnreachable   = errors.New("device is not reachable, try again later")
	ErrNoRecipients  = errors.New("no valid recipients")
	ErrMixedDomains  = errors.New("too many recipients, send to one domain per message")
	ErrDeviceFailure = errors.New("device connection failed")
)

type Session struct {
	router      *Router
	dialer      DeviceDialer
	connections *mailnet.Connections
	inFlight    *mailnet.InFlight
	hostname    string
	peer        string
	logger      *zap.Logger

	from   string
	domain string
	client *smtp.Client
}

func (s *Session) Mail(from string, _ *smtp.MailOptions) error {
	s.release()
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string, _ *smtp.RcptOptions) error {
	domain, err := s.router.Route(to)
	if err != nil {
		s.logger.Info("inbound recipient rejected", zap.String("to", to), zap.Error(err))
		return routeError(err)
	}
	if s.domain == "" {
		if err := s.connect(domain); err != nil {
			return err
		}
	} else if !strings.EqualFold(domain, s.domain) {
		s.logger.Info("inbound second domain deferred",
			zap.String("bound", s.domain), zap.String("deferred", domain))
		return &smtp.SMTPError{
			Code:         452,
			EnhancedCode: smtp.EnhancedCode{4, 5, 3},
			Message:      ErrMixedDomains.Error(),
		}
	}
	if err := s.client.Rcpt(to, nil); err != nil {
		return relayError(err)
	}
	return nil
}

func (s *Session) Data(r io.Reader) error {
	if s.client == nil {
		return &smtp.SMTPError{
			Code:         503,
			EnhancedCode: smtp.EnhancedCode{5, 5, 1},
			Message:      ErrNoRecipients.Error(),
		}
	}
	if !s.inFlight.Acquire() {
		return tryAgain(mailnet.ErrBusy, smtp.EnhancedCode{4, 3, 2})
	}
	defer s.inFlight.Release()

	writer, err := s.client.Data()
	if err != nil {
		return relayError(err)
	}
	if _, err := io.Copy(writer, r); err != nil {
		_ = writer.Close()
		return relayError(err)
	}
	if err := writer.Close(); err != nil {
		s.logger.Info("inbound rejected by device",
			zap.String("domain", s.domain), zap.Error(err))
		return relayError(err)
	}
	s.logger.Info("inbound delivered", zap.String("domain", s.domain))
	return nil
}

func (s *Session) connect(domain string) error {
	connection, err := s.dialer.Dial(domain)
	if err != nil {
		s.logger.Info("cannot reach the device through the tunnel",
			zap.String("domain", domain), zap.Error(err))
		return tryAgain(ErrUnreachable, smtp.EnhancedCode{4, 4, 1})
	}
	client := smtp.NewClient(connection)
	if err := client.Hello(s.hostname); err != nil {
		_ = client.Close()
		return relayError(err)
	}
	if err := client.Mail(s.from, nil); err != nil {
		_ = client.Close()
		return relayError(err)
	}
	s.client = client
	s.domain = domain
	return nil
}

func (s *Session) Reset() {
	s.release()
	s.from = ""
}

func (s *Session) Logout() error {
	s.release()
	s.connections.Release(s.peer)
	return nil
}

func (s *Session) release() {
	if s.client != nil {
		_ = s.client.Quit()
		s.client = nil
	}
	s.domain = ""
}

func relayError(err error) error {
	var relayed *smtp.SMTPError
	if errors.As(err, &relayed) {
		return relayed
	}
	return tryAgain(ErrDeviceFailure, smtp.EnhancedCode{4, 4, 2})
}

func routeError(err error) error {
	switch {
	case errors.Is(err, ErrNoSuchDomain), errors.Is(err, ErrNotAccepted):
		return &smtp.SMTPError{Code: 550, EnhancedCode: smtp.EnhancedCode{5, 1, 2}, Message: err.Error()}
	default:
		return tryAgain(err, smtp.EnhancedCode{4, 3, 0})
	}
}

func tryAgain(err error, code smtp.EnhancedCode) error {
	return &smtp.SMTPError{Code: 451, EnhancedCode: code, Message: err.Error()}
}
