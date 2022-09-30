package proto

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupClients() map[string]RedisClient {
	clients := map[string]RedisClient{}

	redisClient1 := NewMockRedisClient(
		map[string]string{"k6": "value_6", "k213": "value_213", "k32151": "value_32151"},
	)
	clients["localhost:16379"] = redisClient1

	redisClient2 := NewMockRedisClient(
		map[string]string{"k0": "value_0", "k1": "value_1", "k5": "value_5", "k9": "value_9", "k10": "value_10"},
	)
	clients["localhost:26379"] = redisClient2

	redisClient3 := NewMockRedisClient(
		map[string]string{"k2": "value_2", "k3": "value_3", "k4": "value_4", "k7": "value_7", "k8": "value_8"},
	)
	clients["localhost:36379"] = redisClient3

	return clients
}

func TestServerGet(t *testing.T) {
	port := 46379

	redises := setupClients()

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	tests := []struct {
		err  error
		key  string
		want string
	}{
		{key: "k0", want: "value_0", err: nil},
		{key: "k1", want: "value_1", err: nil},
		{key: "k2", want: "value_2", err: nil},
		{key: "k3", want: "value_3", err: nil},
		{key: "k4", want: "value_4", err: nil},
		{key: "k5", want: "value_5", err: nil},
		{key: "k6", want: "value_6", err: nil},
		{key: "k7", want: "value_7", err: nil},
		{key: "k8", want: "value_8", err: nil},
		{key: "k213", want: "value_213", err: nil},
		{key: "k32151", want: "value_32151", err: nil},
		{key: "foo", want: "", err: redis.Nil},
	}

	for _, tc := range tests {
		var ctx = context.Background()
		val, err := client.Get(ctx, tc.key).Result()

		assert.Equal(t, tc.err, err, fmt.Sprintf("GET %s error", tc.key))
		assert.Equal(t, tc.want, val, fmt.Sprintf("GET %s", tc.key))
	}

	server.Stop()
}

func TestServerSet(t *testing.T) {
	port := 56379

	redises := setupClients()

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	val, err := client.Get(ctx, "new_key").Result()

	assert.Equal(t, redis.Nil, err, "they should be equal")
	assert.Equal(t, "", val, "they should be equal")

	err = client.Set(ctx, "new_key", "new value", time.Duration(0)).Err()

	assert.Equal(t, nil, err, "they should be equal")

	val, err = client.Get(ctx, "new_key").Result()

	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "new value", val, "they should be equal")

	server.Stop()
}

func TestServerDel(t *testing.T) {
	port := 36379

	redises := setupClients()

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	val, err := client.Get(ctx, "k1").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "value_1", val, "they should be equal")

	val, err = client.Get(ctx, "k2").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "value_2", val, "they should be equal")

	val, err = client.Get(ctx, "k3").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "value_3", val, "they should be equal")

	deleted := client.Del(ctx, "k1", "k3").Val()

	assert.Equal(t, int64(2), deleted, "they should be equal")

	val, err = client.Get(ctx, "k1").Result()
	assert.Equal(t, redis.Nil, err, "they should be equal")
	assert.Equal(t, "", val, "they should be equal")

	val, err = client.Get(ctx, "k2").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "value_2", val, "they should be equal")

	val, err = client.Get(ctx, "k3").Result()
	assert.Equal(t, redis.Nil, err, "they should be equal")
	assert.Equal(t, "", val, "they should be equal")
	server.Stop()
}

func TestServerKeys(t *testing.T) {
	port := 46379

	redises := setupClients()

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	tests := []struct {
		err  error
		key  string
		want []string
	}{
		{key: "k*", want: []string{"k0", "k1", "k10", "k2", "k213", "k3", "k32151", "k4", "k5", "k6", "k7", "k8", "k9"}, err: nil},
		{key: "k2", want: []string{"k2"}, err: nil},
		{key: "*", want: []string{"k0", "k1", "k10", "k2", "k213", "k3", "k32151", "k4", "k5", "k6", "k7", "k8", "k9"}, err: nil},
		{key: "k3215*", want: []string{"k32151"}, err: nil},
		{key: "k32151", want: []string{"k32151"}, err: nil},
		{key: "foo", want: []string{}, err: nil},
	}

	for _, tc := range tests {
		var ctx = context.Background()
		val, err := client.Keys(ctx, tc.key).Result()

		sort.Slice(val, func(i, j int) bool {
			return val[i] < val[j]
		})

		assert.Equal(t, tc.err, err, fmt.Sprintf("GET %s error", tc.key))
		assert.Equal(t, tc.want, val, fmt.Sprintf("GET %s", tc.key))
	}

	server.Stop()
}
