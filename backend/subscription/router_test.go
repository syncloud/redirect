package subscription

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
)

type ProviderStub struct {
	name        string
	unsubscribe *string
	active      bool
}

func (p *ProviderStub) Unsubscribe(id string) error {
	p.unsubscribe = &id
	return nil
}

func (p *ProviderStub) IsActive(_ string) (bool, error) {
	return p.active, nil
}

func TestRouter_Unsubscribe_PayPal(t *testing.T) {
	paypal := &ProviderStub{name: "paypal"}
	stripe := &ProviderStub{name: "stripe"}
	router := NewRouter(paypal, stripe)

	err := router.Unsubscribe(model.SubscriptionTypePayPal, "sub-1")

	assert.Nil(t, err)
	assert.Equal(t, "sub-1", *paypal.unsubscribe)
	assert.Nil(t, stripe.unsubscribe)
}

func TestRouter_Unsubscribe_Stripe(t *testing.T) {
	paypal := &ProviderStub{name: "paypal"}
	stripe := &ProviderStub{name: "stripe"}
	router := NewRouter(paypal, stripe)

	err := router.Unsubscribe(model.SubscriptionTypeStripe, "sub-2")

	assert.Nil(t, err)
	assert.Equal(t, "sub-2", *stripe.unsubscribe)
	assert.Nil(t, paypal.unsubscribe)
}

func TestRouter_Unsubscribe_Unsupported(t *testing.T) {
	paypal := &ProviderStub{name: "paypal"}
	stripe := &ProviderStub{name: "stripe"}
	router := NewRouter(paypal, stripe)

	err := router.Unsubscribe(model.SubscriptionTypeCrypto, "sub-3")

	assert.NotNil(t, err)
	assert.Nil(t, paypal.unsubscribe)
	assert.Nil(t, stripe.unsubscribe)
}

func TestRouter_IsActive_RoutesByType(t *testing.T) {
	paypal := &ProviderStub{name: "paypal", active: true}
	stripe := &ProviderStub{name: "stripe", active: false}
	router := NewRouter(paypal, stripe)

	paypalActive, err := router.IsActive(model.SubscriptionTypePayPal, "sub-1")
	assert.Nil(t, err)
	assert.True(t, paypalActive)

	stripeActive, err := router.IsActive(model.SubscriptionTypeStripe, "sub-2")
	assert.Nil(t, err)
	assert.False(t, stripeActive)

	_, err = router.IsActive(model.SubscriptionTypeCrypto, "sub-3")
	assert.NotNil(t, err)
}
