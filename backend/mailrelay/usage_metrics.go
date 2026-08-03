package mailrelay

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type UsageMetricsStore interface {
	GetMailRelayUsageAll(yearMonth string) ([]model.MailRelayDomainUsage, error)
}

type UsageMetrics struct {
	store  UsageMetricsStore
	now    func() time.Time
	logger *zap.Logger

	messagesDesc *prometheus.Desc
	bouncesDesc  *prometheus.Desc
}

func NewUsageMetrics(store UsageMetricsStore, logger *zap.Logger) *UsageMetrics {
	return &UsageMetrics{
		store:  store,
		now:    time.Now,
		logger: logger,
		messagesDesc: prometheus.NewDesc(
			"redirect_mail_relay_messages", "Mail relay messages this month, by domain.", []string{"domain"}, nil),
		bouncesDesc: prometheus.NewDesc(
			"redirect_mail_relay_bounces", "Mail relay bounces this month, by domain.", []string{"domain"}, nil),
	}
}

func (m *UsageMetrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.messagesDesc
	ch <- m.bouncesDesc
}

func (m *UsageMetrics) Collect(ch chan<- prometheus.Metric) {
	usage, err := m.store.GetMailRelayUsageAll(m.now().UTC().Format("2006-01"))
	if err != nil {
		m.logger.Warn("cannot read mail relay usage", zap.Error(err))
		return
	}
	for _, entry := range usage {
		ch <- prometheus.MustNewConstMetric(m.messagesDesc, prometheus.GaugeValue, float64(entry.Messages), entry.Name)
		ch <- prometheus.MustNewConstMetric(m.bouncesDesc, prometheus.GaugeValue, float64(entry.Bounces), entry.Name)
	}
}
