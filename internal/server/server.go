package server

import (
	"api-gateway/internal/config"
	"api-gateway/internal/discovery"
	"api-gateway/internal/loadbalancer"
	"api-gateway/internal/proxy"
	"context"
	"log/slog"
	"net/http"
	"os"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	proxy      *proxy.Proxy
	registry   *discovery.Registry
	lb         loadbalancer.LoadBalancer
	logger     *slog.Logger
}

func NewServer(cfg *config.Config, p *proxy.Proxy, registry *discovery.Registry, lb loadbalancer.LoadBalancer, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	srv := &Server{
		config:   cfg,
		proxy:    p,
		registry: registry,
		lb:       lb,
		logger:   logger,
	}

	mux := http.NewServeMux()
	
	for _, route := range cfg.Routes {
		route := route
		mux.HandleFunc(route.Path+"/", func(w http.ResponseWriter, r *http.Request) {
			service := srv.registry.GetService(route.ID)
			if service == nil {
				http.Error(w, "Service not found", http.StatusNotFound)
				return
			}
			
			instance, err := srv.lb.Next(service)
			if err != nil {
				srv.logger.Warn("No healthy instances available", "route", route.ID)
				http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
				return
			}
			
			srv.proxy.ForwardRequest(w, r, instance.URL)
		})
	}

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	srv.httpServer = &http.Server{
		Addr:           cfg.Server.Address,
		Handler:        mux,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	return srv
}

func (s *Server) Start() error {
	s.logger.Info("Starting API Gateway", "address", s.config.Server.Address)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down API Gateway")
	return s.httpServer.Shutdown(ctx)
}
