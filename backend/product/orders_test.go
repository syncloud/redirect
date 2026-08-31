package product

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

type recordingCheckout struct {
	started     *Order
	description string
	paid        bool
	amount      int
	currency    string
}

func (r *recordingCheckout) Start(order *Order, description string) (string, string, error) {
	r.started = order
	r.description = description
	return "PROVIDER1", "https://pay.example/PROVIDER1", nil
}

func (r *recordingCheckout) Paid(string) (bool, int, string, error) {
	currency := r.currency
	if currency == "" {
		currency = Currency
	}
	return r.paid, r.amount, currency, nil
}

type recordingMail struct{ sent *Order }

func (m *recordingMail) SendDeviceOrder(order *Order, device, option string) error {
	m.sent = order
	return nil
}

type memoryStore struct {
	orders map[string]*Order
	byId   map[int64]*Order
	paid   []int64
}

func newStore() *memoryStore {
	return &memoryStore{orders: map[string]*Order{}, byId: map[int64]*Order{}}
}

func (s *memoryStore) InsertOrder(order *Order) (int64, error) {
	stored := *order
	stored.Id = int64(len(s.orders) + 1)
	s.orders[stored.Reference] = &stored
	s.byId[stored.Id] = &stored
	return stored.Id, nil
}

func (s *memoryStore) SetOrderProviderReference(id int64, providerReference string) error {
	s.byId[id].ProviderReference = providerReference
	return nil
}

func (s *memoryStore) only() *Order {
	for _, order := range s.orders {
		return order
	}
	return nil
}

func (s *memoryStore) GetOrderByReference(reference string) (*Order, error) {
	return s.orders[reference], nil
}

func (s *memoryStore) RedactOrders(userId int64) error {
	for _, order := range s.orders {
		if order.UserId == userId {
			order.Name, order.Address, order.City, order.Postcode = "", "", "", ""
		}
	}
	return nil
}

func (s *memoryStore) GetUnpaidOrders(before time.Time) ([]*Order, error) {
	unpaid := []*Order{}
	for _, order := range s.orders {
		if !order.Paid && order.ProviderReference != "" {
			unpaid = append(unpaid, order)
		}
	}
	return unpaid, nil
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

func started(t *testing.T, checkout *recordingCheckout, store *memoryStore, mail *recordingMail) (*Orders, string) {
	t.Helper()
	service := orders(checkout, store, mail)
	reference, err := service.Start(addressed(), "paypal")
	if err != nil {
		t.Fatal(err)
	}
	return service, reference
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
	if store.only().Total != paidTotal {
		t.Fatalf("stored %d", store.only().Total)
	}
}

func TestStartStoresTheOrderBeforeTakingMoney(t *testing.T) {
	checkout := &recordingCheckout{}
	store, mail := newStore(), &recordingMail{}

	reference, err := orders(checkout, store, mail).Start(addressed(), "stripe")
	if err != nil {
		t.Fatal(err)
	}
	if checkout.started.Reference != reference {
		t.Fatal("the payment was opened without the order reference")
	}
	if store.only().ProviderReference != "PROVIDER1" {
		t.Fatalf("provider reference not recorded: %+v", store.only())
	}
}

func TestStartHandsTheProviderOurOwnReference(t *testing.T) {
	checkout := &recordingCheckout{}
	store := newStore()

	reference, err := orders(checkout, store, &recordingMail{}).Start(addressed(), "paypal")
	if err != nil {
		t.Fatal(err)
	}
	if reference == "PROVIDER1" {
		t.Fatal("the caller must get our reference, not the provider's")
	}
	if len(reference) != 36 {
		t.Fatalf("want a uuid, got %q", reference)
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

	service, reference := started(t, checkout, store, mail)
	if err := service.Complete(7, reference); err != nil {
		t.Fatal(err)
	}
	if mail.sent == nil {
		t.Fatal("support was not told")
	}
	if mail.sent.Reference != reference || mail.sent.Provider != "paypal" {
		t.Fatalf("order recorded as %+v", mail.sent)
	}
	if len(store.paid) != 1 {
		t.Fatal("the order was not marked paid")
	}
}

func TestCompleteReadsTheOrderFromTheDatabaseNotTheCaller(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal}
	store, mail := newStore(), &recordingMail{}
	service, reference := started(t, checkout, store, mail)

	if err := service.Complete(7, reference); err != nil {
		t.Fatal(err)
	}
	if mail.sent.Device != "h4" || mail.sent.Option != "1t" || mail.sent.Total != paidTotal {
		t.Fatalf("shipped %+v", mail.sent)
	}
}

func TestCompleteRefusesAnotherAccountsOrder(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal}
	store, mail := newStore(), &recordingMail{}

	service, reference := started(t, checkout, store, mail)
	err := service.Complete(8, reference)
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

	service, reference := started(t, checkout, store, mail)
	err := service.Complete(7, reference)
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

	service, reference := started(t, checkout, store, mail)
	err := service.Complete(7, reference)
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
	service, reference := started(t, checkout, store, mail)

	if err := service.Complete(7, reference); err != nil {
		t.Fatal(err)
	}
	store.only().Paid = true
	mail.sent = nil

	if err := service.Complete(7, reference); err != nil {
		t.Fatal(err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told twice about one order")
	}
}

func TestCompleteRefusesAPaymentInAnotherCurrency(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal, currency: "EUR"}
	store, mail := newStore(), &recordingMail{}

	service, reference := started(t, checkout, store, mail)
	err := service.Complete(7, reference)
	if !errors.Is(err, ErrWrongCurrency) {
		t.Fatalf("want ErrWrongCurrency got %v", err)
	}
	if mail.sent != nil {
		t.Fatal("support must not be told about a euro payment for a sterling price")
	}
}

func TestRedactRemovesThePersonButKeepsTheSale(t *testing.T) {
	checkout := &recordingCheckout{paid: true, amount: paidTotal}
	store, mail := newStore(), &recordingMail{}
	service, reference := started(t, checkout, store, mail)

	if err := service.Redact(7); err != nil {
		t.Fatal(err)
	}

	order := store.orders[reference]
	if order.Name != "" || order.Address != "" || order.City != "" || order.Postcode != "" {
		t.Fatalf("the person is still there: %+v", order)
	}
	if order.Country != "Germany" || order.Total != paidTotal || order.Reference != reference {
		t.Fatalf("the sale was damaged: %+v", order)
	}
}

func TestRedactLeavesOtherPeopleAlone(t *testing.T) {
	checkout := &recordingCheckout{}
	store, mail := newStore(), &recordingMail{}
	service, reference := started(t, checkout, store, mail)

	if err := service.Redact(8); err != nil {
		t.Fatal(err)
	}
	if store.orders[reference].Name == "" {
		t.Fatal("somebody else's order was redacted")
	}
}
