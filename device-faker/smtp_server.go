package main

import (
	"log"
	"time"

	"github.com/emersion/go-smtp"
)

type SmtpServer struct {
	server *smtp.Server
}

func NewSmtpServer(address string, mailbox *Mailbox) *SmtpServer {
	server := smtp.NewServer(smtp.BackendFunc(func(c *smtp.Conn) (smtp.Session, error) {
		return &Session{mailbox: mailbox, connection: c.Conn()}, nil
	}))
	server.Addr = address
	server.Domain = "device.faker"
	server.ReadTimeout = time.Minute
	server.WriteTimeout = time.Minute
	server.AllowInsecureAuth = false
	return &SmtpServer{server: server}
}

func (s *SmtpServer) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Printf("device smtp stopped: %v", err)
		}
	}()
	return nil
}
