package product

import "testing"

func catalog() *Catalog {
	return NewCatalog([]Device{{
		Code:  "h4",
		Name:  "Syncloud H4",
		Board: "odroid-hc4",
		Price: 22900,
		Options: []Option{
			{Code: "120", Name: "120 GB", Extra: 0},
			{Code: "1t", Name: "1 TB", Extra: 8000},
			{Code: "2tx2", Name: "2 TB x 2", Extra: 43000},
		},
	}}, 1500)
}

func TestTotalAddsTheOptionAndShipping(t *testing.T) {
	for _, c := range []struct {
		option string
		want   int
	}{
		{"120", 22900 + 1500},
		{"1t", 22900 + 8000 + 1500},
		{"2tx2", 22900 + 43000 + 1500},
	} {
		got, err := catalog().Total("h4", c.option)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("%s: want %d got %d", c.option, c.want, got)
		}
	}
}

func TestTotalRefusesAnOptionTheDeviceDoesNotHave(t *testing.T) {
	if _, err := catalog().Total("h4", "8t"); err == nil {
		t.Fatal("want an error for an unknown option")
	}
}

func TestTotalRefusesAnUnknownDevice(t *testing.T) {
	if _, err := catalog().Total("n", "120"); err == nil {
		t.Fatal("want an error for an unknown device")
	}
}

func TestTotalIsNeverTakenFromTheRequest(t *testing.T) {
	got, err := catalog().Total("h4", "120")
	if err != nil {
		t.Fatal(err)
	}
	if got != 24400 {
		t.Fatalf("the price must come from the catalog, got %d", got)
	}
}
