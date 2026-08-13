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
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	registry := discovery.NewRegistry()
	for _, route := range cfg.Routes {
		service := discovery.NewService(route.ID, route.Upstreams)
		registry.AddService(service)
	}

	healthChecker := health.NewChecker(registry, logger)
	healthChecker.Start() 

	lb := loadbalancer.NewRoundRobin()
	gatewayProxy := proxy.NewProxy(cfg)
	
	gatewayServer := server.NewServer(cfg, gatewayProxy, registry, lb, logger)

	go func() {
		if err := gatewayServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Received shutdown signal, initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := gatewayServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway shutdown complete")
}
