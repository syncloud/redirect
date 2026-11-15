package grant

import (
	"strconv"
	"strings"
	"time"

	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type Redeemer struct {
	store      Store
	subscriber Subscriber
	logger     *zap.Logger
}

func NewRedeemer(store Store, subscriber Subscriber, logger *zap.Logger) *Redeemer {
	return &Redeemer{store: store, subscriber: subscriber, logger: logger}
}

func (r *Redeemer) Redeem(user *model.User, code string, now time.Time) (*Redemption, error) {
	found, err := r.store.GetGrantByCode(Normalize(code))
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, ErrNoCode
	}
	if !found.Available() {
		return nil, ErrUsedUp
	}
	claimed, err := r.store.ClaimGrantUse(found.Id)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, ErrUsedUp
	}

	redemption := &Redemption{
		GrantId: found.Id,
		UserId:  user.Id,
		Plan:    found.Plan,
		EndsAt:  now.AddDate(0, found.Months, 0),
	}
	id, err := r.store.InsertRedemption(redemption)
	if err != nil {
		return nil, r.giveBack(found.Id, 0, err)
	}
	redemption.Id = id

	err = r.subscriber.Subscribe(user, strconv.FormatInt(id, 10), model.SubscriptionTypeGrant, found.Plan)
	if err != nil {
		return nil, r.giveBack(found.Id, id, err)
	}
	return redemption, nil
}

func (r *Redeemer) giveBack(grantId, redemptionId int64, cause error) error {
	if redemptionId != 0 {
		if err := r.store.DeleteRedemption(redemptionId); err != nil {
			r.logger.Error("unable to drop a redemption", zap.Error(err))
		}
	}
	if err := r.store.ReleaseGrantUse(grantId); err != nil {
		r.logger.Error("unable to give a grant use back", zap.Error(err))
	}
	return cause
}

func Normalize(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
