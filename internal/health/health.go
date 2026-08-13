package health

import (
	"api-gateway/internal/discovery"
	"log/slog"
	"net/http"
	"time"
)

// Checker performs active health checking on service instances
type Checker struct {
	registry *discovery.Registry
	client   *http.Client
	interval time.Duration
	logger   *slog.Logger
}

// NewChecker initializes a new health checker
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

// Start runs the health checker loop in a goroutine
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
			// We check concurrently
			go hc.checkInstance(service.ID, inst)
		}
	}
}

func (hc *Checker) checkInstance(serviceID string, inst *discovery.Instance) {
	// Construct the health check URL. In a real system, this might be configurable per service.
	healthURL := inst.URL.String() + "/health"
	
	resp, err := hc.client.Get(healthURL)
	isHealthy := err == nil && resp.StatusCode == http.StatusOK
	
	if resp != nil {
		resp.Body.Close()
	}

	// If state changed, log and update
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
