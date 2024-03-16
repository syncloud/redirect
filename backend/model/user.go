package model

import "time"

const (
	StatusCreated        int64 = 0
	StatusTrialEmailSent int64 = 1
	StatusLockEmailSent  int64 = 2
	StatusLocked         int64 = 3
	StatusSubscribed     int64 = 4

	SubscriptionTypePayPal = 1
	SubscriptionTypeCrypto = 2
)

type User struct {
	Id                  int64     `json:"-"`
	Email               string    `json:"email,omitempty"`
	PasswordHash        string    `json:"-"`
	Active              bool      `json:"active,omitempty"`
	UpdateToken         string    `json:"update_token,omitempty"`
	NotificationEnabled bool      `json:"notification_enabled"`
	Timestamp           time.Time `json:"timestamp,omitempty"`
	SubscriptionId      *string   `json:"subscription_id,omitempty"`
	SubscriptionType    *int      `json:"subscription_type,omitempty"`
	Status              int64     `json:"status,omitempty"`
	StatusAt            time.Time `json:"status_at,omitempty"`
	RegisteredAt        time.Time `json:"registered_at,omitempty"`
}

func (u *User) IsSubscribed() bool {
	return u.SubscriptionId != nil
}

func (u *User) IsNDaysSinceStatus(now time.Time, days int) bool {
	return u.StatusAt.Before(now.AddDate(0, 0, -days))
}

func (u *User) IsReadyForLockEmail(now time.Time) bool {
	return u.Status == StatusTrialEmailSent && u.IsNDaysSinceStatus(now, 20)
}

func (u *User) IsReadyForAccountLock(now time.Time) bool {
	return u.Status == StatusLockEmailSent && u.IsNDaysSinceStatus(now, 10)
}

func (u *User) IsReadyForAccountRemove(now time.Time) bool {
	return u.Status == StatusLocked && u.IsNDaysSinceStatus(now, 10)
}

func (u *User) IsStatusCreated() bool {
	return u.Status == StatusCreated
}

func (u *User) TrialEmailSent(now time.Time) {
	u.StatusAt = now
	u.Status = StatusTrialEmailSent
}

func (u *User) IsTrialEmailSent() bool {
	return u.Status == StatusTrialEmailSent
}

func (u *User) IsLockEmailSent() bool {
	return u.Status == StatusLockEmailSent
}

func (u *User) LockEmailSent(now time.Time) {
	u.StatusAt = now
	u.Status = StatusLockEmailSent
}

func (u *User) Lock(now time.Time) {
	u.StatusAt = now
	u.Status = StatusLocked
}

func (u *User) IsLocked() bool {
	return u.Status == StatusLocked
}

func (u *User) Subscribe(subscriptionId string, subscriptionType int) {
	u.SubscriptionId = &subscriptionId
	u.SubscriptionType = &subscriptionType
	u.Status = StatusSubscribed
}

func (u *User) UnSubscribe(now time.Time) {
	u.SubscriptionId = nil
	u.SubscriptionType = nil
	u.StatusAt = now
	u.Status = StatusLocked
}

func (u *User) IsPayPal() bool {
	return u.SubscriptionType != nil && *u.SubscriptionType == SubscriptionTypePayPal
}
