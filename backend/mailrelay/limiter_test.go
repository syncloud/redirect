package mailrelay

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func limiterAt(now *time.Time) *Limiter {
	limiter := NewLimiter(Limits{Minute: 2, Hour: 5, Day: 10, Recipients: 3})
	limiter.now = func() time.Time { return *now }
	return limiter
}

func TestLimiter_AllowsUnderLimit(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := limiterAt(&now)
	assert.NoError(t, limiter.Allow("device.syncloud.it", 1))
	assert.NoError(t, limiter.Allow("device.syncloud.it", 1))
}

func TestLimiter_BlocksBurstWithinMinute(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := limiterAt(&now)
	assert.NoError(t, limiter.Allow("device.syncloud.it", 2))
	assert.ErrorIs(t, limiter.Allow("device.syncloud.it", 1), ErrTooManyPerMinute)
}

func TestLimiter_MinuteWindowResets(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := limiterAt(&now)
	assert.NoError(t, limiter.Allow("device.syncloud.it", 2))
	now = now.Add(time.Minute)
	assert.NoError(t, limiter.Allow("device.syncloud.it", 1))
}

func TestLimiter_HourlyLimitSurvivesMinuteResets(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := limiterAt(&now)
	for i := 0; i < 2; i++ {
		assert.NoError(t, limiter.Allow("device.syncloud.it", 2))
		now = now.Add(time.Minute)
	}
	assert.NoError(t, limiter.Allow("device.syncloud.it", 1))
	assert.ErrorIs(t, limiter.Allow("device.syncloud.it", 1), ErrTooManyPerHour)
}

func TestLimiter_IsPerDomain(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := limiterAt(&now)
	assert.NoError(t, limiter.Allow("one.syncloud.it", 2))
	assert.NoError(t, limiter.Allow("two.syncloud.it", 2))
}

func TestLimiter_Recipients(t *testing.T) {
	now := time.Unix(0, 0)
	limiter := limiterAt(&now)
	assert.NoError(t, limiter.AllowRecipients(3))
	assert.ErrorIs(t, limiter.AllowRecipients(4), ErrTooManyRecipient)
}
