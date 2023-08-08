package service

import (
	"github.com/syncloud/redirect/metrics"
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

}
