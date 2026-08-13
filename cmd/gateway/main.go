package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/discovery"
	"api-gateway/internal/health"
	"api-gateway/internal/loadbalancer"
	"api-gateway/internal/proxy"
	"api-gateway/internal/server"
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup Logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. Initialize Service Registry & Health Checker
	registry := discovery.NewRegistry()
	for _, route := range cfg.Routes {
		service := discovery.NewService(route.ID, route.Upstreams)
		registry.AddService(service)
	}

	healthChecker := health.NewChecker(registry, logger)
	healthChecker.Start() // Background health checks

	// 3. Initialize Load Balancer and Proxy
	lb := loadbalancer.NewRoundRobin()
	gatewayProxy := proxy.NewProxy(cfg)
	
	// 4. Initialize Core Server
	gatewayServer := server.NewServer(cfg, gatewayProxy, registry, lb, logger)

	// 5. Start Server in a goroutine
	go func() {
		if err := gatewayServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 4. Graceful Shutdown (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	<-quit
	log.Println("Received shutdown signal, initiating graceful shutdown...")

	// Create a context with a timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := gatewayServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway shutdown complete")
}
