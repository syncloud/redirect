package outbound

import (
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type UserStore interface {
	GetUser(id int64) (*model.User, error)
}

type Tiers struct {
	store  UserStore
	free   int64
	pro    int64
	max    int64
	logger *zap.Logger
}

func NewTiers(store UserStore, free int64, pro int64, max int64, logger *zap.Logger) *Tiers {
	return &Tiers{store: store, free: free, pro: pro, max: max, logger: logger}
}

func (t *Tiers) MessageLimit(userId int64) int64 {
	user, err := t.store.GetUser(userId)
	if err != nil || user == nil || !user.IsSubscribed() {
		return t.free
	}
	if user.Plan != nil && *user.Plan == model.PlanMax {
		return t.max
	}
	return t.pro
}
