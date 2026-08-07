package relay

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeUserStore struct{ user *model.User }

func (f *fakeUserStore) GetUser(_ int64) (*model.User, error) { return f.user, nil }

func limitFor(user *model.User) int64 {
	return NewTiers(&fakeUserStore{user: user}, 50, 2000, 10000, zap.NewNop()).MessageLimit(1)
}

func TestMessageLimit_UnsubscribedGetsFreeAllowance(t *testing.T) {
	assert.Equal(t, int64(50), limitFor(&model.User{}))
}

func TestMessageLimit_UnknownUserGetsFreeAllowance(t *testing.T) {
	assert.Equal(t, int64(50), limitFor(nil))
}

func subscribed(plan *string) *model.User {
	subscription := "sub-1"
	return &model.User{SubscriptionId: &subscription, Plan: plan}
}

func TestMessageLimit_Pro(t *testing.T) {
	assert.Equal(t, int64(2000), limitFor(subscribed(nil)))
}

func TestMessageLimit_Max(t *testing.T) {
	plan := model.PlanMax
	assert.Equal(t, int64(10000), limitFor(subscribed(&plan)))
}
