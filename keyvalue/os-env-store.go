package keyvalue

import (
	"context"
	"os"
	"strings"

	"github.com/appliedres/cloudy"
)

var _ KeyValueStore = (*OsEnvKeyValueStore)(nil)

// System Environment Variables Services
type OsEnvKeyValueStore struct{}

func NewOsEnvKeyValueStore(ctx context.Context, config interface{}) *OsEnvKeyValueStore {
	return &OsEnvKeyValueStore{}
}

func (ce *OsEnvKeyValueStore) Get(name string) (string, error) {
	normKey := NormalizeKey(name)
	envKey := cloudy.ToEnvName(normKey, "")
	value, _ := os.LookupEnv(envKey)
	return value, nil
}

func NewBasicOsEnvKeyValueStore() *OsEnvKeyValueStore {
	return &OsEnvKeyValueStore{}
}

func (ce *OsEnvKeyValueStore) GetAll() (map[string]string, error) {
	m := make(map[string]string)
	kvs := os.Environ()
	for _, kv := range kvs {
		indx := strings.Index(kv, "=")
		if indx > 0 {
			key := kv[:indx]
			val := kv[indx+1:]
			nkey := NormalizeKey(key)
			m[nkey] = val
		}
	}

	return m, nil
}
