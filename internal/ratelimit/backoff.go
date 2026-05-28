package ratelimit

import (
	"context"
	"time"

	"github.com/alparslanahmed/keyword-market-mcp/internal/utils"
)

// RetryConfig controls exponential backoff retries.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// DefaultRetryConfig matches the spec: up to 3 retries with exponential backoff.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{MaxRetries: 3, BaseDelay: 500 * time.Millisecond, MaxDelay: 8 * time.Second}
}

// IsRetryable is the default predicate: retry only errors flagged retryable
// (rate limits, transient network/server errors). Invalid credentials, invalid
// customer ID, and validation errors are never retried.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	return utils.AsAppError(err).Retryable
}

// Retry runs fn with exponential backoff. It stops early on a non-retryable
// error and returns the last error after exhausting attempts.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error, retryable func(error) bool) error {
	if retryable == nil {
		retryable = IsRetryable
	}
	var err error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if err = fn(); err == nil {
			return nil
		}
		if !retryable(err) || attempt == cfg.MaxRetries {
			return err
		}
		delay := cfg.BaseDelay << attempt
		if cfg.MaxDelay > 0 && delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}
