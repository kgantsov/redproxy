package proxy

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/consistent_hashing"
)

type Proxy interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) int64
	Keys(ctx context.Context, pattern string) []string
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

func (c *MockProxy) Keys(ctx context.Context, pattern string) []string {
	keys := []string{}

	for key, _ := range c.store {
		keys = append(keys, key)
	}
	return keys
}

type RedisProxy struct {
	redises           map[string]*redis.Client
	consistentHashing *consistent_hashing.ConsistentHashing
}

func NewRedisProxy(redisesOptions []*redis.Options) *RedisProxy {
	redises := map[string]*redis.Client{}
	for _, redisOptions := range redisesOptions {
		redis := redis.NewClient(redisOptions)
		redises[redisOptions.Addr] = redis
	}

	var nodes []string
	for _, redisOptions := range redisesOptions {
		nodes = append(nodes, redisOptions.Addr)
	}

	consistentHashing := consistent_hashing.NewConsistentHashing(nodes, 10)

	r := &RedisProxy{redises: redises, consistentHashing: consistentHashing}

	return r
}

func (c *RedisProxy) getNode(key string) *redis.Client {
	node := c.consistentHashing.GetNode(key)
	log.Debugf("Got a node `%s` for a key `%s`", node, key)
	return c.redises[node]
}

func (c *RedisProxy) getNodes(keys ...string) map[string]*redis.Client {
	keyClients := map[string]*redis.Client{}

	for _, key := range keys {
		node := c.consistentHashing.GetNode(key)
		log.Debugf("Got a node `%s` for a key `%s`", node, key)
		keyClients[key] = c.redises[node]
	}

	return keyClients
}

func (c *RedisProxy) Get(ctx context.Context, key string) (string, error) {
	value, err := c.getNode(key).Get(ctx, key).Result()

	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *RedisProxy) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := c.getNode(key).Set(ctx, key, value, expiration).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *RedisProxy) Del(ctx context.Context, keys ...string) int64 {
	var res int64

	keyClients := c.getNodes(keys...)

	for _, key := range keys {
		client := keyClients[key]
		res += client.Del(ctx, keys...).Val()
	}

	return res
}

func (c *RedisProxy) Keys(ctx context.Context, pattern string) []string {
	keys := []string{}

	for _, client := range c.redises {
		serverKeys := client.Keys(ctx, pattern).Val()
		keys = append(keys, serverKeys...)
	}

	return keys
}
