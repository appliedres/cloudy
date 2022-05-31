package datastore

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/appliedres/cloudy"
)

func TestFilesystemDS(t *testing.T) {
	ctx := cloudy.StartContext()

	// get the name for a directory but do not create it
	dir := filepath.Join(os.TempDir(), "test-fsds")
	defer cleanup(dir)

	ds := NewFilesystemStore(".test-data", dir)

	BinaryDataStoreTest(t, ctx, ds)
}

func cleanup(dir string) {
	os.RemoveAll(dir)
}
