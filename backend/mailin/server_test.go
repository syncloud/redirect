package mailin

import (
	"io"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"testing"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/mailnet"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeDevice struct {
	port    int
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

func startDevice(t *testing.T, device *fakeDevice) *fakeDevice {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	device.port = listener.Addr().(*net.TCPAddr).Port

	server := gosmtp.NewServer(gosmtp.BackendFunc(func(_ *gosmtp.Conn) (gosmtp.Session, error) {
		return &deviceSession{device: device}, nil
	}))
	server.Domain = "device.syncloud.it"
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { _ = server.Close() })
	return device
}

type fakeStore struct {
	domains map[string]*model.Domain
}

func (f *fakeStore) GetDomainByName(name string) (*model.Domain, error) {
	return f.domains[name], nil
}

func relayed(t *testing.T, domains map[string]*model.Domain) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	address := listener.Addr().String()
	assert.NoError(t, listener.Close())

	router := NewRouter(&fakeStore{domains: domains}, "127.0.0.1")
	server := NewServer(address, "mx.syncloud.it", router, mailnet.NewConnections(0),
		mailnet.NewInFlight(0), mailnet.NewCertificateLoader("", ""), 1024*1024, zap.NewNop())
	assert.NoError(t, server.Start())
	t.Cleanup(func() { _ = server.Close() })
	return address
}

func mailRelayDomain(name string, port int) *model.Domain {
	return &model.Domain{Name: name, MailRelay: true, SmtpPort: &port}
}

func dial(t *testing.T, address string) *smtp.Client {
	t.Helper()
	for i := 0; i < 100; i++ {
		client, err := smtp.Dial(address)
		if err == nil {
			return client
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("inbound server never came up on %s", address)
	return nil
}

func deliver(t *testing.T, address string, recipients ...string) error {
	t.Helper()
	client := dial(t, address)
	defer client.Close()
	if err := client.Mail("sender@example.com"); err != nil {
		return err
	}
	for _, to := range recipients {
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte("Subject: hello\r\n\r\nbody\r\n")); err != nil {
		return err
	}
	return writer.Close()
}

func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), code),
		"expected %s, got %v", code, err)
}

func TestInbound_DeliversToTheDevice(t *testing.T) {
	device := startDevice(t, &fakeDevice{})
	address := relayed(t, map[string]*model.Domain{
		"alice.syncloud.it": mailRelayDomain("alice.syncloud.it", device.port)})

	err := deliver(t, address, "user@alice.syncloud.it")

	assert.NoError(t, err)
	recipients, body := device.got()
	assert.Equal(t, []string{"user@alice.syncloud.it"}, recipients)
	assert.Contains(t, body, "Subject: hello")
}

func TestInbound_UnknownDomainRejected(t *testing.T) {
	address := relayed(t, map[string]*model.Domain{})

	err := deliver(t, address, "user@stranger.syncloud.it")

	assertCode(t, err, "550")
}

func TestInbound_MailRelayOffRejected(t *testing.T) {
	port := 20000
	address := relayed(t, map[string]*model.Domain{
		"alice.syncloud.it": {Name: "alice.syncloud.it", MailRelay: false, SmtpPort: &port}})

	err := deliver(t, address, "user@alice.syncloud.it")

	assertCode(t, err, "550")
}

func TestInbound_DeviceOfflineDeferred(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	closedPort := listener.Addr().(*net.TCPAddr).Port
	assert.NoError(t, listener.Close())

	address := relayed(t, map[string]*model.Domain{
		"alice.syncloud.it": mailRelayDomain("alice.syncloud.it", closedPort)})

	err = deliver(t, address, "user@alice.syncloud.it")

	assertCode(t, err, "451")
}

func TestInbound_DeviceRejectsRecipient(t *testing.T) {
	device := startDevice(t, &fakeDevice{rcptErr: &gosmtp.SMTPError{
		Code: 550, EnhancedCode: gosmtp.EnhancedCode{5, 1, 1}, Message: "user unknown"}})
	address := relayed(t, map[string]*model.Domain{
		"alice.syncloud.it": mailRelayDomain("alice.syncloud.it", device.port)})

	err := deliver(t, address, "nobody@alice.syncloud.it")

	assertCode(t, err, "550")
	assert.Contains(t, err.Error(), "user unknown")
}

func TestInbound_DeviceRejectsMessage(t *testing.T) {
	device := startDevice(t, &fakeDevice{dataErr: &gosmtp.SMTPError{
		Code: 550, EnhancedCode: gosmtp.EnhancedCode{5, 7, 1}, Message: "spam rejected"}})
	address := relayed(t, map[string]*model.Domain{
		"alice.syncloud.it": mailRelayDomain("alice.syncloud.it", device.port)})

	err := deliver(t, address, "user@alice.syncloud.it")

	assertCode(t, err, "550")
	assert.Contains(t, err.Error(), "spam rejected")
}

func TestInbound_SecondDomainDeferred(t *testing.T) {
	alice := startDevice(t, &fakeDevice{})
	bob := startDevice(t, &fakeDevice{})
	address := relayed(t, map[string]*model.Domain{
		"alice.syncloud.it": mailRelayDomain("alice.syncloud.it", alice.port),
		"bob.syncloud.it":   mailRelayDomain("bob.syncloud.it", bob.port)})

	client := dial(t, address)
	defer client.Close()
	assert.NoError(t, client.Mail("sender@example.com"))
	assert.NoError(t, client.Rcpt("user@alice.syncloud.it"))

	err := client.Rcpt("user@bob.syncloud.it")

	assertCode(t, err, "452")

	writer, dataErr := client.Data()
	assert.NoError(t, dataErr)
	_, writeErr := writer.Write([]byte("Subject: hello\r\n\r\nbody\r\n"))
	assert.NoError(t, writeErr)
	assert.NoError(t, writer.Close())

	aliceRecipients, _ := alice.got()
	bobRecipients, _ := bob.got()
	assert.Equal(t, []string{"user@alice.syncloud.it"}, aliceRecipients)
	assert.Empty(t, bobRecipients)
}

func TestInbound_ManyRecipientsOnOneDevice(t *testing.T) {
	device := startDevice(t, &fakeDevice{})
	address := relayed(t, map[string]*model.Domain{
		"alice.syncloud.it": mailRelayDomain("alice.syncloud.it", device.port)})

	err := deliver(t, address, "one@alice.syncloud.it", "two@alice.syncloud.it")

	assert.NoError(t, err)
	recipients, _ := device.got()
	assert.Equal(t, []string{"one@alice.syncloud.it", "two@alice.syncloud.it"}, recipients)
}
