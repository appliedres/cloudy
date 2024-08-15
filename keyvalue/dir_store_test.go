package keyvalue

import (
	"os"
	"testing"
)

func TestDirStore(t *testing.T) {
	store := NewDirectoryKeyValueStoreWFilename("/tmp/arktestdir")
	TestWritableKVStore(t, store, TestStoreNormalForms)
	os.Remove("/tmp/arktestdir")
}
