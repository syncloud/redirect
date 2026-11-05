package event

import (
	"strings"
	"testing"
)

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

func TestSiteCarriesEveryShopAndPlanStep(t *testing.T) {
	events := NewEvents(Site())
	for _, name := range append(Shop(), Plan()...) {
		if !events.Known(name) {
			t.Errorf("%s should be known", name)
		}
	}
}

func TestSiteLosesNothingWhenPlanIsAppended(t *testing.T) {
	if len(Site()) != len(Shop())+len(Plan()) {
		t.Errorf("site has %d names, shop %d and plan %d", len(Site()), len(Shop()), len(Plan()))
	}
}

func TestPlanNamesAreUniqueAndPrefixed(t *testing.T) {
	seen := map[string]bool{}
	for _, name := range Plan() {
		if seen[name] {
			t.Errorf("%s is listed twice", name)
		}
		if !strings.HasPrefix(name, "plan.") {
			t.Errorf("%s is not a plan step", name)
		}
		seen[name] = true
	}
}

func TestShopAndPlanNeverCollide(t *testing.T) {
	shop := map[string]bool{}
	for _, name := range Shop() {
		shop[name] = true
	}
	for _, name := range Plan() {
		if shop[name] {
			t.Errorf("%s is in both lists", name)
		}
	}
}
