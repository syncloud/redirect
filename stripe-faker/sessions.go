package main

import (
	"strconv"
	"sync"
)

type Session struct {
	Id       string
	Currency string
	Amount   int
	Paid     bool
}

type Sessions struct {
	mutex    sync.Mutex
	sessions map[string]*Session
	next     int
}

func NewSessions() *Sessions {
	return &Sessions{sessions: map[string]*Session{}}
}

func (s *Sessions) Add(currency string, amount int) *Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.next++
	session := &Session{Id: "cs_faker_" + strconv.Itoa(s.next), Currency: currency, Amount: amount}
	s.sessions[session.Id] = session
	return session
}

func (s *Sessions) Get(id string) *Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.sessions[id]
}

func (s *Sessions) Pay(id string) *Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	session := s.sessions[id]
	if session != nil {
		session.Paid = true
	}
	return session
}
