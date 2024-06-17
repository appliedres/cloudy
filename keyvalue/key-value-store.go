package keyvalue

import (
	"github.com/appliedres/cloudy"
	"github.com/go-openapi/strfmt"
)

type KeyValueStoreFactory = cloudy.ProviderFactory2[KeyValueStore]

// -- Iterfaces
type KeyValueStore interface {
	Get(key string) (string, error)
	GetAll() (map[string]string, error)
}

type FilteredKeyValueStore interface {
	KeyValueStore
	GetWithPrefix(prefix string) (map[string]string, error)
}

type WritableKeyValueStore interface {
	KeyValueStore
	Set(key string, value string) error
	SetMany(items map[string]string) error
	Delete(key string) error
}

type SecureKeyValueStore interface {
	KeyValueStore
	GetSecure(key string) (strfmt.Password, error)
}

type WritableSecureKeyValueStore interface {
	SecureKeyValueStore
	WritableKeyValueStore
	SetSecure(key string, value strfmt.Password) error
}
