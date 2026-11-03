package label

import "testing"

func TestKeepsAValueItKnows(t *testing.T) {
	allowed := NewAllowed(Languages())
	for _, value := range []string{"en", "de", "zh-CN"} {
		if allowed.Label(value) != value {
			t.Errorf("%s should be kept", value)
		}
	}
}

func TestFoldsAnythingElseIntoOther(t *testing.T) {
	allowed := NewAllowed(Languages())
	for _, value := range []string{"", "EN", "en-GB", "klingon", "de\n"} {
		if allowed.Label(value) != Other {
			t.Errorf("%q should fold into %s", value, Other)
		}
	}
}

func TestEveryLanguageTheAppOffersIsAllowed(t *testing.T) {
	allowed := NewAllowed(Languages())
	for _, value := range Languages() {
		if allowed.Label(value) == Other {
			t.Errorf("%s is offered but not allowed", value)
		}
	}
}
