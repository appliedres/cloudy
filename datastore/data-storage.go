package datastore

import (
	"context"
	"errors"
	"io"

	"github.com/appliedres/cloudy"
)

var ErrInvalidConfiguration = errors.New("invalid configuration object")

var BinaryDataStoreProviders = cloudy.NewProviderRegistry[BinaryDataStore]()
var JsonDataStoreProviders = cloudy.NewProviderRegistry[UntypedJsonDataStore]()
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
