package metrics

import (
	"github.com/syncloud/redirect/db"
	"log"
	"time"
)

type Publisher struct {
	graphiteClient *GraphiteClient
	database       *db.MySql
}

func NewPublisher(
	graphiteClient *GraphiteClient,
	database *db.MySql,
) *Publisher {
	return &Publisher{
		graphiteClient: graphiteClient,
		database:       database,
	}
}

func (p *Publisher) Start() {
	p.graphiteClient.Start()
	devicesGauge := p.graphiteClient.Graphite.NewGauge("db.devices")
	usersGauge := p.graphiteClient.Graphite.NewGauge("db.users")
	go func() {
		for {
			count, err := p.database.GetOnlineDevicesCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				devicesGauge.Set(float64(count))
			}
			count, err = p.database.GetUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				usersGauge.Set(float64(count))
			}
			time.Sleep(10 * time.Second)
		}
	}()
}
