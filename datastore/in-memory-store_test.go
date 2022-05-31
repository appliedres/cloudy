package datastore

import (
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/stretchr/testify/assert"
)

func TestInMemDS(t *testing.T) {
	ctx := cloudy.StartContext()

	ds := NewInMemoryStore()
	err := ds.Open(ctx, nil)
	assert.Nil(t, err, "Should not fail to open")

	BinaryDataStoreTest(t, ctx, ds)
}
