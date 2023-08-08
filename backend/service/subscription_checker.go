package service

import (
	"github.com/syncloud/redirect/metrics"
	"log"
	"time"
)

type SubscriptionChecker struct {
	graphiteClient *metrics.GraphiteClient
}

func New(graphiteClient *metrics.GraphiteClient) *SubscriptionChecker {
	return &SubscriptionChecker{
		graphiteClient: graphiteClient,
	}
}

func (s *SubscriptionChecker) Start() {
	devicesGauge := s.graphiteClient.Graphite.NewGauge("db.devices")
	go func() {
		for {
			count, err := m.database.GetOnlineDevicesCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				devicesGauge.Set(float64(count))
			}
			count, err = m.database.GetUsersCount()
			if err != nil {
				log.Printf("db error %v", err)
			} else {
				usersGauge.Set(float64(count))
			}
			time.Sleep(10 * time.Second)
		}
	}()
}
