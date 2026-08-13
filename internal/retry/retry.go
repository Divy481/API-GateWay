package retry

import (
	"context"
	"math/rand"
	"time"
)

type Policy struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

func (p *Policy) Execute(ctx context.Context, operation func() error) error {
	var err error
	
	for attempt := 0; attempt <= p.MaxRetries; attempt++ {
		err = operation()
		if err == nil {
			return nil
		}
		
		
		if attempt == p.MaxRetries {
			break
		}

		delay := p.calculateDelay(attempt)
		
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	return err
}

func (p *Policy) calculateDelay(attempt int) time.Duration {
	backoff := float64(p.BaseDelay) * float64(int(1)<<attempt)
	
	jitter := (rand.Float64() * 0.2) * backoff
	delay := time.Duration(backoff + jitter)
	
	if delay > p.MaxDelay {
		return p.MaxDelay
	}
	
	return delay
}
