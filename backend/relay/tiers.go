package relay

import (
	"strings"

	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

const SmtpProxySuffix = "-smtp"

type DomainStore interface {
	GetDomainByName(name string) (*model.Domain, error)
	GetUser(id int64) (*model.User, error)
}

type Tiers struct {
	store  DomainStore
	free   int64
	pro    int64
	max    int64
	logger *zap.Logger
}

func NewTiers(store DomainStore, free int64, pro int64, max int64, logger *zap.Logger) *Tiers {
	return &Tiers{store: store, free: free, pro: pro, max: max, logger: logger}
}

func (t *Tiers) OwnerLimit(name string) (int64, int64, bool) {
	domain, err := t.store.GetDomainByName(strings.TrimSuffix(name, SmtpProxySuffix))
	if err != nil || domain == nil {
		return 0, 0, false
	}
	return domain.UserId, t.UserLimit(domain.UserId), true
}

func (t *Tiers) UserLimit(userId int64) int64 {
	user, err := t.store.GetUser(userId)
	if err != nil || user == nil || !user.IsSubscribed() {
		return t.free
	}
	if user.Plan != nil && *user.Plan == model.PlanMax {
		return t.max
	}
	return t.pro
}
