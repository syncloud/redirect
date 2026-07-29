package mailrelay

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeDomains struct{ domain *model.Domain }

func (f *fakeDomains) GetDomain(token string) (*model.Domain, error) {
	if f.domain == nil || f.domain.UpdateToken == nil || *f.domain.UpdateToken != token {
		return nil, nil
	}
	return f.domain, nil
}

type fakePlans struct{ limit int64 }

func (f *fakePlans) MessageLimit(_ int64) int64 { return f.limit }

type fakeUsage struct{ sent int64 }

func (f *fakeUsage) Sent(_ string) (int64, error) { return f.sent, nil }
func (f *fakeUsage) Increment(_ string, count int64) error {
	f.sent += count
	return nil
}

type fakeBlocklist struct{ blocked bool }

func (f *fakeBlocklist) Blocked(_ string) (bool, error) { return f.blocked, nil }

type fakeBlocker struct{ blocked []string }

func (f *fakeBlocker) Block(domain string, _ string) error {
	f.blocked = append(f.blocked, domain)
	return nil
}

func relayWith(limit int64, sent int64, blocked bool) (*Relay, *fakeUsage) {
	token := "the-token"
	usage := &fakeUsage{sent: sent}
	return New(
		&fakeDomains{domain: &model.Domain{Name: "device.syncloud.it", UserId: 7, UpdateToken: &token}},
		&fakePlans{limit: limit},
		usage,
		&fakeBlocklist{blocked: blocked},
		zap.NewNop()), usage
}

func TestAuthorize(t *testing.T) {
	relay, _ := relayWith(100, 0, false)
	domain, err := relay.Authorize("device.syncloud.it", "the-token")
	assert.NoError(t, err)
	assert.Equal(t, "device.syncloud.it", domain.Name)
}

func TestAuthorize_WrongToken(t *testing.T) {
	relay, _ := relayWith(100, 0, false)
	_, err := relay.Authorize("device.syncloud.it", "nope")
	assert.ErrorIs(t, err, ErrUnknownToken)
}

func TestAuthorize_TokenOfAnotherDomain(t *testing.T) {
	relay, _ := relayWith(100, 0, false)
	_, err := relay.Authorize("someone-else.syncloud.it", "the-token")
	assert.ErrorIs(t, err, ErrNotOwned)
}

func TestAuthorize_NotSubscribed(t *testing.T) {
	relay, _ := relayWith(0, 0, false)
	_, err := relay.Authorize("device.syncloud.it", "the-token")
	assert.ErrorIs(t, err, ErrNotAllowed)
}

func TestAuthorize_OverLimit(t *testing.T) {
	relay, _ := relayWith(100, 100, false)
	_, err := relay.Authorize("device.syncloud.it", "the-token")
	assert.ErrorIs(t, err, ErrOverLimit)
}

func TestAuthorize_Blocked(t *testing.T) {
	relay, _ := relayWith(100, 0, true)
	_, err := relay.Authorize("device.syncloud.it", "the-token")
	assert.ErrorIs(t, err, ErrBlocked)
}

func TestAllowed(t *testing.T) {
	relay, _ := relayWith(100, 0, false)
	domain := &model.Domain{Name: "device.syncloud.it"}
	assert.True(t, relay.Allowed(domain, "user@device.syncloud.it"))
	assert.True(t, relay.Allowed(domain, "User@Device.Syncloud.IT"))
	assert.False(t, relay.Allowed(domain, "user@evil.com"))
	assert.False(t, relay.Allowed(domain, "not-an-address"))
}

func TestSent_CountsRecipients(t *testing.T) {
	relay, usage := relayWith(100, 0, false)
	assert.NoError(t, relay.Sent(&model.Domain{Name: "device.syncloud.it"}, 3))
	assert.Equal(t, int64(3), usage.sent)
}

func complaint(t *testing.T, body string) *fakeBlocker {
	t.Helper()
	blocker := &fakeBlocker{}
	recorder := httptest.NewRecorder()
	NewComplaints(blocker, zap.NewNop()).Handle(
		recorder, httptest.NewRequest(http.MethodPost, "/mail/complaint", strings.NewReader(body)))
	assert.Equal(t, http.StatusOK, recorder.Code)
	return blocker
}

func TestComplaint_BlocksSenderDomain(t *testing.T) {
	blocker := complaint(t, `{"Type":"Notification","Message":"{\"notificationType\":\"Complaint\",\"mail\":{\"source\":\"user@device.syncloud.it\"}}"}`)
	assert.Equal(t, []string{"device.syncloud.it"}, blocker.blocked)
}

func TestComplaint_IgnoresDelivery(t *testing.T) {
	blocker := complaint(t, `{"Type":"Notification","Message":"{\"notificationType\":\"Delivery\",\"mail\":{\"source\":\"user@device.syncloud.it\"}}"}`)
	assert.Empty(t, blocker.blocked)
}
