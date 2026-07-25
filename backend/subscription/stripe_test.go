package subscription

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/log"
	"github.com/syncloud/redirect/model"
	"testing"
)

func TestStripe_Enabled(t *testing.T) {
	assert.True(t, NewStripe("sk", "price_m", "price_a", "", "", "s", "c", log.Default()).Enabled())
	assert.False(t, NewStripe("", "price_m", "price_a", "", "", "s", "c", log.Default()).Enabled())
	assert.False(t, NewStripe("sk", "", "price_a", "", "", "s", "c", log.Default()).Enabled())
	assert.False(t, NewStripe("sk", "price_m", "", "", "", "s", "c", log.Default()).Enabled())
}

func TestStripe_MaxEnabled(t *testing.T) {
	assert.True(t, NewStripe("sk", "price_m", "price_a", "price_mm", "price_ma", "s", "c", log.Default()).MaxEnabled())
	assert.False(t, NewStripe("sk", "price_m", "price_a", "", "price_ma", "s", "c", log.Default()).MaxEnabled())
	assert.False(t, NewStripe("sk", "price_m", "price_a", "price_mm", "", "s", "c", log.Default()).MaxEnabled())
}

func TestStripe_PriceId(t *testing.T) {
	stripe := NewStripe("sk", "price_m", "price_a", "price_mm", "price_ma", "s", "c", log.Default())

	monthly, err := stripe.priceId(StripePlanMonthly)
	assert.Nil(t, err)
	assert.Equal(t, "price_m", monthly)

	annual, err := stripe.priceId(StripePlanAnnual)
	assert.Nil(t, err)
	assert.Equal(t, "price_a", annual)

	maxMonthly, err := stripe.priceId(StripePlanMaxMonthly)
	assert.Nil(t, err)
	assert.Equal(t, "price_mm", maxMonthly)

	maxAnnual, err := stripe.priceId(StripePlanMaxAnnual)
	assert.Nil(t, err)
	assert.Equal(t, "price_ma", maxAnnual)

	_, err = stripe.priceId("unknown")
	assert.NotNil(t, err)
}

func TestStripe_TierForPrice(t *testing.T) {
	stripe := NewStripe("sk", "price_m", "price_a", "price_mm", "price_ma", "s", "c", log.Default())
	assert.Equal(t, model.PlanPro, stripe.tierForPrice("price_m"))
	assert.Equal(t, model.PlanPro, stripe.tierForPrice("price_a"))
	assert.Equal(t, model.PlanMax, stripe.tierForPrice("price_mm"))
	assert.Equal(t, model.PlanMax, stripe.tierForPrice("price_ma"))
	assert.Equal(t, model.PlanPro, stripe.tierForPrice(""))
}

func TestStripe_CreateCheckout_NotConfigured(t *testing.T) {
	_, err := NewStripe("", "", "", "", "", "s", "c", log.Default()).CreateCheckout(StripePlanMonthly)
	assert.NotNil(t, err)
}
