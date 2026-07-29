package mailrelay

import (
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type UserStore interface {
	GetUser(id int64) (*model.User, error)
}

// Tiers maps a subscription to a monthly message allowance. The relay is a paid
// feature, so an unsubscribed user gets nothing rather than a small free tier.
type Tiers struct {
	store  UserStore
	pro    int64
	max    int64
	logger *zap.Logger
}

func NewTiers(store UserStore, pro int64, max int64, logger *zap.Logger) *Tiers {
	return &Tiers{store: store, pro: pro, max: max, logger: logger}
}

func (t *Tiers) MessageLimit(userId int64) int64 {
	user, err := t.store.GetUser(userId)
	if err != nil || user == nil || !user.IsSubscribed() {
		return 0
	}
	if user.Plan != nil && *user.Plan == model.PlanMax {
		return t.max
	}
	return t.pro
}
