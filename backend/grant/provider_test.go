package grant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func provider(store Store, at time.Time) *Provider {
	return NewProvider(store, func() time.Time { return at })
}

func TestTheYearCountsAsRunningWhileItHasLongToGo(t *testing.T) {
	store := newStoreStub()
	store.addRedemption(1, now.AddDate(0, 12, 0))

	active, err := provider(store, now).IsActive("1")

	assert.NoError(t, err)
	assert.True(t, active)
}

func TestTheLastFortnightOfTheYearWarnsInsteadOfRunningOutSilently(t *testing.T) {
	store := newStoreStub()
	endsAt := now.AddDate(0, 0, 10)
	store.addRedemption(1, endsAt)

	at := provider(store, now)
	active, err := at.IsActive("1")
	assert.NoError(t, err)
	assert.False(t, active, "the cleaner warns while the paid period has not ended")

	paidThrough, err := at.PaidThrough("1")
	assert.NoError(t, err)
	assert.Equal(t, endsAt, paidThrough)
	assert.True(t, now.Before(paidThrough), "the domain must outlive the warning")
}

func TestOnceTheYearIsUpThePeriodHasEnded(t *testing.T) {
	store := newStoreStub()
	endsAt := now.AddDate(0, 0, -1)
	store.addRedemption(1, endsAt)

	at := provider(store, now)
	active, err := at.IsActive("1")
	assert.NoError(t, err)
	assert.False(t, active)

	paidThrough, err := at.PaidThrough("1")
	assert.NoError(t, err)
	assert.False(t, now.Before(paidThrough))
}

func TestAnUnknownRedemptionIsAnErrorRatherThanAFreeRide(t *testing.T) {
	store := newStoreStub()

	_, err := provider(store, now).IsActive("404")
	assert.Error(t, err)

	_, err = provider(store, now).PaidThrough("404")
	assert.Error(t, err)

	_, err = provider(store, now).IsActive("not-a-number")
	assert.Error(t, err)
}

func TestGivingUpAGrantNeedsNothingCancelledOutside(t *testing.T) {
	assert.NoError(t, provider(newStoreStub(), now).Unsubscribe("1"))
}
