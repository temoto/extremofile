package extremofile

import (
	"math/rand"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseChecksum(t *testing.T) {
	t.Parallel()

	f := func(data []byte) bool {
		b := append(data, checksum(data)...)
		back, err := parse(b)
		return assert.NoError(t, err) && assert.Equal(t, data, back)
	}
	assert.NoError(t, quick.Check(f, nil))
}

func TestParseShort(t *testing.T) {
	t.Parallel()

	f := func(b []byte) bool {
		if len(b) >= checkSize {
			b = b[:checkSize-1]
		}
		back, err := parse(b)
		return assert.Nil(t, back) && assert.Equal(t, errMetaParse, err)
	}
	assert.NoError(t, quick.Check(f, nil))
}

func TestParseCorrupt(t *testing.T) {
	t.Parallel()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	f := func(data []byte) bool {
		b := append(data, checksum(data)...)
		offset := rnd.Uint32() % uint32(len(b))
		b[offset]++
		back, err := parse(b)
		return assert.Nil(t, back) && assert.Equal(t, errCorrupt, err)
	}
	assert.NoError(t, quick.Check(f, nil))
}
