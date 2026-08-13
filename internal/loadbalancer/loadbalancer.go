package loadbalancer

import (
	"api-gateway/internal/discovery"
	"errors"
	"sync/atomic"
)

var ErrNoHealthyInstances = errors.New("no healthy instances available")

type LoadBalancer interface {
	Next(service *discovery.Service) (*discovery.Instance, error)
}

type RoundRobin struct {
	counter uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{}
}

func (rr *RoundRobin) Next(service *discovery.Service) (*discovery.Instance, error) {
	healthy := service.GetHealthyInstances()
	if len(healthy) == 0 {
		return nil, ErrNoHealthyInstances
	}

	idx := atomic.AddUint64(&rr.counter, 1) % uint64(len(healthy))
	return healthy[idx], nil
}

