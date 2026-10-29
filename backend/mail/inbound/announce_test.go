package inbound

import (
	"net"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
)

func TestAnnounce_DeviceSeesTheRelayedClientAddress(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()
	address, stop, err := relayed(tunnelTo("alice.syncloud.it", device),
		map[string]*model.Domain{"alice.syncloud.it": relayDomain("alice.syncloud.it")})
	assert.NoError(t, err)
	defer stop()

	assert.NoError(t, deliver(address, "user@alice.syncloud.it"))

	client, helo := device.seen()
	host, _, err := net.SplitHostPort(client)
	assert.NoError(t, err)
	assert.Equal(t, "203.0.113.1", host)
	assert.Equal(t, "localhost", helo)
}

func TestAnnounce_ClientWithoutHeloIsGreetedAsTheRelay(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()

	helo := announceTo(t, device, Client{Address: "198.51.100.7"})

	assert.Equal(t, "mx.syncloud.it", helo)
	client, _ := device.seen()
	assert.Equal(t, "198.51.100.7:0", client)
}

func TestAnnounce_Ipv6Client(t *testing.T) {
	device, stopDevice, err := startDevice(&fakeDevice{})
	assert.NoError(t, err)
	defer stopDevice()

	announceTo(t, device, Client{Address: "2001:db8::7", Helo: "mail.example.net"})

	client, helo := device.seen()
	assert.Equal(t, "[2001:db8::7]:0", client)
	assert.Equal(t, "mail.example.net", helo)
}

func TestAnnounce_RejectsAnAddressThatIsNotAnIp(t *testing.T) {
	connection, peer := net.Pipe()
	defer func() { _ = connection.Close() }()
	defer func() { _ = peer.Close() }()

	_, err := Announce(connection, "mx.syncloud.it", Client{Address: "unknown"})

	assert.Error(t, err)
}

func announceTo(t *testing.T, device *fakeDevice, client Client) string {
	connection, err := net.Dial("tcp", device.address)
	assert.NoError(t, err)
	helo, err := Announce(connection, "mx.syncloud.it", client)
	assert.NoError(t, err)
	smtpClient, err := smtp.NewClient(connection, "device.syncloud.it")
	assert.NoError(t, err)
	assert.NoError(t, smtpClient.Hello(helo))
	assert.NoError(t, smtpClient.Mail("from@example.com"))
	_ = smtpClient.Quit()
	return helo
}
