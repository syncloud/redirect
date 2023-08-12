package dns

import (
	"github.com/smira/go-statsd"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
	"time"
)

type DatabaseStub struct {
	domain        *model.Domain
	user          *model.User
	domainUpdated bool
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
	d.domainUpdated = true
	return nil
}

type RemoverStub struct {
	deleted bool
}

func (r *RemoverStub) DeleteDomain(userId int64, domainName string) error {
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

type StatsStub struct {
	value int64
}

func (g *StatsStub) Incr(stat string, count int64, tags ...statsd.Tag) {
	g.value += count
}

func TestCleaner_Clean_ActiveSubscribed_NotRemoveDomain(t *testing.T) {
	subscriptionId := "123"
	now := time.Now()
	domain := &model.Domain{Id: 1, Name: "test.com", LastUpdate: &now}
	database := &DatabaseStub{
		domain: domain,
		user:   &model.User{SubscriptionId: &subscriptionId},
	}
	remover := &RemoverStub{}
	mail := &MailStub{}
	stats := &StatsStub{}
	cleaner := NewCleaner(database, remover, mail, stats)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.False(t, remover.deleted)
	assert.False(t, database.domainUpdated)
	assert.Equal(t, &now, database.domain.LastUpdate)
}

func TestCleaner_Clean_NotActiveSubscribed_NotRemoveDomain(t *testing.T) {
	subscriptionId := "123"
	now := time.Now()
	timestamp := now.AddDate(0, -1, -1)
	domain := &model.Domain{Id: 1, Name: "test.com", LastUpdate: &timestamp}
	database := &DatabaseStub{
		domain: domain,
		user:   &model.User{SubscriptionId: &subscriptionId},
	}
	remover := &RemoverStub{}
	mail := &MailStub{}
	stats := &StatsStub{}
	cleaner := NewCleaner(database, remover, mail, stats)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.False(t, remover.deleted)
	assert.True(t, database.domainUpdated)
	assert.Equal(t, &now, database.domain.LastUpdate)
}

func TestCleaner_Clean_Active_NotRemoveDomain(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com", LastUpdate: &now},
		user:   &model.User{},
	}
	remover := &RemoverStub{}
	mail := &MailStub{}
	stats := &StatsStub{}
	cleaner := NewCleaner(database, remover, mail, stats)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.False(t, remover.deleted)
	assert.False(t, database.domainUpdated)
}

func TestCleaner_Clean_NotActive_RemoveDomain(t *testing.T) {
	now := time.Now()
	timestamp := now.AddDate(0, -1, -1)
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com", LastUpdate: &timestamp},
		user:   &model.User{Email: "test"},
	}
	remover := &RemoverStub{}
	mail := &MailStub{}
	stats := &StatsStub{}
	cleaner := NewCleaner(database, remover, mail, stats)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.True(t, remover.deleted)
	assert.False(t, database.domainUpdated)
}

func TestCleaner_Clean_UnknownLastUpdate_RemoveDomain(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com"},
		user:   &model.User{Email: "test"},
	}
	remover := &RemoverStub{}
	mail := &MailStub{}
	stats := &StatsStub{}
	cleaner := NewCleaner(database, remover, mail, stats)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.True(t, remover.deleted)
	assert.False(t, database.domainUpdated)
}

func TestCleaner_Clean_RemoveSendNotification(t *testing.T) {
	now := time.Now()
	timestamp := now.AddDate(0, -1, -1)
	domain := &model.Domain{Id: 1, Name: "test.com", LastUpdate: &timestamp}
	database := &DatabaseStub{
		domain: domain,
		user:   &model.User{},
	}
	remover := &RemoverStub{}
	mail := &MailStub{}
	stats := &StatsStub{}
	cleaner := NewCleaner(database, remover, mail, stats)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.True(t, remover.deleted)
	assert.True(t, mail.sent)
	assert.False(t, database.domainUpdated)
}

func TestCleaner_Clean_NotRemoveNotSendNotification(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{
		domain: &model.Domain{Id: 1, Name: "test.com", LastUpdate: &now},
		user:   &model.User{},
	}
	remover := &RemoverStub{}
	mail := &MailStub{}
	stats := &StatsStub{}
	cleaner := NewCleaner(database, remover, mail, stats)
	err := cleaner.Clean(now)
	assert.Nil(t, err)
	assert.False(t, remover.deleted)
	assert.False(t, mail.sent)
}
