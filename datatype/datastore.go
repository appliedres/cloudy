package datatype

import (
	"context"

	"github.com/appliedres/cloudy/datastore"
)

type JsonDataStore[T any] interface {

	// Open will open the datastore for usage. This should
	// only be done once per datastore
	Open(ctx context.Context, config any) error

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
	Query(ctx context.Context, query *datastore.SimpleQuery) ([]*T, error)

	QueryAndUpdate(ctx context.Context, query *datastore.SimpleQuery, updater func(ctx context.Context, items []*T) ([]*T, error)) ([]*T, error)

	// DeleteAll(ctx context.Context, key []string) error
	// SaveAll(ctx context.Context, item []*T, key []string) error

	// QueryAsMap(ctx context.Context, query *datastore.SimpleQuery) ([]map[string]any, error)
	// QueryTable(ctx context.Context, query *datastore.SimpleQuery) ([][]any, error)
}

type BulkJsonDataStore[T any] interface {
	JsonDataStore[T]

	DeleteAll(ctx context.Context, key []string) error

	SaveAll(ctx context.Context, item []*T, key []string) error

	DeleteQuery(ctx context.Context, query *datastore.SimpleQuery) ([]string, error)
}

type AdvQueryJsonDatastore[T any] interface {
	JsonDataStore[T]

	QueryAsMap(ctx context.Context, query *datastore.SimpleQuery) ([]map[string]any, error)

	QueryTable(ctx context.Context, query *datastore.SimpleQuery) ([][]any, error)

	// DeleteHierarcy(ctx context.Context, parentField string, parentKey string) error
}
