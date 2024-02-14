package repository

import (
	"context"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var InProgressKey = "__in_progress"

type Idempotency interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

func NewRedisIdempotencyRepository(address string) (Idempotency, error) {
	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: strings.Split(address, ","),
	})

	return &RedisIdempotencyRepository{redis: rdb}, nil
}

type RedisIdempotencyRepository struct {
	redis *redis.Client
}

func (r *RedisIdempotencyRepository) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.redis.SetEx(ctx, key, value, ttl).Err()
}

func (r *RedisIdempotencyRepository) Get(ctx context.Context, key string) (string, error) {
	return r.redis.Get(ctx, key).Result()
}
