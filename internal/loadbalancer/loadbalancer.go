package loadbalancer

import (
	"api-gateway/internal/discovery"
	"errors"
	"sync/atomic"
)

var ErrNoHealthyInstances = errors.New("no healthy instances available")

// LoadBalancer interface defines how to select an instance
type LoadBalancer interface {
	Next(service *discovery.Service) (*discovery.Instance, error)
}

// RoundRobin implements a fast, lock-free round-robin load balancer
type RoundRobin struct {
	counter uint64
}

// NewRoundRobin creates a RoundRobin LB
func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

// Next selects the next healthy instance using atomic operations
func (rr *RoundRobin) Next(service *discovery.Service) (*discovery.Instance, error) {
	healthy := service.GetHealthyInstances()
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}

	// Lock-free atomic increment
	idx := atomic.AddUint64(&rr.counter, 1) % uint64(len(healthy))
	return healthy[idx], nil
}

// LeastConnection could be implemented here as well
