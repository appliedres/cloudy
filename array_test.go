package cloudy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayIncludes(t *testing.T) {
	arr := []string{"1", "2", "3"}
	yes := ArrayIncludes(arr, "2")
	assert.True(t, yes, "Should be yes")

	no := ArrayIncludes(arr, "4")
	assert.False(t, no, "Should be no")
}

func TestArrayDisjoint(t *testing.T) {
	arr := []string{"1", "2", "3"}
	other := []string{"1", "2", "4"}
	result := ArrayDisjoint(arr, other)
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0], "4")
}

func TestArrayRemoveAll(t *testing.T) {
	arr := []string{"1", "2", "3"}
	result := ArrayRemoveAll(arr, func(item string) bool {
		return item != "1"
	})
	assert.Equal(t, len(result), 1)
	assert.Equal(t, result[0], "1")
}

func TestArrayIncludesAll(t *testing.T) {
	arr := []string{"1", "2", "3"}
	subset := []string{"1", "2"}
	notsubset := []string{"1", "2", "4"}

	yes := ArrayIncludesAll(arr, subset)
	no := ArrayIncludesAll(arr, notsubset)

	assert.Equal(t, len(yes), 2)
	assert.Equal(t, len(no), 2)

}

func TestArrayFindIndex(t *testing.T) {
	arr := []string{"1", "2", "3"}

	result := ArrayFindIndex(arr, func(item string) bool {
		return item != "1"
	})

	assert.Equal(t, result, 1)
}

func TestArrayFirst(t *testing.T) {
	arr := []string{"1", "2", "3"}

	result, ok := ArrayFirst(arr, func(item string) bool {
		return item != "1"
	})

	assert.Equal(t, result, "2")
	assert.True(t, ok)
}

func TestArrayRemove(t *testing.T) {
	arr := []string{"1", "2", "3"}

	result, ok := ArrayRemove(arr, func(item string) bool {
		return item == "2"
	})

	assert.Equal(t, len(result), 2)
	assert.True(t, ok)
	assert.Equal(t, result[1], "3")
}
