package payment

import "testing"

func TestParseReadsPoundsAndPence(t *testing.T) {
	for _, c := range []struct {
		value string
		want  int
	}{
		{"324.00", 32400},
		{"229.99", 22999},
		{"15", 1500},
		{"0.05", 5},
		{"1234.5", 123450},
	} {
		if got := parse(c.value); got != c.want {
			t.Errorf("%s: want %d got %d", c.value, c.want, got)
		}
	}
}

func TestParseRefusesRubbishRatherThanGuessing(t *testing.T) {
	for _, value := range []string{"", "free", "GBP 12"} {
		if got := parse(value); got != 0 {
			t.Errorf("%q should not parse to %d", value, got)
		}
	}
}
