package cloudy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeekingBuffer(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	buffer := NewSeekingBuffer(data)

	some := make([]byte, 2)

	count, err := buffer.Read(some)
	assert.Equal(t, count, int(2), "Expect 2 bytes read")
	assert.Nil(t, err, "Not expecting error in read")
	assert.Equal(t, buffer.offset, int64(2), "Expect to be at offset 2")

	// Seek from beginning
	seeked, err := buffer.Seek(5, 0)
	assert.Nil(t, err, "Not expecting error in seek")
	assert.Equal(t, seeked, int64(5), "Expect 5")
	assert.Equal(t, buffer.offset, int64(5), "Expect to be at offset 7")

	// Seek from current
	seeked2, err := buffer.Seek(2, 1)
	assert.Nil(t, err, "Not expecting error in seek")
	assert.Equal(t, int64(7), seeked2, "Expect 7")
	assert.Equal(t, int64(7), buffer.offset, "Expect to be at offset 7")

	// Seek from end
	seeked3, err := buffer.Seek(2, 2)
	assert.Nil(t, err, "Not expecting error in seek")
	assert.Equal(t, int64(8), seeked3, "Expect 8")
	assert.Equal(t, buffer.offset, int64(8), "Expect to be at offset 8")

}
