package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

// RedisSlidingWindow implements a distributed rate limiter using Redis
type RedisSlidingWindow struct {
	client *redis.Client
}

// NewRedisSlidingWindow creates a new Redis rate limiter
func NewRedisSlidingWindow(redisAddr string) *RedisSlidingWindow {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
		PoolSize: 100, // optimized for high throughput
	})

	return &RedisSlidingWindow{
		client: rdb,
	}
}

// Allow checks if the request is allowed using a sliding window algorithm in Redis via Lua script
func (r *RedisSlidingWindow) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	// A simple Lua script for sliding window.
	// We use ZSET where score is timestamp, and value is a unique identifier (or just timestamp if we don't have collisions).
	// For high throughput, a token bucket script is often preferred, but sliding window is requested.
	
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local clearBefore = now - window

		redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)
		local count = redis.call('ZCARD', key)
		
		if count < limit then
			redis.call('ZADD', key, now, now .. '-' .. ARGV[4])
			redis.call('EXPIRE', key, math.ceil(window/1000))
			return 1
		else
			return 0
		end
	`

	now := time.Now().UnixMilli()
	// ARGV4 is a random/unique value, we just use nanosecond to avoid ZSET collisions
	uniqueStr := fmt.Sprintf("%d", time.Now().UnixNano()) 

	result, err := r.client.Eval(ctx, script, []string{key}, limit, window.Milliseconds(), now, uniqueStr).Result()
	if err != nil {
		return false, fmt.Errorf("redis eval error: %w", err)
	}

	allowed := result.(int64) == 1
	return allowed, nil
}
