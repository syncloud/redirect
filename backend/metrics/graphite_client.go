package metrics

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/graphite"
	"time"
)

type GraphiteClient struct {
	Graphite         *graphite.Graphite
	prefix, hostname string
	port             int
}

func New(prefix, hostname string, port int) *GraphiteClient {
	return &GraphiteClient{
		Graphite: graphite.New(prefix, log.NewNopLogger()),
		prefix:   prefix,
		hostname: hostname,
		port:     port,
	}
}

func (g *GraphiteClient) Start() {
	report := time.NewTicker(5 * time.Second)
	//defer report.Stop()
	go g.Graphite.SendLoop(
		context.Background(),
		report.C,
		"tcp",
		fmt.Sprintf("%s:%d", g.hostname, g.port),
	)
}
