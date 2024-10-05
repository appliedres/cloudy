package datatype

import (
	"context"
	"errors"

	"github.com/appliedres/cloudy/datastore"
)

const InMemoryinaryStoreID = "memory"

var _ JsonDataStore[string] = (*InMemoryStore[string])(nil)

type InMemoryStore[T any] struct {
	items map[string]*T
}

func NewInMemoryStore[T any]() *InMemoryStore[T] {
	return &InMemoryStore[T]{
		items: make(map[string]*T),
	}
}

func (mem *InMemoryStore[T]) Open(ctx context.Context, config interface{}) error {
	mem.items = make(map[string]*T)
	return nil
}

func (mem *InMemoryStore[T]) Close(ctx context.Context) error {
	mem.items = nil
	return nil
}

func (mem *InMemoryStore[T]) Save(ctx context.Context, data *T, key string) error {
	mem.items[key] = data
	return nil
}

func (mem *InMemoryStore[T]) Get(ctx context.Context, key string) (*T, error) {
	return mem.items[key], nil
}

func (mem *InMemoryStore[T]) Delete(ctx context.Context, key string) error {
	delete(mem.items, key)
	return nil
}
func (mem *InMemoryStore[T]) Exists(ctx context.Context, key string) (bool, error) {
	_, found := mem.items[key]
	return found, nil
}

func (mem *InMemoryStore[T]) GetAll(ctx context.Context) ([]*T, error) {
	rtn := make([]*T, len(mem.items))
	i := 0
	for _, v := range rtn {
		rtn[i] = v
		i++
	}
	return rtn, nil
}

func (mem *InMemoryStore[T]) Query(ctx context.Context, query *datastore.SimpleQuery) ([]*T, error) {
	return nil, errors.New("not implemented")
}

func (mem *InMemoryStore[T]) QueryAndUpdate(ctx context.Context, query *datastore.SimpleQuery, updater func(ctx context.Context, items []*T) ([]*T, error)) ([]*T, error) {
	return nil, errors.New("not implemented")
}

func (mem *InMemoryStore[T]) SaveAll(ctx context.Context, items []*T, key []string) error {
	for i, key := range key {
		mem.items[key] = items[i]
	}
	return nil
}

func (mem *InMemoryStore[T]) DeleteAll(ctx context.Context, key []string) error {
	for _, k := range key {
		_ = mem.Delete(ctx, k)
	}
	return nil
}

func (mem *InMemoryStore[T]) QueryAsMap(ctx context.Context, query *datastore.SimpleQuery) ([]map[string]any, error) {
	return nil, errors.New("not implemented")
}
func (mem *InMemoryStore[T]) QueryTable(ctx context.Context, query *datastore.SimpleQuery) ([][]interface{}, error) {
	return nil, errors.New("not implemented")
}
