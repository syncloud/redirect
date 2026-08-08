package main

import (
	"errors"
	"io"
	"net"

	"github.com/emersion/go-smtp"
)

var ErrConnectionDropped = errors.New("connection dropped on purpose")

type Session struct {
	mailbox    *Mailbox
	connection net.Conn
	recipients []string
}

func (s *Session) Mail(_ string, _ *smtp.MailOptions) error {
	return nil
}

func (s *Session) Rcpt(to string, _ *smtp.RcptOptions) error {
	if s.mailbox.Behaviour().Rcpt == Reject {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "no such user here",
		}
	}
	s.recipients = append(s.recipients, to)
	return nil
}

func (s *Session) Data(reader io.Reader) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	switch s.mailbox.Behaviour().Data {
	case Reject:
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 6, 0},
			Message:      "message refused",
		}
	case Drop:
		_ = s.connection.Close()
		return ErrConnectionDropped
	}
	s.mailbox.Add(Message{Recipients: s.recipients, Body: string(body)})
	return nil
}

func (s *Session) Reset() {
	s.recipients = nil
}

func (s *Session) Logout() error {
	return nil
}
