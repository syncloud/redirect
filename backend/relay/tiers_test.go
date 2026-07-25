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
