package proto

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtoHandleRequest(t *testing.T) {
	store := map[string]string{"k1": "v1", "k2": "2", "k3": "value", "year": "2022"}

	tests := []struct {
		store   map[string]string
		command string
		want    string
	}{
		{store: store, command: "*2\r\n$3\r\nGET\r\n$2\r\nk1\r\n", want: "+v1\r\n"},
		{store: store, command: "*2\r\n$3\r\nGET\r\n$2\r\nk2\r\n", want: "+2\r\n"},
		{store: store, command: "*2\r\n$3\r\nGET\r\n$2\r\nk3\r\n", want: "+value\r\n"},
		{store: store, command: "*2\r\n$3\r\nGET\r\n$4\r\nyear\r\n", want: "+2022\r\n"},
		{store: store, command: "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n", want: "$-1\r\n"},
		{store: store, command: "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", want: "+OK\r\n"},
		// {store: store, command: "*2\r\n$3\r\nDEL\r\n$2\r\nk1\r\n", want: ":1\r\n"},
	}

	for _, tc := range tests {
		redisClient := NewMockRedisClient(tc.store)
		redises := map[string]RedisClient{}
		redises["localhost:6379"] = redisClient
		_proxy := NewRedisProxy(redises)

		buf := new(bytes.Buffer)
		redisProto := NewProto(_proxy, strings.NewReader(tc.command), buf)
		redisProto.HandleRequest()
		assert.Equal(t, buf.String(), tc.want, "they should be equal")
	}
}
