package mailin

import (
	"errors"
	"io"
	"net"
	"strings"
	"time"

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
	connections *mailnet.Connections
	inFlight    *mailnet.InFlight
	hostname    string
	dialTimeout time.Duration
	peer        string
	logger      *zap.Logger

	from   string
	route  *Route
	client *smtp.Client
}

func (s *Session) Mail(from string, _ *smtp.MailOptions) error {
	s.release()
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string, _ *smtp.RcptOptions) error {
	route, err := s.router.Route(to)
	if err != nil {
		s.logger.Info("inbound recipient rejected", zap.String("to", to), zap.Error(err))
		return routeError(err)
	}
	if s.route == nil {
		if err := s.connect(route); err != nil {
			return err
		}
	} else if !strings.EqualFold(route.Domain, s.route.Domain) {
		s.logger.Info("inbound second domain deferred",
			zap.String("bound", s.route.Domain), zap.String("deferred", route.Domain))
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
			zap.String("domain", s.route.Domain), zap.Error(err))
		return relayError(err)
	}
	s.logger.Info("inbound delivered", zap.String("domain", s.route.Domain))
	return nil
}

func (s *Session) connect(route *Route) error {
	connection, err := net.DialTimeout("tcp", route.Address, s.dialTimeout)
	if err != nil {
		s.logger.Info("inbound device is not reachable",
			zap.String("domain", route.Domain), zap.Error(err))
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
	s.route = route
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
	s.route = nil
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
	case errors.Is(err, ErrNoDeviceRoute):
		return tryAgain(ErrNoDeviceRoute, smtp.EnhancedCode{4, 3, 0})
	default:
		return tryAgain(err, smtp.EnhancedCode{4, 3, 0})
	}
}

func tryAgain(err error, code smtp.EnhancedCode) error {
	return &smtp.SMTPError{Code: 451, EnhancedCode: code, Message: err.Error()}
}
