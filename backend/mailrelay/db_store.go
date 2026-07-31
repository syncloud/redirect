package mailrelay

import "time"

type Store interface {
	GetMailRelayMessages(name string, yearMonth string) (int64, error)
	GetMailRelayUsageForUser(userId int64, yearMonth string) (int64, error)
	AddMailRelayMessages(name string, yearMonth string, messages int64) error
	GetMailRelayBounces(name string, yearMonth string) (int64, error)
	AddMailRelayBounces(name string, yearMonth string, bounces int64) error
	IsMailRelayBlocked(name string) (bool, error)
	BlockMailRelay(name string, reason string) error
}

type DbStore struct {
	store Store
	now   func() time.Time
}

func NewDbStore(store Store) *DbStore {
	return &DbStore{store: store, now: time.Now}
}

func (d *DbStore) yearMonth() string {
	return d.now().UTC().Format("2006-01")
}

func (d *DbStore) Sent(domain string) (int64, error) {
	return d.store.GetMailRelayMessages(domain, d.yearMonth())
}

func (d *DbStore) SentByUser(userId int64) (int64, error) {
	return d.store.GetMailRelayUsageForUser(userId, d.yearMonth())
}

func (d *DbStore) Increment(domain string, count int64) error {
	return d.store.AddMailRelayMessages(domain, d.yearMonth(), count)
}

func (d *DbStore) Bounced(domain string) (int64, error) {
	return d.store.GetMailRelayBounces(domain, d.yearMonth())
}

func (d *DbStore) Bounce(domain string, count int64) error {
	return d.store.AddMailRelayBounces(domain, d.yearMonth(), count)
}

func (d *DbStore) Blocked(domain string) (bool, error) {
	return d.store.IsMailRelayBlocked(domain)
}

func (d *DbStore) Block(domain string, reason string) error {
	return d.store.BlockMailRelay(domain, reason)
}
