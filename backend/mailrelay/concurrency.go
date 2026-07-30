package mailrelay

import (
	"fmt"
	"sync"
)

var (
	ErrTooManyConnections = fmt.Errorf("too many connections, try again shortly")
	ErrBusy               = fmt.Errorf("relay is busy, try again shortly")
)

type Connections struct {
	limit   int
	mutex   sync.Mutex
	perPeer map[string]int
}

func NewConnections(limit int) *Connections {
	return &Connections{limit: limit, perPeer: map[string]int{}}
}

func (c *Connections) Acquire(peer string) bool {
	if c.limit <= 0 {
		return true
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.perPeer[peer] >= c.limit {
		return false
	}
	c.perPeer[peer]++
	return true
}

func (c *Connections) Release(peer string) {
	if c.limit <= 0 {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.perPeer[peer] <= 1 {
		delete(c.perPeer, peer)
		return
	}
	c.perPeer[peer]--
}

type InFlight struct {
	slots chan struct{}
}

func NewInFlight(limit int) *InFlight {
	if limit <= 0 {
		return &InFlight{}
	}
	return &InFlight{slots: make(chan struct{}, limit)}
}

func (i *InFlight) Acquire() bool {
	if i.slots == nil {
		return true
	}
	select {
	case i.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (i *InFlight) Release() {
	if i.slots == nil {
		return
	}
	select {
	case <-i.slots:
	default:
	}
}
