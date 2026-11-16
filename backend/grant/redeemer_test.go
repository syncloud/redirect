package grant

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

var now = time.Date(2026, 11, 15, 12, 0, 0, 0, time.UTC)

func redeemer(store Store, subscriber Subscriber) *Redeemer {
	return NewRedeemer(store, subscriber, zap.NewNop())
}

func TestRedeemGivesTheBuyerTheYearTheyPaidFor(t *testing.T) {
	store := newStoreStub()
	store.add("AAAA-BBBB-CCCC", 12, 1)
	subscriber := &subscriberStub{}
	user := &model.User{Id: 7}

	redemption, err := redeemer(store, subscriber).Redeem(user, "AAAA-BBBB-CCCC", now)

	assert.NoError(t, err)
	assert.Equal(t, now.AddDate(0, 12, 0), redemption.EndsAt)
	assert.Equal(t, model.SubscriptionTypeGrant, subscriber.kind)
	assert.Equal(t, "pro", subscriber.plan)
	assert.True(t, user.IsGrant())
	assert.True(t, user.IsAutoManaged())
}

func TestRedeemSpendsOneUseOfTheCode(t *testing.T) {
	store := newStoreStub()
	given := store.add("AAAA-BBBB-CCCC", 12, 1)

	_, err := redeemer(store, &subscriberStub{}).Redeem(&model.User{Id: 7}, "AAAA-BBBB-CCCC", now)

	assert.NoError(t, err)
	assert.Equal(t, 0, given.UsesLeft)
}

func TestRedeemRefusesTheSameCodeTwiceWhenOnlyOneWasPaidFor(t *testing.T) {
	store := newStoreStub()
	store.add("AAAA-BBBB-CCCC", 12, 1)
	service := redeemer(store, &subscriberStub{})

	_, err := service.Redeem(&model.User{Id: 7}, "AAAA-BBBB-CCCC", now)
	assert.NoError(t, err)

	_, err = service.Redeem(&model.User{Id: 8}, "AAAA-BBBB-CCCC", now)
	assert.ErrorIs(t, err, ErrUsedUp)
}

func TestRedeemLetsAResellerCodeServeEveryDeviceItWasBoughtFor(t *testing.T) {
	store := newStoreStub()
	store.add("ACME-2026-GOLD", 3, 3)
	service := redeemer(store, &subscriberStub{})

	for id := int64(1); id <= 3; id++ {
		redemption, err := service.Redeem(&model.User{Id: id}, "ACME-2026-GOLD", now)
		assert.NoError(t, err)
		assert.Equal(t, now.AddDate(0, 3, 0), redemption.EndsAt,
			"a reseller batch carries its own period, not the shop's")
	}

	_, err := service.Redeem(&model.User{Id: 4}, "ACME-2026-GOLD", now)
	assert.ErrorIs(t, err, ErrUsedUp)
}

func TestRedeemRefusesACodeWeNeverIssued(t *testing.T) {
	store := newStoreStub()

	_, err := redeemer(store, &subscriberStub{}).Redeem(&model.User{Id: 7}, "NOPE-NOPE-NOPE", now)

	assert.ErrorIs(t, err, ErrNoCode)
}

func TestRedeemReadsACodeTheWayItWasTyped(t *testing.T) {
	store := newStoreStub()
	store.add("AAAA-BBBB-CCCC", 12, 1)

	_, err := redeemer(store, &subscriberStub{}).Redeem(&model.User{Id: 7}, "  aaaa-bbbb-cccc ", now)

	assert.NoError(t, err)
}

func TestRedeemKeepsTheUseWhenTheSubscriptionCouldNotBeTaken(t *testing.T) {
	store := newStoreStub()
	given := store.add("AAAA-BBBB-CCCC", 12, 1)
	subscriber := &subscriberStub{err: errors.New("you have an existing subscription")}

	_, err := redeemer(store, subscriber).Redeem(&model.User{Id: 7}, "AAAA-BBBB-CCCC", now)

	assert.Error(t, err)
	assert.Equal(t, 1, given.UsesLeft, "a code must survive a failed redemption")
	assert.Len(t, store.deleted, 1)
	assert.Empty(t, store.redemptions)
}
