package mailrelay

import "time"

type UsageStore interface {
	GetMailRelayUsageForUser(userId int64, yearMonth string) (int64, error)
	IsMailRelayEnabledForUser(userId int64) (bool, error)
}

// AccountUsage reports what a whole account has sent this month, across all of its
// devices, since the allowance is per plan rather than per device.
type AccountUsage struct {
	store UsageStore
	tiers *Tiers
	now   func() time.Time
}

func NewAccountUsage(store UsageStore, tiers *Tiers) *AccountUsage {
	return &AccountUsage{store: store, tiers: tiers, now: time.Now}
}

func (u *AccountUsage) UsedMessages(userId int64) (int64, error) {
	return u.store.GetMailRelayUsageForUser(userId, u.now().UTC().Format("2006-01"))
}

func (u *AccountUsage) Enabled(userId int64) (bool, error) {
	return u.store.IsMailRelayEnabledForUser(userId)
}

func (u *AccountUsage) LimitMessages(userId int64) int64 {
	return u.tiers.MessageLimit(userId)
}
