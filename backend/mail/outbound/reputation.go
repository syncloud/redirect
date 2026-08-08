package outbound

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

const (
	bounceMetric    = "Reputation.BounceRate"
	complaintMetric = "Reputation.ComplaintRate"
)

type MetricSource interface {
	Rate(metric string) (float64, error)
}

type Reputation struct {
	source   MetricSource
	interval time.Duration
	logger   *zap.Logger

	mutex     sync.Mutex
	bounce    float64
	complaint float64

	bounceDesc    *prometheus.Desc
	complaintDesc *prometheus.Desc
}

func NewReputation(source MetricSource, interval time.Duration, logger *zap.Logger) *Reputation {
	return &Reputation{
		source:   source,
		interval: interval,
		logger:   logger,
		bounceDesc: prometheus.NewDesc(
			"redirect_ses_bounce_rate", "SES bounce rate, sending is paused at 0.1.", nil, nil),
		complaintDesc: prometheus.NewDesc(
			"redirect_ses_complaint_rate", "SES complaint rate, account is reviewed at 0.001.", nil, nil),
	}
}

func (r *Reputation) Start() error {
	r.poll()
	go func() {
		for range time.Tick(r.interval) {
			r.poll()
		}
	}()
	return nil
}

func (r *Reputation) poll() {
	bounce, err := r.source.Rate(bounceMetric)
	if err != nil {
		r.logger.Warn("cannot read ses bounce rate", zap.Error(err))
		return
	}
	complaint, err := r.source.Rate(complaintMetric)
	if err != nil {
		r.logger.Warn("cannot read ses complaint rate", zap.Error(err))
		return
	}
	r.mutex.Lock()
	r.bounce = bounce
	r.complaint = complaint
	r.mutex.Unlock()
}

func (r *Reputation) Describe(ch chan<- *prometheus.Desc) {
	ch <- r.bounceDesc
	ch <- r.complaintDesc
}

func (r *Reputation) Collect(ch chan<- prometheus.Metric) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ch <- prometheus.MustNewConstMetric(r.bounceDesc, prometheus.GaugeValue, r.bounce)
	ch <- prometheus.MustNewConstMetric(r.complaintDesc, prometheus.GaugeValue, r.complaint)
}
