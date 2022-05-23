package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/appliedres/cloudy/datastore"
	"github.com/stretchr/testify/assert"
)

type TestItem struct {
	ID   string
	Name string
}

func JsonDataStoreTest(t *testing.T, ctx context.Context, ds datastore.JsonDataStore[TestItem]) {
	testDoc := &TestItem{
		ID:   "12345",
		Name: "TEST",
	}
	fmt.Println("Saving")
	err := ds.Save(ctx, testDoc, testDoc.ID)
	assert.Nil(t, err, "Should not get an error saving to the database")

	// Exists
	fmt.Println("Checking Existence")

	exists, err := ds.Exists(ctx, testDoc.ID)
	assert.Nil(t, err, "Should not get an error")
	assert.True(t, exists, "Should exist")

	// Retrieve
	fmt.Println("Getting")
	testDoc2, err := ds.Get(ctx, testDoc.ID)
	assert.Nil(t, err, "Should not get an error saving to the database")
	assert.NotNil(t, testDoc2, "Item should be found")
	assert.Equal(t, testDoc2.ID, testDoc.ID, "IDs should be equal")

	// Delete
	fmt.Println("Deleteing")

	ds.Delete(ctx, testDoc.ID)
	assert.Nil(t, err, "Should not get an error")

	// Exists
	fmt.Println("Checking Existence")

	exists2, err := ds.Exists(ctx, testDoc.ID)
	assert.Nil(t, err, "Should not get an error")
	assert.False(t, exists2, "Should NOT exist")

	// Retrieve
	fmt.Println("Getting Missing")

	item3, err := ds.Get(ctx, testDoc.ID)
	assert.Nil(t, err, "Should not get an error saving to the database")
	assert.Nil(t, item3, "Should not get an error ")

	fmt.Println("Done")
}
