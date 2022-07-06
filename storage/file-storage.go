package storage

import "context"

// StorageArea is an abstract container / bucket representation
type FileShare struct {
	Name string
}

// ObjectStorageManager manages storage areas. T
type FileStorageManager interface {
	List(ctx context.Context) ([]*FileShare, error)
	Get(ctx context.Context, key string) (*FileShare, error)
	Exists(ctx context.Context, key string) (bool, error)
	Create(ctx context.Context, key string, tags map[string]string) (*FileShare, error)
	Delete(ctx context.Context, key string) error
}
