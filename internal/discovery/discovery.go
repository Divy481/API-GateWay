package discovery

import (
	"net/url"
	"sync"
)

// Instance represents a single backend service instance
type Instance struct {
	URL     *url.URL
	Healthy bool
	Active  int64 // active connections (for LeastConn)
}

// Service represents a cluster of instances for a specific route
type Service struct {
	ID        string
	mu        sync.RWMutex
	Instances []*Instance
}

// NewService creates a new Service registry
func NewService(id string, urls []string) *Service {
	instances := make([]*Instance, 0, len(urls))
	for _, rawURL := range urls {
		parsed, _ := url.Parse(rawURL)
		if parsed != nil {
			instances = append(instances, &Instance{
				URL:     parsed,
				Healthy: true, // Assume healthy initially
			})
		}
	}
	return &Service{
		ID:        id,
		Instances: instances,
	}
}

// GetHealthyInstances returns only the instances currently marked healthy
func (s *Service) GetHealthyInstances() []*Instance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	healthy := make([]*Instance, 0, len(s.Instances))
	for _, inst := range s.Instances {
		if inst.Healthy {
			healthy = append(healthy, inst)
		}
	}
	return healthy
}

// UpdateHealth updates the health status of an instance by URL
func (s *Service) UpdateHealth(targetURL *url.URL, isHealthy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, inst := range s.Instances {
		if inst.URL.String() == targetURL.String() {
			inst.Healthy = isHealthy
			break
		}
	}
}

// Registry holds all services
type Registry struct {
	Services map[string]*Service
}

// NewRegistry initializes a registry
func NewRegistry() *Registry {
	return &Registry{
		Services: make(map[string]*Service),
	}
}

// AddService adds a service to the registry
func (r *Registry) AddService(s *Service) {
	r.Services[s.ID] = s
}

// GetService retrieves a service by ID
func (r *Registry) GetService(id string) *Service {
	return r.Services[id]
}
