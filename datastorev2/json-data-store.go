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

type DatastoreEventHandler interface {
	OnOpen()
	OnClose()
	OnSave(item interface{}, key string)
	OnDelete(key string)
	OnGet(item interface{}, key string)
	OnGetAll(items []interface{})
}

type ObservableDatastore struct {
	ds        JsonDataStore
	listeners []DatastoreEventHandler
}

func (ods *ObservableDatastore) AddListener(l DatastoreEventHandler) {
	ods.listeners = append(ods.listeners, l)
}

func (ods *ObservableDatastore) RemoveListener(l DatastoreEventHandler) {
	for i := len(ods.listeners) - 1; i >= 0; i-- {
		if ods.listeners[i] == l {
			ods.listeners = append(ods.listeners[:i], ods.listeners[i+1:]...)
		}
	}
}

// Open will open the datastore for usage. This should
// only be done once per datastore
func (ods *ObservableDatastore) Open(ctx context.Context, config interface{}) error {
	err := ods.ds.Open(ctx, config)
	for _, l := range ods.listeners {
		l.OnOpen()
	}
	return err
}

// Close should be called to cleannly close the datastore
func (ods *ObservableDatastore) Close(ctx context.Context) error {
	err := ods.ds.Close(ctx)
	for _, l := range ods.listeners {
		l.OnClose()
	}
	return err
}

// Save stores an item in the datastore. There is no difference
// between an insert and an update.
func (ods *ObservableDatastore) Save(ctx context.Context, item interface{}, key string) error {
	err := ods.ds.Save(ctx, item, key)
	for _, l := range ods.listeners {
		l.OnSave(item, key)
	}
	return err
}

// Get retrieves an item by it's unique id
func (ods *ObservableDatastore) Get(ctx context.Context, key string) (interface{}, error) {
	item, err := ods.ds.Get(ctx, key)
	for _, l := range ods.listeners {
		l.OnGet(item, key)
	}
	return item, err
}

// Gets all the items in the store.
func (ods *ObservableDatastore) GetAll(ctx context.Context, page *models.Page) ([]interface{}, *models.Page, error) {
	rtn, page, err := ods.ds.GetAll(ctx, page)
	for _, l := range ods.listeners {
		l.OnGetAll(rtn)
	}
	return rtn, page, err
}

// Deletes an item
func (ods *ObservableDatastore) Delete(ctx context.Context, key string) error {
	err := ods.ds.Delete(ctx, key)
	for _, l := range ods.listeners {
		l.OnDelete(key)
	}
	return err
}

// Checks to see if a key exists
func (ods *ObservableDatastore) Exists(ctx context.Context, key string) (bool, error) {
	return ods.ds.Exists(ctx, key)
}

// Sends a simple Query
func (ods *ObservableDatastore) Query(ctx context.Context, query *SimpleQuery) ([]interface{}, error) {
	return ods.ds.Query(ctx, query)
}

// interface{}
// Hook for the datastore to call when the table is created
func (ods *ObservableDatastore) SetOnCreate(fn OnCreateFn) {
	ods.ds.SetOnCreate(fn)
}
