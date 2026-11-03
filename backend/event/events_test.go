package event

import "testing"

func TestKnowsWhatItWasGiven(t *testing.T) {
	events := NewEvents([]string{ShopView, ShopPayStripe})
	for _, name := range []string{ShopView, ShopPayStripe} {
		if !events.Known(name) {
			t.Errorf("%s should be known", name)
		}
	}
}

func TestRefusesAnythingElse(t *testing.T) {
	events := NewEvents(Shop())
	for _, name := range []string{"", "shop.View", "shop.view ", "anything", "../etc/passwd"} {
		if events.Known(name) {
			t.Errorf("%q should not be known", name)
		}
	}
}

func TestRefusesEverythingWhenNothingIsConfigured(t *testing.T) {
	if NewEvents(nil).Known(ShopView) {
		t.Error("an empty configuration must not accept events")
	}
}

func TestShopNamesAreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range Shop() {
		if seen[name] {
			t.Errorf("%s is listed twice", name)
		}
		seen[name] = true
	}
}
