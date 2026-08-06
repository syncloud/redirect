package mailin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/syncloud/redirect/mailnet"
	"go.uber.org/zap"
)

var (
	ErrUnreachable   = errors.New("device is not reachable, try again later")
	ErrNoTunnel      = errors.New("device has no inbound tunnel, try again later")
	ErrNoRecipients  = errors.New("no valid recipients")
	ErrMixedDomains  = errors.New("too many recipients, send to one domain per message")
	ErrDeviceFailure = errors.New("device connection failed")
)

const (
	// the port the device's postfix listens on behind the tunnel; frps strips
	// it when matching the CONNECT host, but it has to be well formed
	deviceSmtpPort = 25

	maxConnectResponse = 4096
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
	connection, err := net.DialTimeout("tcp", route.Muxer, s.dialTimeout)
	if err != nil {
		s.logger.Info("cannot reach the tunnel multiplexer",
			zap.String("muxer", route.Muxer), zap.Error(err))
		return tryAgain(ErrUnreachable, smtp.EnhancedCode{4, 4, 1})
	}
	if err := connectToDevice(connection, route.Domain, s.dialTimeout); err != nil {
		_ = connection.Close()
		s.logger.Info("device has no tunnel open",
			zap.String("domain", route.Domain), zap.Error(err))
		return tryAgain(ErrNoTunnel, smtp.EnhancedCode{4, 4, 1})
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

// connectToDevice asks frps for the device's tunnel by name. The reply is read
// a byte at a time on purpose: buffering would swallow the greeting postfix
// sends the moment the stream is joined, and go-smtp's client takes the
// connection rather than a reader, so those bytes could not be handed back.
func connectToDevice(connection net.Conn, domain string, timeout time.Duration) error {
	if err := connection.SetDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	defer func() { _ = connection.SetDeadline(time.Time{}) }()

	target := net.JoinHostPort(domain, strconv.Itoa(deviceSmtpPort))
	request := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", target, target)
	if _, err := connection.Write([]byte(request)); err != nil {
		return err
	}

	var head []byte
	one := make([]byte, 1)
	for len(head) < maxConnectResponse {
		if _, err := io.ReadFull(connection, one); err != nil {
			return err
		}
		head = append(head, one[0])
		if bytes.HasSuffix(head, []byte("\r\n\r\n")) {
			break
		}
	}
	status := string(head)
	if end := strings.IndexByte(status, '\n'); end > 0 {
		status = strings.TrimSpace(status[:end])
	}
	if !strings.Contains(status, " 200 ") {
		return fmt.Errorf("tunnel refused the connection: %s", status)
	}
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
	default:
		return tryAgain(err, smtp.EnhancedCode{4, 3, 0})
	}
}

func tryAgain(err error, code smtp.EnhancedCode) error {
	return &smtp.SMTPError{Code: 451, EnhancedCode: code, Message: err.Error()}
}
