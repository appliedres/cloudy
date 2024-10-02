package datastore

import (
	"context"
	"errors"
	"io"

	"github.com/appliedres/cloudy"
)

const InMemoryinaryStoreID = "memory"

var _ UntypedJsonDataStore = (*InMemoryStore)(nil)

func init() {
	BinaryDataStoreProviders.Register(InMemoryinaryStoreID, &InMemoryStoreFactory{})
}

type InMemoryStoreFactory struct{}

func (f *InMemoryStoreFactory) Create(cfg interface{}) (BinaryDataStore, error) {
	return NewInMemoryStore(), nil
}

func (f *InMemoryStoreFactory) FromEnv(env *cloudy.Environment) (interface{}, error) {
	return nil, nil
}

type InMemoryStore struct {
	items map[string][]byte
	fn    OnCreateDS
}

func NewInMemoryStore() *InMemoryStore {
	return new(InMemoryStore)
}

func (mem *InMemoryStore) Open(ctx context.Context, config interface{}) error {
	mem.items = make(map[string][]byte)
	if mem.fn != nil {
		err := mem.fn(ctx, mem)
		if err != nil {
			return err
		}
	}
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
	out, err := io.ReadAll(data)
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

func (mem *InMemoryStore) GetAll(ctx context.Context) ([][]byte, error) {
	rtn := make([][]byte, len(mem.items))
	i := 0
	for _, v := range rtn {
		rtn[i] = v
		i++
	}
	return rtn, nil
}

func (mem *InMemoryStore) OnCreate(fn OnCreateDS) {
	mem.fn = fn
}

func (mem *InMemoryStore) Query(ctx context.Context, query *SimpleQuery) ([][]byte, error) {
	return nil, errors.New("not implemented")
}

func (mem *InMemoryStore) QueryAndUpdate(ctx context.Context, query *SimpleQuery, updater func(ctx context.Context, items [][]byte) ([][]byte, error)) ([][]byte, error) {
	return nil, errors.New("not implemented")
}

func (mem *InMemoryStore) SaveAll(ctx context.Context, items [][]byte, key []string) error {
	for i, key := range key {
		mem.items[key] = items[i]
	}
	return nil
}

func (mem *InMemoryStore) DeleteAll(ctx context.Context, key []string) error {
	for _, k := range key {
		_ = mem.Delete(ctx, k)
	}
	return nil
}

func (mem *InMemoryStore) QueryAsMap(ctx context.Context, query *SimpleQuery) ([]map[string]any, error) {
	return nil, errors.New("not implemented")
}
func (mem *InMemoryStore) QueryTable(ctx context.Context, query *SimpleQuery) ([][]interface{}, error) {
	return nil, errors.New("not implemented")
}
