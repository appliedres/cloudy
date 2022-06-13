package datastore

import (
	"context"
	"io"
	"io/ioutil"
)

const InMemoryinaryStoreID = "memory"

func init() {
	BinaryDataStoreProviders.Register(InMemoryinaryStoreID, &InMemoryStoreFactory{})
}

type InMemoryStoreFactory struct{}

func (f *InMemoryStoreFactory) Create(cfg interface{}) (BinaryDataStore, error) {
	return NewInMemoryStore(), nil
}

func (f *InMemoryStoreFactory) ToConfig(config map[string]interface{}) (interface{}, error) {
	return nil, nil
}

type InMemoryStore struct {
	items map[string][]byte
}

func NewInMemoryStore() *InMemoryStore {
	return new(InMemoryStore)
}

func (mem *InMemoryStore) Open(ctx context.Context, config interface{}) error {
	mem.items = make(map[string][]byte)
	return nil
}

func (mem *InMemoryStore) Close(ctx context.Context) error {
	mem.items = nil
	return nil
}

func (mem *InMemoryStore) Save(ctx context.Context, data []byte, key string) error {
	mem.items[key] = data
	return nil
}
func (mem *InMemoryStore) SaveStream(ctx context.Context, data io.ReadCloser, key string) (int64, error) {
	out, err := ioutil.ReadAll(data)
	if err != nil {
		return 0, err
	}
	return int64(len(out)), mem.Save(ctx, out, key)
}

func (mem *InMemoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	return mem.items[key], nil
}

func (mem *InMemoryStore) Delete(ctx context.Context, key string) error {
	delete(mem.items, key)
	return nil
}
func (mem *InMemoryStore) Exists(ctx context.Context, key string) (bool, error) {
	_, found := mem.items[key]
	return found, nil
}
