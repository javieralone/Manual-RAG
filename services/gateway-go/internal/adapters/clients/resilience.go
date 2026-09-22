package clients

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrCircuitOpen = errors.New("dependency circuit breaker is open")

type ResilienceConfig struct {
	Retries     int
	Backoff     time.Duration
	MaxFailures int
	ResetAfter  time.Duration
}

func configuredResilience(values []ResilienceConfig) ResilienceConfig {
	if len(values) > 0 {
		return values[0]
	}
	return ResilienceConfig{MaxFailures: 5, ResetAfter: 30 * time.Second}
}

type circuitBreaker struct {
	mu          sync.Mutex
	failures    int
	openedAt    time.Time
	maxFailures int
	resetAfter  time.Duration
}

func newCircuitBreaker(maxFailures int, resetAfter time.Duration) *circuitBreaker {
	return &circuitBreaker{maxFailures: maxFailures, resetAfter: resetAfter}
}

func (b *circuitBreaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openedAt.IsZero() {
		return true
	}
	if time.Since(b.openedAt) >= b.resetAfter {
		b.openedAt = time.Time{}
		b.failures = 0
		return true
	}
	return false
}

func (b *circuitBreaker) success() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.openedAt = time.Time{}
}

func (b *circuitBreaker) failure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.maxFailures {
		b.openedAt = time.Now()
	}
}

func retryDelay(ctx context.Context, base time.Duration, attempt int) error {
	delay := base * time.Duration(1<<attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
