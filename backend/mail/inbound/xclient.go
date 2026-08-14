package inbound

import (
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
)

const Unavailable = "[UNAVAILABLE]"

type Client struct {
	Address string
	Name    string
	Helo    string
}

type greeted struct {
	net.Conn
	reader io.Reader
}

func (g *greeted) Read(p []byte) (int, error) {
	return g.reader.Read(p)
}

func Announce(connection net.Conn, hostname string, client Client) (net.Conn, error) {
	text := textproto.NewConn(connection)
	if _, _, err := text.ReadResponse(220); err != nil {
		return nil, err
	}
	if err := text.PrintfLine("EHLO %s", hostname); err != nil {
		return nil, err
	}
	_, capabilities, err := text.ReadResponse(250)
	if err != nil {
		return nil, err
	}
	if supports(capabilities, "XCLIENT") {
		if err := text.PrintfLine("XCLIENT %s", attributes(client)); err != nil {
			return nil, err
		}
		if _, _, err := text.ReadResponse(220); err != nil {
			return nil, err
		}
	}
	greeting := strings.NewReader(fmt.Sprintf("220 %s ESMTP\r\n", hostname))
	return &greeted{Conn: connection, reader: io.MultiReader(greeting, text.R)}, nil
}

func supports(capabilities string, extension string) bool {
	for _, line := range strings.Split(capabilities, "\n") {
		name, _, _ := strings.Cut(strings.TrimSpace(line), " ")
		if strings.EqualFold(name, extension) {
			return true
		}
	}
	return false
}

func attributes(client Client) string {
	values := []string{"ADDR=" + or(client.Address), "NAME=" + or(client.Name)}
	if client.Helo != "" {
		values = append(values, "HELO="+client.Helo)
	}
	return strings.Join(values, " ")
}

func or(value string) string {
	if value == "" {
		return Unavailable
	}
	return value
}
