package health

import (
	"api-gateway/internal/discovery"
	"log/slog"
	"net/http"
	"time"
)

type Checker struct {
	registry *discovery.Registry
	client   *http.Client
	interval time.Duration
	logger   *slog.Logger
}

func NewChecker(registry *discovery.Registry, logger *slog.Logger) *Checker {
	return &Checker{
		registry: registry,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
		interval: 10 * time.Second,
		logger:   logger,
	}
}

func (hc *Checker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		for range ticker.C {
			hc.checkAll()
		}
	}()
}

func (hc *Checker) checkAll() {
	for _, service := range hc.registry.Services {
		for _, inst := range service.Instances {
			go hc.checkInstance(service.ID, inst)
		}
	}
}

func (hc *Checker) checkInstance(serviceID string, inst *discovery.Instance) {
	healthURL := inst.URL.String() + "/health"
	
	resp, err := hc.client.Get(healthURL)
	isHealthy := err == nil && resp.StatusCode == http.StatusOK
	
	if resp != nil {
		resp.Body.Close()
	}

	if inst.Healthy != isHealthy {
		hc.logger.Info("Instance health changed",
			"service", serviceID,
			"url", inst.URL.String(),
			"healthy", isHealthy,
		)
		service := hc.registry.GetService(serviceID)
		if service != nil {
			service.UpdateHealth(inst.URL, isHealthy)
		}
	}
}
