package inbound

import (
	"bufio"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
)

type fakeResolver struct {
	names map[string]string
}

func (f *fakeResolver) Name(address string) string {
	return f.names[address]
}

func testResolver() Resolver {
	return &fakeResolver{names: map[string]string{"203.0.113.1": "mail.example.net"}}
}

type xclientDevice struct {
	address    string
	advertise  bool
	mutex      sync.Mutex
	announced  string
	greetings  int
	helos      []string
	recipients []string
	body       string
}

func startXclientDevice(advertise bool) (*xclientDevice, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	device := &xclientDevice{address: listener.Addr().String(), advertise: advertise}
	go func() {
		for {
			connection, err := listener.Accept()
			if err != nil {
				return
			}
			go device.serve(connection)
		}
	}()
	return device, func() { _ = listener.Close() }, nil
}

func (d *xclientDevice) serve(connection net.Conn) {
	defer func() { _ = connection.Close() }()
	reader := bufio.NewReader(connection)
	write := func(line string) { _, _ = connection.Write([]byte(line + "\r\n")) }

	write("220 device.syncloud.it ESMTP")
	d.count()
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.TrimRight(line, "\r\n")
		upper := strings.ToUpper(command)
		switch {
		case strings.HasPrefix(upper, "EHLO"):
			d.addHelo(strings.TrimSpace(command[len("EHLO"):]))
			if d.advertise {
				write("250-device.syncloud.it")
				write("250 XCLIENT ADDR NAME HELO")
			} else {
				write("250 device.syncloud.it")
			}
		case strings.HasPrefix(upper, "XCLIENT"):
			d.record(strings.TrimSpace(command[len("XCLIENT"):]))
			write("220 device.syncloud.it ESMTP")
			d.count()
		case strings.HasPrefix(upper, "MAIL FROM"):
			write("250 2.1.0 Ok")
		case strings.HasPrefix(upper, "RCPT TO"):
			d.addRecipient(command)
			write("250 2.1.5 Ok")
		case strings.HasPrefix(upper, "DATA"):
			write("354 End data with <CR><LF>.<CR><LF>")
			d.readBody(reader)
			write("250 2.0.0 Ok")
		case strings.HasPrefix(upper, "QUIT"):
			write("221 2.0.0 Bye")
			return
		default:
			write("250 2.0.0 Ok")
		}
	}
}

func (d *xclientDevice) readBody(reader *bufio.Reader) {
	var body strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if strings.TrimRight(line, "\r\n") == "." {
			break
		}
		body.WriteString(line)
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.body = body.String()
}

func (d *xclientDevice) record(attributes string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.announced = attributes
}

func (d *xclientDevice) count() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.greetings++
}

func (d *xclientDevice) addHelo(name string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.helos = append(d.helos, name)
}

func (d *xclientDevice) greeted() []string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return append([]string(nil), d.helos...)
}

func (d *xclientDevice) addRecipient(command string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.recipients = append(d.recipients, command)
}

func (d *xclientDevice) got() (string, int, string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.announced, d.greetings, d.body
}

func TestXclient_AnnouncesTheRealClientToTheDevice(t *testing.T) {
	device, stopDevice, err := startXclientDevice(true)
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(&fakeDialer{devices: map[string]string{"alice.syncloud.it": device.address}},
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "user@alice.syncloud.it"))

	announced, _, body := device.got()
	assert.Contains(t, announced, "ADDR=203.0.113.1")
	assert.Contains(t, announced, "NAME=mail.example.net")
	assert.Contains(t, announced, "HELO=localhost")
	assert.Contains(t, body, "Subject: hello")
}

func TestXclient_DeviceIsGreetedWithTheClientHeloAfterTheReset(t *testing.T) {
	device, stopDevice, err := startXclientDevice(true)
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(&fakeDialer{devices: map[string]string{"alice.syncloud.it": device.address}},
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "user@alice.syncloud.it"))

	assert.Equal(t, []string{"mx.syncloud.it", "localhost"}, device.greeted())
}

func TestXclient_DeviceWithoutTheExtensionKeepsTheRelayHelo(t *testing.T) {
	device, stopDevice, err := startXclientDevice(false)
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(&fakeDialer{devices: map[string]string{"alice.syncloud.it": device.address}},
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "user@alice.syncloud.it"))

	assert.Equal(t, []string{"mx.syncloud.it", "mx.syncloud.it"}, device.greeted())
}

func TestXclient_DeviceWithoutTheExtensionStillGetsTheMessage(t *testing.T) {
	device, stopDevice, err := startXclientDevice(false)
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(&fakeDialer{devices: map[string]string{"alice.syncloud.it": device.address}},
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "user@alice.syncloud.it"))

	announced, greetings, body := device.got()
	assert.Equal(t, "", announced)
	assert.Equal(t, 1, greetings)
	assert.Contains(t, body, "Subject: hello")
}

func TestXclient_UnresolvedClientIsReportedUnavailable(t *testing.T) {
	device, stopDevice, err := startXclientDevice(true)
	assert.NoError(t, err)
	defer stopDevice()
	connection, err := net.Dial("tcp", device.address)
	assert.NoError(t, err)
	defer func() { _ = connection.Close() }()

	announced, helo, err := Announce(connection, "mx.syncloud.it", Client{Address: "198.51.100.7"})
	assert.NoError(t, err)
	assert.Equal(t, "mx.syncloud.it", helo)
	defer func() { _ = announced.Close() }()

	attributes, greetings, _ := device.got()
	assert.Contains(t, attributes, "ADDR=198.51.100.7")
	assert.Contains(t, attributes, "NAME="+Unavailable)
	assert.NotContains(t, attributes, "HELO=")
	assert.Equal(t, 2, greetings)
}
