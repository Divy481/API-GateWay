# High-Throughput Distributed API Gateway

A production-grade, highly optimized API Gateway written in Go. This project demonstrates advanced distributed systems engineering, emphasizing high throughput, fault tolerance, and comprehensive observability.

## Architecture

```text
                         Internet
                            │
                            ▼
                  ┌──────────────────┐
                  │   Load Balancer  │ (Kubernetes Service)
                  └────────┬─────────┘
                           │
                           ▼
              ┌─────────────────────────┐
              │      API Gateway        │
              │                         │
              │  HTTP Server            │
              │  Router                 │
              │  Authentication         │
              │  Rate Limiter           │
              │  Load Balancer          │
              │  Reverse Proxy          │
              │  Circuit Breaker        │
              │  Retry                  │
              │  Cache                  │
              │  Metrics / Tracing      │
              └────────────┬────────────┘
                           │
          ┌────────────────┼────────────────┐
          ▼                ▼                ▼
    User Service      Order Service    Payment Service
          │                │                │
          └────────────────┼────────────────┘
                           │
                     Redis/PostgreSQL
```

---

## 🚀 Performance Benchmarks & Targets

The gateway is designed to maintain **low single-digit millisecond overhead** even under load. 

**Local Microbenchmark (Core Proxy Overhead):**
```text
BenchmarkProxy-4   	   23994	     41969 ns/op	   40282 B/op	      86 allocs/op
```
*Proxying overhead per request is kept consistently under ~0.04 ms locally, driven by aggressive connection pooling and tuning of the underlying `http.Transport`.*

**Target SLOs:**
- **Availability:** 99.9%
- **P95 Latency:** < 100 ms
- **P99 Latency:** < 250 ms
- **5xx Error Rate:** < 0.1%

---

## 🏗️ Major Engineering Components & Trade-offs

### 1. High-Performance Reverse Proxy (`internal/proxy`)
- **Problem:** Standard HTTP proxies open new TCP connections per request, destroying throughput under heavy load.
- **Design & Implementation:** We rely on `httputil.ReverseProxy` powered by a heavily customized `http.Transport`. 
- **Optimization:** We tuned `MaxIdleConns` (1000) and `MaxIdleConnsPerHost` (100) to keep sockets alive. `ExpectContinueTimeout` and `TLSHandshakeTimeout` are tightly bounded.
- **Concurrency & Failure Modes:** Lock-free processing per request. The primary failure mode is socket exhaustion (TIME_WAIT spikes), mitigated by `IdleConnTimeout`.

### 2. Lock-Free Load Balancing (`internal/loadbalancer`)
- **Problem:** A Round-Robin load balancer requires state (which instance is next). Mutexes here create massive contention.
- **Design & Implementation:** Implemented a lock-free load balancer using `sync/atomic`. 
- **Concurrency:** `atomic.AddUint64(&rr.counter, 1)` safely handles concurrent request assignments without blocking goroutines.
- **Failure Modes:** If all instances fail, the load balancer correctly returns `ErrNoHealthyInstances`, which the HTTP server maps to a 503.

### 3. Distributed Rate Limiter (`internal/ratelimit`)
- **Problem:** Multiple API gateway instances must enforce global rate limits without extreme latency overhead.
- **Design & Implementation:** Redis-backed sliding window using atomic Lua scripting. 
- **Optimization:** Lua scripting guarantees atomic execution in Redis, reducing the network round-trips from 3 (Get, Add, Expire) to exactly 1. We chose sliding window over token bucket to demonstrate granular time-based control.

### 4. Circuit Breaker (`internal/circuitbreaker`)
- **Problem:** Cascading failures. A slow backend ties up gateway goroutines, eventually crashing the gateway.
- **Design & Implementation:** An atomic state machine (`Closed`, `Open`, `Half-Open`). 
- **Concurrency:** Uses `sync/atomic` (CAS operations) for state transitions to avoid read/write locks on the hot path, only acquiring a `sync.Mutex` on rare state changes.
- **Failure Modes:** When open, it fails fast, allowing the gateway to shed load instantly. 

### 5. Service Discovery & Active Health Checking (`internal/health`)
- **Implementation:** Background goroutines proactively ping upstream `/health` endpoints. 
- **Concurrency:** Uses `sync.RWMutex` to update instance health. Reads are highly concurrent (via `RLock`), while writes only occur if health state flips.

---

## 📊 Observability (RED & USE Methodology)

Fully instrumented using OpenTelemetry and Prometheus. 

1. **Metrics (Prometheus):**
   - **Rate:** `gateway_requests_total`
   - **Errors:** 5xx rate tracked via HTTP status dimensions.
   - **Duration:** `gateway_request_duration_seconds` (Histogram for P50, P95, P99).
2. **Distributed Tracing (Jaeger):**
   - W3C Trace Context propagation.
   - Answers: *"Did the 800ms latency happen in the proxy, Redis, or the Upstream?"*
3. **Structured Logging:**
   - Zero-allocation JSON logging via `slog`, automatically binding `trace_id` and `span_id`.

---

## 🧪 Failure Testing Scenarios

The system is built to survive the following chaotic events:
1. **Upstream Timeout/Crash:** The background health checker detects the failure and drains traffic. Existing connections trigger the **Circuit Breaker** to open, preventing retry storms.
2. **Traffic Spike:** Redis Sliding Window blocks excessive requests, responding with HTTP 429, preventing backend saturation.
3. **Transient Network Loss:** The **Retry Policy** uses exponential backoff and 20% Jitter to smoothly recover missed packets without amplifying the outage.

---

## 🛠️ Quick Start

**1. Spin up the Observability Stack (Redis, Prometheus, Grafana, Jaeger)**
```bash
cd monitoring
docker-compose up -d
```

**2. Run Dummy Backends (Simulated Services)**
```bash
go run tests/dummy_backend/main.go 8081 &
go run tests/dummy_backend/main.go 8082 &
go run tests/dummy_backend/main.go 8083 &
```

**3. Start the API Gateway**
```bash
go run cmd/gateway/main.go
```

**4. Test the Routing**
```bash
# Triggers Round-Robin between 8081 and 8082
curl http://localhost:8080/api/users/ 
```

**Kubernetes Deployment**
```bash
kubectl apply -f deployments/kubernetes/
```