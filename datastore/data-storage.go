package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"io"

	"github.com/appliedres/cloudy"
)

var InvalidConfiguration = errors.New("invalid configuration object")

var BinaryDataStoreProviders = cloudy.NewProviderRegistry[BinaryDataStore]()
var JsonDataStoreProviders = cloudy.NewProviderRegistry[JsonDataStore[any]]()
var IndexerProviders = cloudy.NewProviderRegistry[Indexer[any]]()

type BinaryDataStore interface {
	Open(ctx context.Context, config interface{}) error
	Close(ctx context.Context) error

	Save(ctx context.Context, data []byte, key string) error
	SaveStream(ctx context.Context, data io.ReadCloser, key string) (int64, error)
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type Indexer[T any] interface {
	Open(ctx context.Context, config interface{}) error
	Close(ctx context.Context) error

	Index(ctx context.Context, id string, data []byte) error
	Remove(ctx context.Context, id string) error
	Search(ctx context.Context, query interface{}) ([]T, error)
}

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
	err = json.Unmarshal(data, &v)
	if err != nil {
		return zero, err
	}
	return v, nil
}
