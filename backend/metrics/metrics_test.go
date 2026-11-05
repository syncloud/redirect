package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func series(t *testing.T, m *Metrics, name string) map[string]float64 {
	registry := prometheus.NewPedanticRegistry()
	assert.Nil(t, registry.Register(m))
	families, err := registry.Gather()
	assert.Nil(t, err)
	found := map[string]float64{}
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			key := ""
			for _, pair := range metric.GetLabel() {
				if key != "" {
					key += ","
				}
				key += pair.GetName() + "=" + pair.GetValue()
			}
			found[key] = metric.GetCounter().GetValue()
		}
	}
	return found
}

func TestDomainAcquire_SeparatesAFirstTimeDeviceFromAReinstall(t *testing.T) {
	m := New()

	m.DomainAcquire(DomainNew, OwnerFree)
	m.DomainAcquire(DomainExisting, OwnerSubscriber)
	m.DomainAcquire(DomainExisting, OwnerSubscriber)

	assert.Equal(t, map[string]float64{
		"domain=new,owner=free":            1,
		"domain=existing,owner=subscriber": 2,
	}, series(t, m, "redirect_domain_acquire_total"))
}

func TestDomainAcquire_IsAbsentUntilADeviceClaimsSomething(t *testing.T) {
	assert.Empty(t, series(t, New(), "redirect_domain_acquire_total"))
}
