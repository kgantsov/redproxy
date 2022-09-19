package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v9"
)

type Client interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) int64
}

type MockClient struct {
	store map[string]string
}

func NewMockClient(store map[string]string) *MockClient {
	r := &MockClient{store: store}

	return r
}

func (c *MockClient) Get(ctx context.Context, key string) (string, error) {
	value, ok := c.store[key]

	if !ok {
		return "", errors.New("don't exist")
	}

	return value, nil
}

func (c *MockClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.store[key] = value.(string)

	return nil
}

func (c *MockClient) Del(ctx context.Context, keys ...string) int64 {
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

type RedisClient struct {
	redis *redis.Client
}

func NewRedisClient(host, port, password string, db int) *RedisClient {
	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	r := &RedisClient{redis: redis}

	return r
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	value, err := c.redis.Get(ctx, key).Result()

	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.redis.Set(ctx, key, value, expiration).Err()

	if err != nil {
		return err
	}

	return nil
}

func (c *RedisClient) Del(ctx context.Context, keys ...string) int64 {
	res := c.redis.Del(ctx, keys...).Val()

	return res
}
