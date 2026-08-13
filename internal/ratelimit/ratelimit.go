package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type RedisSlidingWindow struct {
	client *redis.Client
}

func NewRedisSlidingWindow(redisAddr string) *RedisSlidingWindow {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", 
		DB:       0,  
		PoolSize: 100, 
	})

	return &RedisSlidingWindow{
		client: rdb,
	}
}

func (r *RedisSlidingWindow) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	
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
	uniqueStr := fmt.Sprintf("%d", time.Now().UnixNano()) 

	result, err := r.client.Eval(ctx, script, []string{key}, limit, window.Milliseconds(), now, uniqueStr).Result()
	if err != nil {
		return false, fmt.Errorf("redis eval error: %w", err)
	}

	allowed := result.(int64) == 1
	return allowed, nil
}
