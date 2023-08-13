package model

import "time"

const (
	StatusCreated        int64 = 0
	StatusTrialEmailSent int64 = 1
	StatusLockEmailSent  int64 = 2
	StatusLocked         int64 = 3
	StatusSubscribed     int64 = 4
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
	Status              int64     `json:"status,omitempty"`
	RegisteredAt        time.Time `json:"registered_at,omitempty"`
}

func (u *User) IsSubscribed() bool {
	return u.SubscriptionId != nil
}

func (u *User) IsNDaysSinceRegistration(now time.Time, days int) bool {
	return u.RegisteredAt.Before(now.AddDate(0, 0, -days))
}

func (u *User) IsReadyForLockEmail(now time.Time) bool {
	return u.Status == StatusTrialEmailSent && u.IsNDaysSinceRegistration(now, 20)
}

func (u *User) IsReadyForAccountLock(now time.Time) bool {
	return u.Status == StatusLockEmailSent && u.IsNDaysSinceRegistration(now, 30)
}

func (u *User) IsStatusCreated() bool {
	return u.Status == StatusCreated
}

func (u *User) TrialEmailSent() {
	u.Status = StatusTrialEmailSent
}

func (u *User) IsTrialEmailSent() bool {
	return u.Status == StatusTrialEmailSent
}

func (u *User) IsLockEmailSent() bool {
	return u.Status == StatusLockEmailSent
}

func (u *User) LockEmailSent() {
	u.Status = StatusLockEmailSent
}

func (u *User) Lock() {
	u.Status = StatusLocked
}

func (u *User) IsLocked() bool {
	return u.Status == StatusLocked
}

func (u *User) Subscribe(subscriptionId string) {
	u.SubscriptionId = &subscriptionId
	u.Status = StatusSubscribed
}
