package proxy_test

import (
	"api-gateway/internal/config"
	"api-gateway/internal/proxy"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func BenchmarkProxy(b *testing.B) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer backend.Close()
	
	backendURL, _ := url.Parse(backend.URL)

	cfg := &config.Config{
		Server: config.ServerConfig{
			MaxIdleConns:          1000,
			MaxIdleConnsPerHost:   100,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			DialTimeout:           5 * time.Second,
		},
	}

	p := proxy.NewProxy(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.ForwardRequest(w, r, backendURL)
	})
	
	req := httptest.NewRequest(http.MethodGet, "http://localhost/", nil)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)
		}
	})
}
