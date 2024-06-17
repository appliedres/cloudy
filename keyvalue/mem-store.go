package keyvalue

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Jeffail/gabs/v2"
	"github.com/appliedres/cloudy"
	"github.com/hashicorp/go-multierror"
)

// Compile time interface checks
var _ WritableKeyValueStore = (*InMemoryKeyValueStore)(nil)

// var _ KeyValueStoreFactory = (*InMemoryKeyValueStore)(nil)

// // Factory
// type FileKeyValueStoreFactory struct{}

// func (f *FileKeyValueStoreFactory) NewConfig() interface{} {
// 	return &InMemoryKeyValueStore{}
// }

// func (f *FileKeyValueStoreFactory) New(ctx context.Context, config any) (KeyValueStore, error) {
// 	return NewFileKeyValueStore(ctx, config)
// }

// Store
type InMemoryKeyValueStore struct {
	lock sync.RWMutex
	data map[string]string
}

func NewInMemoryKeyValueStore(ctx context.Context, config interface{}) (*InMemoryKeyValueStore, error) {
	return &InMemoryKeyValueStore{
		data: make(map[string]string),
	}, nil
}

func NewDefaultInMemoryKeyValueStore() *InMemoryKeyValueStore {
	return &InMemoryKeyValueStore{
		data: make(map[string]string),
	}
}

func (fs *InMemoryKeyValueStore) GetAll() (map[string]string, error) {
	return fs.data, nil
}

func (fs *InMemoryKeyValueStore) Get(name string) (string, error) {

	fs.lock.Lock()
	defer fs.lock.Unlock()
	key := NormalizeKey(name)
	rtn := fs.data[key]
	if rtn != "" {
		return rtn, nil
	}

	rtn = fs.data[name]
	return rtn, nil
}

func (fs *InMemoryKeyValueStore) Set(name string, value string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	key := NormalizeKey(name)
	fs.data[key] = value

	return nil
}

func (fs *InMemoryKeyValueStore) SetMany(many map[string]string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	for k, value := range many {
		key := NormalizeKey(k)
		fs.data[key] = value
	}

	return nil
}

func (fs *InMemoryKeyValueStore) Delete(name string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	key := NormalizeKey(name)
	delete(fs.data, key)

	return nil
}

func (fs *InMemoryKeyValueStore) SaveAsJson(filename string) error {
	c := gabs.New()

	var merr *multierror.Error
	for k, v := range fs.data {
		_, err := c.SetP(v, k)
		merr = multierror.Append(merr, err)
	}
	err := merr.ErrorOrNil()
	if err != nil {
		return err
	}
	return os.WriteFile(filename, c.BytesIndent("", "  "), 0600)
}

func (fs *InMemoryKeyValueStore) SaveAsEnv(filename string) error {
	var sb strings.Builder
	for k, v := range fs.data {
		key := cloudy.ToEnvName(k, "")
		sb.WriteString(fmt.Sprintf("%v=%v\n", key, v))
	}
	return os.WriteFile(filename, []byte(sb.String()), 0600)
}
