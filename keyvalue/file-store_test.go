package keyvalue

import (
	"os"
	"testing"
)

func TestFileStore(t *testing.T) {
	storeEnv := NewFileKeyValueStoreWFilename("test.env")
	TestWritableKVStore(t, storeEnv, TestStoreNormalForms)
	os.Remove("test.env")

	storeJson := NewFileKeyValueStoreWFilename("test.json")
	TestWritableKVStore(t, storeJson, TestStoreNormalForms)
	os.Remove("test.json")
}
