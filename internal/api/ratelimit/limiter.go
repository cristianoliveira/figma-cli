package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/logging"
	"golang.org/x/time/rate"
)

// Limiter defines an interface for rate limiting.
type Limiter interface {
	// Wait blocks until the request can proceed without exceeding rate limits.
	Wait(ctx context.Context) error
	// Update updates the limiter's state based on response headers.
	Update(headers http.Header)
}

// TokenBucketLimiter implements a token bucket rate limiter with per-minute and per-month limits.
type TokenBucketLimiter struct {
	mu sync.Mutex

	// per-minute token bucket
	minuteLimiter *rate.Limiter
	// monthly limit tracking
	monthlyLimit int
	monthlyCount int
	monthStart   time.Time // start of current month

	// current limits (can be overridden by headers)
	perMinute int
	perMonth  int

	// tier and seat type
	tier     int
	seatType string

	logger logging.Logger

	// nowFunc provides the current time; can be overridden in tests
	nowFunc func() time.Time
}

// NewTokenBucketLimiter creates a new TokenBucketLimiter for the given tier and seat type.
// If logger is nil, the default logger will be used.
func NewTokenBucketLimiter(tier int, seatType string, logger logging.Logger) *TokenBucketLimiter {
	if logger == nil {
		logger = logging.Default()
	}
	limits := GetLimits(tier, seatType)
	// Create per-minute limiter with burst = per minute limit
	// refill rate = limit per second
	minuteLimiter := rate.NewLimiter(rate.Limit(limits.PerMinute)/60.0, limits.PerMinute)
	nowFunc := time.Now
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

// Wait implements Limiter.
func (l *TokenBucketLimiter) Wait(ctx context.Context) error {
	l.mu.Lock()
	// Check monthly limit if applicable
	if l.monthlyLimit > 0 {
		l.resetMonthIfNeeded()
		if l.monthlyCount >= l.monthlyLimit {
			l.mu.Unlock()
			return fmt.Errorf("monthly rate limit exceeded (%d/%d)", l.monthlyCount, l.monthlyLimit)
		}
	}
	l.mu.Unlock()

	// Wait for per-minute token bucket
	if err := l.minuteLimiter.Wait(ctx); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	// Warn when per-minute token bucket is low (below 20% capacity)
	availableTokens := l.minuteLimiter.Tokens()
	if availableTokens < float64(l.perMinute)*0.2 {
		l.logger.Warn(ctx, "Approaching per-minute rate limit",
			logging.NewField("available_tokens", availableTokens),
			logging.NewField("limit", l.perMinute),
			logging.NewField("percentage", 20),
		)
	}
	// Increment monthly counter after acquiring token
	if l.monthlyLimit > 0 {
		l.resetMonthIfNeeded()
		l.monthlyCount++
		// Warn when approaching 80% of monthly limit
		if l.monthlyCount >= int(float64(l.monthlyLimit)*0.8) {
			l.logger.Warn(ctx, "Approaching monthly rate limit",
				logging.NewField("current", l.monthlyCount),
				logging.NewField("limit", l.monthlyLimit),
				logging.NewField("percentage", 80),
			)
		}
	}
	return nil
}

// Update implements Limiter.
func (l *TokenBucketLimiter) Update(headers http.Header) {
	parser := &DefaultHeaderParser{}
	limit, remaining, reset, retryAfter := parser.Parse(headers)
	if limit > 0 {
		l.mu.Lock()
		// Update per-minute limit based on header
		if limit != l.perMinute {
			l.perMinute = limit
			l.minuteLimiter.SetLimit(rate.Limit(limit) / 60.0)
			l.minuteLimiter.SetBurst(limit)
			l.logger.Info(context.Background(), "Updated per-minute rate limit based on header",
				logging.NewField("new_limit", limit),
			)
		}
		l.mu.Unlock()
	}
	if retryAfter > 0 {
		l.logger.Info(context.Background(), "Server suggests retry after",
			logging.NewField("retry_after", retryAfter),
		)
		// We could adjust the limiter's wait time, but for simplicity we just log.
		// The retry logic in transport will respect Retry-After via error.
	}
	// remaining and reset can be used for logging or more sophisticated tracking
	_ = remaining
	_ = reset
}

// resetMonthIfNeeded resets monthly count if we've moved into a new month.
// Must be called with lock held.
func (l *TokenBucketLimiter) resetMonthIfNeeded() {
	now := time.Now()
	if l.nowFunc != nil {
		now = l.nowFunc()
	}
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	if currentMonth.After(l.monthStart) {
		l.logger.Debug(context.Background(), "New month detected, resetting monthly count",
			logging.NewField("previous_count", l.monthlyCount),
		)
		l.monthlyCount = 0
		l.monthStart = currentMonth
	}
}

// HeaderParser extracts rate limit information from HTTP headers.
type HeaderParser interface {
	Parse(headers http.Header) (limit, remaining int, reset time.Time, retryAfter time.Duration)
}

// DefaultHeaderParser implements HeaderParser for Figma API headers.
type DefaultHeaderParser struct{}

func (p *DefaultHeaderParser) Parse(headers http.Header) (limit, remaining int, reset time.Time, retryAfter time.Duration) {
	// Parse X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset
	limit = parseIntHeader(headers, "X-RateLimit-Limit", 0)
	remaining = parseIntHeader(headers, "X-RateLimit-Remaining", 0)
	reset = parseTimeHeader(headers, "X-RateLimit-Reset", time.Time{})
	// Parse Retry-After
	retryAfter = parseRetryAfter(headers)
	return
}

// Helper functions copied from internal/api/error_utils.go to avoid circular dependency.
func parseIntHeader(headers http.Header, key string, defaultValue int) int {
	value := headers.Get(key)
	if value == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

func parseTimeHeader(headers http.Header, key string, defaultValue time.Time) time.Time {
	value := headers.Get(key)
	if value == "" {
		return defaultValue
	}
	// Try parsing as Unix timestamp (seconds)
	if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Unix(unix, 0)
	}
	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	// Try parsing as HTTP-date
	if t, err := http.ParseTime(value); err == nil {
		return t
	}
	return defaultValue
}

func parseRetryAfter(headers http.Header) time.Duration {
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}
	// Try parsing as seconds (integer)
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}
	// Try parsing as HTTP-date (RFC 7231)
	if t, err := http.ParseTime(retryAfter); err == nil {
		now := time.Now()
		if t.After(now) {
			return t.Sub(now)
		}
		return 0
	}
	return 0
}
