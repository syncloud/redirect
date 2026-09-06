package product

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

type settlerStub struct {
	unpaid    []*Order
	before    time.Time
	settled   []string
	abandoned []int64
	stuck     []string
	failWith  error
}

func (s *settlerStub) Abandon(order *Order) error {
	s.abandoned = append(s.abandoned, order.Id)
	return nil
}

func (s *settlerStub) Stuck(order *Order, reason string) error {
	s.stuck = append(s.stuck, reason)
	return nil
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
	r := NewReconciler(stub, time.Minute, 2*time.Minute, 24*time.Hour, zap.NewNop())
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

func TestGivesUpOnAnOrderNobodyEverPaid(t *testing.T) {
	now := time.Unix(1000000, 0)
	stub := &settlerStub{
		unpaid:   []*Order{{Id: 7, CreatedAt: now.Add(-48 * time.Hour)}},
		failWith: ErrNotPaid,
	}
	reconciler(stub).Run()

	if len(stub.abandoned) != 1 || stub.abandoned[0] != 7 {
		t.Fatalf("abandoned %v", stub.abandoned)
	}
}

func TestKeepsWaitingOnAnOrderThatIsStillFresh(t *testing.T) {
	now := time.Unix(1000000, 0)
	stub := &settlerStub{
		unpaid:   []*Order{{Id: 7, CreatedAt: now.Add(-1 * time.Hour)}},
		failWith: ErrNotPaid,
	}
	reconciler(stub).Run()

	if len(stub.abandoned) != 0 {
		t.Fatalf("gave up too early on %v", stub.abandoned)
	}
}

func TestTellsSupportWhenAnOrderCannotBeSettled(t *testing.T) {
	stub := &settlerStub{
		unpaid:   []*Order{{Id: 7}},
		failWith: errors.New("provider said yes but the database said no"),
	}
	reconciler(stub).Run()

	if len(stub.stuck) != 1 {
		t.Fatalf("support was told %d times", len(stub.stuck))
	}
	if len(stub.abandoned) != 0 {
		t.Fatal("a stuck order must not be quietly abandoned")
	}
}
