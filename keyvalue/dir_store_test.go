package keyvalue

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirStore(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "arkloudtest*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	store := NewDirectoryKeyValueStoreWFilename(tmpDir)
	TestWritableKVStore(t, store, TestStoreNormalForms)
}

func TestDirStoreRead(t *testing.T) {

	tmpDir, err := os.MkdirTemp("", "arkloudtest*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})

	value1 := "value1"
	value2 := "value2"
	value3 := "value3"

	err = os.WriteFile(filepath.Join(tmpDir, "my-value-name"), []byte(value1), 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "my.value.name2"), []byte(value2), 0600)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "MY_ENV_VALUE"), []byte(value3), 0600)
	if err != nil {
		t.Fatal(err)
	}

	store := NewDirectoryKeyValueStoreWFilename(tmpDir)
	all, err := store.GetAll()
	if err != nil {
		t.Fatal(err)
	}

	v1 := all["my.value.name"]
	v2 := all["my.value.name2"]
	v3 := all["my.env.value"]
	assert.Equal(t, v1, value1)
	assert.Equal(t, v2, value2)
	assert.Equal(t, v3, value3)
}
