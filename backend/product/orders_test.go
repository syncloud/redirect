package product

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

type recordingCheckout struct {
	started     *Order
	description string
	paid        bool
	amount      int
}

func (r *recordingCheckout) Start(order *Order, description string) (string, error) {
	r.started = order
	r.description = description
	return "REF1", nil
}

func (r *recordingCheckout) Paid(string) (bool, int, error) {
	return r.paid, r.amount, nil
}

type recordingMail struct{ sent *Order }

func (m *recordingMail) SendDeviceOrder(order *Order, device, option string) error {
	m.sent = order
	return nil
}

func orders(checkout *recordingCheckout, mail *recordingMail) *Orders {
	return NewOrders(catalog(), NewCheckouts(checkout, checkout), mail, zap.NewNop())
}

func addressed() *Order {
	return &Order{
		Email: "a@b.c", Device: "h4", Option: "1t",
		Name: "A B", Address: "1 Road", City: "Town", Postcode: "X1", Country: "Germany",
	}
}

func TestStartPricesFromTheCatalogNotTheRequest(t *testing.T) {
	checkout, mail := &recordingCheckout{}, &recordingMail{}
	order := addressed()
	order.Total = 1

	if _, err := orders(checkout, mail).Start(order, "stripe"); err != nil {
		t.Fatal(err)
	}
	if checkout.started.Total != 22900+8000+1500 {
		t.Fatalf("charged %d", checkout.started.Total)
	}
}

func TestStartRefusesAnIncompleteAddress(t *testing.T) {
	order := addressed()
	order.City = ""
	if _, err := orders(&recordingCheckout{}, &recordingMail{}).Start(order, "stripe"); err == nil {
		t.Fatal("want an error when the address is incomplete")
	}
}

func TestCompleteTellsSupportOncePaid(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: 22900 + 8000 + 1500}
	mail := &recordingMail{}

	if err := orders(checkout, mail).Complete(addressed(), "paypal", "REF1"); err != nil {
		t.Fatal(err)
	}
	if mail.sent == nil {
		t.Fatal("support was not told")
	}
	if mail.sent.Reference != "REF1" || mail.sent.PaidWith != "paypal" {
		t.Fatalf("order recorded as %+v", mail.sent)
	}
}

func TestCompleteRefusesWhenNothingWasPaid(t *testing.T) {
	checkout := &recordingCheckout{paid: false}
	mail := &recordingMail{}

	err := orders(checkout, mail).Complete(addressed(), "stripe", "REF1")
	if !errors.Is(err, ErrNotPaid) {
		t.Fatalf("want ErrNotPaid got %v", err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told about an unpaid order")
	}
}

func TestCompleteRefusesWhenTheAmountIsShort(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: 100}
	mail := &recordingMail{}

	err := orders(checkout, mail).Complete(addressed(), "stripe", "REF1")
	if !errors.Is(err, ErrWrongAmount) {
		t.Fatalf("want ErrWrongAmount got %v", err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told about an underpaid order")
	}
}
