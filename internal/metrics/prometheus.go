package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	GatewayRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of HTTP requests processed by the gateway",
		},
		[]string{"method", "path", "status"},
	)

	GatewayRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Latency of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	UpstreamRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "upstream_latency_seconds",
			Help:    "Latency of upstream HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"upstream"},
	)

	RateLimitRejectedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_rejected_total",
			Help: "Total number of requests rejected by the rate limiter",
		},
		[]string{"route"},
	)

	CircuitBreakerOpenTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_open_total",
			Help: "Total number of times a circuit breaker opened",
		},
		[]string{"service"},
	)
)
