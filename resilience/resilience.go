package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// Backoff calculates how long to wait before the next retry.
type Backoff struct {
	Base       time.Duration
	Max        time.Duration
	Multiplier float64
	Jitter     bool
}

var DefaultBackoff = Backoff{
	Base:       100 * time.Millisecond,
	Max:        30 * time.Second,
	Multiplier: 2.0,
	Jitter:     true,
}

func (b Backoff) Duration(n int) time.Duration {
	d := float64(b.Base) * math.Pow(b.Multiplier, float64(n))

	if d > float64(b.Max) {
		d = float64(b.Max)
	}

	if b.Jitter {
		d = d/2 + rand.Float64()*d/2
	}

	return time.Duration(d)
}

// RetryConfig controls how retries behave.
type RetryConfig struct {
	MaxAttempts int
	Backoff     Backoff
}

var DefaultRetry = RetryConfig{
	MaxAttempts: 3,
	Backoff:     DefaultBackoff,
}

// Retry calls fn repeatedly until it succeeds or runs out of attempts.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		wait := cfg.Backoff.Duration(attempt)
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("all %d attempts failed: %w", cfg.MaxAttempts, lastErr)
}

type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker stops calling a failing service to give it time to recover.
type CircuitBreaker struct {
	maxFailures      int
	timeout          time.Duration
	state            circuitState
	failures         int
	lastFailure      time.Time
	halfOpenInFlight bool // true when a trial request is already running
	mu               sync.Mutex
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
		state:       stateClosed,
	}
}

var ErrCircuitOpen = errors.New("circuit breaker is open")

// Execute runs fn if the circuit allows it.
// In half-open state only one trial request is allowed at a time.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	switch cb.state {
	case stateOpen:
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.state = stateHalfOpen
			cb.halfOpenInFlight = true
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}

	case stateHalfOpen:
		if cb.halfOpenInFlight {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		cb.halfOpenInFlight = true
	}

	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures || cb.state == stateHalfOpen {
			cb.state = stateOpen
		} else {
			cb.state = stateClosed
		}

		cb.halfOpenInFlight = false
		return err
	}

	// Any success resets the consecutive failure count.
	cb.failures = 0
	cb.state = stateClosed
	cb.halfOpenInFlight = false
	return nil
}
