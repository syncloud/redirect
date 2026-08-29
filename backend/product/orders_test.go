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

type memoryStore struct {
	orders map[string]*Order
	paid   []int64
}

func newStore() *memoryStore {
	return &memoryStore{orders: map[string]*Order{}}
}

func (s *memoryStore) InsertOrder(order *Order) (int64, error) {
	order.Id = int64(len(s.orders) + 1)
	s.orders[order.Reference] = order
	return order.Id, nil
}

func (s *memoryStore) GetOrderByReference(reference string) (*Order, error) {
	return s.orders[reference], nil
}

func (s *memoryStore) MarkOrderPaid(id int64) error {
	s.paid = append(s.paid, id)
	return nil
}

func orders(checkout *recordingCheckout, store *memoryStore, mail *recordingMail) *Orders {
	return NewOrders(catalog(), NewCheckouts(checkout, checkout), store, mail, zap.NewNop())
}

func addressed() *Order {
	return &Order{
		UserId: 7, Email: "a@b.c", Device: "h4", Option: "1t",
		Name: "A B", Address: "1 Road", City: "Town", Postcode: "X1", Country: "Germany",
	}
}

const paidTotal = 22900 + 8000 + 1500

func started(t *testing.T, checkout *recordingCheckout, store *memoryStore, mail *recordingMail) *Orders {
	t.Helper()
	service := orders(checkout, store, mail)
	if _, err := service.Start(addressed(), "paypal"); err != nil {
		t.Fatal(err)
	}
	return service
}

func TestStartPricesFromTheCatalogNotTheRequest(t *testing.T) {
	checkout, store, mail := &recordingCheckout{}, newStore(), &recordingMail{}
	order := addressed()
	order.Total = 1

	if _, err := orders(checkout, store, mail).Start(order, "stripe"); err != nil {
		t.Fatal(err)
	}
	if checkout.started.Total != paidTotal {
		t.Fatalf("charged %d", checkout.started.Total)
	}
	if store.orders["REF1"].Total != paidTotal {
		t.Fatalf("stored %d", store.orders["REF1"].Total)
	}
}

func TestStartRefusesAnIncompleteAddress(t *testing.T) {
	order := addressed()
	order.City = ""
	if _, err := orders(&recordingCheckout{}, newStore(), &recordingMail{}).Start(order, "stripe"); err == nil {
		t.Fatal("want an error when the address is incomplete")
	}
}

func TestCompleteTellsSupportOncePaid(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal}
	store, mail := newStore(), &recordingMail{}

	if err := started(t, checkout, store, mail).Complete(7, "REF1"); err != nil {
		t.Fatal(err)
	}
	if mail.sent == nil {
		t.Fatal("support was not told")
	}
	if mail.sent.Reference != "REF1" || mail.sent.Provider != "paypal" {
		t.Fatalf("order recorded as %+v", mail.sent)
	}
	if len(store.paid) != 1 {
		t.Fatal("the order was not marked paid")
	}
}

func TestCompleteReadsTheOrderFromTheDatabaseNotTheCaller(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal}
	store, mail := newStore(), &recordingMail{}
	service := started(t, checkout, store, mail)

	if err := service.Complete(7, "REF1"); err != nil {
		t.Fatal(err)
	}
	if mail.sent.Device != "h4" || mail.sent.Option != "1t" || mail.sent.Total != paidTotal {
		t.Fatalf("shipped %+v", mail.sent)
	}
}

func TestCompleteRefusesAnotherAccountsOrder(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal}
	store, mail := newStore(), &recordingMail{}

	err := started(t, checkout, store, mail).Complete(8, "REF1")
	if !errors.Is(err, ErrNoOrder) {
		t.Fatalf("want ErrNoOrder got %v", err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told about someone else's order")
	}
}

func TestCompleteRefusesAReferenceWeNeverIssued(t *testing.T) {
	err := orders(&recordingCheckout{paid: true}, newStore(), &recordingMail{}).Complete(7, "MADEUP")
	if !errors.Is(err, ErrNoOrder) {
		t.Fatalf("want ErrNoOrder got %v", err)
	}
}

func TestCompleteRefusesWhenNothingWasPaid(t *testing.T) {
	checkout := &recordingCheckout{paid: false}
	store, mail := newStore(), &recordingMail{}

	err := started(t, checkout, store, mail).Complete(7, "REF1")
	if !errors.Is(err, ErrNotPaid) {
		t.Fatalf("want ErrNotPaid got %v", err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told about an unpaid order")
	}
}

func TestCompleteRefusesWhenTheAmountIsShort(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: 100}
	store, mail := newStore(), &recordingMail{}

	err := started(t, checkout, store, mail).Complete(7, "REF1")
	if !errors.Is(err, ErrWrongAmount) {
		t.Fatalf("want ErrWrongAmount got %v", err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told about an underpaid order")
	}
}

func TestCompleteIsIdempotent(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal}
	store, mail := newStore(), &recordingMail{}
	service := started(t, checkout, store, mail)

	if err := service.Complete(7, "REF1"); err != nil {
		t.Fatal(err)
	}
	store.orders["REF1"].Paid = true
	mail.sent = nil

	if err := service.Complete(7, "REF1"); err != nil {
		t.Fatal(err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told twice about one order")
	}
}
