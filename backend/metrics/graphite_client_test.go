package metrics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGraphiteClient_GaugeSet(t *testing.T) {
	client := New("test", "test", 0)
	client.GaugeSet("test", 1)
	client.GaugeSet("test", 1)
	assert.Len(t, client.gauges, 1)
}

func TestGraphiteClient_CounterAdd(t *testing.T) {
	client := New("test", "test", 0)
	client.CounterAdd("test", 1)
	client.CounterAdd("test", 1)
	assert.Len(t, client.counters, 1)
}
