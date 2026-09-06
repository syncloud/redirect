package product

import "testing"

func TestMoneyKeepsTwoDecimals(t *testing.T) {
	for _, c := range []struct {
		pence int
		want  string
	}{
		{32400, "324.00"},
		{22900, "229.00"},
		{5, "0.05"},
		{1500, "15.00"},
		{22999, "229.99"},
	} {
		if got := Money(c.pence); got != c.want {
			t.Errorf("%d: want %s got %s", c.pence, c.want, got)
		}
	}
}
