package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines the caching interface
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new redis cache client
func NewRedisCache(redisAddr string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", 
		DB:       1, // Use DB 1 for cache (DB 0 for rate limit)
		PoolSize: 100,
	})

	return &RedisCache{client: rdb}
}

// Get retrieves an item
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}
	return val, nil
}

// Set stores an item
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}
