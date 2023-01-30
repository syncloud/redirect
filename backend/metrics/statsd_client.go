package metrics

import "github.com/smira/go-statsd"

type StatsdClient interface {
	Incr(stat string, count int64, tags ...statsd.Tag)
	Gauge(stat string, value int64, tags ...statsd.Tag)
}
