package datastore

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BinaryDataStoreTest(t *testing.T, ctx context.Context, ds BinaryDataStore) {
	// Random Byte Data
	data := make([]byte, 4000)
	rand.Read(data) // #nosec G404 For Testing only

	id := "my-test-data"

	fmt.Println("Saving")
	err := ds.Save(ctx, data, id)
	assert.Nil(t, err, "Should not get an error saving to the database")

	// Exists
	fmt.Println("Checking Existence")

	exists, err := ds.Exists(ctx, id)
	assert.Nil(t, err, "Should not get an error")
	assert.True(t, exists, "Should exist")

	// Retrieve
	fmt.Println("Getting")
	data2, err := ds.Get(ctx, id)
	assert.Nil(t, err, "Should not get an error saving to the database")
	assert.NotNil(t, data2, "Item should be found")
	assert.Equal(t, data, data, "IDs should be equal")

	// Delete
	fmt.Println("Deleteing")

	ds.Delete(ctx, id)
	assert.Nil(t, err, "Should not get an error")

	// Exists
	fmt.Println("Checking Existence")

	exists2, err := ds.Exists(ctx, id)
	assert.Nil(t, err, "Should not get an error")
	assert.False(t, exists2, "Should NOT exist")

	// Retrieve
	fmt.Println("Getting Missing")

	item3, err := ds.Get(ctx, id)
	assert.Nil(t, err, "Should not get an error retrieving")
	assert.Nil(t, item3, "Should not get an error ")

	fmt.Println("Done")
}
