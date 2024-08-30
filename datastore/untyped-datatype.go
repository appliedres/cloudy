package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/appliedres/cloudy"
	"github.com/hashicorp/go-multierror"
)

type UDatatype struct {
	Name      string
	Prefix    string
	IDField   string
	GetIDFunc func(dt *UDatatype, item interface{}) string
	SetIDFunc func(dt *UDatatype, item interface{}, id string) string

	DataStore UntypedJsonDataStore
	ItemType  interface{}

	BeforeSave []UBeforeSaveInterceptor
	AfterSave  []UAfterSaveInterceptor
	BeforeGet  []UBeforeGetInterceptor
	AfterGet   []UAfterGetInterceptor

	initialized        bool
	OnCreateDS         OnCreateDS
	OnConnectionChange func()
}

// Go is really crazy in how it handles Generics.. So in order to call all the "non typed" methods
// on each Udatatype we need an interface

type UBeforeSaveInterceptor interface {
	BeforeSave(ctx context.Context, dt *UDatatype, item interface{}) (interface{}, error)
}

type BeforeSaveFunc func(context.Context, *UDatatype, interface{}) (interface{}, error)

func (f BeforeSaveFunc) BeforeSave(ctx context.Context, dt *UDatatype, item interface{}) (interface{}, error) {
	return f(ctx, dt, item)
}

type UAfterSaveInterceptor interface {
	AfterSave(ctx context.Context, dt *UDatatype, item interface{}) (interface{}, error)
}

type UBeforeGetInterceptor interface {
	BeforeGet(ctx context.Context, dt *UDatatype, ID string) (string, error)
}

type UAfterGetInterceptor interface {
	AfterGet(ctx context.Context, dt *UDatatype, item interface{}) (interface{}, error)
}

func (dt *UDatatype) GetAll(ctx context.Context) ([]interface{}, error) {
	var err error
	var output []interface{}

	err = dt.initIfNeeded(ctx)
	if err != nil {
		return output, err
	}
	if dt.DataStore == nil {
		return nil, errors.New("No Datastore Configured")
	}

	// Load the item
	data, err := dt.DataStore.GetAll(ctx)
	if err != nil {
		return output, err
	}

	// Run the interceptors, fail on error
	var merr *multierror.Error
	for _, rawBytes := range data {

		v := cloudy.NewInstancePtr(dt.ItemType)
		err = json.Unmarshal(rawBytes, v)
		if err != nil {
			merr = multierror.Append(merr, err)
			continue
		}

		output = append(output, v)
		for _, interceptor := range dt.AfterGet {
			_, err := interceptor.AfterGet(ctx, dt, v)
			if err != nil {
				merr = multierror.Append(merr, err)
			}
		}
	}

	// All good, now Return
	return output, merr.ErrorOrNil()
}

func (dt *UDatatype) Get(ctx context.Context, ID string) (interface{}, error) {
	var err error
	var output interface{}

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

// Retrieves the Raw Data from the data store and then attempts
// to unmarshall it using reflection
func (dt *UDatatype) GetRaw(ctx context.Context, ID string) (interface{}, error) {

	err := dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}
	if dt.DataStore == nil {
		return nil, errors.New("Datastore not initialized yet")
	}

	rawBytes, err := dt.DataStore.Get(ctx, ID)
	if err != nil {
		return rawBytes, err
	}
	if rawBytes == nil {
		return nil, nil
	}

	v := cloudy.NewInstancePtr(dt.ItemType)
	err = json.Unmarshal(rawBytes, v)

	return v, err
}

func (dt *UDatatype) Save(ctx context.Context, item interface{}) (interface{}, error) {
	var err error
	var v interface{}
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

func (dt *UDatatype) SaveRaw(ctx context.Context, item interface{}) (interface{}, error) {
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}
	if dt.DataStore == nil {
		return nil, errors.New("Datastore not initialized")
	}

	key := dt.GetID(ctx, item)
	if key == "" || key == "<invalid Value>" {
		return nil, cloudy.Error(ctx, "No ID Set")
	}

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return nil, err
	}

	err = dt.DataStore.Save(ctx, data, key)
	if err != nil {
		return nil, err
	}

	// if dt.Indexer != nil {
	// 	err = dt.Indexer.Index(ctx, key, data)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return item, nil
}

func (dt *UDatatype) NativeQuery(ctx context.Context, query interface{}) ([]interface{}, error) {
	var zero []interface{}
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return zero, err
	}

	// if dt.Indexer != nil {
	// 	return dt.Indexer.Search(ctx, query)
	// }

	return zero, errors.New("no Indexer configured")
}

func (dt *UDatatype) Query(ctx context.Context, query *SimpleQuery) ([][]byte, error) {
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}

	if dt.DataStore == nil {
		return nil, cloudy.Error(ctx, "dt.Datastore %s is nil", dt.Name)
	}

	return dt.DataStore.Query(ctx, query)
}

func (dt *UDatatype) GetID(ctx context.Context, item interface{}) string {
	if dt.GetIDFunc != nil {
		return dt.GetIDFunc(dt, item)
	}

	idField := "ID"
	if dt.IDField != "" {
		idField = dt.IDField
	}

	return cloudy.GetFieldString(item, idField)
}

func (dt *UDatatype) Delete(ctx context.Context, key string) error {
	return dt.DataStore.Delete(ctx, key)
}

func (dt *UDatatype) SetID(ctx context.Context, item interface{}, id string) {
	if dt.SetIDFunc != nil {
		dt.SetIDFunc(dt, item, id)
	}

	idField := "ID"
	if dt.IDField != "" {
		idField = dt.IDField
	}

	cloudy.SetFieldString(item, idField, id)
}

func (dt *UDatatype) GenerateID() string {
	return cloudy.GenerateId(dt.Prefix, 15)
}

func (dt *UDatatype) Exists(ctx context.Context, id string) (bool, error) {
	return dt.DataStore.Exists(ctx, id)
}

func (dt *UDatatype) initIfNeeded(ctx context.Context) error {
	// cloudy.Info(ctx, "dt.initIfNeeded %s", dt.Name)

	if dt.initialized {
		// cloudy.Info(ctx, "dt.initIfNeeded already initialized")
		return nil
	}
	dt.initialized = true

	if dt.DataStore != nil {
		dt.DataStore.OnCreate(dt.OnCreateDS)
		err := dt.DataStore.Open(ctx, nil)
		if err != nil {
			return err
		}
		if dt.OnConnectionChange != nil {
			dt.OnConnectionChange()
		}
	}

	// if dt.Indexer != nil {
	// 	err := dt.Indexer.Open(ctx, nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// cloudy.Info(ctx, "dt.initIfNeeded complete")
	return nil
}

func (dt *UDatatype) Initialize(ctx context.Context) error {
	cloudy.Info(ctx, "dt.Initialize %s", dt.Name)

	if dt.DataStore != nil {
		err := dt.DataStore.Open(ctx, nil)
		if err != nil {
			return err
		}
	}

	// if dt.Indexer != nil {
	// 	err := dt.Indexer.Open(ctx, nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

func (dt *UDatatype) Shutdown(ctx context.Context) error {
	if dt.DataStore != nil {
		err := dt.DataStore.Close(ctx)
		if err != nil {
			return err
		}
	}

	// if dt.Indexer != nil {
	// 	err := dt.Indexer.Close(ctx)
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

type UDatatypeOption = func(dt *UDatatype)

func NewUDatatype(name string, table string, objectType interface{}, options ...UDatatypeOption) *UDatatype {
	dt := &UDatatype{
		Name:     name,
		ItemType: objectType,
	}
	for _, opt := range options {
		opt(dt)
	}
	return dt
}

type DatatypeTypedOperations[T any] interface {
	Name() string
	Get(ctx context.Context, id string) (*T, error)
	GetAll(ctx context.Context) ([]*T, error)
	Save(ctx context.Context, item *T) (*T, error)
	Query(ctx context.Context, query *SimpleQuery) ([]*T, error)
	Delete(ctx context.Context, id string) error
	IsReady() bool
	SetOnConnectionChange(fn func())
}

type datatypeTypedOperationsImpl[T any] struct {
	dt *UDatatype
}

func AsTypedDatatype[T any](dt *UDatatype) DatatypeTypedOperations[T] {
	return &datatypeTypedOperationsImpl[T]{dt: dt}
}

func (impl *datatypeTypedOperationsImpl[T]) IsReady() bool {
	return true
}

func (impl *datatypeTypedOperationsImpl[T]) Name() string {
	return impl.dt.Name
}

func (impl *datatypeTypedOperationsImpl[T]) Delete(ctx context.Context, id string) error {
	return impl.dt.Delete(ctx, id)
}

func (impl *datatypeTypedOperationsImpl[T]) Get(ctx context.Context, id string) (*T, error) {
	data, err := impl.dt.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	rtn, isType := data.(*T)
	if !isType {
		return nil, fmt.Errorf("unable to cast %v to %v", reflect.TypeOf(data), reflect.TypeOf(impl.dt.ItemType))
	}
	return rtn, err
}

func (impl *datatypeTypedOperationsImpl[T]) GetAll(ctx context.Context) ([]*T, error) {
	data, err := impl.dt.DataStore.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	var results []*T
	for _, item := range data {
		var obj T
		if err := json.Unmarshal(item, &obj); err != nil {
			return nil, fmt.Errorf("unable to unmarshal item: %v", err)
		}
		results = append(results, &obj)
	}
	return results, nil
}

func (impl *datatypeTypedOperationsImpl[T]) Save(ctx context.Context, item *T) (*T, error) {
	itemSaved, err := impl.dt.Save(ctx, item)
	if err != nil {
		return nil, err
	}
	rtn, isType := itemSaved.(*T)
	if !isType {
		return nil, fmt.Errorf("unable to cast %v to %v", reflect.TypeOf(itemSaved), reflect.TypeOf(impl.dt.ItemType))
	}
	return rtn, err
}

func (impl *datatypeTypedOperationsImpl[T]) SetOnConnectionChange(fn func()) {
	impl.dt.OnConnectionChange = fn
}

func (impl *datatypeTypedOperationsImpl[T]) Query(ctx context.Context, query *SimpleQuery) ([]*T, error) {
	results, err := impl.dt.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if results == nil {
		return nil, nil
	}
	rtn := make([]*T, len(results))
	for i, data := range results {
		r := cloudy.NewInstancePtr(impl.dt.ItemType)
		err := json.Unmarshal(data, r)
		if err == nil {
			rtn[i] = r.(*T)
		}
	}
	return rtn, err
}

func WithIdField(idField string) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.IDField = idField
	}
}

func WithCreateFn(fn OnCreateDS) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.OnCreateDS = fn
	}
}

func WithSetId(fn func(dt *UDatatype, item interface{}, id string) string) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.SetIDFunc = fn
	}
}

func WithGetId(fn func(dt *UDatatype, item interface{}) string) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.GetIDFunc = fn
	}
}

func WithBeforeSave(fn UBeforeSaveInterceptor) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.BeforeSave = append(dt.BeforeSave, fn)
	}
}

func WithBeforeGet(fn UBeforeGetInterceptor) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.BeforeGet = append(dt.BeforeGet, fn)
	}
}

func WithAfterSave(fn UAfterSaveInterceptor) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.AfterSave = append(dt.AfterSave, fn)
	}
}

func WithAfterGet(fn UAfterGetInterceptor) UDatatypeOption {
	return func(dt *UDatatype) {
		dt.AfterGet = append(dt.AfterGet, fn)
	}
}
