package mailrelay

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeSender struct {
	err  error
	sent int
}

func (f *fakeSender) Send(_ string, recipients []string, _ []byte) error {
	if f.err != nil {
		return f.err
	}
	f.sent += len(recipients)
	return nil
}

type fakeScanner struct{ err error }

func (f *fakeScanner) Scan(_ string, _ []string, _ string, _ []byte) error { return f.err }

type relayUnderTest struct {
	address string
	sender  *fakeSender
	scanner *fakeScanner
}

func defaultRelay() *Relay {
	token := "the-token"
	return New(
		&fakeDomains{domain: &model.Domain{Name: "device.syncloud.it", UserId: 7, UpdateToken: &token}},
		&fakePlans{limit: 100},
		&fakeUsage{},
		&fakeBlocklist{},
		&fakeWarner{},
		zap.NewNop())
}

func startRelay(t *testing.T, limits Limits, sender *fakeSender, scanner *fakeScanner,
	connections *Connections, inFlight *InFlight) *relayUnderTest {
	t.Helper()
	return startRelayWith(t, defaultRelay(), limits, sender, scanner, connections, inFlight)
}

func startRelayWith(t *testing.T, relay *Relay, limits Limits, sender *fakeSender, scanner *fakeScanner,
	connections *Connections, inFlight *InFlight) *relayUnderTest {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	address := listener.Addr().String()

	server := NewServer(address, "syncloud.it", relay, sender, scanner,
		NewLimiter(limits), connections, inFlight, NewCertificateLoader("", ""), 1024*1024, zap.NewNop())
	assert.NoError(t, listener.Close())
	assert.NoError(t, server.Start())
	t.Cleanup(func() { _ = server.Close() })

	return &relayUnderTest{address: address, sender: sender, scanner: scanner}
}

func dial(t *testing.T, r *relayUnderTest) (*smtp.Client, error) {
	t.Helper()
	var err error
	for i := 0; i < 100; i++ {
		var client *smtp.Client
		client, err = smtp.Dial(r.address)
		if err == nil {
			return client, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return nil, err
}

func send(t *testing.T, r *relayUnderTest, recipients ...string) error {
	t.Helper()
	client, err := dial(t, r)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Auth(smtp.PlainAuth("", "device.syncloud.it", "the-token", "127.0.0.1")); err != nil {
		return err
	}
	if err := client.Mail("user@device.syncloud.it"); err != nil {
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
	if _, err := writer.Write([]byte("Subject: test\r\n\r\nbody\r\n")); err != nil {
		return err
	}
	return writer.Close()
}

func defaultLimits() Limits {
	return Limits{Minute: 100, Hour: 100, Day: 100, Recipients: 2}
}

func relayFor(t *testing.T, sender *fakeSender, scanner *fakeScanner) *relayUnderTest {
	return startRelay(t, defaultLimits(), sender, scanner, NewConnections(0), NewInFlight(0))
}

func assertCode(t *testing.T, err error, code string) {
	t.Helper()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), code),
		fmt.Sprintf("expected reply starting with %s, got %q", code, err.Error()))
}

func TestRelayServer_DeliversAuthenticatedMail(t *testing.T) {
	sender := &fakeSender{}
	r := relayFor(t, sender, &fakeScanner{})
	assert.NoError(t, send(t, r, "someone@example.com"))
	assert.Equal(t, 1, sender.sent)
}

func authWith(t *testing.T, r *relayUnderTest, login string, password string) error {
	t.Helper()
	client, err := dial(t, r)
	assert.NoError(t, err)
	defer client.Close()
	return client.Auth(smtp.PlainAuth("", login, password, "127.0.0.1"))
}

func serverFor(t *testing.T, relay *Relay) *relayUnderTest {
	t.Helper()
	return startRelayWith(t, relay, defaultLimits(), &fakeSender{}, &fakeScanner{},
		NewConnections(0), NewInFlight(0))
}

func TestRelayServer_RejectsWrongPassword(t *testing.T) {
	r := relayFor(t, &fakeSender{}, &fakeScanner{})
	assertCode(t, authWith(t, r, "device.syncloud.it", "wrong"), "535")
}

func TestRelayServer_RejectsForeignLogin(t *testing.T) {
	r := relayFor(t, &fakeSender{}, &fakeScanner{})
	assertCode(t, authWith(t, r, "someone-else.syncloud.it", "the-token"), "535")
}

func TestRelayServer_BlockedIsPermanent(t *testing.T) {
	relay, _ := relayWith(100, 0, true)
	assertCode(t, authWith(t, serverFor(t, relay), "device.syncloud.it", "the-token"), "535")
}

func TestRelayServer_PlanWithoutRelayIsPermanent(t *testing.T) {
	relay, _ := relayWith(0, 0, false)
	assertCode(t, authWith(t, serverFor(t, relay), "device.syncloud.it", "the-token"), "535")
}

func TestRelayServer_OverMonthlyLimitIsTemporary(t *testing.T) {
	relay, _ := relayWith(10, 10, false)
	assertCode(t, authWith(t, serverFor(t, relay), "device.syncloud.it", "the-token"), "454")
}

func TestRelayServer_RejectsForeignSender(t *testing.T) {
	r := relayFor(t, &fakeSender{}, &fakeScanner{})
	client, err := dial(t, r)
	assert.NoError(t, err)
	defer client.Close()
	assert.NoError(t, client.Auth(smtp.PlainAuth("", "device.syncloud.it", "the-token", "127.0.0.1")))
	assertCode(t, client.Mail("someone@elsewhere.com"), "550")
}

func TestRelayServer_SpamIsPermanent(t *testing.T) {
	r := relayFor(t, &fakeSender{}, &fakeScanner{err: ErrRejectedAsSpam})
	assertCode(t, send(t, r, "someone@example.com"), "550")
}

func TestRelayServer_SendFailureIsTemporary(t *testing.T) {
	r := relayFor(t, &fakeSender{err: fmt.Errorf("ses is throttling")}, &fakeScanner{})
	assertCode(t, send(t, r, "someone@example.com"), "451")
}

func TestRelayServer_ScannerOutageIsTemporary(t *testing.T) {
	r := relayFor(t, &fakeSender{}, &fakeScanner{err: fmt.Errorf("spam filter unavailable")})
	assertCode(t, send(t, r, "someone@example.com"), "451")
}

func TestRelayServer_RateLimitIsTemporary(t *testing.T) {
	limits := defaultLimits()
	limits.Minute = 1
	r := startRelay(t, limits, &fakeSender{}, &fakeScanner{}, NewConnections(0), NewInFlight(0))
	assert.NoError(t, send(t, r, "someone@example.com"))
	assertCode(t, send(t, r, "someone@example.com"), "451")
}

func TestRelayServer_TooManyRecipientsIsPermanent(t *testing.T) {
	r := relayFor(t, &fakeSender{}, &fakeScanner{})
	assertCode(t, send(t, r, "a@example.com", "b@example.com", "c@example.com"), "550")
}

func TestRelayServer_AtCapacityIsTemporary(t *testing.T) {
	inFlight := NewInFlight(1)
	assert.True(t, inFlight.Acquire())
	r := startRelay(t, defaultLimits(), &fakeSender{}, &fakeScanner{}, NewConnections(0), inFlight)
	assertCode(t, send(t, r, "someone@example.com"), "451")
}

func TestRelayServer_ConnectionCap(t *testing.T) {
	r := startRelay(t, defaultLimits(), &fakeSender{}, &fakeScanner{}, NewConnections(1), NewInFlight(0))

	first, err := dial(t, r)
	assert.NoError(t, err)
	defer first.Close()
	assert.NoError(t, first.Hello("device.syncloud.it"))

	second, err := dial(t, r)
	assert.NoError(t, err)
	defer second.Close()
	assertCode(t, second.Hello("device.syncloud.it"), "451")
}
