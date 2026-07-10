package subscription

import (
	"fmt"
	"github.com/syncloud/redirect/model"
)

type Provider interface {
	Unsubscribe(id string) error
	IsActive(id string) (bool, error)
}

type Router struct {
	paypal Provider
	stripe Provider
}

func NewRouter(paypal Provider, stripe Provider) *Router {
	return &Router{
		paypal: paypal,
		stripe: stripe,
	}
}

func (r *Router) Unsubscribe(subscriptionType int, id string) error {
	switch subscriptionType {
	case model.SubscriptionTypePayPal:
		return r.paypal.Unsubscribe(id)
	case model.SubscriptionTypeStripe:
		return r.stripe.Unsubscribe(id)
	default:
		return fmt.Errorf("unsupported subscription type: %d", subscriptionType)
	}
}

func (r *Router) IsActive(subscriptionType int, id string) (bool, error) {
	switch subscriptionType {
	case model.SubscriptionTypePayPal:
		return r.paypal.IsActive(id)
	case model.SubscriptionTypeStripe:
		return r.stripe.IsActive(id)
	default:
		return false, fmt.Errorf("unsupported subscription type: %d", subscriptionType)
	}
}
