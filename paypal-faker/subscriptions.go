package main

import (
	"strings"
	"sync"
)

type Subscriptions struct {
	mutex     sync.Mutex
	overrides map[string]string
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{overrides: map[string]string{}}
}

func (s *Subscriptions) Revise(id string, planId string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.overrides[id] = planId
}

func (s *Subscriptions) PlanId(id string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if planId, ok := s.overrides[id]; ok {
		return planId
	}
	if strings.HasPrefix(id, "PAYPALSUB~") {
		parts := strings.Split(strings.TrimPrefix(id, "PAYPALSUB~"), "~")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}
	return "P-FAKERPLAN"
}
