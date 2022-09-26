package proto

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponserSendError(t *testing.T) {
	tests := []struct {
		value error
		want  string
	}{
		{value: errors.New("Not found"), want: "-ERR Not found\r\n"},
		{value: errors.New("Parsing error"), want: "-ERR Parsing error\r\n"},
	}

	for _, tc := range tests {
		buf := new(bytes.Buffer)
		responser := NewResponser(buf)

		responser.SendError(tc.value)

		assert.Equal(t, buf.String(), tc.want, "they should be equal")
	}
}

func TestSendPong(t *testing.T) {
	tests := []struct {
		want string
	}{
		{want: "+PONG\r\n"},
	}

	for _, tc := range tests {
		buf := new(bytes.Buffer)
		responser := NewResponser(buf)

		responser.SendPong()

		assert.Equal(t, buf.String(), tc.want, "they should be equal")
	}
}

func TestResponserSendInt(t *testing.T) {
	tests := []struct {
		want  string
		value int64
	}{
		{value: -123, want: ":-123\r\n"},
		{value: 0, want: ":0\r\n"},
		{value: 123, want: ":123\r\n"},
		{value: 72863872136, want: ":72863872136\r\n"},
	}

	for _, tc := range tests {
		buf := new(bytes.Buffer)
		responser := NewResponser(buf)

		responser.SendInt(tc.value)

		assert.Equal(t, buf.String(), tc.want, "they should be equal")
	}
}

func TestResponserSendStr(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "foo", want: "+foo\r\n"},
		{value: "OK", want: "+OK\r\n"},
		{value: "This is a test string", want: "+This is a test string\r\n"},
		{value: "", want: "+\r\n"},
	}

	for _, tc := range tests {
		buf := new(bytes.Buffer)
		responser := NewResponser(buf)

		responser.SendStr(tc.value)

		assert.Equal(t, buf.String(), tc.want, "they should be equal")
	}
}

func TestResponserSendArr(t *testing.T) {
	tests := []struct {
		want  string
		value []string
	}{
		{value: []string{"foo"}, want: "*1\r\n$3\r\nfoo\r\n"},
		{value: []string{""}, want: "*1\r\n$0\r\n\r\n"},
		{value: []string{"v1", "v2", "3"}, want: "*3\r\n$2\r\nv1\r\n$2\r\nv2\r\n$1\r\n3\r\n"},
		{value: []string{}, want: "*0\r\n"},
	}

	for _, tc := range tests {
		buf := new(bytes.Buffer)
		responser := NewResponser(buf)

		responser.SendArr(tc.value)

		assert.Equal(t, buf.String(), tc.want, "they should be equal")
	}
}
