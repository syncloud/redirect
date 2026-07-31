package relay

type UsageStore interface {
	GetRelayUsageForUser(userId int64, yearMonth string) (int64, error)
	IsRelayEnabledForUser(userId int64) (bool, error)
}

type Usage struct {
	store UsageStore
	tiers *Tiers
}

func NewUsage(store UsageStore, tiers *Tiers) *Usage {
	return &Usage{store: store, tiers: tiers}
}

func (u *Usage) UsedBytes(userId int64) (int64, error) {
	return u.store.GetRelayUsageForUser(userId, month())
}

func (u *Usage) Enabled(userId int64) (bool, error) {
	return u.store.IsRelayEnabledForUser(userId)
}

func (u *Usage) LimitBytes(userId int64) int64 {
	return u.tiers.UserLimit(userId)
}
