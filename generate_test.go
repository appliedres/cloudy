package cloudy

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePassword(t *testing.T) {
	p := GeneratePassword(8, 2, 2, 2)

	assert.Equal(t, len(p), 8)
	assert.Regexp(t, regexp.MustCompile(".*[0-9].*[0-9].*"), p, "Expecting 2 digits")
	assert.Regexp(t, regexp.MustCompile(".*[A-Z].*[A-Z].*"), p, "Expecting 2 uppercase")
	assert.Regexp(t, regexp.MustCompile(".*[!@#$%&*].*[!@#$%&*].*"), p, "Expecting 2 special")

	p = GeneratePasswordNoSpecial(8, 2, 2)

	assert.Equal(t, len(p), 8)
	assert.Regexp(t, regexp.MustCompile(".*[0-9].*[0-9].*"), p, "Expecting 2 digits")
	assert.Regexp(t, regexp.MustCompile(".*[A-Z].*[A-Z].*"), p, "Expecting 2 uppercase")
	assert.Regexp(t, regexp.MustCompile(".*[!@#$%&*]{0}.*"), p, "Expecting 0 special")

}

func TestGeneratedD(t *testing.T) {
	id1 := GenerateId("ID", 10)
	id2 := GenerateId("ID", 10)

	assert.Equal(t, len(id1), 10)
	assert.Equal(t, len(id2), 10)
	assert.NotEqual(t, id1, id2)
}

func TestHashId(t *testing.T) {
	id := HashId("id", "1234124", "asdfasfdas", "gfsdgfsdg")
	assert.NotNil(t, id)
}
