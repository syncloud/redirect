package mailrelay

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"go.uber.org/zap"
)

type fakeUsageMetricsStore struct {
	usage     []model.MailRelayDomainUsage
	err       error
	yearMonth string
}

func (f *fakeUsageMetricsStore) GetMailRelayUsageAll(yearMonth string) ([]model.MailRelayDomainUsage, error) {
	f.yearMonth = yearMonth
	return f.usage, f.err
}

type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Now() time.Time {
	return f.now
}

func usageMetricsFor(store UsageMetricsStore) *UsageMetrics {
	return NewUsageMetrics(store, &fakeClock{now: time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC)}, zap.NewNop())
}

func samples(t *testing.T, metrics *UsageMetrics) map[string]float64 {
	t.Helper()
	registry := prometheus.NewRegistry()
	registry.MustRegister(metrics)
	families, err := registry.Gather()
	assert.NoError(t, err)
	values := map[string]float64{}
	for _, family := range families {
		for _, metric := range family.GetMetric() {
			domain := ""
			for _, label := range metric.GetLabel() {
				if label.GetName() == "domain" {
					domain = label.GetValue()
				}
			}
			values[fmt.Sprintf("%s{%s}", family.GetName(), domain)] = metric.GetGauge().GetValue()
		}
	}
	return values
}

func TestUsageMetrics_PerDomain(t *testing.T) {
	store := &fakeUsageMetricsStore{usage: []model.MailRelayDomainUsage{
		{Name: "one.syncloud.it", Messages: 12, Bounces: 1},
		{Name: "two.syncloud.it", Messages: 3, Bounces: 0},
	}}
	values := samples(t, usageMetricsFor(store))

	assert.Equal(t, 12.0, values["redirect_mail_relay_messages{one.syncloud.it}"])
	assert.Equal(t, 1.0, values["redirect_mail_relay_bounces{one.syncloud.it}"])
	assert.Equal(t, 3.0, values["redirect_mail_relay_messages{two.syncloud.it}"])
	assert.Equal(t, "2026-08", store.yearMonth)
}

func TestUsageMetrics_StoreErrorEmitsNothing(t *testing.T) {
	store := &fakeUsageMetricsStore{err: fmt.Errorf("database is down")}
	assert.Empty(t, samples(t, usageMetricsFor(store)))
}
