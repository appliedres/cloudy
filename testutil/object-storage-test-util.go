package testutil

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"testing"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/storage"
	"github.com/stretchr/testify/assert"
)

func TestObjectStorageManager(t *testing.T, mgr storage.ObjectStorageManager) {
	ctx := cloudy.StartContext()

	meta := map[string]string{
		"TEST_TAG": "TEST_TAG_VALUE",
	}

	// Exists
	exists, err := mgr.Exists(ctx, "arkloud-object-storage-test")
	assert.Nil(t, err)

	if exists {
		err = mgr.Delete(ctx, "arkloud-object-storage-test")
		assert.Nil(t, err)
	}

	// Create
	osm, err := mgr.Create(ctx, "arkloud-object-storage-test", false, meta)
	assert.Nil(t, err)
	assert.NotNil(t, osm)

	// List
	all, err := mgr.List(ctx)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(all), 1, "Expected there to be at least one item")

	// Get
	osm2, err := mgr.Get(ctx, "arkloud-object-storage-test")
	assert.Nil(t, err)
	assert.NotNil(t, osm2)

	// Test
	if osm2 != nil {
		TestObjectStorage(t, osm2)
	}

	// Delete
	err = mgr.Delete(ctx, "arkloud-object-storage-test")
	assert.Nil(t, err)
}

func TestObjectStorage(t *testing.T, osm storage.ObjectStorage) {
	ctx := cloudy.StartContext()

	// Make random data
	blk := make([]byte, 1024)
	_, err := rand.Read(blk)
	assert.Nil(t, err)

	buf := bytes.NewBuffer(blk)

	tags := map[string]string{
		"TEST_OBJECT_TAG": "TEST_OBJECT_TAG_VALUE",
	}
	err = osm.Upload(ctx, "test-key", buf, tags)
	assert.Nil(t, err)

	exists, err := osm.Exists(ctx, "test-key")
	assert.Nil(t, err)
	assert.True(t, exists)

	downloaded, err := osm.Download(ctx, "test-key")
	assert.Nil(t, err)
	assert.NotNil(t, downloaded)
	data, err := ioutil.ReadAll(downloaded)
	assert.Nil(t, err)
	assert.Equal(t, data, blk)

	items, prefixes, err := osm.List(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, len(items), 1)
	assert.Equal(t, len(prefixes), 0)

	err = osm.Delete(ctx, "test-key")
	assert.Nil(t, err)

}
