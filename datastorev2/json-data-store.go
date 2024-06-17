package datastorev2

import (
	"context"

	"github.com/appliedres/cloudy/models"
)

// JsonDataStore stores data structures as JSON. The type argument can
// be any struct that the json package can marshal. The type argument
// must NOT be a pointer type. The config information is based on
// the driver being used. This will typically contain connection
// information and the necessary information for generating the table
// or index.
type JsonDataStore interface {

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
	GetAll(ctx context.Context, page *models.Page) ([]interface{}, *models.Page, error)

	// Deletes an item
	Delete(ctx context.Context, key string) error

	// Checks to see if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Sends a simple Query
	Query(ctx context.Context, query *SimpleQuery) ([]interface{}, error)

	// interface{}
	// Hook for the datastore to call when the table is created
	SetOnCreate(fn OnCreateFn)
}

type OnCreateFn = func(ctx context.Context, ds JsonDataStore) error

type NativeQuerable interface {
	NativeQuery(ctx context.Context, query interface{}) (interface{}, error)
}

var jsonstores = make(map[string]JsonDataStoreFactory)

type JsonDataStoreFactory interface {
	Create(cfg interface{}) (JsonDataStore, error)
	ToConfig(cfgMap map[string]interface{}) (interface{}, error)
}

func RegisterJsonDatastore(providerName string, factory JsonDataStoreFactory) {
	jsonstores[providerName] = factory
}

func NewJsonDatastore(providerName string, cfg interface{}) (JsonDataStore, error) {
	factory := jsonstores[providerName]
	ds, err := factory.Create(cfg)
	if err != nil {
		return nil, err
	}
	return ds, nil
}
