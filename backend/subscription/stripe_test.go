package subscription

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/log"
	"testing"
)

func TestStripe_Enabled(t *testing.T) {
	assert.True(t, NewStripe("sk", "price_m", "price_a", "s", "c", log.Default()).Enabled())
	assert.False(t, NewStripe("", "price_m", "price_a", "s", "c", log.Default()).Enabled())
	assert.False(t, NewStripe("sk", "", "price_a", "s", "c", log.Default()).Enabled())
	assert.False(t, NewStripe("sk", "price_m", "", "s", "c", log.Default()).Enabled())
}

func TestStripe_PriceId(t *testing.T) {
	stripe := NewStripe("sk", "price_m", "price_a", "s", "c", log.Default())

	monthly, err := stripe.priceId(StripePlanMonthly)
	assert.Nil(t, err)
	assert.Equal(t, "price_m", monthly)

	annual, err := stripe.priceId(StripePlanAnnual)
	assert.Nil(t, err)
	assert.Equal(t, "price_a", annual)

	_, err = stripe.priceId("unknown")
	assert.NotNil(t, err)
}

func TestStripe_CreateCheckout_NotConfigured(t *testing.T) {
	_, err := NewStripe("", "", "", "s", "c", log.Default()).CreateCheckout(StripePlanMonthly)
	assert.NotNil(t, err)
}
