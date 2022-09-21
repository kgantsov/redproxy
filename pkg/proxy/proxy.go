package proxy

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v9"
)

type Proxy interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) int64
}

type MockProxy struct {
	store map[string]string
}

func NewMockProxy(store map[string]string) *MockProxy {
	r := &MockProxy{store: store}

	return r
}

func (c *MockProxy) Get(ctx context.Context, key string) (string, error) {
	value, ok := c.store[key]

	if !ok {
		return "", errors.New("don't exist")
	}

	return value, nil
}

func (c *MockProxy) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.store[key] = value.(string)

	return nil
}

func (c *MockProxy) Del(ctx context.Context, keys ...string) int64 {
	var res int64

	for _, key := range keys {
		_, ok := c.store[key]

		if ok {
			res++
			delete(c.store, key)
		}
	}
	return res
}

type RedisProxy struct {
	redises []*redis.Client
}

func NewRedisProxy(redisesOptions []*redis.Options) *RedisProxy {
	var redises []*redis.Client
	for _, redisOptions := range redisesOptions {
		redis := redis.NewClient(redisOptions)
		redises = append(redises, redis)
	}

	r := &RedisProxy{redises: redises}

	return r
}

func (c *RedisProxy) Get(ctx context.Context, key string) (string, error) {
	value, err := c.redises[0].Get(ctx, key).Result()

	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *RedisProxy) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.redises[0].Set(ctx, key, value, expiration).Err()

	if err != nil {
		return err
	}

	return nil
}

func (c *RedisProxy) Del(ctx context.Context, keys ...string) int64 {
	res := c.redises[0].Del(ctx, keys...).Val()

	return res
}
