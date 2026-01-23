package ratelimit

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/stretchr/testify/assert"
)

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
	limiter := NewTokenBucketLimiter(1, "view_collab", logger)
	// view_collab has monthly limit of 6
	// We can't easily test month boundaries, but we can verify that
	// the monthly counter increments.
	// We'll just ensure Wait doesn't error for a few requests.
	ctx := context.Background()
	for i := 0; i < 6; i++ {
		err := limiter.Wait(ctx)
		assert.NoError(t, err)
	}
	// The 7th request should fail because monthly limit reached.
	// However note that monthly limit resets at month start; we assume same month.
	// This test may fail if month changes between runs, but unlikely.
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
