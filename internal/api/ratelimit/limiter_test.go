package ratelimit

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// newTokenBucketLimiterWithNow creates a limiter with a custom now function for testing.
func newTokenBucketLimiterWithNow(tier int, seatType string, logger logging.Logger, nowFunc func() time.Time) *TokenBucketLimiter {
	limits := GetLimits(tier, seatType)
	minuteLimiter := rate.NewLimiter(rate.Limit(limits.PerMinute)/60.0, limits.PerMinute)
	now := nowFunc()
	tb := &TokenBucketLimiter{
		minuteLimiter: minuteLimiter,
		monthlyLimit:  limits.PerMonth,
		monthlyCount:  0,
		monthStart:    time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),
		perMinute:     limits.PerMinute,
		perMonth:      limits.PerMonth,
		tier:          tier,
		seatType:      seatType,
		logger:        logger,
		nowFunc:       nowFunc,
	}
	return tb
}

func TestTokenBucketLimiter_Wait(t *testing.T) {
	logger := logging.NewNopLogger()
	limiter := NewTokenBucketLimiter(1, "dev_full", logger)

	ctx := context.Background()
	// First request should succeed immediately
	start := time.Now()
	err := limiter.Wait(ctx)
	assert.NoError(t, err)
	assert.True(t, time.Since(start) < 100*time.Millisecond)

	// Subsequent requests may be throttled if we exceed rate,
	// but with burst = 120, two requests should still pass
	err = limiter.Wait(ctx)
	assert.NoError(t, err)
}

func TestTokenBucketLimiter_MonthlyLimit(t *testing.T) {
	logger := logging.NewNopLogger()
	// Use a fixed time to avoid flakiness at month boundaries
	fixedTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	nowFunc := func() time.Time { return fixedTime }
	limiter := newTokenBucketLimiterWithNow(1, "view_collab", logger, nowFunc)
	// view_collab has monthly limit of 6
	ctx := context.Background()
	for i := 0; i < 6; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}
	// The 7th request should fail because monthly limit reached.
	err := limiter.Wait(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monthly rate limit exceeded")
}

func TestTokenBucketLimiter_UpdateHeaders(t *testing.T) {
	logger := logging.NewNopLogger()
	limiter := NewTokenBucketLimiter(1, "dev_full", logger)
	headers := http.Header{}
	headers.Set("X-RateLimit-Limit", "200")
	headers.Set("Retry-After", "30")
	limiter.Update(headers)
	// No easy way to verify internal state changed; we just ensure no panic.
}

func TestTokenBucketLimiter_MonthlyReset(t *testing.T) {
	logger := logging.NewNopLogger()
	// Use a mutable clock for testing month transitions
	currentTime := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	nowFunc := func() time.Time { return currentTime }
	limiter := newTokenBucketLimiterWithNow(1, "view_collab", logger, nowFunc)
	ctx := context.Background()
	// Use up the monthly limit (6 requests)
	for i := 0; i < 6; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}
	// 7th request should fail
	err := limiter.Wait(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monthly rate limit exceeded")
	// Advance to next month
	currentTime = time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	// Now monthly limit should be reset, next request should succeed
	err = limiter.Wait(ctx)
	assert.NoError(t, err)
	// Count should now be 1 for the new month
	// We cannot directly check monthlyCount, but we can use up remaining 5 requests
	for i := 0; i < 5; i++ {
		err = limiter.Wait(ctx)
		assert.NoError(t, err)
	}
	// 6th request in new month should fail again
	err = limiter.Wait(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monthly rate limit exceeded")
}

func TestGetLimits(t *testing.T) {
	tests := []struct {
		tier      int
		seatType  string
		wantMin   int
		wantMonth int
	}{
		{1, "dev_full", 120, 0},
		{2, "dev_full", 300, 0},
		{3, "dev_full", 600, 0},
		{4, "dev_full", 1200, 0},
		{1, "view_collab", 120, 6},
		{2, "view_collab", 300, 6},
		{3, "view_collab", 600, 6},
		{4, "view_collab", 1200, 6},
	}
	for _, tt := range tests {
		limits := GetLimits(tt.tier, tt.seatType)
		assert.Equal(t, tt.wantMin, limits.PerMinute, "tier %d seat %s per minute", tt.tier, tt.seatType)
		assert.Equal(t, tt.wantMonth, limits.PerMonth, "tier %d seat %s per month", tt.tier, tt.seatType)
	}
}
