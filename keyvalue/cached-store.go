package keyvalue

import (
	"sync"
	"time"

	"github.com/go-openapi/strfmt"
)

var _ KeyValueStore = (*CachedKeyValueStore)(nil)
var _ WritableKeyValueStore = (*CachedWritableKeyValueStore)(nil)
var _ SecureKeyValueStore = (*CachedSecureKeyValueStore)(nil)
var _ WritableSecureKeyValueStore = (*CachedWritableSecureKeyValueStore)(nil)

type CachedKeyValueStore struct {
	lastReadError error
	lock          sync.RWMutex
	Source        KeyValueStore
	ttl           time.Duration
	expiration    time.Time
	cachedMap     map[string]string
}

func NewCachedKVStore(source KeyValueStore) *CachedKeyValueStore {
	return &CachedKeyValueStore{
		ttl:    5 * time.Minute,
		Source: source,
	}
}

func (ce *CachedKeyValueStore) checkExpired() {
	ce.lock.RLock()
	defer ce.lock.RUnlock()

	if time.Now().After(ce.expiration) {
		ce.lock.Lock()
		defer ce.lock.Unlock()

		items, err := ce.Source.GetAll()
		ce.lastReadError = err
		if err == nil {
			ce.cachedMap = items
			ce.expiration = time.Now().Add(ce.ttl)
		}
	}
}

func (ce *CachedKeyValueStore) Get(name string) (string, error) {
	ce.checkExpired()

	ce.lock.RLock()
	defer ce.lock.RUnlock()
	normKey := NormalizeKey(name)
	return ce.cachedMap[normKey], ce.lastReadError
}

func (ce *CachedKeyValueStore) GetAll() (map[string]string, error) {
	ce.checkExpired()

	ce.lock.RLock()
	defer ce.lock.RUnlock()

	return ce.cachedMap, ce.lastReadError
}

type CachedWritableKeyValueStore struct {
	*CachedKeyValueStore
	writableSource WritableKeyValueStore
}

func NewWritableCachedKVStore(source WritableKeyValueStore) *CachedWritableKeyValueStore {
	return &CachedWritableKeyValueStore{
		CachedKeyValueStore: &CachedKeyValueStore{
			ttl:    5 * time.Minute,
			Source: source,
		},
		writableSource: source,
	}
}

func (ce *CachedWritableKeyValueStore) Set(name string, value string) error {
	ce.checkExpired()

	ce.lock.Lock()
	defer ce.lock.Unlock()

	normKey := NormalizeKey(name)

	err := ce.writableSource.Set(normKey, value)
	if err != nil {
		return err
	}
	ce.cachedMap[normKey] = value
	return nil
}

func (ce *CachedWritableKeyValueStore) Delete(name string) error {
	ce.checkExpired()

	ce.lock.Lock()
	defer ce.lock.Unlock()

	normKey := NormalizeKey(name)
	err := ce.writableSource.Delete(normKey)
	if err != nil {
		return err
	}
	delete(ce.cachedMap, normKey)
	return nil
}

func (ce *CachedWritableKeyValueStore) SetMany(many map[string]string) error {
	ce.checkExpired()

	ce.lock.Lock()
	defer ce.lock.Unlock()

	err := ce.writableSource.SetMany(many)
	if err != nil {
		return err
	}
	for k, v := range many {
		normKey := NormalizeKey(k)
		ce.cachedMap[normKey] = v
	}
	return nil
}

type CachedSecureKeyValueStore struct {
	*CachedKeyValueStore
	secureStore SecureKeyValueStore
}

func NewCachedSecureKeyValueStoree(source SecureKeyValueStore) *CachedSecureKeyValueStore {
	return &CachedSecureKeyValueStore{
		CachedKeyValueStore: &CachedKeyValueStore{
			ttl:    5 * time.Minute,
			Source: source,
		},
		secureStore: source,
	}
}

func (ce *CachedSecureKeyValueStore) GetSecure(name string) (strfmt.Password, error) {
	ce.checkExpired()

	ce.lock.RLock()
	defer ce.lock.RUnlock()

	normKey := NormalizeKey(name)
	cached := ce.cachedMap[normKey]
	if cached != "" {
		return strfmt.Password(cached), ce.lastReadError
	}

	return ce.GetSecure(name)
}

type CachedWritableSecureKeyValueStore struct {
	*CachedSecureKeyValueStore
	writableStore WritableSecureKeyValueStore
}

func NewCachedWritableSecureKeyValueStore(source WritableSecureKeyValueStore) *CachedWritableSecureKeyValueStore {
	return &CachedWritableSecureKeyValueStore{
		CachedSecureKeyValueStore: &CachedSecureKeyValueStore{
			CachedKeyValueStore: &CachedKeyValueStore{
				ttl:    5 * time.Minute,
				Source: source,
			},
			secureStore: source,
		},
		writableStore: source,
	}
}

func (ce *CachedWritableSecureKeyValueStore) Delete(name string) error {
	return ce.writableStore.Delete(name)
}

func (ce *CachedWritableSecureKeyValueStore) Set(name string, value string) error {
	return ce.writableStore.Set(name, value)
}

func (ce *CachedWritableSecureKeyValueStore) SetMany(many map[string]string) error {
	return ce.writableStore.SetMany(many)
}

func (ce *CachedWritableSecureKeyValueStore) SetSecure(name string, value strfmt.Password) error {
	ce.checkExpired()

	ce.lock.Lock()
	defer ce.lock.Unlock()

	normKey := NormalizeKey(name)
	err := ce.writableStore.SetSecure(normKey, value)
	if err != nil {
		return err
	}
	ce.cachedMap[normKey] = value.String()
	return nil
}
