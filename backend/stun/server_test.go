package stun

import (
	"net"
	"testing"
	"time"

	"github.com/pion/stun/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestBindingReturnsSourceAddress(t *testing.T) {
	serverConn, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	server := NewServer(serverConn, zap.NewNop())
	go server.Serve()
	defer serverConn.Close()

	client, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer client.Close()

	request := stun.MustBuild(stun.TransactionID, stun.BindingRequest, stun.Fingerprint)
	_, err = client.WriteTo(request.Raw, serverConn.LocalAddr())
	require.NoError(t, err)

	require.NoError(t, client.SetReadDeadline(time.Now().Add(5*time.Second)))
	buf := make([]byte, 1500)
	n, _, err := client.ReadFrom(buf)
	require.NoError(t, err)

	response := &stun.Message{Raw: buf[:n]}
	require.NoError(t, response.Decode())
	assert.Equal(t, stun.BindingSuccess, response.Type)
	assert.Equal(t, request.TransactionID, response.TransactionID)
	assert.NoError(t, stun.Fingerprint.Check(response))

	var mapped stun.XORMappedAddress
	require.NoError(t, mapped.GetFrom(response))
	local := client.LocalAddr().(*net.UDPAddr)
	assert.True(t, local.IP.Equal(mapped.IP))
	assert.Equal(t, local.Port, mapped.Port)
}

func TestIgnoresNonStun(t *testing.T) {
	_, ok := Respond([]byte("hello, not a stun message"), &net.UDPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 5})
	assert.False(t, ok)
}

func TestIgnoresNonBinding(t *testing.T) {
	indication := stun.MustBuild(stun.TransactionID, stun.NewType(stun.MethodBinding, stun.ClassIndication))
	_, ok := Respond(indication.Raw, &net.UDPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 5})
	assert.False(t, ok)
}

func TestMapsIpv6Source(t *testing.T) {
	request := stun.MustBuild(stun.TransactionID, stun.BindingRequest)
	source := &net.UDPAddr{IP: net.ParseIP("2001:db8::1"), Port: 41641}
	raw, ok := Respond(request.Raw, source)
	require.True(t, ok)

	response := &stun.Message{Raw: raw}
	require.NoError(t, response.Decode())
	var mapped stun.XORMappedAddress
	require.NoError(t, mapped.GetFrom(response))
	assert.True(t, source.IP.Equal(mapped.IP))
	assert.Equal(t, 41641, mapped.Port)
}
