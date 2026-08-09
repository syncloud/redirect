package inbound

import (
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"testing"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/pires/go-proxyproto"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/mail"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeDevice struct {
	address string
	rcptErr error
	dataErr error

	mutex      sync.Mutex
	from       string
	recipients []string
	body       string
}

type deviceSession struct {
	device *fakeDevice
}

func (s *deviceSession) Mail(from string, _ *gosmtp.MailOptions) error {
	s.device.mutex.Lock()
	defer s.device.mutex.Unlock()
	s.device.from = from
	return nil
}

func (s *deviceSession) Rcpt(to string, _ *gosmtp.RcptOptions) error {
	if s.device.rcptErr != nil {
		return s.device.rcptErr
	}
	s.device.mutex.Lock()
	defer s.device.mutex.Unlock()
	s.device.recipients = append(s.device.recipients, to)
	return nil
}

func (s *deviceSession) Data(r io.Reader) error {
	body, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if s.device.dataErr != nil {
		return s.device.dataErr
	}
	s.device.mutex.Lock()
	defer s.device.mutex.Unlock()
	s.device.body = string(body)
	return nil
}

func (s *deviceSession) Reset()        {}
func (s *deviceSession) Logout() error { return nil }

func (d *fakeDevice) got() ([]string, string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return append([]string{}, d.recipients...), d.body
}

func startDevice(device *fakeDevice) (*fakeDevice, func(), error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	server := gosmtp.NewServer(gosmtp.BackendFunc(func(_ *gosmtp.Conn) (gosmtp.Session, error) {
		return &deviceSession{device: device}, nil
	}))
	server.Domain = "device.syncloud.it"
	go func() { _ = server.Serve(listener) }()
	device.address = listener.Addr().String()
	return device, func() { _ = server.Close() }, nil
}

type fakeDialer struct {
	devices map[string]string
}

func (d *fakeDialer) Dial(domain string) (net.Conn, error) {
	address, ok := d.devices[domain]
	if !ok {
		return nil, fmt.Errorf("no tunnel for %s", domain)
	}
	return net.Dial("tcp", address)
}

func tunnelTo(domain string, device *fakeDevice) *fakeDialer {
	return &fakeDialer{devices: map[string]string{domain: device.address}}
}

type fakeTraffic struct {
	mutex    sync.Mutex
	over     bool
	recorded map[string]int64
}

func (f *fakeTraffic) OverLimit(_ string) bool {
	return f.over
}

func (f *fakeTraffic) Record(domain string, bytes int64) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.recorded == nil {
		f.recorded = map[string]int64{}
	}
	f.recorded[domain] += bytes
}

func (f *fakeTraffic) bytes(domain string) int64 {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.recorded[domain]
}

type fakeStore struct {
	domains map[string]*model.Domain
}

func (f *fakeStore) GetDomainByName(name string) (*model.Domain, error) {
	return f.domains[name], nil
}

func relayed(dialer DeviceDialer, domains map[string]*model.Domain) (string, func(), error) {
	address, stop, _, err := relayedWith(dialer, domains, mail.NewConnections(0), &fakeTraffic{})
	return address, stop, err
}

func relayedWith(dialer DeviceDialer, domains map[string]*model.Domain,
	connections *mail.Connections, traffic Traffic) (string, func(), *fakeTraffic, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, nil, err
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		return "", nil, nil, err
	}

	router := NewRouter(&fakeStore{domains: domains})
	server := NewServer(address, "mx.syncloud.it", router, dialer, connections,
		mail.NewInFlight(0), traffic, 1024*1024, zap.NewNop())
	if err := server.Start(); err != nil {
		return "", nil, nil, err
	}
	stop := func() { _ = server.Close() }
	if err := waitListening(address); err != nil {
		stop()
		return "", nil, nil, err
	}
	counter, _ := traffic.(*fakeTraffic)
	return address, stop, counter, nil
}

func waitListening(address string) error {
	var err error
	for i := 0; i < 100; i++ {
		var connection net.Conn
		connection, err = net.DialTimeout("tcp", address, 100*time.Millisecond)
		if err == nil {
			return connection.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("inbound server never came up on %s: %w", address, err)
}

func relayDomain(name string) *model.Domain {
	return &model.Domain{Name: name, Relay: true}
}

func dial(address string) (*smtp.Client, error) {
	var err error
	for i := 0; i < 100; i++ {
		var client *smtp.Client
		client, err = proxyDial(address, "203.0.113.1")
		if err == nil {
			return client, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return nil, fmt.Errorf("inbound server never came up on %s: %w", address, err)
}

func deliver(address string, recipients ...string) error {
	client, err := dial(address)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Mail("sender@example.com"); err != nil {
		return err
	}
	for _, to := range recipients {
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}
	writer, dataErr := client.Data()
	if dataErr != nil {
		return dataErr
	}
	if _, err := writer.Write([]byte("Subject: hello\r\n\r\nbody\r\n")); err != nil {
		return err
	}
	return writer.Close()
}

func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), code), "expected %s, got %v", code, err)
}

func TestInbound_DeliversToTheDevice(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(tunnelTo("alice.syncloud.it", device), map[string]*model.Domain{
		"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "user@alice.syncloud.it"))

	recipients, body := device.got()
	assert.Equal(t, []string{"user@alice.syncloud.it"}, recipients)
	assert.Contains(t, body, "Subject: hello")
}

func TestInbound_UnknownDomainRejected(t *testing.T) {
	address, stop, err := relayed(&fakeDialer{}, map[string]*model.Domain{})
	assert.NoError(t, err)
	defer stop()

	assertCode(t, deliver(address, "user@stranger.syncloud.it"), "550")
}

func TestInbound_RelayOffRejected(t *testing.T) {
	address, stop, err := relayed(&fakeDialer{}, map[string]*model.Domain{
		"alice.syncloud.it": {Name: "alice.syncloud.it", Relay: false}})
	assert.NoError(t, err)
	defer stop()

	assertCode(t, deliver(address, "user@alice.syncloud.it"), "550")
}

func TestInbound_NoTunnelDeferred(t *testing.T) {
	address, stop, err := relayed(&fakeDialer{}, map[string]*model.Domain{
		"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assertCode(t, deliver(address, "user@alice.syncloud.it"), "451")
}

func TestInbound_DeviceRejectsRecipient(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{rcptErr: &gosmtp.SMTPError{
		Code: 550, EnhancedCode: gosmtp.EnhancedCode{5, 1, 1}, Message: "user unknown"}})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(tunnelTo("alice.syncloud.it", device), map[string]*model.Domain{
		"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	deliverErr := deliver(address, "nobody@alice.syncloud.it")

	assertCode(t, deliverErr, "550")
	assert.Contains(t, deliverErr.Error(), "user unknown")
}

func TestInbound_DeviceRejectsMessage(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{dataErr: &gosmtp.SMTPError{
		Code: 550, EnhancedCode: gosmtp.EnhancedCode{5, 7, 1}, Message: "spam rejected"}})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(tunnelTo("alice.syncloud.it", device), map[string]*model.Domain{
		"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	deliverErr := deliver(address, "user@alice.syncloud.it")

	assertCode(t, deliverErr, "550")
	assert.Contains(t, deliverErr.Error(), "spam rejected")
}

func TestInbound_SecondDomainDeferred(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(
		&fakeDialer{devices: map[string]string{
			"alice.syncloud.it": device.address, "bob.syncloud.it": device.address}},
		map[string]*model.Domain{
			"alice.syncloud.it": relayDomain("alice.syncloud.it"),
			"bob.syncloud.it":   relayDomain("bob.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	client, err := dial(address)
	assert.NoError(t, err)
	defer client.Close()
	assert.NoError(t, client.Mail("sender@example.com"))
	assert.NoError(t, client.Rcpt("user@alice.syncloud.it"))

	assertCode(t, client.Rcpt("user@bob.syncloud.it"), "452")

	writer, dataErr := client.Data()
	assert.NoError(t, dataErr)
	_, writeErr := writer.Write([]byte("Subject: hello\r\n\r\nbody\r\n"))
	assert.NoError(t, writeErr)
	assert.NoError(t, writer.Close())

	recipients, _ := device.got()
	assert.Equal(t, []string{"user@alice.syncloud.it"}, recipients)
}

func TestInbound_ManyRecipientsOnOneDevice(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(tunnelTo("alice.syncloud.it", device), map[string]*model.Domain{
		"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "one@alice.syncloud.it", "two@alice.syncloud.it"))

	recipients, _ := device.got()
	assert.Equal(t, []string{"one@alice.syncloud.it", "two@alice.syncloud.it"}, recipients)
}

func proxyDial(address string, client string) (*smtp.Client, error) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	header := &proxyproto.Header{
		Version:           2,
		Command:           proxyproto.PROXY,
		TransportProtocol: proxyproto.TCPv4,
		SourceAddr:        &net.TCPAddr{IP: net.ParseIP(client), Port: 40000},
		DestinationAddr:   &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10025},
	}
	if _, err := header.WriteTo(connection); err != nil {
		return nil, err
	}
	return smtp.NewClient(connection, "mx.syncloud.it")
}

func TestInbound_ProxyProtocolKeepsSendersApart(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, _, err := relayedWith(tunnelTo("alice.syncloud.it", device),
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")},
		mail.NewConnections(1), &fakeTraffic{})
	assert.NoError(t, err)
	defer stop()

	first, err := proxyDial(address, "203.0.113.7")
	assert.NoError(t, err)
	defer first.Close()
	second, err := proxyDial(address, "198.51.100.9")
	assert.NoError(t, err)
	defer second.Close()

	assert.NoError(t, first.Mail("sender@example.com"))
	assert.NoError(t, second.Mail("sender@example.com"))
}

func TestInbound_ProxyProtocolLimitsOneSender(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, _, err := relayedWith(tunnelTo("alice.syncloud.it", device),
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")},
		mail.NewConnections(1), &fakeTraffic{})
	assert.NoError(t, err)
	defer stop()

	first, err := proxyDial(address, "203.0.113.7")
	assert.NoError(t, err)
	defer first.Close()
	second, err := proxyDial(address, "203.0.113.7")
	assert.NoError(t, err)
	defer second.Close()

	assert.NoError(t, first.Mail("sender@example.com"))
	assert.Error(t, second.Mail("sender@example.com"))
}

func TestInbound_ProxyProtocolPolicy(t *testing.T) {
	loopback, err := fromLoopbackOnly(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1})
	assert.NoError(t, err)
	assert.Equal(t, proxyproto.REQUIRE, loopback)

	remote, err := fromLoopbackOnly(&net.TCPAddr{IP: net.ParseIP("203.0.113.7"), Port: 1})
	assert.NoError(t, err)
	assert.Equal(t, proxyproto.REJECT, remote)
}

func TestInbound_PortConflictStopsTheApi(t *testing.T) {
	held, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	defer held.Close()

	server := NewServer(held.Addr().String(), "mx.syncloud.it",
		NewRouter(&fakeStore{domains: map[string]*model.Domain{}}), &fakeDialer{},
		mail.NewConnections(0), mail.NewInFlight(0), &fakeTraffic{}, 1024*1024, zap.NewNop())

	assert.Error(t, server.Start())
}

func TestInbound_CountsTheBytesItRelays(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, traffic, err := relayedWith(tunnelTo("alice.syncloud.it", device),
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")},
		mail.NewConnections(0), &fakeTraffic{})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "user@alice.syncloud.it"))

	assert.Greater(t, traffic.bytes("alice.syncloud.it"), int64(0))
}

func TestInbound_OverTheLimitIsDeferred(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, traffic, err := relayedWith(tunnelTo("alice.syncloud.it", device),
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")},
		mail.NewConnections(0), &fakeTraffic{over: true})
	assert.NoError(t, err)
	defer stop()

	assertCode(t, deliver(address, "user@alice.syncloud.it"), "451")
	assert.Equal(t, int64(0), traffic.bytes("alice.syncloud.it"))
}
