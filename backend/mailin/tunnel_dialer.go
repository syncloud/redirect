package mailin

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	// the port the device's postfix listens on behind the tunnel; frps strips
	// it when matching the CONNECT host, but it has to be well formed
	deviceSmtpPort = 25

	maxConnectResponse = 4096
)

// TunnelDialer reaches a device through the frps multiplexer, which routes on
// the name in an http CONNECT rather than giving every device a port.
type TunnelDialer struct {
	muxer   string
	timeout time.Duration
}

func NewTunnelDialer(muxer string, timeout time.Duration) *TunnelDialer {
	return &TunnelDialer{muxer: muxer, timeout: timeout}
}

func (d *TunnelDialer) Dial(domain string) (net.Conn, error) {
	connection, err := net.DialTimeout("tcp", d.muxer, d.timeout)
	if err != nil {
		return nil, err
	}
	if err := d.connect(connection, domain); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return connection, nil
}

// connect asks the multiplexer for the device's tunnel by name. The reply is
// read a byte at a time on purpose: buffering would swallow the greeting
// postfix sends the moment the stream is joined, and go-smtp's client takes the
// connection rather than a reader, so those bytes could not be handed back.
func (d *TunnelDialer) connect(connection net.Conn, domain string) error {
	if err := connection.SetDeadline(time.Now().Add(d.timeout)); err != nil {
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
