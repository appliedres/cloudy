package testutil

import (
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/storage"
	"github.com/stretchr/testify/assert"
)

func TestFileShareStorageManager(t *testing.T, mgr storage.FileStorageManager) {
	ctx := cloudy.StartContext()

	meta := map[string]string{
		"TEST_TAG": "TEST_TAG_VALUE",
	}

	// Exists
	exists, err := mgr.Exists(ctx, "file-storage-test")
	assert.Nil(t, err)

	if exists {
		err = mgr.Delete(ctx, "file-storage-test")
		assert.Nil(t, err)
	}

	// Create
	osm, err := mgr.Create(ctx, "file-storage-test", meta)
	assert.Nil(t, err)
	assert.NotNil(t, osm)

	// List
	all, err := mgr.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(all))

	// Get
	osm2, err := mgr.Get(ctx, "file-storage-test")
	assert.Nil(t, err)
	assert.NotNil(t, osm2)

	// Delete
	err = mgr.Delete(ctx, "file-storage-test")
	assert.Nil(t, err)

}
