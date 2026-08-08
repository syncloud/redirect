package caddysmtp

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type fakeUpstream struct {
	conn        net.Conn
	connections int
	remote      net.Addr
}

func (u *fakeUpstream) Connect(remote net.Addr, _ net.Addr) (net.Conn, error) {
	u.connections++
	u.remote = remote
	if u.conn == nil {
		return nil, errors.New("no upstream")
	}
	return u.conn, nil
}

type backendResult struct {
	received []string
	err      error
}

func runBackend(conn net.Conn, greeting string, replies ...string) <-chan backendResult {
	results := make(chan backendResult, 1)
	go func() {
		result := backendResult{}
		defer func() {
			_ = conn.Close()
			results <- result
		}()
		if result.err = writeLines(conn, greeting); result.err != nil {
			return
		}
		reader := newReader(conn)
		for _, reply := range replies {
			line, err := readLine(reader)
			if err != nil {
				result.err = err
				return
			}
			result.received = append(result.received, line)
			if result.err = writeLines(conn, reply); result.err != nil {
				return
			}
		}
	}()
	return results
}

func startConversation(upstream Upstream, tlsConfig *tls.Config) (net.Conn, <-chan error) {
	client, server := net.Pipe()
	errs := make(chan error, 1)
	go func() {
		conversation := NewConversation("mx.test", upstream,
			func() *tls.Config { return tlsConfig }, zap.NewNop())
		errs <- conversation.Serve(server)
	}()
	return client, errs
}

func selfSigned() (*tls.Config, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "mx.test"},
		DNSNames:     []string{"mx.test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{{
		Certificate: [][]byte{der},
		PrivateKey:  key,
	}}}, nil
}

func readLines(reader *bufio.Reader, count int) ([]string, error) {
	lines := make([]string, 0, count)
	for i := 0; i < count; i++ {
		line, err := readLine(reader)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func TestConversation_OffersStartTlsAfterEhlo(t *testing.T) {
	client, _ := startConversation(&fakeUpstream{}, nil)
	defer client.Close()
	reader := newReader(client)

	greeting, err := readLine(reader)
	require.NoError(t, err)
	assert.Equal(t, "220 mx.test ESMTP", greeting)

	require.NoError(t, writeLines(client, "EHLO sender.test"))
	capabilities, err := readLines(reader, 5)
	require.NoError(t, err)
	assert.Equal(t, "250-mx.test", capabilities[0])
	assert.Equal(t, "250 STARTTLS", capabilities[4])
}

func TestConversation_ReplaysGreetingAndRelaysCleartextSession(t *testing.T) {
	backend, upstreamSide := net.Pipe()
	backendResults := runBackend(backend, "220 backend ESMTP",
		"250-backend\r\n250 OK", "250 2.1.0 Sender ok")
	upstream := &fakeUpstream{conn: upstreamSide}

	client, _ := startConversation(upstream, nil)
	defer client.Close()
	reader := newReader(client)

	_, err := readLines(reader, 1)
	require.NoError(t, err)
	require.NoError(t, writeLines(client, "EHLO sender.test"))
	_, err = readLines(reader, 5)
	require.NoError(t, err)

	require.NoError(t, writeLines(client, "MAIL FROM:<a@sender.test>"))
	reply, err := readLine(reader)
	require.NoError(t, err)
	assert.Equal(t, "250 2.1.0 Sender ok", reply)

	result := <-backendResults
	require.NoError(t, result.err)
	assert.Equal(t, []string{"EHLO sender.test", "MAIL FROM:<a@sender.test>"}, result.received)
	assert.Equal(t, 1, upstream.connections)
}

func TestConversation_UpgradesToTlsThenRelays(t *testing.T) {
	tlsConfig, err := selfSigned()
	require.NoError(t, err)

	backend, upstreamSide := net.Pipe()
	backendResults := runBackend(backend, "220 backend ESMTP", "250 backend")
	upstream := &fakeUpstream{conn: upstreamSide}

	client, _ := startConversation(upstream, tlsConfig)
	defer client.Close()
	reader := newReader(client)

	_, err = readLines(reader, 1)
	require.NoError(t, err)
	require.NoError(t, writeLines(client, "EHLO sender.test"))
	_, err = readLines(reader, 5)
	require.NoError(t, err)

	require.NoError(t, writeLines(client, "STARTTLS"))
	ready, err := readLine(reader)
	require.NoError(t, err)
	assert.Equal(t, "220 2.0.0 Ready to start TLS", ready)

	secure := tls.Client(client, &tls.Config{InsecureSkipVerify: true})
	require.NoError(t, secure.Handshake())

	require.NoError(t, writeLines(secure, "EHLO sender.test"))
	reply, err := readLine(newReader(secure))
	require.NoError(t, err)
	assert.Equal(t, "250 backend", reply)

	result := <-backendResults
	require.NoError(t, result.err)
	assert.Equal(t, []string{"EHLO sender.test"}, result.received)
}

func TestConversation_RejectsDataPipelinedBeforeTheUpgrade(t *testing.T) {
	upstream := &fakeUpstream{}
	client, errs := startConversation(upstream, nil)
	defer client.Close()
	reader := newReader(client)

	_, err := readLines(reader, 1)
	require.NoError(t, err)
	require.NoError(t, writeLines(client, "EHLO sender.test"))
	_, err = readLines(reader, 5)
	require.NoError(t, err)

	require.NoError(t, writeLines(client, "STARTTLS", "MAIL FROM:<a@sender.test>"))
	assert.ErrorIs(t, <-errs, ErrPipelinedUpgrade)
	assert.Equal(t, 0, upstream.connections)
}

func TestConversation_AnswersQuitWithoutConnectingUpstream(t *testing.T) {
	upstream := &fakeUpstream{}
	client, errs := startConversation(upstream, nil)
	defer client.Close()
	reader := newReader(client)

	_, err := readLines(reader, 1)
	require.NoError(t, err)
	require.NoError(t, writeLines(client, "QUIT"))
	bye, err := readLine(reader)
	require.NoError(t, err)
	assert.Equal(t, "221 2.0.0 Bye", bye)

	require.NoError(t, <-errs)
	assert.Equal(t, 0, upstream.connections)
}
