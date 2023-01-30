package metrics

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/graphite"
	"time"
)

func Start(prefix, hostname string, port int) *graphite.Graphite {
	g := graphite.New(prefix, log.NewNopLogger())
	report := time.NewTicker(5 * time.Second)
	//defer report.Stop()
	go g.SendLoop(context.Background(), report.C, "tcp", fmt.Sprintf("%s:%d", hostname, port))
	return g
}
