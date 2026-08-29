package product

import (
	"errors"
	"testing"
)

type stubCheckout struct{ name string }

func (s stubCheckout) Start(*Order, string) (string, error) { return s.name, nil }
func (s stubCheckout) Paid(string) (bool, int, error)       { return true, 0, nil }

func TestPicksTheProviderAsked(t *testing.T) {
	checkouts := NewCheckouts(stubCheckout{"paypal"}, stubCheckout{"stripe"})
	for _, name := range []string{"paypal", "stripe"} {
		got, err := checkouts.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		ref, _ := got.Start(nil, "")
		if ref != name {
			t.Errorf("asked for %s got %s", name, ref)
		}
	}
}

func TestRefusesAProviderWeDoNotHave(t *testing.T) {
	checkouts := NewCheckouts(stubCheckout{"paypal"}, stubCheckout{"stripe"})
	for _, name := range []string{"", "PayPal", "bitcoin", "stripe "} {
		_, err := checkouts.Get(name)
		if !errors.Is(err, ErrUnknownProvider) {
			t.Errorf("%q should not be a provider", name)
		}
	}
}
