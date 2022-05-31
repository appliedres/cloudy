package cloudy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToGabs(t *testing.T) {
	element := struct {
		Name      string
		Branch    string
		Language  string
		Particles int
	}{
		Name:      "Pikachu",
		Branch:    "ECE",
		Language:  "C++",
		Particles: 498,
	}

	result, err := ToGabs(element)
	assert.Nil(t, err)
	assert.NotNil(t, result.S("Name"))

	element.Name = "BADD"

	err = FromGabs(result, &element)
	assert.Nil(t, err)
	assert.Equal(t, element.Name, "Pikachu")
}
