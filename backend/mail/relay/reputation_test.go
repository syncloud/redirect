package relay

import (
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakeMetrics struct {
	rates map[string]float64
	err   error
}

func (f *fakeMetrics) Rate(metric string) (float64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.rates[metric], nil
}

func gathered(t *testing.T, reputation *Reputation) map[string]float64 {
	t.Helper()
	registry := prometheus.NewRegistry()
	registry.MustRegister(reputation)
	families, err := registry.Gather()
	assert.NoError(t, err)
	values := map[string]float64{}
	for _, family := range families {
		values[family.GetName()] = family.GetMetric()[0].GetGauge().GetValue()
	}
	return values
}

func TestReputation_PublishesRates(t *testing.T) {
	source := &fakeMetrics{rates: map[string]float64{
		bounceMetric:    0.02,
		complaintMetric: 0.0005,
	}}
	reputation := NewReputation(source, time.Hour, zap.NewNop())
	reputation.poll()

	values := gathered(t, reputation)
	assert.Equal(t, 0.02, values["redirect_ses_bounce_rate"])
	assert.Equal(t, 0.0005, values["redirect_ses_complaint_rate"])
}

func TestReputation_KeepsLastKnownOnError(t *testing.T) {
	source := &fakeMetrics{rates: map[string]float64{bounceMetric: 0.02, complaintMetric: 0.001}}
	reputation := NewReputation(source, time.Hour, zap.NewNop())
	reputation.poll()

	source.err = fmt.Errorf("cloudwatch is down")
	reputation.poll()

	values := gathered(t, reputation)
	assert.Equal(t, 0.02, values["redirect_ses_bounce_rate"])
}
