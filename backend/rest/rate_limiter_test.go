package rest

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRateLimiterAllowsUpToLimit(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)
	now := time.Now()
	assert.True(t, limiter.Allow("1.2.3.4", now))
	assert.True(t, limiter.Allow("1.2.3.4", now))
	assert.True(t, limiter.Allow("1.2.3.4", now))
}

func TestRateLimiterBlocksOverLimit(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)
	now := time.Now()
	for i := 0; i < 3; i++ {
		limiter.Allow("1.2.3.4", now)
	}
	assert.False(t, limiter.Allow("1.2.3.4", now.Add(time.Second)))
}

func TestRateLimiterKeysAreIndependent(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	now := time.Now()
	assert.True(t, limiter.Allow("1.2.3.4", now))
	assert.False(t, limiter.Allow("1.2.3.4", now))
	assert.True(t, limiter.Allow("5.6.7.8", now))
}

func TestRateLimiterRecoversAfterWindow(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	now := time.Now()
	assert.True(t, limiter.Allow("1.2.3.4", now))
	assert.False(t, limiter.Allow("1.2.3.4", now.Add(30*time.Second)))
	assert.True(t, limiter.Allow("1.2.3.4", now.Add(2*time.Minute)))
}

func TestRateLimiterBlockedRequestDoesNotExtendWindow(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	now := time.Now()
	assert.True(t, limiter.Allow("1.2.3.4", now))
	for i := 1; i < 60; i++ {
		limiter.Allow("1.2.3.4", now.Add(time.Duration(i)*time.Second))
	}
	assert.True(t, limiter.Allow("1.2.3.4", now.Add(61*time.Second)))
}

func TestRateLimiterEvictsStaleKeys(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	now := time.Now()
	for i := 0; i <= rateLimiterMaxKeys; i++ {
		limiter.Allow(string(rune(i)), now)
	}
	limiter.Allow("trigger", now.Add(2*time.Minute))
	assert.LessOrEqual(t, len(limiter.hits), 1)
}

func TestRateLimiterBurstMatchesObservedBotPattern(t *testing.T) {
	limiter := NewRateLimiter(anonymousBurstLimit, anonymousBurstWindow)
	now := time.Now()
	assert.True(t, limiter.Allow("1.2.3.4", now))
	assert.True(t, limiter.Allow("1.2.3.4", now.Add(time.Second)))
	assert.True(t, limiter.Allow("1.2.3.4", now.Add(2*time.Second)))
	assert.False(t, limiter.Allow("1.2.3.4", now.Add(3*time.Second)))
}
