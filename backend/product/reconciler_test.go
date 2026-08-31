package product

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

type settlerStub struct {
	unpaid   []*Order
	before   time.Time
	settled  []string
	failWith error
}

func (s *settlerStub) Unpaid(before time.Time) ([]*Order, error) {
	s.before = before
	return s.unpaid, nil
}

func (s *settlerStub) Settle(order *Order) error {
	if s.failWith != nil {
		return s.failWith
	}
	s.settled = append(s.settled, order.Reference)
	return nil
}

func reconciler(stub *settlerStub) *Reconciler {
	r := NewReconciler(stub, time.Minute, 2*time.Minute, zap.NewNop())
	r.now = func() time.Time { return time.Unix(1000000, 0) }
	return r
}

func TestSettlesOrdersThatWerePaidAfterAll(t *testing.T) {
	stub := &settlerStub{unpaid: []*Order{
		{Reference: "one", Provider: "stripe"},
		{Reference: "two", Provider: "paypal"},
	}}

	reconciler(stub).Run()

	if len(stub.settled) != 2 {
		t.Fatalf("settled %v", stub.settled)
	}
}

func TestLeavesRecentOrdersAlone(t *testing.T) {
	stub := &settlerStub{}
	r := reconciler(stub)

	r.Run()

	want := time.Unix(1000000, 0).Add(-2 * time.Minute)
	if !stub.before.Equal(want) {
		t.Fatalf("asked for orders before %v, want %v", stub.before, want)
	}
}

func TestKeepsGoingWhenAnOrderIsStillUnpaid(t *testing.T) {
	stub := &settlerStub{
		unpaid:   []*Order{{Reference: "one"}, {Reference: "two"}},
		failWith: ErrNotPaid,
	}

	reconciler(stub).Run()

	if len(stub.settled) != 0 {
		t.Fatalf("nothing should settle, got %v", stub.settled)
	}
}

func TestKeepsGoingWhenAProviderFails(t *testing.T) {
	stub := &settlerStub{
		unpaid:   []*Order{{Reference: "one"}},
		failWith: errors.New("provider is down"),
	}

	reconciler(stub).Run()
}
