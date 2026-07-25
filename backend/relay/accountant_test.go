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

type fakeDirectory struct {
	owners map[string]int64
	limits map[int64]int64
}

func (d *fakeDirectory) OwnerLimit(name string) (int64, int64, bool) {
	userId, ok := d.owners[name]
	if !ok {
		return 0, 0, false
	}
	return userId, d.limits[userId], true
}

func oneUser(limit int64, names ...string) *fakeDirectory {
	d := &fakeDirectory{owners: map[string]int64{}, limits: map[int64]int64{1: limit}}
	for _, name := range names {
		d.owners[name] = 1
	}
	return d
}

type noopWarner struct{}

func (noopWarner) Warn(_ int64, _ int64, _ int64) error { return nil }

type recordingWarner struct {
	warned []int64
}

func (w *recordingWarner) Warn(userId int64, _ int64, _ int64) error {
	w.warned = append(w.warned, userId)
	return nil
}

func newAccountant(directory Directory, source TrafficSource, db RelayDb) *Accountant {
	a := NewAccountant(source, db, directory, noopWarner{}, time.Minute, zap.NewNop())
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
	a := newAccountant(oneUser(0, "alice"), source, db)

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
	a := newAccountant(oneUser(0, "alice"), source, &fakeRelayDb{})
	a.poll() // baseline 5000
	a.poll() // reset -> delta = 200
	assert.Equal(t, int64(200), a.monthly["alice"])
}

func TestAccountant_OverLimit(t *testing.T) {
	source := &fakeSource{values: []map[string]int64{
		{"alice": 0},
		{"alice": 4096},
	}}
	a := newAccountant(oneUser(4096, "alice"), source, &fakeRelayDb{})
	a.poll() // baseline 0
	assert.False(t, a.OverLimit("alice"))
	a.poll() // +4096 -> at limit
	assert.True(t, a.OverLimit("alice"))
}

func TestAccountant_LimitIsPerUserNotPerDevice(t *testing.T) {
	directory := &fakeDirectory{
		owners: map[string]int64{"alice": 1, "bob": 1, "carol": 2},
		limits: map[int64]int64{1: 4096, 2: 4096},
	}
	source := &fakeSource{values: []map[string]int64{
		{"alice": 0, "bob": 0, "carol": 0},
		{"alice": 3000, "bob": 2000, "carol": 1000},
	}}
	a := newAccountant(directory, source, &fakeRelayDb{})
	a.poll()
	a.poll()

	assert.True(t, a.OverLimit("alice"))
	assert.True(t, a.OverLimit("bob"))
	assert.False(t, a.OverLimit("carol"))
}

func TestAccountant_WarnsOnceAt80Percent(t *testing.T) {
	directory := &fakeDirectory{
		owners: map[string]int64{"alice": 1},
		limits: map[int64]int64{1: 1000},
	}
	source := &fakeSource{values: []map[string]int64{
		{"alice": 0},
		{"alice": 850},
		{"alice": 950},
	}}
	a := newAccountant(directory, source, &fakeRelayDb{})
	warner := &recordingWarner{}
	a.warner = warner

	a.poll()
	assert.Empty(t, warner.warned)
	a.poll()
	a.poll()
	assert.Equal(t, []int64{1}, warner.warned)
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
