package main

import "github.com/smira/go-statsd"

type TestStatsdClient struct {
}

func (s TestStatsdClient) Incr(stat string, count int64, tags ...statsd.Tag) {

}
