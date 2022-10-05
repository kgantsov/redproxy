package proto

import (
	"context"
	"time"

	"github.com/go-redis/redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/kgantsov/redproxy/pkg/consistent_hashing"
)

type RedisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Append(ctx context.Context, key, value string) *redis.IntCmd
	IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd
	DecrBy(ctx context.Context, key string, decrement int64) *redis.IntCmd
	Keys(ctx context.Context, pattern string) *redis.StringSliceCmd
}

type RedisProxy struct {
	clients           map[string]RedisClient
	consistentHashing *consistent_hashing.ConsistentHashing
}

func NewRedisProxy(clients map[string]RedisClient) *RedisProxy {
	nodes := make([]string, 0)

	for addr, _ := range clients {
		nodes = append(nodes, addr)
	}

	consistentHashing := consistent_hashing.NewConsistentHashing(nodes, 10)

	r := &RedisProxy{clients: clients, consistentHashing: consistentHashing}

	return r
}

func (c *RedisProxy) getNode(key string) RedisClient {
	node := c.consistentHashing.GetNode(key)
	log.Debugf("Got a node `%s` for a key `%s`", node, key)

	return c.clients[node]
}

func (c *RedisProxy) getNodes(keys ...string) map[string]RedisClient {
	keyClients := map[string]RedisClient{}

	for _, key := range keys {
		node := c.consistentHashing.GetNode(key)
		log.Debugf("Got a node `%s` for a key `%s`", node, key)
		keyClients[key] = c.clients[node]
	}

	return keyClients
}

func (c *RedisProxy) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.getNode(key).Get(ctx, key)
}

func (c *RedisProxy) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.getNode(key).Set(ctx, key, value, expiration)
}

func (c *RedisProxy) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	var res int64

	keyClients := c.getNodes(keys...)

	for _, key := range keys {
		client := keyClients[key]
		res += client.Del(ctx, keys...).Val()
	}

	cmd := &redis.IntCmd{}
	cmd.SetVal(res)

	return cmd
}

func (c *RedisProxy) Append(ctx context.Context, key, value string) *redis.IntCmd {
	return c.getNode(key).Append(ctx, key, value)
}

func (c *RedisProxy) IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd {
	return c.getNode(key).IncrBy(ctx, key, value)
}

func (c *RedisProxy) DecrBy(ctx context.Context, key string, decrement int64) *redis.IntCmd {
	return c.getNode(key).DecrBy(ctx, key, decrement)
}

func (c *RedisProxy) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	keys := []string{}

	for _, client := range c.clients {
		serverKeys := client.Keys(ctx, pattern).Val()
		keys = append(keys, serverKeys...)
	}

	cmd := &redis.StringSliceCmd{}
	cmd.SetVal(keys)

	return cmd
}
