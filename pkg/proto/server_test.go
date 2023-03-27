package proto

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/kgantsov/redproxy/pkg/consistent_hashing"
	"github.com/stretchr/testify/assert"
)

func setupClients(n int) map[string]RedisClient {
	startPort := 6379
	nodes := make([]string, 0)
	clients := map[string]RedisClient{}

	var ctx = context.Background()

	for i := 0; i < n; i++ {
		node := fmt.Sprintf("localhost:%d", startPort)
		clientRedis := redis.NewClient(&redis.Options{
			Addr:     node,
			Password: "",
			DB:       0,
		})
		clients[node] = clientRedis
		clientRedis.FlushAll(ctx)

		nodes = append(nodes, node)

		startPort++
	}

	consistentHashing := consistent_hashing.NewConsistentHashing(nodes, 10)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		node := consistentHashing.GetNode(key)
		client := clients[node]
		fmt.Printf("=====> %s %s\n", node, key)
		client.Set(ctx, key, value, time.Duration(0))
	}

	return clients
}

func TestServerGet(t *testing.T) {
	port := 46379

	redises := setupClients(3)

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
		{key: "foo", want: "", err: redis.Nil},
		{key: "k20", want: "", err: redis.Nil},
		{key: "foodsadsadcx", want: "", err: redis.Nil},
	}

	for i := 0; i < 20; i++ {
		tests = append(
			tests,
			struct {
				err  error
				key  string
				want string
			}{key: fmt.Sprintf("key_%d", i), want: fmt.Sprintf("value_%d", i), err: nil},
		)
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

	redises := setupClients(3)

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

	redises := setupClients(3)

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	for i := 0; i < 20; i++ {
		val, err := client.Get(ctx, fmt.Sprintf("key_%d", i)).Result()
		assert.Equal(t, nil, err, "they should be equal")
		assert.Equal(t, fmt.Sprintf("value_%d", i), val, "they should be equal")
	}

	deleted := client.Del(ctx, "key_0", "key_1", "key_2", "key_3", "key_4").Val()

	assert.Equal(t, int64(5), deleted, "they should be equal")

	for i := 0; i < 20; i++ {
		val, err := client.Get(ctx, fmt.Sprintf("key_%d", i)).Result()

		if i < 5 {
			assert.Equal(t, redis.Nil, err, "they should be equal")
			assert.Equal(t, "", val, "they should be equal")
		} else {
			assert.Equal(t, nil, err, "they should be equal")
			assert.Equal(t, fmt.Sprintf("value_%d", i), val, "they should be equal")
		}
	}

	server.Stop()
}

func TestServerAppend(t *testing.T) {
	port := 56379

	redises := setupClients(3)

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	val, err := client.Get(ctx, "key_0").Result()

	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "value_0", val, "they should be equal")

	err = client.Append(ctx, "key_0", "_long_suffix_0").Err()

	assert.Equal(t, nil, err, "they should be equal")

	val, err = client.Get(ctx, "key_0").Result()

	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "value_0_long_suffix_0", val, "they should be equal")

	server.Stop()
}

func TestServerIncrDecr(t *testing.T) {
	port := 56379

	redises := setupClients(3)

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	err := client.Set(ctx, "counter", 1, time.Duration(0)).Err()
	assert.Equal(t, nil, err, "they should be equal")

	val, _ := client.Get(ctx, "counter").Result()
	assert.Equal(t, "1", val, "they should be equal")

	err = client.Incr(ctx, "counter").Err()
	assert.Equal(t, nil, err, "they should be equal")

	val, _ = client.Get(ctx, "counter").Result()
	assert.Equal(t, "2", val, "they should be equal")

	err = client.IncrBy(ctx, "counter", 8).Err()
	assert.Equal(t, nil, err, "they should be equal")

	val, _ = client.Get(ctx, "counter").Result()
	assert.Equal(t, "10", val, "they should be equal")

	err = client.Decr(ctx, "counter").Err()
	assert.Equal(t, nil, err, "they should be equal")

	val, _ = client.Get(ctx, "counter").Result()
	assert.Equal(t, "9", val, "they should be equal")

	err = client.DecrBy(ctx, "counter", 4).Err()
	assert.Equal(t, nil, err, "they should be equal")

	val, _ = client.Get(ctx, "counter").Result()
	assert.Equal(t, "5", val, "they should be equal")

	server.Stop()
}

func TestServerKeys(t *testing.T) {
	port := 46379

	redises := setupClients(3)

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	all := []string{}
	for i := 0; i < 20; i++ {
		all = append(all, fmt.Sprintf("key_%d", i))
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i] < all[j]
	})

	tests := []struct {
		err  error
		key  string
		want []string
	}{
		{key: "k*", want: all, err: nil},
		{key: "key_*", want: all, err: nil},
		{key: "key_2", want: []string{"key_2"}, err: nil},
		{key: "*", want: all, err: nil},
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

func TestServerMGet(t *testing.T) {
	port := 46379

	redises := setupClients(3)

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
		keys []string
		want []interface{}
	}{
		{
			keys: []string{"key_1", "key_2", "key_3", "do not exist", "key_4", "KJDK", "key_8"},
			want: []interface{}{"value_1", "value_2", "value_3", "", "value_4", "", "value_8"},
			err:  nil,
		},
	}

	r := redises["localhost:6379"]

	for _, tc := range tests {
		var ctx = context.Background()
		fmt.Printf("=====> )))) -----> %#v", r.MGet(ctx, tc.keys...).Val())
		val, err := client.MGet(ctx, tc.keys...).Result()
		fmt.Printf("=====> )))) %s %+v", val, r.MGet(ctx, tc.keys...).Val())

		assert.Equal(t, tc.err, err, fmt.Sprintf("MGET %s error", tc.keys))
		assert.Equal(t, tc.want, val, fmt.Sprintf("MGET %s", tc.keys))
	}

	server.Stop()
}

func TestServerMSet(t *testing.T) {
	port := 46379

	redises := setupClients(3)

	_proxy := NewRedisProxy(redises)
	server := NewServer(_proxy, port)

	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	tests := []struct {
		err     error
		keyVals []string
		want    []interface{}
	}{
		{
			keyVals: []string{"key1", "value1", "key2", "value2", "foo", "bar", "IOKL", "cxzcsd123"},
			want:    []interface{}{"value1", "value2", "bar", "cxzcsd123"},
			err:     nil,
		},
		{
			keyVals: []string{"test_key", "this is the value"},
			want:    []interface{}{"this is the value"},
			err:     nil,
		},
	}

	for _, tc := range tests {
		var ctx = context.Background()

		err := client.MSet(
			ctx, tc.keyVals,
		).Err()
		assert.Equal(t, nil, err)

		keys := []string{}
		for i := 0; i < len(tc.keyVals); i += 2 {
			keys = append(keys, tc.keyVals[i])
		}

		val, err := client.MGet(ctx, keys...).Result()

		assert.Equal(t, tc.err, err, fmt.Sprintf("MGET %s error", keys))
		assert.Equal(t, tc.want, val, fmt.Sprintf("MGET %s", keys))
	}

	server.Stop()
}
