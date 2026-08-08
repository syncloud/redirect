package inbound

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// startFakeMuxer answers CONNECT for the names it knows and joins the stream to
// upstream, the way frps tcpmux does. Only this file needs it: everything else
// takes a DeviceDialer.
func startFakeMuxer(upstream string, known ...string) (string, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	names := map[string]bool{}
	for _, n := range known {
		names[n] = true
	}
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer connection.Close()
				request, err := http.ReadRequest(bufio.NewReader(connection))
				if err != nil || request.Method != http.MethodConnect {
					return
				}
				host := request.Host
				if h, _, err := net.SplitHostPort(host); err == nil {
					host = h
				}
				if !names[host] {
					_, _ = connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
					return
				}
				device, err := net.Dial("tcp", upstream)
				if err != nil {
					return
				}
				defer device.Close()
				if _, err := connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
					return
				}
				go func() { _, _ = io.Copy(device, connection) }()
				_, _ = io.Copy(connection, device)
			}()
		}
	}()
	return listener.Addr().String(), func() { _ = listener.Close() }, nil
}

// greeter answers with an smtp banner the moment a connection is joined, which
// is what postfix does and what the CONNECT reply must not swallow
func startGreeter(banner string) (string, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go func() {
				defer connection.Close()
				_, _ = connection.Write([]byte(banner))
				_, _ = io.Copy(io.Discard, connection)
			}()
		}
	}()
	return listener.Addr().String(), func() { _ = listener.Close() }, nil
}

func TestTunnelDialer_JoinsTheDeviceStream(t *testing.T) {
	device, stopDevice, err := startGreeter("220 device ready\r\n")
	assert.NoError(t, err)
	defer stopDevice()
	muxer, stopMuxer, err := startFakeMuxer(device, "alice.syncloud.it")
	assert.NoError(t, err)
	defer stopMuxer()

	connection, err := NewTunnelDialer(muxer, 5*time.Second).Dial("alice.syncloud.it")

	assert.NoError(t, err)
	defer connection.Close()

	// the banner must survive the handshake, not be read ahead and lost
	banner := make([]byte, len("220 device ready\r\n"))
	_, err = io.ReadFull(connection, banner)
	assert.NoError(t, err)
	assert.Equal(t, "220 device ready\r\n", string(banner))
}

func TestTunnelDialer_UnknownDeviceRefused(t *testing.T) {
	device, stopDevice, err := startGreeter("220 device ready\r\n")
	assert.NoError(t, err)
	defer stopDevice()
	muxer, stopMuxer, err := startFakeMuxer(device, "alice.syncloud.it")
	assert.NoError(t, err)
	defer stopMuxer()

	_, err = NewTunnelDialer(muxer, 5*time.Second).Dial("bob.syncloud.it")

	assert.Error(t, err)
}

func TestTunnelDialer_NoMultiplexer(t *testing.T) {
	_, err := NewTunnelDialer("127.0.0.1:1", time.Second).Dial("alice.syncloud.it")

	assert.Error(t, err)
}
