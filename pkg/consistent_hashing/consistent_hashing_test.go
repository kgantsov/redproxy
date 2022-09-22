package consistent_hashing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponserSendError(t *testing.T) {
	ch := NewConsistentHashing([]string{"host-0", "host-1", "host-2"}, 10)

	assert.Equal(t, ch.GetNode("key_1"), "host-2", "they should be equal")
	assert.Equal(t, ch.GetNode("key_2"), "host-0", "they should be equal")
	assert.Equal(t, ch.GetNode("key_3"), "host-0", "they should be equal")
	assert.Equal(t, ch.GetNode("key_4"), "host-2", "they should be equal")
	assert.Equal(t, ch.GetNode("key_5"), "host-1", "they should be equal")

	ch = NewConsistentHashing([]string{"host-0", "host-1", "host-2", "host-3"}, 10)
	assert.Equal(t, ch.GetNode("key_1"), "host-2", "they should be equal")
	assert.Equal(t, ch.GetNode("key_2"), "host-0", "they should be equal")
	assert.Equal(t, ch.GetNode("key_3"), "host-0", "they should be equal")
	assert.Equal(t, ch.GetNode("key_4"), "host-2", "they should be equal")
	assert.Equal(t, ch.GetNode("key_5"), "host-1", "they should be equal")
}
