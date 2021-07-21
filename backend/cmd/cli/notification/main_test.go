package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"io/ioutil"
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

func Test(t *testing.T) {
	sentFile := tempFile().Name()
	defer os.Remove(sentFile)
	errors := map[string]bool{
		"test2@example.com": true,
	}
	mail := &MailStub{errors: errors, sent: make(map[string]bool)}
	notification := NewNotification(&DbStub{}, mail, sentFile, "*")
	notification.Send()

	assert.Equal(t, 1, len(mail.sent))
	assert.Contains(t, mail.sent, "test1@example.com")

	mail.errors = make(map[string]bool)
	mail.sent = make(map[string]bool)
	notification.Send()
	assert.Equal(t, 1, len(mail.sent))
	assert.Contains(t, mail.sent, "test2@example.com")

	content, err := ioutil.ReadFile(sentFile)
	assert.Nil(t, err)
	fmt.Println(string(content))
}

func tempFile() *os.File {
	tmpFile, err := ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}
	return tmpFile
}
