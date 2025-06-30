package datastore

import (
	"context"
	"errors"
	"time"
)

var _ JsonDataStore[any] = (*InMemoryTypedStore[any])(nil)

type DatastoreRecordTyped[T any] struct {
	RowMetadata
	Data *T
}

type InMemoryTypedStore[T any] struct {
	records map[string]*DatastoreRecordTyped[T]
	fn      func(ctx context.Context, ds JsonDataStore[T]) error
}

func NewInMemoryTypedStore[T any]() *InMemoryTypedStore[T] {
	return new(InMemoryTypedStore[T])
}

func (mem *InMemoryTypedStore[T]) Open(ctx context.Context, config interface{}) error {
	mem.records = make(map[string]*DatastoreRecordTyped[T])
	if mem.fn != nil {
		err := mem.fn(ctx, mem)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mem *InMemoryTypedStore[T]) Close(ctx context.Context) error {
	mem.records = nil
	return nil
}

func (mem *InMemoryTypedStore[T]) Save(ctx context.Context, data *T, key string) error {
	rec := &DatastoreRecordTyped[T]{
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

func (mem *InMemoryTypedStore[T]) GetMetadata(ctx context.Context, key ...string) ([]*RowMetadata, error) {
	rtn := make([]*RowMetadata, len(key))
	for _, k := range key {
		rec := mem.records[k]
		if rec != nil {
			rtn = append(rtn, &rec.RowMetadata)
		}
	}
	return rtn, nil
}

func (mem *InMemoryTypedStore[T]) Get(ctx context.Context, key string) (*T, error) {
	found := mem.records[key]
	if found != nil {
		return found.Data, nil
	}
	return nil, nil
}

func (mem *InMemoryTypedStore[T]) Delete(ctx context.Context, key string) error {
	delete(mem.records, key)
	return nil
}
func (mem *InMemoryTypedStore[T]) Exists(ctx context.Context, key string) (bool, error) {
	_, found := mem.records[key]
	return found, nil
}

func (mem *InMemoryTypedStore[T]) GetAll(ctx context.Context) ([]*T, error) {
	rtn := make([]*T, len(mem.records))
	i := 0
	for _, v := range mem.records {
		rtn[i] = v.Data
		i++
	}
	return rtn, nil
}

func (mem *InMemoryTypedStore[T]) OnCreate(fn func(ctx context.Context, ds JsonDataStore[T]) error) {
	mem.fn = fn
}

func (mem *InMemoryTypedStore[T]) Query(ctx context.Context, query *SimpleQuery) ([]*T, error) {
	return nil, errors.New("not implemented")
}

func (mem *InMemoryTypedStore[T]) QueryAndUpdate(ctx context.Context, query *SimpleQuery, updater func(ctx context.Context, items []*T) ([]*T, error)) ([]*T, error) {
	return nil, errors.New("not implemented")
}

func (mem *InMemoryTypedStore[T]) SaveAll(ctx context.Context, items []*T, key []string) error {
	for i, key := range key {
		mem.Save(ctx, items[i], key)
	}
	return nil
}

func (mem *InMemoryTypedStore[T]) DeleteAll(ctx context.Context, key []string) error {
	for _, k := range key {
		_ = mem.Delete(ctx, k)
	}
	return nil
}

func (mem *InMemoryTypedStore[T]) QueryAsMap(ctx context.Context, query *SimpleQuery) ([]map[string]any, error) {
	return nil, errors.New("not implemented")
}

func (mem *InMemoryTypedStore[T]) QueryTable(ctx context.Context, query *SimpleQuery) ([][]interface{}, error) {
	return nil, errors.New("not implemented")
}
