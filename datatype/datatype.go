package datatype

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/datastore"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// type any struct{}

type Datatype[T any] struct {
	Name      string
	Table     string
	Prefix    string
	IDField   string
	GetIDFunc func(dt *Datatype[T], item *T) string
	SetIDFunc func(dt *Datatype[T], item *T, id string) string

	DataStore JsonDataStore[T]

	// BeforeSave []BeforeSaveInterceptor[T]
	BeforeSave  []InterceptItem[T]
	AfterSave   []InterceptItem[T]
	AfterGet    []InterceptItem[T]
	AfterDelete []AfterDeleteFunc[T]

	initialized        bool
	OnConnectionChange func()
}

func NewDatatype[T any](name string, table string, options ...func(dt *Datatype[T])) *Datatype[T] {
	dt := &Datatype[T]{
		Name:  name,
		Table: table,
	}
	for _, opt := range options {
		opt(dt)
	}
	return dt
}

type BeforeSaveInterceptor[T any] interface {
	BeforeSave(ctx context.Context, dt *Datatype[T], item *T) (*T, error)
}

type InterceptItem[T any] func(ctx context.Context, dt *Datatype[T], item *T) (*T, error)

type AfterSaveInterceptor[T any] interface {
	AfterSave(ctx context.Context, dt *Datatype[T], item *T) (*T, error)
}

type AfterGetInterceptor[T any] interface {
	AfterGet(ctx context.Context, dt *Datatype[T], item *T) (*T, error)
}

type AfterDeleteInterceptor[T any] interface {
	AfterDelete(ctx context.Context, dt *Datatype[T], key []string) error
}
type AfterDeleteFunc[T any] func(ctx context.Context, dt *Datatype[T], key []string) error

func (dt *Datatype[T]) interceptGet(ctx context.Context, item *T) (*T, error) {
	var me *multierror.Error
	for _, fn := range dt.AfterGet {
		_, err := fn(ctx, dt, item)
		if err != nil {
			me = multierror.Append(me, err)
		}
	}
	return item, me.ErrorOrNil()
}
func (dt *Datatype[T]) interceptBeforeSave(ctx context.Context, item *T) (*T, error) {
	var me *multierror.Error
	for _, fn := range dt.BeforeSave {
		_, err := fn(ctx, dt, item)
		if err != nil {
			me = multierror.Append(me, err)
		}
	}
	return item, me.ErrorOrNil()
}
func (dt *Datatype[T]) interceptAfterSave(ctx context.Context, item *T) (*T, error) {
	var me *multierror.Error
	for _, fn := range dt.AfterSave {
		_, err := fn(ctx, dt, item)
		if err != nil {
			me = multierror.Append(me, err)
		}
	}
	return item, me.ErrorOrNil()
}

func (dt *Datatype[T]) interceptAfterDelete(ctx context.Context, key []string) error {
	var me *multierror.Error
	for _, fn := range dt.AfterDelete {
		err := fn(ctx, dt, key)
		if err != nil {
			me = multierror.Append(me, err)
		}
	}
	return me.ErrorOrNil()
}

func (dt *Datatype[T]) SetDatastore(ds JsonDataStore[T]) {
	dt.DataStore = ds
}

func (dt *Datatype[T]) GetAll(ctx context.Context) ([]*T, error) {
	var err error
	var output []*T

	err = dt.initIfNeeded(ctx)
	if err != nil {
		return output, err
	}
	if dt.DataStore == nil {
		return nil, errors.New("No Datastore Configured")
	}

	// Load the item
	items, err := dt.DataStore.GetAll(ctx)
	if err != nil {
		return output, err
	}

	// Run the interceptors, fail on error
	var merr *multierror.Error
	for _, item := range items {
		item, err = dt.interceptGet(ctx, item)
		if err != nil {
			merr = multierror.Append(merr, err)
			continue
		}
		output = append(output, item)
	}

	// All good, now Return
	return output, merr.ErrorOrNil()
}

func (dt *Datatype[T]) Get(ctx context.Context, ID string) (*T, error) {
	var err error
	var output *T
	err = dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}

	// Load the item
	output, err = dt.GetRaw(ctx, ID)
	if err != nil {
		return nil, err
	}

	return dt.interceptGet(ctx, output)
}

// Retrieves the Raw Data from the data store and then attempts
// to unmarshall it using reflection
func (dt *Datatype[T]) GetRaw(ctx context.Context, ID string) (*T, error) {

	err := dt.initIfNeeded(ctx)
	if err != nil {
		return nil, err
	}
	if dt.DataStore == nil {
		return nil, errors.New("Datastore not initialized yet")
	}

	return dt.DataStore.Get(ctx, ID)
}

func (dt *Datatype[T]) Save(ctx context.Context, item *T) (*T, error) {
	var err error
	var v *T
	err = dt.initIfNeeded(ctx)
	if err != nil {
		return v, err
	}

	item, err = dt.interceptBeforeSave(ctx, item)
	if err != nil {
		return item, err
	}

	// Run the internal Save
	item, err = dt.SaveRaw(ctx, item)
	if err != nil {
		return item, err
	}

	return dt.interceptAfterSave(ctx, item)
}

func (dt *Datatype[T]) SaveRaw(ctx context.Context, item *T) (*T, error) {
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return item, err
	}
	if dt.DataStore == nil {
		return item, errors.New("Datastore not initialized")
	}

	id := dt.GetID(ctx, item)
	return item, dt.DataStore.Save(ctx, item, id)
}

func (dt *Datatype[T]) ToRaw(ctx context.Context, item *T) ([]byte, string, error) {
	key := dt.GetID(ctx, item)
	if key == "" || key == "<invalid Value>" {
		return nil, "", cloudy.Error(ctx, "No ID Set")
	}

	data, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return nil, key, err
	}

	return data, key, err
}

func (dt *Datatype[T]) NativeQuery(ctx context.Context, query *T) ([]*T, error) {
	var zero []*T
	err := dt.initIfNeeded(ctx)
	if err != nil {
		return zero, err
	}

	// if dt.Indexer != nil {
	// 	return dt.Indexer.Search(ctx, query)
	// }

	return zero, errors.New("no Indexer configured")
}

func (dt *Datatype[T]) Query(ctx context.Context, query *datastore.SimpleQuery) ([]*T, error) {
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
	err := dt.DataStore.Delete(ctx, key)
	if err != nil {
		return err
	}
	return dt.interceptAfterDelete(ctx, []string{key})
}

func (dt *Datatype[T]) DeleteAll(ctx context.Context, keys []string) error {
	bulkDs, isBulk := dt.DataStore.(BulkJsonDataStore[T])
	if isBulk {
		err := bulkDs.DeleteAll(ctx, keys)
		if err != nil {
			return err
		}
		return dt.interceptAfterDelete(ctx, keys)
	}

	for _, key := range keys {
		err := dt.DataStore.Delete(ctx, key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dt *Datatype[T]) GetIDs(ctx context.Context, items []*T) []string {
	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = dt.GetID(ctx, item)
	}
	return keys
}

func (dt *Datatype[T]) SaveAll(ctx context.Context, items []*T) error {
	bulkDs, isBulk := dt.DataStore.(BulkJsonDataStore[T])
	if isBulk {
		keys := dt.GetIDs(ctx, items)
		return bulkDs.SaveAll(ctx, items, keys)
	}

	for _, item := range items {
		_, err := dt.Save(ctx, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dt *Datatype[T]) FromByte(ctx context.Context, data []byte) (*T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return &v, err
}

func (dt *Datatype[T]) QueryAndUpdate(ctx context.Context, query *datastore.SimpleQuery, fn func(ctx context.Context, items []*T) ([]*T, error)) ([]*T, error) {
	itemsRaw, err := dt.DataStore.QueryAndUpdate(ctx, query, func(ctx context.Context, items []*T) ([]*T, error) {
		return fn(ctx, items)
	})
	return itemsRaw, err
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
	if dt.DataStore == nil {
		return fmt.Errorf("initIfNeeded dt.Datastore %s is nil", dt.Name)
	}

	if dt.initialized {
		return nil
	}

	cloudy.Info(ctx, "dt.initIfNeeded %s", dt.Name)

	err := dt.Initialize(ctx)
	if err != nil {
		return err
	}

	if dt.OnConnectionChange != nil {
		dt.OnConnectionChange()
	}
	return nil
}

func (dt *Datatype[T]) Initialize(ctx context.Context) error {
	cloudy.Info(ctx, "dt.Initialize %s", dt.Name)

	if dt.DataStore == nil {
		return fmt.Errorf("Initialize dt.Datastore %s is nil", dt.Name)
	}

	err := dt.DataStore.Open(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "Datastore Open")
	}
	dt.initialized = true

	return nil
}

func (dt *Datatype[T]) Shutdown(ctx context.Context) error {
	if dt.DataStore != nil {
		err := dt.DataStore.Close(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dt *Datatype[T]) QueryAsMap(ctx context.Context, query *datastore.SimpleQuery) ([]map[string]any, error) {
	advDS, isAdv := dt.DataStore.(AdvQueryJsonDatastore[T])
	if isAdv {
		return advDS.QueryAsMap(ctx, query)
	}

	// If we are NOT an advanced query JSON Datastore then we have to do this the HARD way
	// We have to GET all the items from a query and then pull out what we need
	cols := query.Colums
	query.Colums = []string{}

	items, err := dt.DataStore.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	rtn := make([]map[string]any, len(items))
	for i, item := range items {

		m := make(map[string]any)
		container, err := dt.toGabs(item)
		if err != nil {
			return nil, err
		}

		for _, col := range cols {
			c := container.Path(col)
			if c != nil {
				m[col] = c.Data()
			}
		}
		rtn[i] = m
	}
	return rtn, nil
}

func (dt *Datatype[T]) toGabs(item *T) (*gabs.Container, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return gabs.ParseJSON(data)
}

func (dt *Datatype[T]) QueryTable(ctx context.Context, query *datastore.SimpleQuery) ([][]any, error) {
	advDS, isAdv := dt.DataStore.(AdvQueryJsonDatastore[T])
	if isAdv {
		return advDS.QueryTable(ctx, query)
	}

	// If we are NOT an advanced query JSON Datastore then we have to do this the HARD way
	// We have to GET all the items from a query and then pull out what we need
	cols := query.Colums
	query.Colums = []string{}

	items, err := dt.DataStore.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	rtn := make([][]any, len(items))
	for i, item := range items {
		row := make([]any, len(cols))
		container, err := dt.toGabs(item)
		if err != nil {
			return nil, err
		}

		for j, col := range cols {
			c := container.Path(col)
			if c != nil {
				row[j] = c.Data()
			}
		}
		rtn[i] = row
	}
	return rtn, nil
}

func (dt *Datatype[T]) IsReady() bool {
	return dt.DataStore != nil && dt.initialized
}

func (dt *Datatype[T]) AddIfMissing(ctx context.Context, item *T) (bool, error) {
	id := dt.GetID(ctx, item)
	exists, err := dt.Exists(ctx, id)
	if err != nil {
		return false, err
	}
	if exists {
		return false, nil
	}

	_, err = dt.Save(ctx, item)
	return true, err
}

// type Datatype2Option[T any] = func(dt *Datatype2[T])

func WithIdField[T any](idField string) func(dt *Datatype[T]) {
	return func(dt *Datatype[T]) {
		dt.IDField = idField
	}
}

func WithSetId[T any](fn func(dt *Datatype[T], item *T, id string) string) func(dt *Datatype[T]) {
	return func(dt *Datatype[T]) {
		dt.SetIDFunc = fn
	}
}

func WithGetId[T any](fn func(dt *Datatype[T], item *T) string) func(dt *Datatype[T]) {
	return func(dt *Datatype[T]) {
		dt.GetIDFunc = fn
	}
}

func WithBeforeSave[T any](fn InterceptItem[T]) func(dt *Datatype[T]) {
	return func(dt *Datatype[T]) {
		dt.BeforeSave = append(dt.BeforeSave, fn)
	}
}

func WithAfterSave[T any](fn InterceptItem[T]) func(dt *Datatype[T]) {
	return func(dt *Datatype[T]) {
		dt.AfterSave = append(dt.AfterSave, fn)
	}
}

func WithAfterGet[T any](fn InterceptItem[T]) func(dt *Datatype[T]) {
	return func(dt *Datatype[T]) {
		dt.AfterGet = append(dt.AfterGet, fn)
	}
}

func WithAfterDelete[T any](fn AfterDeleteFunc[T]) func(dt *Datatype[T]) {
	return func(dt *Datatype[T]) {
		dt.AfterDelete = append(dt.AfterDelete, fn)
	}
}
