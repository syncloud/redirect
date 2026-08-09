package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeUsageStore struct {
	domains []string
	traffic map[string]int64
	asked   []string
}

func (f *fakeUsageStore) GetUserDomainNames(_ int64) ([]string, error) {
	return f.domains, nil
}

func (f *fakeUsageStore) GetRelayTraffic(names []string, _ string) (int64, error) {
	f.asked = names
	var total int64
	for _, name := range names {
		total += f.traffic[name]
	}
	return total, nil
}

func (f *fakeUsageStore) IsRelayEnabledForUser(_ int64) (bool, error) {
	return true, nil
}

func TestUsage_CountsEveryDomainOfTheUser(t *testing.T) {
	store := &fakeUsageStore{
		domains: []string{"alice.syncloud.it", "spare.syncloud.it"},
		traffic: map[string]int64{
			"alice.syncloud.it":         100,
			"spare.syncloud.it":         3,
			"somebody-else.syncloud.it": 9000,
		},
	}

	used, err := NewUsage(store, nil).UsedBytes(1)

	assert.NoError(t, err)
	assert.Equal(t, int64(103), used)
}

func TestUsage_AsksForTheDomainsOfTheUser(t *testing.T) {
	store := &fakeUsageStore{domains: []string{"alice.syncloud.it"}}

	_, err := NewUsage(store, nil).UsedBytes(1)

	assert.NoError(t, err)
	assert.Equal(t, []string{"alice.syncloud.it"}, store.asked)
}

func TestUsage_NoDomainsMeansNoTraffic(t *testing.T) {
	store := &fakeUsageStore{}

	used, err := NewUsage(store, nil).UsedBytes(1)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), used)
	assert.Empty(t, store.asked)
}
