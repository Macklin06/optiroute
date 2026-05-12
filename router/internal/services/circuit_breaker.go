package services

import (
	"sync"
	"time"
)

type CircuitState int

const (
	StateClosed   CircuitState = iota
	StateOpen     CircuitState = iota
	StateHalfOpen CircuitState = iota
)

type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failureCount int
	maxFailures  int
	openedAt     time.Time
	timeout      time.Duration
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) CanRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.openedAt) > cb.timeout {
			cb.state = StateHalfOpen
			return true
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = StateClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	if cb.failureCount >= cb.maxFailures {
		cb.state = StateOpen
		cb.openedAt = time.Now()
	}
}
