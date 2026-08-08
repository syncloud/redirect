package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeTiersStore struct {
	users map[int64]*model.User
}

func (f *fakeTiersStore) GetDomainByName(_ string) (*model.Domain, error) {
	return nil, nil
}

func (f *fakeTiersStore) GetUser(id int64) (*model.User, error) {
	return f.users[id], nil
}

func subscribedUser(plan string) *model.User {
	id := "sub"
	u := &model.User{SubscriptionId: &id}
	if plan != "" {
		p := plan
		u.Plan = &p
	}
	return u
}

func TestTiers_UserLimit(t *testing.T) {
	store := &fakeTiersStore{users: map[int64]*model.User{
		1: {},
		2: subscribedUser(""),
		3: subscribedUser(model.PlanPro),
		4: subscribedUser(model.PlanMax),
	}}
	tiers := NewTiers(store, 1, 10, 100, zap.NewNop())

	assert.Equal(t, int64(1), tiers.UserLimit(1))
	assert.Equal(t, int64(10), tiers.UserLimit(2))
	assert.Equal(t, int64(10), tiers.UserLimit(3))
	assert.Equal(t, int64(100), tiers.UserLimit(4))
	assert.Equal(t, int64(1), tiers.UserLimit(99))
}

type domainTiersStore struct {
	fakeTiersStore
	domains map[string]*model.Domain
	asked   []string
}

func (f *domainTiersStore) GetDomainByName(name string) (*model.Domain, error) {
	f.asked = append(f.asked, name)
	return f.domains[name], nil
}

func newDomainStore() *domainTiersStore {
	return &domainTiersStore{
		fakeTiersStore: fakeTiersStore{users: map[int64]*model.User{7: subscribedUser(model.PlanMax)}},
		domains:        map[string]*model.Domain{"alice.syncloud.it": {Name: "alice.syncloud.it", UserId: 7}},
	}
}

func TestTiers_OwnerLimit_WebProxy(t *testing.T) {
	store := newDomainStore()
	tiers := NewTiers(store, 1, 10, 100, zap.NewNop())

	userId, limit, ok := tiers.OwnerLimit("alice.syncloud.it")

	assert.True(t, ok)
	assert.Equal(t, int64(7), userId)
	assert.Equal(t, int64(100), limit)
}

func TestTiers_OwnerLimit_SmtpProxyBelongsToSameUser(t *testing.T) {
	store := newDomainStore()
	tiers := NewTiers(store, 1, 10, 100, zap.NewNop())

	userId, limit, ok := tiers.OwnerLimit("alice.syncloud.it" + SmtpProxySuffix)

	assert.True(t, ok)
	assert.Equal(t, int64(7), userId)
	assert.Equal(t, int64(100), limit)
	assert.Equal(t, []string{"alice.syncloud.it"}, store.asked)
}

func TestTiers_OwnerLimit_UnknownProxy(t *testing.T) {
	store := newDomainStore()
	tiers := NewTiers(store, 1, 10, 100, zap.NewNop())

	_, _, ok := tiers.OwnerLimit("stranger.syncloud.it" + SmtpProxySuffix)

	assert.False(t, ok)
}
