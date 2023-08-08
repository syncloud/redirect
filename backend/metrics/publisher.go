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
	domainsGauge := p.graphiteClient.Graphite.NewGauge("db.domains")
	allUsersGauge := p.graphiteClient.Graphite.NewGauge("db.users.all")
	activeUsersGauge := p.graphiteClient.Graphite.NewGauge("db.users.active")
	subscribedUsersGauge := p.graphiteClient.Graphite.NewGauge("db.users.subscribed")
	deadUsersGauge := p.graphiteClient.Graphite.NewGauge("db.users.dead")
	go func() {
		for {
			count, err := p.database.GetOnlineDevicesCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				devicesGauge.Set(float64(count))
			}

			count, err = p.database.GetAllUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				allUsersGauge.Set(float64(count))
			}

			count, err = p.database.GetActiveUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				activeUsersGauge.Set(float64(count))
			}

			count, err = p.database.GetSubscribedUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				subscribedUsersGauge.Set(float64(count))
			}

			count, err = p.database.Get2MonthOldActiveUsersWithoutDomainCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				deadUsersGauge.Set(float64(count))
			}

			count, err = p.database.GetDomainCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				domainsGauge.Set(float64(count))
			}

			time.Sleep(10 * time.Second)
		}
	}()
}
