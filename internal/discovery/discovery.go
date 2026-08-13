package discovery

import (
	"net/url"
	"sync"
)

type Instance struct {
	URL     *url.URL
	Healthy bool
	Active  int64 
}

type Service struct {
	ID        string
	mu        sync.RWMutex
	Instances []*Instance
}

func NewService(id string, urls []string) *Service {
	instances := make([]*Instance, 0, len(urls))
	for _, rawURL := range urls {
		parsed, _ := url.Parse(rawURL)
		if parsed != nil {
			instances = append(instances, &Instance{
				URL:     parsed,
				Healthy: true, 
			})
		}
	}
	return &Service{
		ID:        id,
		Instances: instances,
	}
}

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

type Registry struct {
	Services map[string]*Service
}

func NewRegistry() *Registry {
	return &Registry{
		Services: make(map[string]*Service),
	}
}

func (r *Registry) AddService(s *Service) {
	r.Services[s.ID] = s
}

func (r *Registry) GetService(id string) *Service {
	return r.Services[id]
}
