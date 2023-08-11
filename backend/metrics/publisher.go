package metrics

import (
	"github.com/syncloud/redirect/db"
	"log"
	"time"
)

type Graphite interface {
	GaugeSet(name string, value float64)
}

type Publisher struct {
	graphite Graphite
	database *db.MySql
}

func NewPublisher(
	graphite Graphite,
	database *db.MySql,
) *Publisher {
	return &Publisher{
		graphite: graphite,
		database: database,
	}
}

func (p *Publisher) Start() error {
	go func() {
		for {
			count, err := p.database.GetOnlineDevicesCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				p.graphite.GaugeSet("db.devices", float64(count))
			}

			count, err = p.database.GetAllUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				p.graphite.GaugeSet("db.users.all", float64(count))
			}

			count, err = p.database.GetActiveUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				p.graphite.GaugeSet("db.users.active", float64(count))
			}

			count, err = p.database.GetSubscribedUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				p.graphite.GaugeSet("db.users.subscribed", float64(count))
			}

			count, err = p.database.Get2MonthOldActiveUsersWithoutDomainCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				p.graphite.GaugeSet("db.users.dead", float64(count))
			}

			count, err = p.database.GetDomainCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				p.graphite.GaugeSet("db.domains", float64(count))
			}

			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}
