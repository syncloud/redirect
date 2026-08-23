package rest

import (
	"sync"
	"time"
)

const rateLimiterMaxKeys = 10000

type RateLimiter struct {
	mutex  sync.Mutex
	hits   map[string][]time.Time
	limit  int
	window time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		hits:   make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

func (r *RateLimiter) Allow(key string, now time.Time) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.hits) > rateLimiterMaxKeys {
		r.evict(now)
	}

	recent := r.recent(r.hits[key], now)
	if len(recent) >= r.limit {
		r.hits[key] = recent
		return false
	}

	r.hits[key] = append(recent, now)
	return true
}

func (r *RateLimiter) recent(times []time.Time, now time.Time) []time.Time {
	cutoff := now.Add(-r.window)
	kept := make([]time.Time, 0, len(times))
	for _, t := range times {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	return kept
}

func (r *RateLimiter) evict(now time.Time) {
	for key, times := range r.hits {
		recent := r.recent(times, now)
		if len(recent) == 0 {
			delete(r.hits, key)
		} else {
			r.hits[key] = recent
		}
	}
}
