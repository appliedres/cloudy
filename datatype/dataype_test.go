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
