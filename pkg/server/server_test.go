package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/kgantsov/redproxy/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func TestServerGet(t *testing.T) {
	proxy := proxy.NewMockProxy(
		map[string]string{"k1": "v1", "k2": "2", "k3": "value", "year": "2022"},
	)
	port := 46379

	server := NewServer(proxy, port)
	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	tests := []struct {
		key  string
		want string
		err  error
	}{
		{key: "k1", want: "v1", err: nil},
		{key: "k2", want: "2", err: nil},
		{key: "k3", want: "value", err: nil},
		{key: "year", want: "2022", err: nil},
		{key: "foo", want: "", err: redis.Nil},
	}

	for _, tc := range tests {
		var ctx = context.Background()
		val, err := client.Get(ctx, tc.key).Result()

		assert.Equal(t, tc.err, err, fmt.Sprintf("GET %s error", tc.key))
		assert.Equal(t, tc.want, val, fmt.Sprintf("GET %s", tc.key))
	}
}

func TestServerSet(t *testing.T) {
	proxy := proxy.NewMockProxy(
		map[string]string{"k1": "v1", "k2": "2", "k3": "value", "year": "2022"},
	)
	port := 56379

	server := NewServer(proxy, port)
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
}

func TestServerDel(t *testing.T) {
	proxy := proxy.NewMockProxy(
		map[string]string{"k1": "v1", "k2": "2", "k3": "value", "year": "2022"},
	)
	port := 36379

	server := NewServer(proxy, port)
	go server.ListenAndServe()

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%d", port),
		Password: "",
		DB:       0,
	})

	var ctx = context.Background()

	val, err := client.Get(ctx, "k1").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "v1", val, "they should be equal")

	val, err = client.Get(ctx, "k2").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "2", val, "they should be equal")

	val, err = client.Get(ctx, "k3").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "value", val, "they should be equal")

	deleted := client.Del(ctx, "k1", "k3").Val()

	assert.Equal(t, int64(2), deleted, "they should be equal")

	val, err = client.Get(ctx, "k1").Result()
	assert.Equal(t, redis.Nil, err, "they should be equal")
	assert.Equal(t, "", val, "they should be equal")

	val, err = client.Get(ctx, "k2").Result()
	assert.Equal(t, nil, err, "they should be equal")
	assert.Equal(t, "2", val, "they should be equal")

	val, err = client.Get(ctx, "k3").Result()
	assert.Equal(t, redis.Nil, err, "they should be equal")
	assert.Equal(t, "", val, "they should be equal")
}
