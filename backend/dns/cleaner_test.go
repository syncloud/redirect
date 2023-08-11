package dns

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
	"time"
)

type DatabaseStub struct {
	domain *model.Domain
	user   *model.User
}

func (d *DatabaseStub) GetUser(id int64) (*model.User, error) {
	return d.user, nil
}

func (d *DatabaseStub) GetDomainTokenUpdatedBefore(before time.Time) (string, error) {
	return "token", nil
}

func (d *DatabaseStub) GetDomainByToken(token string) (*model.Domain, error) {
	return d.domain, nil
}

func (d *DatabaseStub) UpdateDomain(domain *model.Domain) error {
	d.domain = domain
	return nil
}

type RemoverStub struct {
	deleted bool
}

func (r *RemoverStub) DeleteDomainRecords(domain *model.Domain) error {
	r.deleted = true
	return nil
}

type MailStub struct {
	sent bool
}

func (m *MailStub) SendDnsCleanNotification(to string, userDomain string) error {
	m.sent = true
	return nil
}

type GraphiteStub struct {
	value float64
}

func (g *GraphiteStub) CounterAdd(name string, value float64) {
	g.value += value
}

func TestCleaner_Clean_NotRemoveDNSForSubscribedUsers(t *testing.T) {
	subscriptionId := "123"
	now := time.Now()
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com", LastUpdate: &now},
		user:   &model.User{SubscriptionId: &subscriptionId},
	}
	dns := &RemoverStub{}
	mail := &MailStub{}
	graphite := &GraphiteStub{}
	cleaner := NewCleaner(database, dns, mail, graphite)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.False(t, dns.deleted)
}

func TestCleaner_Clean_NotRemoveDNSLessThanAMonthOfInactivity(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com", LastUpdate: &now},
		user:   &model.User{},
	}
	dns := &RemoverStub{}
	mail := &MailStub{}
	graphite := &GraphiteStub{}
	cleaner := NewCleaner(database, dns, mail, graphite)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.False(t, dns.deleted)
}

func TestCleaner_Clean_RemoveDNSMoreThanAMonthOfInactivity(t *testing.T) {
	now := time.Now()
	timestamp := now.AddDate(0, -1, -1)
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com", LastUpdate: &timestamp},
		user:   &model.User{Email: "test"},
	}
	dns := &RemoverStub{}
	mail := &MailStub{}
	graphite := &GraphiteStub{}
	cleaner := NewCleaner(database, dns, mail, graphite)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.True(t, dns.deleted)
}

func TestCleaner_Clean_RemoveDNSUnknownLastUpdate(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com"},
		user:   &model.User{Email: "test"},
	}
	dns := &RemoverStub{}
	mail := &MailStub{}
	graphite := &GraphiteStub{}
	cleaner := NewCleaner(database, dns, mail, graphite)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.True(t, dns.deleted)
}

func TestCleaner_Clean_RemoveUpdateTime(t *testing.T) {
	now := time.Now()
	timestamp := now.AddDate(0, -1, -1)
	domain := &model.Domain{Id: 1, Name: "test.com", LastUpdate: &timestamp}
	database := &DatabaseStub{
		domain: domain,
		user:   &model.User{},
	}
	dns := &RemoverStub{}
	mail := &MailStub{}
	graphite := &GraphiteStub{}
	cleaner := NewCleaner(database, dns, mail, graphite)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.True(t, dns.deleted)
	assert.Equal(t, &now, domain.LastUpdate)
}

func TestCleaner_Clean_RemoveSendNotification(t *testing.T) {
	now := time.Now()
	timestamp := now.AddDate(0, -1, -1)
	domain := &model.Domain{Id: 1, Name: "test.com", LastUpdate: &timestamp}
	database := &DatabaseStub{
		domain: domain,
		user:   &model.User{},
	}
	dns := &RemoverStub{}
	mail := &MailStub{}
	graphite := &GraphiteStub{}
	cleaner := NewCleaner(database, dns, mail, graphite)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.True(t, dns.deleted)
	assert.True(t, mail.sent)
}

func TestCleaner_Clean_NotRemoveNotSendNotification(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com", LastUpdate: &now},
		user:   &model.User{},
	}
	dns := &RemoverStub{}
	mail := &MailStub{}
	graphite := &GraphiteStub{}
	cleaner := NewCleaner(database, dns, mail, graphite)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.False(t, dns.deleted)
	assert.False(t, mail.sent)
}
