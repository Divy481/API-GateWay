package proxy

import (
	"api-gateway/internal/config"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// Proxy represents the API Gateway reverse proxy
type Proxy struct {
	Transport *http.Transport
}

// NewProxy creates a new highly-tuned reverse proxy
func NewProxy(cfg *config.Config) *Proxy {
	// 1. Tuned http.Transport for high-throughput and connection pooling
	// This is critical for reverse proxy performance. It avoids creating new TCP
	// connections for every request and instead reuses idle ones.
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.Server.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true, // Attempt to use HTTP/2 with upstream
		MaxIdleConns:          cfg.Server.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.Server.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.Server.MaxConnsPerHost,
		IdleConnTimeout:       cfg.Server.IdleConnTimeout,
		ResponseHeaderTimeout: cfg.Server.ResponseHeaderTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Buffer sizing could be explicitly set to optimize for latency or throughput
		// ReadBufferSize:  32 * 1024,
		// WriteBufferSize: 32 * 1024,
	}

	return &Proxy{
		Transport: transport,
	}
}

// ForwardRequest handles the actual proxying of the request to a chosen upstream URL
func (p *Proxy) ForwardRequest(w http.ResponseWriter, r *http.Request, target *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Use our highly tuned transport instead of the default transport
	proxy.Transport = p.Transport
	
	// Modify the Director to also pass the original context
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// Set headers for tracing and context
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		// Delete X-Forwarded-For if we want to reset it (or append to it)
		// It's usually better to let the reverse proxy handle it automatically 
		// but we can enforce it.
	}
	
	// Error handling if backend is unavailable
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		// Log the error (we'll integrate structured logging here later)
		// For now, return a 502 Bad Gateway
		http.Error(rw, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}
