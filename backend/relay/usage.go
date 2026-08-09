package relay

type UsageStore interface {
	GetUserDomainNames(userId int64) ([]string, error)
	GetRelayTraffic(names []string, yearMonth string) (int64, error)
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
	domains, err := u.store.GetUserDomainNames(userId)
	if err != nil {
		return 0, err
	}
	return u.store.GetRelayTraffic(domains, month())
}

func (u *Usage) Enabled(userId int64) (bool, error) {
	return u.store.IsRelayEnabledForUser(userId)
}

func (u *Usage) LimitBytes(userId int64) int64 {
	return u.tiers.UserLimit(userId)
}
