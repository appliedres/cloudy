package datastore

import (
	"context"
	"encoding/json"

	"github.com/appliedres/cloudy"
)

// JsonDataStore stores data structures as JSON. The type argument can
// be any struct that the json package can marshal. The type argument
// must NOT be a pointer type. The config information is based on
// the driver being used. This will typically contain connection
// information and the necessary information for generating the table
// or index.
type JsonDataStore[T any] interface {

	// Open will open the datastore for usage. This should
	// only be done once per datastore
	Open(ctx context.Context, config interface{}) error

	// Close should be called to cleannly close the datastore
	Close(ctx context.Context) error

	// Save stores an item in the datastore. There is no difference
	// between an insert and an update.
	Save(ctx context.Context, item *T, key string) error

	// Get retrieves an item by it's unique id
	Get(ctx context.Context, key string) (*T, error)

	// Gets all the items in the store.
	GetAll(ctx context.Context) ([]*T, error)

	// Deletes an item
	Delete(ctx context.Context, key string) error

	// Checks to see if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Sends a simple Query
	Query(ctx context.Context, query *SimpleQuery) ([]*T, error)
}

type NativeQuerable[T any] interface {
	NativeQuery(ctx context.Context, query interface{}) (interface{}, error)
}

type JsonDataStoreAdapter[T any] struct {
	DS    BinaryDataStore
	Model T
}

func ToJsonDataStore[T any](ds BinaryDataStore) *JsonDataStoreAdapter[T] {
	return &JsonDataStoreAdapter[T]{
		DS: ds,
	}
}

func (j *JsonDataStoreAdapter[T]) Save(ctx context.Context, item T, key string) error {
	data, err := json.MarshalIndent(item, "", "   ")
	if err != nil {
		return err
	}

	return j.DS.Save(ctx, data, key)
}

func (j *JsonDataStoreAdapter[T]) Get(ctx context.Context, key string) (T, error) {
	var zero T
	data, err := j.DS.Get(ctx, key)
	if err != nil {
		return zero, err
	}
	if data == nil {
		return zero, nil
	}

	v, err := cloudy.NewT[T]()
	if err != nil {
		return zero, err
	}
	err = json.Unmarshal(data, &v)
	if err != nil {
		return zero, err
	}
	return v, nil
}

// var jsonstores = make(map[string]JsonDataStoreFactory[any])

// type JsonDataStoreFactory[M any] interface {
// 	Create(cfg interface{}) (JsonDataStore[M], error)
// 	ToConfig(cfgMap map[string]interface{}) (interface{}, error)
// }

// func RegisterJsonDatastore(providerName string, factory JsonDataStoreFactory[any]) {
// 	jsonstores[providerName] = factory
// }

// func NewJsonDatastore[M any](providerName string, cfg interface{}) (JsonDataStore[M], error) {
// 	factory := jsonstores[providerName]
// 	ds, err := factory.Create(cfg)
// 	return ds.(JsonDataStore[M]), err
// }

func NewTypedStore[T any](store UntypedJsonDataStore) JsonDataStore[T] {
	return &TypedJsonStore[T]{ds: store}
}

type TypedJsonStore[T any] struct {
	ds UntypedJsonDataStore
}

// Open will open the datastore for usage. This should
// only be done once per datastore
func (ts *TypedJsonStore[T]) Open(ctx context.Context, config interface{}) error {
	return ts.ds.Open(ctx, config)
}

// Close should be called to cleannly close the datastore
func (ts *TypedJsonStore[T]) Close(ctx context.Context) error {
	return ts.ds.Close(ctx)
}

// Save stores an item in the datastore. There is no difference
// between an insert and an update.
func (ts *TypedJsonStore[T]) Save(ctx context.Context, item *T, key string) error {
	return ts.ds.Save(ctx, item, key)
}

// Get retrieves an item by it's unique id
func (ts *TypedJsonStore[T]) Get(ctx context.Context, key string) (*T, error) {
	var zero *T
	item, err := ts.ds.Get(ctx, key)
	if err != nil {
		return zero, err
	}
	return item.(*T), err
}

// Gets all the items in the store.
func (ts *TypedJsonStore[T]) GetAll(ctx context.Context) ([]*T, error) {
	results, err := ts.ds.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	rtn := make([]*T, len(results))
	for i, v := range results {
		rtn[i] = v.(*T)
	}
	return rtn, err
}

// Deletes an item
func (ts *TypedJsonStore[T]) Delete(ctx context.Context, key string) error {
	return ts.ds.Delete(ctx, key)
}

// Checks to see if a key exists
func (ts *TypedJsonStore[T]) Exists(ctx context.Context, key string) (bool, error) {
	return ts.ds.Exists(ctx, key)
}

// Sends a simple Query
func (ts *TypedJsonStore[T]) Query(ctx context.Context, query *SimpleQuery) ([]*T, error) {
	results, err := ts.ds.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	rtn := make([]*T, len(results))
	for i, v := range results {
		rtn[i] = v.(*T)
	}
	return rtn, err
}

type UntypedJsonDataStore interface {

	// Open will open the datastore for usage. This should
	// only be done once per datastore
	Open(ctx context.Context, config interface{}) error

	// Close should be called to cleannly close the datastore
	Close(ctx context.Context) error

	// Save stores an item in the datastore. There is no difference
	// between an insert and an update.
	Save(ctx context.Context, item interface{}, key string) error

	// Get retrieves an item by it's unique id
	Get(ctx context.Context, key string) (interface{}, error)

	// Gets all the items in the store.
	GetAll(ctx context.Context) ([]interface{}, error)

	// Deletes an item
	Delete(ctx context.Context, key string) error

	// Checks to see if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Sends a simple Query
	Query(ctx context.Context, query *SimpleQuery) ([]interface{}, error)
}

var jsonstores = make(map[string]JsonDataStoreFactory)

type JsonDataStoreFactory interface {
	Create(cfg interface{}) (UntypedJsonDataStore, error)
	ToConfig(cfgMap map[string]interface{}) (interface{}, error)
}

func RegisterJsonDatastore(providerName string, factory JsonDataStoreFactory) {
	jsonstores[providerName] = factory
}

func NewJsonDatastore[M any](providerName string, cfg interface{}) (JsonDataStore[M], error) {
	factory := jsonstores[providerName]
	ds, err := factory.Create(cfg)
	if err != nil {
		return nil, err
	}
	typed := NewTypedStore[M](ds)
	return typed, nil
}
