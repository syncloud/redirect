package smtp

import (
	"gopkg.in/gomail.v2"
)

type Smtp struct {
	host     string
	port     int
	tls      bool
	login    string
	password string
}

func NewSmtp(host string, port int, tls bool, login string, password string) *Smtp {
	return &Smtp{
		host:     host,
		port:     port,
		tls:      tls,
		login:    login,
		password: password,
	}
}

func (s Smtp) Send(from string, contentType string, body string, subject string, to ...string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", from)
	message.SetHeader("To", to...)
	message.SetHeader("Subject", subject)
	message.SetBody(contentType, body)
	dialer := gomail.Dialer{Host: s.host, Port: s.port}
	if s.tls {
		dialer.Username = s.login
		dialer.Password = s.password
	}
	err := dialer.DialAndSend(message)
	return err
}
