package retry

import (
	"context"
	"math/rand"
	"time"
)

// Policy defines how retries should be executed
type Policy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// Execute runs the operation with exponential backoff and jitter
func (p *Policy) Execute(ctx context.Context, operation func() error) error {
	var err error
	
	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}
		
		// Determine if we should retry based on the error
		// (In a real system, you'd only retry 5xx, timeouts, network errors)
		
		if attempt == p.MaxRetries {
			break
		}

		delay := p.calculateDelay(attempt)
		
		select {
		case <-time.After(delay):
			// Wait and retry
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return err
}

func (p *Policy) calculateDelay(attempt int) time.Duration {
	// Exponential backoff
	backoff := float64(p.BaseDelay) * float64(int(1)<<attempt)
	
	// Add Jitter (up to 20%)
	jitter := (rand.Float64() * 0.2) * backoff
	delay := time.Duration(backoff + jitter)
	
	if delay > p.MaxDelay {
		return p.MaxDelay
	}
	
	return delay
}
