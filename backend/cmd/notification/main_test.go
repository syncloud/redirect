package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"log"
	"os"
	"testing"
)

type MailStub struct {
	errors map[string]bool
	sent   map[string]bool
}

func (m MailStub) SendReleaseAnnouncement(to string) error {
	if m.errors[to] {
		return fmt.Errorf("test mail error")
	}
	m.sent[to] = true
	return nil
}

type DbStub struct {
}

func (d DbStub) GetUsersByField(field string, value string) ([]*model.User, error) {
	return []*model.User{
		{Email: "test1@example.com", NotificationEnabled: true},
		{Email: "test2@example.com", NotificationEnabled: true},
		{Email: "test3@example.com", NotificationEnabled: false},
	}, nil
}

func TestOnlyRetryErrorsOnSecondRun(t *testing.T) {
	sentFile := tempFile().Name()
	defer os.Remove(sentFile)
	mail1 := &MailStub{
		errors: map[string]bool{"test2@example.com": true},
		sent:   make(map[string]bool),
	}
	notification1 := NewNotification(&DbStub{}, mail1, sentFile, "*")
	notification1.Send()

	assert.Equal(t, 1, len(mail1.sent))
	assert.Contains(t, mail1.sent, "test1@example.com")

	mail2 := &MailStub{
		errors: make(map[string]bool),
		sent:   make(map[string]bool),
	}
	notification2 := NewNotification(&DbStub{}, mail2, sentFile, "*")
	notification2.Send()
	assert.Equal(t, 1, len(mail2.sent))
	assert.Contains(t, mail2.sent, "test2@example.com")

	content, err := os.ReadFile(sentFile)
	assert.Nil(t, err)
	fmt.Println(string(content))
}

func tempFile() *os.File {
	tmpFile, err := os.CreateTemp("", "")
	if err != nil {
		log.Fatal(err)
	}
	return tmpFile
}
