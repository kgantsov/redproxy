package proto

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kgantsov/redproxy/pkg/proxy"
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
		{store: store, command: "*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n", want: "+\r\n"},
		{store: store, command: "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", want: "+OK\r\n"},
		{store: store, command: "*2\r\n$3\r\nDEL\r\n$2\r\nk1\r\n", want: ":1\r\n"},
	}

	for _, tc := range tests {
		proxy := proxy.NewMockProxy(tc.store)

		buf := new(bytes.Buffer)
		redisProto := NewProto(proxy, strings.NewReader(tc.command), buf)
		redisProto.HandleRequest()
		assert.Equal(t, buf.String(), tc.want, "they should be equal")
	}
}
