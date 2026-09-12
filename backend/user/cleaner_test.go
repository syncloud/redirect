package user

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/service"
	"testing"
	"time"
)

type DatabaseStub struct {
	user            *model.User
	userDeleted     int64
	actionDeleted   int64
	actions         []*model.Action
	deleteActionErr error
}

func (d *DatabaseStub) DeleteUser(userId int64) error {
	d.userDeleted = userId
	return nil
}

func (d *DatabaseStub) DeleteActions(userId int64) error {
	d.actionDeleted = userId
	return nil
}

func (d *DatabaseStub) DeleteActionsCreatedBefore(actionTypeId uint64, before time.Time) (int64, error) {
	if d.deleteActionErr != nil {
		return 0, d.deleteActionErr
	}
	var kept []*model.Action
	deleted := int64(0)
	for _, action := range d.actions {
		if action.ActionTypeId == actionTypeId && action.CreatedAt.Before(before) {
			deleted++
			continue
		}
		kept = append(kept, action)
	}
	d.actions = kept
	return deleted, nil
}

func (d *DatabaseStub) UpdateUser(user *model.User) error {
	d.user = user
	return nil
}

func (d *DatabaseStub) GetUser(_ int64) (*model.User, error) {
	return d.user, nil
}

func (d *DatabaseStub) GetNextUserId(_ int64) (int64, error) {
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
	trial        bool
	lockSoon     bool
	locked       bool
	removed      bool
	unsubscribed bool
}

func (m *MailStub) SendPlanUnSubscribed(_ string) error {
	m.unsubscribed = true
	return nil
}

func (m *MailStub) SendAccountRemoved(_ string) error {
	m.removed = true
	return nil
}

func (m *MailStub) SendAccountLockSoon(_ string) error {
	m.lockSoon = true
	return nil
}

func (m *MailStub) SendAccountLocked(_ string) error {
	m.locked = true
	return nil
}

func (m *MailStub) SendTrial(_ string) error {
	m.trial = true
	return nil
}

type CheckerStub struct {
	active bool
}

func (c *CheckerStub) IsActive(_ int, _ string) (bool, error) {
	return c.active, nil
}

type RemoverStub struct {
	domainsRemoved       bool
	domainsRemovedUserId int64
}

func (r *RemoverStub) DeleteAllDomains(userId int64) error {
	r.domainsRemoved = true
	r.domainsRemovedUserId = userId
	return nil
}

func newTestCleaner(database Database, state State, mail Mail, remover Remover, checker SubscriptionChecker) *Cleaner {
	return NewCleaner(database, state, mail, remover, checker, service.ActionPassword, time.Hour, true, log.Default())
}

func TestCleaner_Purge_ExpiredPasswordActions(t *testing.T) {
	now := time.Now()
	expired := &model.Action{Id: 1, ActionTypeId: service.ActionPassword, CreatedAt: now.Add(-2 * time.Hour)}
	fresh := &model.Action{Id: 2, ActionTypeId: service.ActionPassword, CreatedAt: now.Add(-time.Minute)}
	activation := &model.Action{Id: 3, ActionTypeId: service.ActionActivate, CreatedAt: now.Add(-10000 * time.Hour)}
	database := &DatabaseStub{actions: []*model.Action{expired, fresh, activation}}

	cleaner := newTestCleaner(database, &StateStub{}, &MailStub{}, &RemoverStub{}, &CheckerStub{})
	err := cleaner.PurgeExpiredPasswordActions(now)

	assert.NoError(t, err)
	assert.Equal(t, []*model.Action{fresh, activation}, database.actions)
}

func TestCleaner_Purge_Error_DoesNotStopTheUserCursor(t *testing.T) {
	now := time.Now()
	subscriptionId := "1"
	database := &DatabaseStub{
		user:            &model.User{Id: 2, SubscriptionId: &subscriptionId},
		deleteActionErr: errors.New("db is down"),
	}
	state := &StateStub{userId: 1}

	cleaner := newTestCleaner(database, state, &MailStub{}, &RemoverStub{}, &CheckerStub{active: true})
	purgeErr := cleaner.PurgeExpiredPasswordActions(now)
	cleanErr := cleaner.Clean(now)

	assert.Error(t, purgeErr)
	assert.NoError(t, cleanErr)
	assert.Equal(t, int64(2), state.userId)
}

func TestCleaner_Clean_Subscribed_Skip(t *testing.T) {
	now := time.Now()
	subscriptionId := "1"
	database := &DatabaseStub{user: &model.User{Id: 2, SubscriptionId: &subscriptionId}}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	checker := &CheckerStub{active: true}
	cleaner := newTestCleaner(database, state, mail, remover, checker)
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
	user.Lock(now)
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
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
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
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
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
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
	user.TrialEmailSent(now)
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
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
	user := &model.User{Id: 2}
	user.TrialEmailSent(now.AddDate(0, 0, -21))
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsLockEmailSent())
	assert.False(t, mail.trial)
	assert.True(t, mail.lockSoon)
	assert.False(t, mail.locked)
}

func TestCleaner_Clean_StatusLockSoonSent_LessThan10Days_Skip(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2}
	user.LockEmailSent(now.AddDate(0, 0, -9))
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsLockEmailSent())
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.False(t, mail.locked)
	assert.False(t, remover.domainsRemoved)
}

func TestCleaner_Clean_StatusLockSoonSent_MoreThan10Days_Lock(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2}
	user.LockEmailSent(now.AddDate(0, 0, -11))
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.True(t, user.IsLocked())
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.True(t, mail.locked)
	assert.True(t, remover.domainsRemoved)
	assert.Equal(t, int64(2), remover.domainsRemovedUserId)
}

func TestCleaner_Clean_StatusLocked_MoreThan10Days_Remove(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2}
	user.Lock(now.AddDate(0, 0, -11))
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.True(t, remover.domainsRemoved)
	assert.Equal(t, int64(2), remover.domainsRemovedUserId)
	assert.Equal(t, int64(2), database.actionDeleted)
	assert.Equal(t, int64(2), database.userDeleted)
}

func TestCleaner_Clean_StatusLocked_LessThan10Days_NotRemove(t *testing.T) {
	now := time.Now()
	user := &model.User{Id: 2}
	user.Lock(now.AddDate(0, 0, -9))
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.False(t, remover.domainsRemoved)
	assert.Equal(t, int64(0), remover.domainsRemovedUserId)
	assert.Equal(t, int64(0), database.actionDeleted)
	assert.Equal(t, int64(0), database.userDeleted)
}

func TestCleaner_Clean_Crypto_SkipForNow(t *testing.T) {
	now := time.Now()
	subscriptionId := "1"
	subscriptionType := model.SubscriptionTypeCrypto
	user := &model.User{Id: 2, SubscriptionId: &subscriptionId, SubscriptionType: &subscriptionType}
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	checker := &CheckerStub{}

	cleaner := newTestCleaner(database, state, mail, remover, checker)
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.False(t, mail.locked)
	assert.False(t, mail.unsubscribed)
	assert.False(t, remover.domainsRemoved)
	assert.Equal(t, int64(0), remover.domainsRemovedUserId)
	assert.Equal(t, int64(0), database.actionDeleted)
	assert.Equal(t, int64(0), database.userDeleted)
}

func TestCleaner_Clean_PayPalUnsubscribe_Lock(t *testing.T) {
	now := time.Now()
	subscriptionId := "1"
	subscriptionType := model.SubscriptionTypePayPal
	user := &model.User{Id: 2, SubscriptionId: &subscriptionId, SubscriptionType: &subscriptionType}
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	checker := &CheckerStub{}

	cleaner := newTestCleaner(database, state, mail, remover, checker)
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.True(t, user.IsLocked())
	assert.False(t, mail.trial)
	assert.False(t, mail.lockSoon)
	assert.True(t, mail.unsubscribed)
	assert.True(t, remover.domainsRemoved)
	assert.Equal(t, int64(2), remover.domainsRemovedUserId)
}

func TestCleaner_Clean_StripeInactive_Lock(t *testing.T) {
	now := time.Now()
	subscriptionId := "1"
	subscriptionType := model.SubscriptionTypeStripe
	user := &model.User{Id: 2, SubscriptionId: &subscriptionId, SubscriptionType: &subscriptionType}
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	checker := &CheckerStub{active: false}

	cleaner := newTestCleaner(database, state, mail, remover, checker)
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.True(t, user.IsLocked())
	assert.True(t, mail.unsubscribed)
	assert.True(t, remover.domainsRemoved)
	assert.Equal(t, int64(2), remover.domainsRemovedUserId)
}

func TestCleaner_Clean_StripeActive_Skip(t *testing.T) {
	now := time.Now()
	subscriptionId := "1"
	subscriptionType := model.SubscriptionTypeStripe
	user := &model.User{Id: 2, SubscriptionId: &subscriptionId, SubscriptionType: &subscriptionType}
	database := &DatabaseStub{user: user}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	checker := &CheckerStub{active: true}

	cleaner := newTestCleaner(database, state, mail, remover, checker)
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), state.userId)
	assert.False(t, user.IsLocked())
	assert.False(t, mail.unsubscribed)
	assert.False(t, remover.domainsRemoved)
}

func TestCleaner_Clean_Last_ResetToZero(t *testing.T) {
	now := time.Now()
	database := &DatabaseStub{user: nil}
	state := &StateStub{userId: 1}
	mail := &MailStub{}
	remover := &RemoverStub{}
	cleaner := newTestCleaner(database, state, mail, remover, &CheckerStub{})
	err := cleaner.Clean(now)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), state.userId)
}
