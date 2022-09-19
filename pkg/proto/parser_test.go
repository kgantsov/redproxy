package proto

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserParseCommand(t *testing.T) {
	tests := []struct {
		command string
		want    *Command
	}{
		{command: "*2\r\n$3\r\nGET\r\n$2\r\nk1\r\n", want: &Command{Name: "GET", Args: []string{"k1"}}},
		{command: "*2\r\n$3\r\nGET\r\n$2\r\nk2\r\n", want: &Command{Name: "GET", Args: []string{"k2"}}},
		{command: "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", want: &Command{Name: "SET", Args: []string{"foo", "bar"}}},
		{command: "*2\r\n$3\r\nDEL\r\n$2\r\nk1\r\n", want: &Command{Name: "DEL", Args: []string{"k1"}}},
		{command: "*3\r\n$3\r\nDEL\r\n$2\r\nk1\r\n$2\r\nk2\r\n", want: &Command{Name: "DEL", Args: []string{"k1", "k2"}}},
	}

	for _, tc := range tests {
		parser := NewParser(bufio.NewReader(strings.NewReader(tc.command)))

		cmd, _ := parser.ParseCommand()

		assert.Equal(t, cmd, tc.want, "they should be equal")
	}
}
