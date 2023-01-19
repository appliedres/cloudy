package cloudy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapKeyStr(t *testing.T) {
	var firstlast = map[string]interface{}{
		"Bauer": "John",
		"Doe":   "Jane", // last comma is a must
	}

	k1, ok1 := MapKeyStr(firstlast, "bauer", false)
	assert.Equal(t, k1, "")
	assert.False(t, ok1)

	k2, ok2 := MapKeyStr(firstlast, "bauer", true)
	assert.Equal(t, k2, "John")
	assert.True(t, ok2)

	k3, ok3 := MapKeyStr(firstlast, "Bauer", false)
	assert.Equal(t, k3, "John")
	assert.True(t, ok3)

	k4, ok4 := MapKeyStr(firstlast, "Bauer", true)
	assert.Equal(t, k4, "John")
	assert.True(t, ok4)

}

func TestMapKey(t *testing.T) {
	var firstlast = map[string]string{
		"Bauer": "John",
		"Doe":   "Jane", // last comma is a must
	}

	k1, ok1 := MapKey(firstlast, "bauer", false)
	assert.Equal(t, k1, "")
	assert.False(t, ok1)

	k2, ok2 := MapKey(firstlast, "bauer", true)
	assert.Equal(t, k2, "John")
	assert.True(t, ok2)

	k3, ok3 := MapKey(firstlast, "Bauer", false)
	assert.Equal(t, k3, "John")
	assert.True(t, ok3)

	k4, ok4 := MapKey(firstlast, "Bauer", true)
	assert.Equal(t, k4, "John")
	assert.True(t, ok4)

}

func TestStringP(t *testing.T) {
	test := "TEST"
	testp := StringP(test)
	assert.Equal(t, test, *testp)
}

func TestBoolP(t *testing.T) {
	test := true
	testp := BoolP(test)
	assert.Equal(t, test, *testp)
}

func TestStringFromPWithDefault(t *testing.T) {
	test := "TEST"
	testp := StringFromPWithDefault(&test, "")
	assert.Equal(t, test, testp)
}

func TestBoolFromP(t *testing.T) {
	test := true
	testp := BoolFromP(&test)
	assert.Equal(t, test, testp)
}
