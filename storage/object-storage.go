package storage

import (
	"context"
	"io"

	"github.com/appliedres/cloudy"
)

var ObjectStorageProviders = cloudy.NewProviderRegistry[ObjectStorageManager]()

type StoredObject struct {
	Key  string
	Tags map[string]string
	Size int64
	MD5  string
}

type StoredPrefix struct {
	Key string
}

// StorageArea is an abstract container / bucket representation
type StorageArea struct {
	Name string
	Tags map[string]string
}

// ObjectStorageManager represents a "storage account".
type ObjectStorageManager interface {
	Exists(ctx context.Context, key string) (bool, error)
	List(ctx context.Context) ([]*StorageArea, error)
	GetItem(ctx context.Context, key string) (*StorageArea, error)
	Get(ctx context.Context, key string) (ObjectStorage, error)
	Create(ctx context.Context, key string, openToPublic bool, tags map[string]string) (ObjectStorage, error)
	Delete(ctx context.Context, key string) error
}

type ObjectStorage interface {
	Upload(ctx context.Context, key string, data io.Reader, tags map[string]string) error
	Exists(ctx context.Context, key string) (bool, error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]*StoredObject, []*StoredPrefix, error)
	UpdateMetadata(ctx context.Context, key string, tags map[string]string) error
}
