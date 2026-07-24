package relay

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, int64(0), db.stored["alice"])
	a.poll() // +500
	a.poll() // +1500
	assert.Equal(t, int64(2000), db.stored["alice"])
}

func TestAccountant_CounterResetCountsCurrentValue(t *testing.T) {
	source := &fakeSource{values: []map[string]int64{
		{"alice": 5000},
		{"alice": 200}, // frps restarted, counter reset
	}}
	a := newAccountant(0, source, &fakeRelayDb{})
	a.poll() // baseline 5000
	a.poll() // reset -> delta = 200
	assert.Equal(t, int64(200), a.monthly["alice"])
}

func TestAccountant_OverLimit(t *testing.T) {
	source := &fakeSource{values: []map[string]int64{
		{"alice": 0},
		{"alice": 4096},
	}}
	a := newAccountant(4096, source, &fakeRelayDb{})
	a.poll() // baseline 0
	assert.False(t, a.OverLimit("alice"))
	a.poll() // +4096 -> at limit
	assert.True(t, a.OverLimit("alice"))
}

func TestParseTraffic(t *testing.T) {
	text := `# HELP frp_server_traffic_in
frp_server_traffic_in{name="alice.syncloud.it",type="https"} 1200
frp_server_traffic_out{name="alice.syncloud.it",type="https"} 800
frp_server_traffic_in{name="bob.syncloud.it",type="https"} 50
something_else{name="alice.syncloud.it"} 999`
	totals := parseTraffic(strings.NewReader(text))
	assert.Equal(t, int64(2000), totals["alice.syncloud.it"])
	assert.Equal(t, int64(50), totals["bob.syncloud.it"])
}
