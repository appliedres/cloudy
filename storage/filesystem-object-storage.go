package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/appliedres/cloudy"
)

func NewFilesystemObjectStorage(rootDir string) *FilesystemObjectStorage {
	return &FilesystemObjectStorage{
		rootDir: rootDir,
	}
}

type FilesystemObjectStorage struct {
	rootDir string
}

func (fso *FilesystemObjectStorage) Upload(ctx context.Context, key string, data io.Reader, tags map[string]string) error {
	fullpath := filepath.Join(fso.rootDir, key)
	dir := filepath.Dir(fullpath)

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	file, err := os.Create(fullpath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, data)
	if err != nil {
		return err
	}

	closer, canClose := data.(io.ReadCloser)
	if canClose {
		_ = closer.Close()
	}

	return file.Close()
}

func (fso *FilesystemObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullpath := filepath.Join(fso.rootDir, key)
	return cloudy.Exists(fullpath)
}
func (fso *FilesystemObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullpath := filepath.Join(fso.rootDir, key)
	file, err := os.Open(fullpath)
	return file, err
}
func (fso *FilesystemObjectStorage) Delete(ctx context.Context, key string) error {
	fullpath := filepath.Join(fso.rootDir, key)
	return os.Remove(fullpath)
}
func (fso *FilesystemObjectStorage) List(ctx context.Context, prefix string) ([]*StoredObject, []*StoredPrefix, error) {
	return nil, nil, nil
}

func (fso *FilesystemObjectStorage) UpdateMetadata(ctx context.Context, key string, tags map[string]string) error {
	return nil
}