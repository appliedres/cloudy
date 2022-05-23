package datastore

import (
	"context"
	"io"
	"io/ioutil"
)

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
