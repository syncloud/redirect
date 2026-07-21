package relay

import (
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

type fakeSource struct {
	values []map[string]int64
	call   int
}

func (f *fakeSource) Fetch() (map[string]int64, error) {
	if f.call >= len(f.values) {
		return f.values[len(f.values)-1], nil
	}
	v := f.values[f.call]
	f.call++
	return v, nil
}

type fakeRelayDb struct {
	stored map[string]int64
}

func (f *fakeRelayDb) AddRelayTraffic(name string, yearMonth string, bytes int64) error {
	if f.stored == nil {
		f.stored = map[string]int64{}
	}
	f.stored[name] += bytes
	return nil
}

func (f *fakeRelayDb) GetRelayTrafficMonth(yearMonth string) (map[string]int64, error) {
	return map[string]int64{}, nil
}

func newAccountant(limit int64, source TrafficSource, db RelayDb) *Accountant {
	a := NewAccountant(source, db, limit, time.Minute, zap.NewNop())
	a.month = month()
	return a
}

func TestAccountant_BaselineThenAccumulatesDelta(t *testing.T) {
	source := &fakeSource{values: []map[string]int64{
		{"alice": 1000},
		{"alice": 1500},
		{"alice": 3000},
	}}
	db := &fakeRelayDb{}
	a := newAccountant(0, source, db)

	a.poll() // baseline at 1000, nothing added
	if db.stored["alice"] != 0 {
		t.Fatalf("expected 0 after baseline, got %d", db.stored["alice"])
	}
	a.poll() // +500
	a.poll() // +1500
	if db.stored["alice"] != 2000 {
		t.Fatalf("expected 2000 accumulated, got %d", db.stored["alice"])
	}
}

func TestAccountant_CounterResetCountsCurrentValue(t *testing.T) {
	source := &fakeSource{values: []map[string]int64{
		{"alice": 5000},
		{"alice": 200}, // frps restarted, counter reset
	}}
	a := newAccountant(0, source, &fakeRelayDb{})
	a.poll() // baseline 5000
	a.poll() // reset -> delta = 200
	if a.monthly["alice"] != 200 {
		t.Fatalf("expected 200 after reset, got %d", a.monthly["alice"])
	}
}

func TestAccountant_OverLimit(t *testing.T) {
	source := &fakeSource{values: []map[string]int64{
		{"alice": 0},
		{"alice": 4096},
	}}
	a := newAccountant(4096, source, &fakeRelayDb{})
	a.poll() // baseline 0
	if a.OverLimit("alice") {
		t.Fatal("should not be over limit at baseline")
	}
	a.poll() // +4096 -> at limit
	if !a.OverLimit("alice") {
		t.Fatal("should be over limit after exceeding")
	}
}

func TestParseTraffic(t *testing.T) {
	text := `# HELP frp_server_traffic_in
frp_server_traffic_in{name="alice.syncloud.it",type="https"} 1200
frp_server_traffic_out{name="alice.syncloud.it",type="https"} 800
frp_server_traffic_in{name="bob.syncloud.it",type="https"} 50
something_else{name="alice.syncloud.it"} 999`
	totals := parseTraffic(strings.NewReader(text))
	if totals["alice.syncloud.it"] != 2000 {
		t.Fatalf("expected 2000 for alice, got %d", totals["alice.syncloud.it"])
	}
	if totals["bob.syncloud.it"] != 50 {
		t.Fatalf("expected 50 for bob, got %d", totals["bob.syncloud.it"])
	}
}
