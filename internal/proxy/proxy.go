package proxy

import (
	"api-gateway/internal/config"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Proxy struct {
	Transport *http.Transport
}

func NewProxy(cfg *config.Config) *Proxy {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.Server.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true, 
		MaxIdleConns:          cfg.Server.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.Server.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.Server.MaxConnsPerHost,
		IdleConnTimeout:       cfg.Server.IdleConnTimeout,
		ResponseHeaderTimeout: cfg.Server.ResponseHeaderTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		
	}

	return &Proxy{
		Transport: transport,
	}
}

func (p *Proxy) ForwardRequest(w http.ResponseWriter, r *http.Request, target *url.URL) {
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	proxy.Transport = p.Transport
	
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		
	}
	
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		
		http.Error(rw, "Bad Gateway", http.StatusBadGateway)
	}

	proxy.ServeHTTP(w, r)
}
