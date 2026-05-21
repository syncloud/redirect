package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const rogueWindow = 24 * time.Hour

type rogueKey struct {
	platformVersion string
	token           string
}

type RogueDevices struct {
	mu   sync.Mutex
	seen map[rogueKey]time.Time
	desc *prometheus.Desc
	now  func() time.Time
}

func NewRogueDevices() *RogueDevices {
	return &RogueDevices{
		seen: make(map[rogueKey]time.Time),
		desc: prometheus.NewDesc(
			"redirect_rogue_devices",
			"Unique rogue devices (distinct update tokens) seen in the last 24h, by platform_version.",
			[]string{"platform_version"}, nil),
		now: time.Now,
	}
}

func (r *RogueDevices) Mark(platformVersion, token string) {
	r.mu.Lock()
	r.seen[rogueKey{platformVersion, token}] = r.now()
	r.mu.Unlock()
}

func (r *RogueDevices) Describe(ch chan<- *prometheus.Desc) {
	ch <- r.desc
}

func (r *RogueDevices) Collect(ch chan<- prometheus.Metric) {
	cutoff := r.now().Add(-rogueWindow)
	counts := map[string]int64{}
	r.mu.Lock()
	for k, t := range r.seen {
		if t.Before(cutoff) {
			delete(r.seen, k)
			continue
		}
		counts[k.platformVersion]++
	}
	r.mu.Unlock()
	for pv, n := range counts {
		ch <- prometheus.MustNewConstMetric(r.desc, prometheus.GaugeValue, float64(n), pv)
	}
}
