package datastore

import (
	"context"

	"github.com/appliedres/cloudy"
)

type IdGenerator[T any] struct{}

func (idgen *IdGenerator[T]) BeforeSave(ctx context.Context, dt *Datatype[T], item *T) (*T, error) {
	id := dt.GetID(ctx, item)
	if id == "" {
		id = cloudy.GenerateId(dt.Prefix, 15)
		dt.SetID(ctx, item, id)
	}

	return item, nil
}
