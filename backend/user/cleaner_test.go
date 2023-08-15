package user

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/model"
	"testing"
	"time"
)

type DatabaseStub struct {
	user *model.User
}

func (d *DatabaseStub) UpdateUser(user *model.User) error {
	d.user = user
	return nil
}

func (d *DatabaseStub) GetUser(id int64) (*model.User, error) {
	return d.user, nil
}

func (d *DatabaseStub) GetNextUserId(after int64) (int64, error) {
	if d.user == nil {
		return 0, nil
	}
	return d.user.Id, nil
}

type StateStub struct {
	userId int64
}

func (s *StateStub) Get() (int64, error) {
	return s.userId, nil
}

func (s *StateStub) Set(userId int64) error {
	s.userId = userId
	return nil
}

type MailStub struct {
	trial    bool
	lockSoon bool
	locked   bool
}

func (m *MailStub) SendAccountLockSoon(to string) error {
	m.lockSoon = true
	return nil
}

func (m *MailStub) SendAccountLocked(to string) error {
	m.locked = true
	return nil
}

func (m *MailStub) SendTrial(to string) error {
	m.trial = true
	return nil
}

type RemoverStub struct {
	domainsRemoved bool
}

func (r *RemoverStub) DeleteAllDomains(userId int64) error {
	r.domainsRemoved = true
	return nil
}

func TestCleaner_Clean_Subscribed_Skip(t *testing.T) {
	now := time.Now()
	subscriptionId := "1"
	database := &DatabaseStub{user: &model.User{Id: 2, SubscriptionId: &subscriptionId}}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)
}

func TestCleaner_Clean_Locked_Skip(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2}
	user.Lock()
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)
}

func TestCleaner_Clean_StatusCreated_SendTrial(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2, RegisteredAt: now}
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsTrialEmailSent())
	assert.True(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)
}

func TestCleaner_Clean_StatusCreated_SendTrial_Once(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2, RegisteredAt: now}
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsTrialEmailSent())
	assert.True(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)

	mail.trial = false
	err = cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)
}

func TestCleaner_Clean_StatusTrialSent_LessThan20Days_Skip(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2, RegisteredAt: now.AddDate(0, 0, -19)}
	user.TrialEmailSent()
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsTrialEmailSent())
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)
}

func TestCleaner_Clean_StatusTrialSent_MoreThan20Days_SendLockEmail(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2, RegisteredAt: now.AddDate(0, 0, -21)}
	user.TrialEmailSent()
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsLockEmailSent())
	assert.False(t, mail.trial)
	assert.True(t, mail.lockSoon)
	assert.False(t, mail.locked)
}

func TestCleaner_Clean_StatusLockSoonSent_LessThan30Days_Skip(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2, RegisteredAt: now.AddDate(0, 0, -29)}
	user.LockEmailSent()
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsLockEmailSent())
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)
	assert.False(t, remover.domainsRemoved)
}

func TestCleaner_Clean_StatusLockSoonSent_MoreThan30Days_Lock(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2, RegisteredAt: now.AddDate(0, 0, -31)}
	user.LockEmailSent()
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsLocked())
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.True(t, mail.locked)
	assert.True(t, remover.domainsRemoved)
}

func TestCleaner_Clean_Last_ResetToZero(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{user: nil}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := NewCleaner(database, state, mail, remover, true, log.Default())
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), state.userId)
}
