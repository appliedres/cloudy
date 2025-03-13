package datatype

import (
	"context"
	"testing"

	"github.com/appliedres/cloudy/datastore"
	"github.com/stretchr/testify/require"
)

func TestDatatype(t *testing.T) {
	ctx := context.Background()

	dt := NewDatatype[datastore.TestItem]("test", "test")
	ds := NewInMemoryStore[datastore.TestItem]()
	dt.SetDatastore(ds)

	item := &datastore.TestItem{
		ID:   "1234",
		Name: "MyName",
	}

	_, err := dt.Save(ctx, item)
	require.NoError(t, err)

	dt2 := NewDatatype[datastore.TestItem]("test", "test", WithBeforeSave(func(ctx context.Context, dt *Datatype[datastore.TestItem], item *datastore.TestItem) (*datastore.TestItem, error) {
		item.ID = "4567"
		return item, nil
	}))
	dt2.SetDatastore(ds)

	item2 := &datastore.TestItem{
		ID:   "1234",
		Name: "MyName",
	}

	_, err = dt2.Save(ctx, item2)
	require.NoError(t, err)

}

func TestDTInterceptors(t *testing.T) {
	ctx := context.Background()
	ops := make(map[string]int)
	ds := NewInMemoryStore[datastore.TestItem]()
	dt := NewDatatype("test", "test",
		WithBeforeSave(func(ctx context.Context, dt *Datatype[datastore.TestItem], item *datastore.TestItem) (*datastore.TestItem, error) {
			ops["beforeSave"]++
			return item, nil
		}),
		WithAfterSave(func(ctx context.Context, dt *Datatype[datastore.TestItem], item *datastore.TestItem) (*datastore.TestItem, error) {
			ops["afterSave"]++
			return item, nil
		}),
		WithAfterGet(func(ctx context.Context, dt *Datatype[datastore.TestItem], item *datastore.TestItem) (*datastore.TestItem, error) {
			ops["afterGet"]++
			return item, nil
		}),
		WithAfterDelete(func(ctx context.Context, dt *Datatype[datastore.TestItem], keys []string) error {
			ops["afterDelete"]++
			return nil
		}),
	)
	dt.SetDatastore(ds)

	item := &datastore.TestItem{
		ID:   "1234",
		Name: "MyName",
	}

	_, err := dt.Save(ctx, item)
	require.NoError(t, err)
	require.Equal(t, 1, ops["beforeSave"], "beforeSave should be called once")
	require.Equal(t, 1, ops["afterSave"], "afterSave should be called once")

	got, err := dt.Get(ctx, item.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, 1, ops["afterGet"], "afterGet should be called once")

	err = dt.Delete(ctx, item.ID)
	require.NoError(t, err)
	require.Equal(t, 1, ops["afterDelete"], "afterDelete should be called once")

}
