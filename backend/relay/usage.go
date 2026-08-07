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
	return u.store.GetRelayTraffic(ProxyNames(domains), month())
}

// ProxyNames lists the frps proxies a set of domains can have: one carrying
// app traffic and one carrying inbound mail, both counting against the same
// allowance.
func ProxyNames(domains []string) []string {
	names := make([]string, 0, len(domains)*2)
	for _, domain := range domains {
		names = append(names, domain, domain+SmtpProxySuffix)
	}
	return names
}

func (u *Usage) Enabled(userId int64) (bool, error) {
	return u.store.IsRelayEnabledForUser(userId)
}

func (u *Usage) LimitBytes(userId int64) int64 {
	return u.tiers.UserLimit(userId)
}
