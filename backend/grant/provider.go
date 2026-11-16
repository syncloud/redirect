package grant

import (
	"fmt"
	"strconv"
	"time"
)

const noticeDays = 14

type Provider struct {
	store Store
	now   func() time.Time
}

func NewProvider(store Store, now func() time.Time) *Provider {
	return &Provider{store: store, now: now}
}

func (p *Provider) IsActive(id string) (bool, error) {
	redemption, err := p.find(id)
	if err != nil {
		return false, err
	}
	return p.now().Before(redemption.EndsAt.AddDate(0, 0, -noticeDays)), nil
}

func (p *Provider) PaidThrough(id string) (time.Time, error) {
	redemption, err := p.find(id)
	if err != nil {
		return time.Time{}, err
	}
	return redemption.EndsAt, nil
}

func (p *Provider) Unsubscribe(_ string) error {
	return nil
}

func (p *Provider) find(id string) (*Redemption, error) {
	parsed, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("not a redemption id: %s", id)
	}
	redemption, err := p.store.GetRedemption(parsed)
	if err != nil {
		return nil, err
	}
	if redemption == nil {
		return nil, fmt.Errorf("no redemption %d", parsed)
	}
	return redemption, nil
}
