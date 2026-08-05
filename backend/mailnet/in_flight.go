package mailnet

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
