package metrics

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/graphite"
	"sync"
	"time"
)

type GraphiteClient struct {
	Graphite         *graphite.Graphite
	prefix, hostname string
	port             int
	counters         map[string]*graphite.Counter
	gauges           map[string]*graphite.Gauge
	mtx              sync.RWMutex
}

func New(prefix, hostname string, port int) *GraphiteClient {
	return &GraphiteClient{
		Graphite: graphite.New(fmt.Sprintf("%s.", prefix), log.NewNopLogger()),
		prefix:   prefix,
		hostname: hostname,
		port:     port,
		gauges:   make(map[string]*graphite.Gauge),
		counters: make(map[string]*graphite.Counter),
	}
}

func (g *GraphiteClient) Start() error {
	report := time.NewTicker(5 * time.Second)
	//defer report.Stop()
	go g.Graphite.SendLoop(
		context.Background(),
		report.C,
		"tcp",
		fmt.Sprintf("%s:%d", g.hostname, g.port),
	)
	return nil
}

func (g *GraphiteClient) GaugeSet(name string, value float64) {
	g.mtx.Lock()
	if _, ok := g.gauges[name]; !ok {
		gauge := g.Graphite.NewGauge(name)
		g.gauges[name] = gauge
	}
	g.mtx.Unlock()
	g.gauges[name].Set(value)
}

func (g *GraphiteClient) CounterAdd(name string, value float64) {
	g.mtx.Lock()
	if _, ok := g.counters[name]; !ok {
		counter := g.Graphite.NewCounter(name)
		g.counters[name] = counter
	}
	g.mtx.Unlock()
	g.counters[name].Add(value)
}
