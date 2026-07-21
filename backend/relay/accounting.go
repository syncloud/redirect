package relay

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type TrafficSource interface {
	Fetch() (map[string]int64, error)
}

type RelayDb interface {
	AddRelayTraffic(name string, yearMonth string, bytes int64) error
	GetRelayTrafficMonth(yearMonth string) (map[string]int64, error)
}

type Accountant struct {
	source   TrafficSource
	db       RelayDb
	limit    int64
	interval time.Duration
	logger   *zap.Logger

	mu      sync.Mutex
	month   string
	lastRaw map[string]int64
	monthly map[string]int64
	over    map[string]bool

	trafficDesc *prometheus.Desc
	overDesc    *prometheus.Desc
}

func NewAccountant(source TrafficSource, db RelayDb, limit int64, interval time.Duration, logger *zap.Logger) *Accountant {
	return &Accountant{
		source:      source,
		db:          db,
		limit:       limit,
		interval:    interval,
		logger:      logger,
		lastRaw:     map[string]int64{},
		monthly:     map[string]int64{},
		over:        map[string]bool{},
		trafficDesc: prometheus.NewDesc("redirect_relay_traffic_bytes", "Relay traffic this month, by proxy.", []string{"proxy"}, nil),
		overDesc:    prometheus.NewDesc("redirect_relay_over_limit", "1 if the proxy is over its monthly traffic limit.", []string{"proxy"}, nil),
	}
}

func month() string {
	return time.Now().Format("2006-01")
}

func (a *Accountant) Start() error {
	a.mu.Lock()
	a.month = month()
	if seed, err := a.db.GetRelayTrafficMonth(a.month); err == nil {
		a.monthly = seed
		a.recomputeOver()
	} else {
		a.logger.Warn("relay accounting seed failed", zap.Error(err))
	}
	a.mu.Unlock()
	go a.loop()
	return nil
}

func (a *Accountant) loop() {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for range ticker.C {
		a.poll()
	}
}

func (a *Accountant) poll() {
	raw, err := a.source.Fetch()
	if err != nil {
		a.logger.Warn("relay traffic fetch failed", zap.Error(err))
		return
	}
	current := month()

	a.mu.Lock()
	defer a.mu.Unlock()

	if current != a.month {
		a.month = current
		a.monthly = map[string]int64{}
		a.over = map[string]bool{}
	}

	for name, cur := range raw {
		last, seen := a.lastRaw[name]
		a.lastRaw[name] = cur
		if !seen {
			continue
		}
		delta := cur - last
		if delta < 0 {
			delta = cur
		}
		if delta <= 0 {
			continue
		}
		a.monthly[name] += delta
		if err := a.db.AddRelayTraffic(name, current, delta); err != nil {
			a.logger.Warn("relay traffic persist failed", zap.String("proxy", name), zap.Error(err))
		}
	}
	a.recomputeOver()
}

func (a *Accountant) recomputeOver() {
	a.over = map[string]bool{}
	if a.limit <= 0 {
		return
	}
	for name, bytes := range a.monthly {
		if bytes >= a.limit {
			a.over[name] = true
		}
	}
}

func (a *Accountant) OverLimit(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.over[name]
}

func (a *Accountant) Describe(ch chan<- *prometheus.Desc) {
	ch <- a.trafficDesc
	ch <- a.overDesc
}

func (a *Accountant) Collect(ch chan<- prometheus.Metric) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for name, bytes := range a.monthly {
		ch <- prometheus.MustNewConstMetric(a.trafficDesc, prometheus.GaugeValue, float64(bytes), name)
		value := 0.0
		if a.over[name] {
			value = 1.0
		}
		ch <- prometheus.MustNewConstMetric(a.overDesc, prometheus.GaugeValue, value, name)
	}
}
