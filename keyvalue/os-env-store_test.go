package keyvalue

import (
	"os"
	"testing"

	"github.com/appliedres/cloudy"
)

func TestOsEnvKeyValueStore(t *testing.T) {
	store := NewBasicOsEnvKeyValueStore()

	for k, v := range TestStoreNormalForms {
		key := cloudy.ToEnvName(k, "")
		os.Setenv(key, v)
	}

	TestKVStore(t, store, TestStoreNormalForms)
}
