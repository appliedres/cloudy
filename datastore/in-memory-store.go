package datastore

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/appliedres/cloudy"
)

const InMemoryinaryStoreID = "memory"

var _ UntypedJsonDataStore = (*InMemoryStore)(nil)

type DatastoreRecord struct {
	RowMetadata
	Data []byte `json:"data"`
}

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
	records map[string]*DatastoreRecord
	fn      OnCreateDS
}

func NewInMemoryStore() *InMemoryStore {
	return new(InMemoryStore)
}

func (mem *InMemoryStore) Open(ctx context.Context, config interface{}) error {
	mem.records = make(map[string]*DatastoreRecord)
	if mem.fn != nil {
		err := mem.fn(ctx, mem)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *InMemoryStore) Close(ctx context.Context) error {
	mem.records = nil
	return nil
}

func (mem *InMemoryStore) Save(ctx context.Context, data []byte, key string) error {
	rec := &DatastoreRecord{
		Data: data,
		RowMetadata: RowMetadata{
			Key:         key,
			Version:     1,
			DateCreated: time.Now(),
			LastUpdated: time.Now(),
		},
	}
	mem.records[key] = rec
	return nil
}
func (mem *InMemoryStore) SaveStream(ctx context.Context, data io.ReadCloser, key string) (int64, error) {
	out, err := io.ReadAll(data)
	if err != nil {
		return 0, err
	}
	return int64(len(out)), mem.Save(ctx, out, key)
}

func (mem *InMemoryStore) GetMetadata(ctx context.Context, key ...string) ([]*RowMetadata, error) {
	rtn := make([]*RowMetadata, len(key))
	for _, k := range key {
		rec := mem.records[k]
		if rec != nil {
			rtn = append(rtn, &rec.RowMetadata)
		}
	}
	return rtn, nil
}

func (mem *InMemoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	found := mem.records[key]
	if found != nil {
		return found.Data, nil
	}
	return nil, nil
}

func (mem *InMemoryStore) Delete(ctx context.Context, key string) error {
	delete(mem.records, key)
	return nil
}
func (mem *InMemoryStore) Exists(ctx context.Context, key string) (bool, error) {
	_, found := mem.records[key]
	return found, nil
}

func (mem *InMemoryStore) GetAll(ctx context.Context) ([][]byte, error) {
	rtn := make([][]byte, len(mem.records))
	i := 0
	for _, v := range mem.records {
		rtn[i] = v.Data
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
		mem.Save(ctx, items[i], key)
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
