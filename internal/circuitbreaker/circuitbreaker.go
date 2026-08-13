package circuitbreaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// State represents the circuit breaker state
type State int32

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker protects a service by halting traffic when it's failing
type CircuitBreaker struct {
	state          State
	failures       uint64
	maxFailures    uint64
	resetTimeout   time.Duration
	lastFailure    atomic.Value // holds time.Time
	mu             sync.Mutex
}

// New creates a new CircuitBreaker
func New(maxFailures uint64, resetTimeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
	cb.lastFailure.Store(time.Time{})
	return cb
}

// Execute runs the given function if the circuit is closed or half-open
func (cb *CircuitBreaker) Execute(req func() error) error {
	state := atomic.LoadInt32((*int32)(&cb.state))

	if state == int32(StateOpen) {
		lastFailure := cb.lastFailure.Load().(time.Time)
		if time.Since(lastFailure) > cb.resetTimeout {
			// Transition to Half-Open
			if atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(StateOpen), int32(StateHalfOpen)) {
				// We are in half-open state, proceed to try
			} else {
				return ErrCircuitOpen
			}
		} else {
			return ErrCircuitOpen
		}
	}

	err := req()
	
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

func (cb *CircuitBreaker) recordFailure() {
	cb.lastFailure.Store(time.Now())
	state := atomic.LoadInt32((*int32)(&cb.state))
	
	if state == int32(StateHalfOpen) {
		// Immediately open the circuit again
		atomic.StoreInt32((*int32)(&cb.state), int32(StateOpen))
		return
	}

	failures := atomic.AddUint64(&cb.failures, 1)
	if failures >= cb.maxFailures {
		// Transition to Open
		atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(StateClosed), int32(StateOpen))
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	atomic.StoreUint64(&cb.failures, 0)
	state := atomic.LoadInt32((*int32)(&cb.state))
	if state == int32(StateHalfOpen) {
		// Transition back to Closed
		atomic.StoreInt32((*int32)(&cb.state), int32(StateClosed))
	}
}
