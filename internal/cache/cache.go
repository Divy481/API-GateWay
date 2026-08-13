package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisAddr string) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", 
		DB:       1, 
		PoolSize: 100,
	})

	return &RedisCache{client: rdb}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil 
		}
		return nil, err
	}
	return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}
