package rest

import "github.com/smira/go-statsd"

type StatsdClient interface {
	Incr(stat string, count int64, tags ...statsd.Tag)
}
