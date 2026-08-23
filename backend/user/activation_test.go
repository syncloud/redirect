package user

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
	"testing"
	"time"
)

type ActivationDatabaseStub struct {
	pending     []*model.PendingActivation
	sent        []uint64
	incremented []uint64
	queryError  error
}

func (d *ActivationDatabaseStub) GetPendingActivations(_ uint64, _ int, _ int) ([]*model.PendingActivation, error) {
	if d.queryError != nil {
		return nil, d.queryError
	}
	return d.pending, nil
}

func (d *ActivationDatabaseStub) MarkActionSent(actionId uint64, _ time.Time) error {
	d.sent = append(d.sent, actionId)
	return nil
}

func (d *ActivationDatabaseStub) IncrementActionAttempts(actionId uint64) error {
	d.incremented = append(d.incremented, actionId)
	return nil
}

type ActivationMailStub struct {
	failFor map[string]bool
	sentTo  []string
}

func (m *ActivationMailStub) SendActivate(to string, _ string) error {
	if m.failFor[to] {
		return fmt.Errorf("smtp unavailable")
	}
	m.sentTo = append(m.sentTo, to)
	return nil
}

func sender(d *ActivationDatabaseStub, m *ActivationMailStub) *ActivationSender {
	return NewActivationSender(d, m, true, zap.NewNop())
}

func TestSendsPendingActivations(t *testing.T) {
	d := &ActivationDatabaseStub{pending: []*model.PendingActivation{
		{ActionId: 1, Token: "t1", Email: "one@example.com"},
		{ActionId: 2, Token: "t2", Email: "two@example.com"},
	}}
	m := &ActivationMailStub{failFor: map[string]bool{}}

	sent, err := sender(d, m).SendPending(time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 2, sent)
	assert.Equal(t, []string{"one@example.com", "two@example.com"}, m.sentTo)
	assert.Equal(t, []uint64{1, 2}, d.sent)
	assert.Empty(t, d.incremented)
}

func TestNothingPending(t *testing.T) {
	d := &ActivationDatabaseStub{}
	m := &ActivationMailStub{failFor: map[string]bool{}}

	sent, err := sender(d, m).SendPending(time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 0, sent)
}

func TestFailedSendIsNotMarkedSent(t *testing.T) {
	d := &ActivationDatabaseStub{pending: []*model.PendingActivation{
		{ActionId: 7, Token: "t7", Email: "bad@example.com"},
	}}
	m := &ActivationMailStub{failFor: map[string]bool{"bad@example.com": true}}

	sent, err := sender(d, m).SendPending(time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 0, sent)
	assert.Empty(t, d.sent)
	assert.Equal(t, []uint64{7}, d.incremented)
}

func TestOneFailureDoesNotBlockTheRest(t *testing.T) {
	d := &ActivationDatabaseStub{pending: []*model.PendingActivation{
		{ActionId: 1, Token: "t1", Email: "bad@example.com"},
		{ActionId: 2, Token: "t2", Email: "good@example.com"},
	}}
	m := &ActivationMailStub{failFor: map[string]bool{"bad@example.com": true}}

	sent, err := sender(d, m).SendPending(time.Now())

	assert.NoError(t, err)
	assert.Equal(t, 1, sent)
	assert.Equal(t, []string{"good@example.com"}, m.sentTo)
	assert.Equal(t, []uint64{2}, d.sent)
	assert.Equal(t, []uint64{1}, d.incremented)
}

func TestQueryErrorIsReturned(t *testing.T) {
	d := &ActivationDatabaseStub{queryError: fmt.Errorf("database down")}
	m := &ActivationMailStub{failFor: map[string]bool{}}

	sent, err := sender(d, m).SendPending(time.Now())

	assert.Error(t, err)
	assert.Equal(t, 0, sent)
	assert.Empty(t, m.sentTo)
}

func TestDisabledSenderDoesNotStart(t *testing.T) {
	d := &ActivationDatabaseStub{pending: []*model.PendingActivation{
		{ActionId: 1, Token: "t1", Email: "one@example.com"},
	}}
	m := &ActivationMailStub{failFor: map[string]bool{}}

	assert.NoError(t, NewActivationSender(d, m, false, zap.NewNop()).Start())
	assert.Empty(t, m.sentTo)
}
