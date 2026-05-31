package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter serializes calls to a minimum interval, implementing a simple
// requests-per-second cap for Google Ads keyword-planning calls.
type Limiter struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

// NewLimiter builds a limiter at the given requests-per-second. Values <= 0
// default to 1 rps.
func NewLimiter(rps float64) *Limiter {
	if rps <= 0 {
		rps = 1
	}
	return &Limiter{interval: time.Duration(float64(time.Second) / rps)}
}

// Wait blocks until the next call slot is available or the context is done.
func (l *Limiter) Wait(ctx context.Context) error {
	l.mu.Lock()
	now := time.Now()
	next := l.last.Add(l.interval)
	if now.Before(next) {
		wait := next.Sub(now)
		l.last = next
		l.mu.Unlock()
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}
	l.last = now
	l.mu.Unlock()
	return nil
}
