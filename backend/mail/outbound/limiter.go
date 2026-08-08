package outbound

import (
	"fmt"
	"sync"
	"time"
)

var (
	ErrTooManyPerMinute = fmt.Errorf("sending too fast, try again shortly")
	ErrTooManyPerHour   = fmt.Errorf("hourly sending limit reached")
	ErrTooManyPerDay    = fmt.Errorf("daily sending limit reached")
	ErrTooManyRecipient = fmt.Errorf("too many recipients for one message")
)

type Limiter struct {
	limits   Limits
	mutex    sync.Mutex
	counters map[string]*counters
	now      func() time.Time
}

func NewLimiter(limits Limits) *Limiter {
	return &Limiter{limits: limits, counters: map[string]*counters{}, now: time.Now}
}

func (l *Limiter) AllowRecipients(count int) error {
	if l.limits.Recipients > 0 && count > l.limits.Recipients {
		return ErrTooManyRecipient
	}
	return nil
}

func (l *Limiter) Allow(domain string, messages int64) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := l.now()
	c, found := l.counters[domain]
	if !found {
		c = &counters{}
		l.counters[domain] = c
	}
	if err := allow(&c.minute, now, time.Minute, l.limits.Minute, messages, ErrTooManyPerMinute); err != nil {
		return err
	}
	if err := allow(&c.hour, now, time.Hour, l.limits.Hour, messages, ErrTooManyPerHour); err != nil {
		return err
	}
	return allow(&c.day, now, 24*time.Hour, l.limits.Day, messages, ErrTooManyPerDay)
}

func allow(w *window, now time.Time, size time.Duration, limit int64, messages int64, over error) error {
	if limit <= 0 {
		return nil
	}
	if now.Sub(w.start) >= size {
		w.start = now
		w.count = 0
	}
	if w.count+messages > limit {
		return over
	}
	w.count += messages
	return nil
}
