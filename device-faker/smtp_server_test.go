package main

import (
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func freeAddress() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	address := listener.Addr().String()
	return address, listener.Close()
}

func startServer(mailbox *Mailbox) (string, error) {
	address, err := freeAddress()
	if err != nil {
		return "", err
	}
	if err := NewSmtpServer(address, mailbox).Start(); err != nil {
		return "", err
	}
	for i := 0; i < 50; i++ {
		connection, err := net.DialTimeout("tcp", address, time.Second)
		if err == nil {
			return address, connection.Close()
		}
		time.Sleep(20 * time.Millisecond)
	}
	return "", err
}

func send(address string, recipient string, body string) error {
	return smtp.SendMail(address, nil, "sender@example.com", []string{recipient}, []byte(body))
}

func TestDevice_RecordsDeliveredMail(t *testing.T) {
	mailbox := NewMailbox()
	address, err := startServer(mailbox)
	require.NoError(t, err)

	require.NoError(t, send(address, "user@device.test", "Subject: hello\r\n\r\nbody\r\n"))

	messages := mailbox.Messages()
	require.Len(t, messages, 1)
	assert.Contains(t, messages[0].Recipients[0], "user@device.test")
	assert.Contains(t, messages[0].Body, "hello")
}

func TestDevice_RejectsRecipientWhenAsked(t *testing.T) {
	mailbox := NewMailbox()
	mailbox.SetBehaviour(Behaviour{Rcpt: Reject})
	address, err := startServer(mailbox)
	require.NoError(t, err)

	err = send(address, "user@device.test", "Subject: hello\r\n\r\nbody\r\n")

	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "550"), err.Error())
	assert.Empty(t, mailbox.Messages())
}

func TestDevice_RejectsMessageWhenAsked(t *testing.T) {
	mailbox := NewMailbox()
	mailbox.SetBehaviour(Behaviour{Data: Reject})
	address, err := startServer(mailbox)
	require.NoError(t, err)

	err = send(address, "user@device.test", "Subject: hello\r\n\r\nbody\r\n")

	require.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "554"), err.Error())
	assert.Empty(t, mailbox.Messages())
}

func TestDevice_ResetClearsMessagesAndBehaviour(t *testing.T) {
	mailbox := NewMailbox()
	address, err := startServer(mailbox)
	require.NoError(t, err)
	require.NoError(t, send(address, "user@device.test", "Subject: hello\r\n\r\nbody\r\n"))
	mailbox.SetBehaviour(Behaviour{Rcpt: Reject})

	mailbox.Reset()

	assert.Empty(t, mailbox.Messages())
	assert.Equal(t, Behaviour{Rcpt: Accept, Data: Accept}, mailbox.Behaviour())
}
