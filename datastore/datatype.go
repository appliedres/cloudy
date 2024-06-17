package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/appliedres/cloudy"
)

var PrefixSeparator = "-"
var Datatypes *DatatypeCollection

func init() {
	Datatypes = NewDatatypeCollection()
}

type DatatypeCollection struct {
	byName   map[string]*Datatype[any]
	byPrefix map[string]*Datatype[any]
}

func NewDatatypeCollection() *DatatypeCollection {
	dtc := new(DatatypeCollection)
	dtc.byName = make(map[string]*Datatype[any])
	dtc.byPrefix = make(map[string]*Datatype[any])
	return dtc
}

func (dtc *DatatypeCollection) Add(dt *Datatype[any]) error {
	if dt.Indexer == nil {
		return errors.New("No Indexer for " + dt.Name)
	}
	if dt.DataStore == nil {
		return errors.New("No Data Store for " + dt.Name)
	}

	dtc.byName[strings.ToLower(dt.Name)] = dt
	dtc.byPrefix[strings.ToLower(dt.Prefix)] = dt

	return nil
}

func (dtc *DatatypeCollection) Initialize(ctx context.Context) error {
	for _, dt := range dtc.byName {
		err := dt.Initialize(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dtc *DatatypeCollection) Shutdown(ctx context.Context) {
	for _, dt := range dtc.byName {
		err := dt.Shutdown(ctx)
		if err != nil {
			_ = cloudy.Error(ctx, "Error shutting down %v, %v", dt.Name, err)
		}
	}
}

func (dtc *DatatypeCollection) FindByName(name string) *Datatype[any] {
	search := strings.ToLower(name)
	return dtc.byName[search]
}

func (dtc *DatatypeCollection) FindForId(id string) *Datatype[any] {
	lower := strings.ToLower(id)
	if PrefixSeparator != "" {
		parts := strings.Split(lower, PrefixSeparator)
		if len(parts) >= 2 {
			return dtc.byPrefix[parts[0]]
		} else {
			// likely invalid id
			return nil
		}
	}

	// Backup in case there is no prefix separator, not recommeded
	for prefix, dt := range dtc.byPrefix {
		if strings.HasPrefix(lower, prefix) {
			return dt
		}
	}
	return nil
}

func (dtc *DatatypeCollection) FindByPrefix(prefix string) *Datatype[any] {
	search := strings.ToLower(prefix)
	return dtc.byPrefix[search]
}

type Datatype[T any] struct {
	Name      string
	Prefix    string
	IDField   string
	GetIDFunc func(dt *Datatype[T], item *T) string
	SetIDFunc func(dt *Datatype[T], item *T, id string) string

	DataStore JsonDataStore[T]
	Indexer   Indexer[T]
	ItemType  interface{}

	BeforeSave []BeforeSaveInterceptor[T]
	AfterSave  []AfterSaveInterceptor[T]
	BeforeGet  []BeforeGetInterceptor[T]
	AfterGet   []AfterGetInterceptor[T]

	initialized bool
}

// Go is really crazy in how it handles Generics.. So in order to call all the "non typed" methods
// on each datatype we need an interface

// Handles Raw Files (including Images)
type Filetype struct {
}

type LoadInput struct {
	Ctx  context.Context
	Id   string
	User string
}

type LoadOutput struct {
	Ctx  context.Context
	Id   string
	User string
	Item interface{}
}

type BeforeSaveInterceptor[T any] interface {
	BeforeSave(ctx context.Context, dt *Datatype[T], item *T) (*T, error)
}

type AfterSaveInterceptor[T any] interface {
	AfterSave(ctx context.Context, dt *Datatype[T], item *T) (*T, error)
}

type BeforeGetInterceptor[T any] interface {
	BeforeGet(ctx context.Context, dt *Datatype[T], ID string) (string, error)
}

type AfterGetInterceptor[T any] interface {
	AfterGet(ctx context.Context, dt *Datatype[T], item *T) (*T, error)
}

func (dt *Datatype[T]) Get(ctx context.Context, ID string) (*T, error) {
	var err error
	var output *T

	err = dt.initIfNeeded(ctx)
	if err != nil {
		return output, err
	}

	// Run the interceptors, fail on error
	for _, interceptor := range dt.BeforeGet {
		ID, err = interceptor.BeforeGet(ctx, dt, ID)
		if err != nil {
			return output, err
		}
	}

	// Load the item
	output, err = dt.GetRaw(ctx, ID)
	if err != nil {
		return output, err
	}

	// Run the interceptors, fail on error
	for _, interceptor := range dt.AfterGet {
		output, err = interceptor.AfterGet(ctx, dt, output)
		if err != nil {
			return output, err
		}
	}

	// All good, now Return
	return output, nil
}

func (dt *Datatype[T]) GetAll(ctx context.Context) ([]*T, error) {
	cloudy.Info(ctx, "Datatype.GetAll %s", dt.Name)

	var err error
	var output []*T

	err = dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}

	output, err = dt.DataStore.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Run the interceptors, fail on error
	for _, interceptor := range dt.AfterGet {
		for i, item := range output {
			output[i], err = interceptor.AfterGet(ctx, dt, item)
			if err != nil {
				return nil, err
			}
		}
	}

	// All good, now Return
	return output, nil
}

// Retrieves the Raw Data from the data store and then attempts
// to unmarshall it using reflection
func (dt *Datatype[T]) GetRaw(ctx context.Context, ID string) (*T, error) {
	var v *T
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return v, err
	}

	v, err = dt.DataStore.Get(ctx, ID)
	if err != nil {
		return v, err
	}
	return v, nil
}

func (dt *Datatype[T]) Save(ctx context.Context, item *T) (*T, error) {
	var err error
	var v *T
	err = dt.initIfNeeded(ctx)
	if err != nil {
		return v, err
	}

	// Run the interceptors, fail on error
	for _, interceptor := range dt.BeforeSave {
		item, err = interceptor.BeforeSave(ctx, dt, item)
		if err != nil {
			return item, err
		}
	}

	// Run the internal Save
	item, err = dt.SaveRaw(ctx, item)
	if err != nil {
		return item, err
	}

	// Run the post save hooks
	for _, interceptor := range dt.AfterSave {
		item, err = interceptor.AfterSave(ctx, dt, item)
		if err != nil {
			return item, err
		}
	}

	return item, err
}

func (dt *Datatype[T]) SaveRaw(ctx context.Context, item *T) (*T, error) {
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}

	key := dt.GetID(ctx, item)
	if key == "" {
		return nil, cloudy.Error(ctx, "No ID Set")
	}

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return nil, err
	}

	err = dt.DataStore.Save(ctx, item, key)
	if err != nil {
		return nil, err
	}

	if dt.Indexer != nil {
		err = dt.Indexer.Index(ctx, key, data)
		if err != nil {
			return nil, err
		}
	}

	return item, nil
}

func (dt *Datatype[T]) NativeQuery(ctx context.Context, query interface{}) ([]T, error) {
	var zero []T

	err := dt.initIfNeeded(ctx)
	if err != nil {
		return zero, err
	}

	if dt.Indexer != nil {
		return dt.Indexer.Search(ctx, query)
	}

	return zero, errors.New("no Indexer configured")
}

func (dt *Datatype[T]) Query(ctx context.Context, query *SimpleQuery) ([]*T, error) {
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}

	return dt.DataStore.Query(ctx, query)
}

func (dt *Datatype[T]) GetID(ctx context.Context, item *T) string {
	if dt.GetIDFunc != nil {
		return dt.GetIDFunc(dt, item)
	}

	idField := "ID"
	if dt.IDField != "" {
		idField = dt.IDField
	}

	return cloudy.GetFieldString(item, idField)
}

func (dt *Datatype[T]) Delete(ctx context.Context, key string) error {
	return dt.DataStore.Delete(ctx, key)
}

func (dt *Datatype[T]) SetID(ctx context.Context, item *T, id string) {
	if dt.SetIDFunc != nil {
		dt.SetIDFunc(dt, item, id)
	}

	idField := "ID"
	if dt.IDField != "" {
		idField = dt.IDField
	}

	cloudy.SetFieldString(item, idField, id)
}

func (dt *Datatype[T]) GenerateID() string {
	return cloudy.GenerateId(dt.Prefix, 15)
}

func (dt *Datatype[T]) Exists(ctx context.Context, id string) (bool, error) {
	return dt.DataStore.Exists(ctx, id)
}

func (dt *Datatype[T]) initIfNeeded(ctx context.Context) error {
	// cloudy.Info(ctx, "dt.initIfNeeded %s", dt.Name)

	if dt.initialized {
		// cloudy.Info(ctx, "dt.initIfNeeded already initialized")
		return nil
	}

	if dt.DataStore != nil {
		err := dt.DataStore.Open(ctx, nil)
		if err != nil {
			return err
		}
	}

	if dt.Indexer != nil {
		err := dt.Indexer.Open(ctx, nil)
		if err != nil {
			return err
		}
	}

	// cloudy.Info(ctx, "dt.initIfNeeded complete")
	dt.initialized = true
	return nil
}

func (dt *Datatype[T]) Initialize(ctx context.Context) error {
	cloudy.Info(ctx, "dt.Initialize %s", dt.Name)

	if dt.DataStore != nil {
		err := dt.DataStore.Open(ctx, nil)
		if err != nil {
			return err
		}
	}

	if dt.Indexer != nil {
		err := dt.Indexer.Open(ctx, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dt *Datatype[T]) Shutdown(ctx context.Context) error {
	if dt.DataStore != nil {
		err := dt.DataStore.Close(ctx)
		if err != nil {
			return err
		}
	}

	if dt.Indexer != nil {
		err := dt.Indexer.Close(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
