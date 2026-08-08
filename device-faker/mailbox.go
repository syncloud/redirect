package main

import "sync"

const (
	Accept = "accept"
	Reject = "reject"
	Drop   = "drop"
)

type Message struct {
	Recipients []string `json:"recipients"`
	Body       string   `json:"body"`
}

type Behaviour struct {
	Rcpt string `json:"rcpt"`
	Data string `json:"data"`
}

type Mailbox struct {
	mutex     sync.Mutex
	messages  []Message
	behaviour Behaviour
}

func NewMailbox() *Mailbox {
	return &Mailbox{behaviour: Behaviour{Rcpt: Accept, Data: Accept}}
}

func (m *Mailbox) Add(message Message) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = append(m.messages, message)
}

func (m *Mailbox) Messages() []Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	messages := make([]Message, len(m.messages))
	copy(messages, m.messages)
	return messages
}

func (m *Mailbox) Behaviour() Behaviour {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.behaviour
}

func (m *Mailbox) SetBehaviour(behaviour Behaviour) {
	if behaviour.Rcpt == "" {
		behaviour.Rcpt = Accept
	}
	if behaviour.Data == "" {
		behaviour.Data = Accept
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.behaviour = behaviour
}

func (m *Mailbox) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.messages = nil
	m.behaviour = Behaviour{Rcpt: Accept, Data: Accept}
}
