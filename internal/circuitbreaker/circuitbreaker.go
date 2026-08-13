package circuitbreaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type State int32

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	state          State
	failures       uint64
	maxFailures    uint64
	resetTimeout   time.Duration
	lastFailure    atomic.Value 
	mu             sync.Mutex
}

func New(maxFailures uint64, resetTimeout time.Duration) *CircuitBreaker {
	cb := &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
	cb.lastFailure.Store(time.Time{})
	return cb
}

func (cb *CircuitBreaker) Execute(req func() error) error {
	state := atomic.LoadInt32((*int32)(&cb.state))

	if state == int32(StateOpen) {
		lastFailure := cb.lastFailure.Load().(time.Time)
		if time.Since(lastFailure) > cb.resetTimeout {
			if atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(StateOpen), int32(StateHalfOpen)) {
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
		atomic.StoreInt32((*int32)(&cb.state), int32(StateOpen))
		return
	}

	failures := atomic.AddUint64(&cb.failures, 1)
	if failures >= cb.maxFailures {
		atomic.CompareAndSwapInt32((*int32)(&cb.state), int32(StateClosed), int32(StateOpen))
	}
}

func (cb *CircuitBreaker) recordSuccess() {
	atomic.StoreUint64(&cb.failures, 0)
	state := atomic.LoadInt32((*int32)(&cb.state))
	if state == int32(StateHalfOpen) {
		atomic.StoreInt32((*int32)(&cb.state), int32(StateClosed))
	}
}
