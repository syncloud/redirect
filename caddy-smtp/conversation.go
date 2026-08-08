package caddysmtp

import (
	"bufio"
	"crypto/tls"
	"errors"
	"net"
	"time"

	"go.uber.org/zap"
)

const negotiationTimeout = 2 * time.Minute

var ErrPipelinedUpgrade = errors.New("client pipelined data before the tls upgrade")

type Conversation struct {
	hostname  string
	upstream  Upstream
	tlsConfig func() *tls.Config
	logger    *zap.Logger
}

func NewConversation(hostname string, upstream Upstream, tlsConfig func() *tls.Config,
	logger *zap.Logger) *Conversation {
	return &Conversation{hostname: hostname, upstream: upstream, tlsConfig: tlsConfig, logger: logger}
}

func (c *Conversation) Serve(client net.Conn) error {
	remote, local := client.RemoteAddr(), client.LocalAddr()
	if err := client.SetDeadline(time.Now().Add(negotiationTimeout)); err != nil {
		return err
	}
	reader := newReader(client)
	if err := writeLines(client, "220 "+c.hostname+" ESMTP"); err != nil {
		return err
	}
	greeting := ""
	for {
		line, err := readLine(reader)
		if err != nil {
			return err
		}
		switch command(line) {
		case "EHLO":
			greeting = line
			if err := writeLines(client, c.capabilities()...); err != nil {
				return err
			}
		case "HELO":
			greeting = line
			if err := writeLines(client, "250 "+c.hostname); err != nil {
				return err
			}
		case "STARTTLS":
			if reader.Buffered() > 0 {
				return ErrPipelinedUpgrade
			}
			if err := writeLines(client, "220 2.0.0 Ready to start TLS"); err != nil {
				return err
			}
			secure := tls.Server(client, c.tlsConfig())
			if err := secure.Handshake(); err != nil {
				return err
			}
			return c.relay(secure, newReader(secure), remote, local, "", "")
		case "QUIT":
			return writeLines(client, "221 2.0.0 Bye")
		default:
			return c.relay(client, reader, remote, local, greeting, line)
		}
	}
}

func (c *Conversation) capabilities() []string {
	return []string{
		"250-" + c.hostname,
		"250-PIPELINING",
		"250-8BITMIME",
		"250-ENHANCEDSTATUSCODES",
		"250 STARTTLS",
	}
}

func (c *Conversation) relay(client net.Conn, clientReader *bufio.Reader, remote net.Addr,
	local net.Addr, greeting string, pending string) error {
	server, err := c.upstream.Connect(remote, local)
	if err != nil {
		return err
	}
	serverReader := newReader(server)
	if _, err := readResponse(serverReader); err != nil {
		_ = server.Close()
		return err
	}
	if greeting != "" {
		if err := writeLines(server, greeting); err != nil {
			_ = server.Close()
			return err
		}
		if _, err := readResponse(serverReader); err != nil {
			_ = server.Close()
			return err
		}
	}
	if pending != "" {
		if err := writeLines(server, pending); err != nil {
			_ = server.Close()
			return err
		}
	}
	if err := client.SetDeadline(time.Time{}); err != nil {
		_ = server.Close()
		return err
	}
	return pipe(client, clientReader, server, serverReader)
}
