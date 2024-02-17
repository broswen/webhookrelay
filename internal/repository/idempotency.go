package repository

import (
	"context"
	"errors"
	"github.com/broswen/webhookrelay/internal/retry"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	ErrNoKey = errors.New("error no idempotency key found")
)

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
	_, err := retry.NewRetry(time.Millisecond*50, 3, func() (any, error, bool) {
		return nil, r.redis.SetEx(ctx, key, value, ttl).Err(), true
	})()
	return err
}

func (r *RedisIdempotencyRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := retry.NewRetry(time.Millisecond*50, 3, func() (string, error, bool) {
		val, err := r.redis.Get(ctx, key).Result()
		if err != nil {
			return "", err, !errors.Is(err, redis.Nil)
		}
		return val, err, false
	})()
	if err != nil && errors.Is(err, redis.Nil) {
		return "", ErrNoKey
	}
	return val, err
}
