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
	// Create
	osm, err := mgr.Create(ctx, "object-storage-test", false, meta)
	assert.Nil(t, err)
	assert.NotNil(t, osm)

	// List
	all, err := mgr.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(all))

	// Get
	osm2, err := mgr.Get(ctx, "object-storage-test")
	assert.Nil(t, err)
	assert.NotNil(t, osm2)

	// Test
	TestObjectStorage(t, osm)

	// Delete
	err = mgr.Delete(ctx, "object-storage-test")
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

	items, err := osm.List(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, len(items), 1)

	err = osm.Delete(ctx, "test-key")
	assert.Nil(t, err)

}