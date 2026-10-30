package main

import (
	"log"
	"net"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/pires/go-proxyproto"
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
	tcp, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return err
	}
	listener := &proxyproto.Listener{Listener: tcp, Policy: func(net.Addr) (proxyproto.Policy, error) {
		return proxyproto.REQUIRE, nil
	}}
	go func() {
		if err := s.server.Serve(listener); err != nil {
			log.Printf("device smtp stopped: %v", err)
		}
	}()
	return nil
}
