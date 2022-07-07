package storage

import (
	"context"
	"io"
)

type StoredObject struct {
	Key  string
	Tags map[string]string
	Size int64
	MD5  string
}

// StorageArea is an abstract container / bucket representation
type StorageArea struct {
	Name string
}

// ObjectStorageManager manages storage areas. T
type ObjectStorageManager interface {
	Exists(ctx context.Context, key string) (bool, error)
	List(ctx context.Context) ([]*StorageArea, error)
	Get(ctx context.Context, key string) (ObjectStorage, error)
	Create(ctx context.Context, key string, openToPublic bool, tags map[string]string) (ObjectStorage, error)
	Delete(ctx context.Context, key string) error
}

type ObjectStorage interface {
	Upload(ctx context.Context, key string, data io.Reader, tags map[string]string) error
	Exists(ctx context.Context, key string) (bool, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]*StoredObject, error)
}
