package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	var files []*StoredObject
	var dirs []*StoredPrefix

	entries, err := os.ReadDir(fso.rootDir)
	if err != nil {
		return files, dirs, err
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	for _, entry := range entries {
		err = fso.listinternal(entry, "/", prefix, &files, &dirs)
		if err != nil {
			return files, dirs, err
		}
	}

	return files, dirs, nil
}

func (fso *FilesystemObjectStorage) listinternal(entry os.DirEntry, path string, prefixFilter string, files *[]*StoredObject, dirs *[]*StoredPrefix) error {
	fpath := filepath.Join(path, entry.Name())

	if entry.IsDir() {
		fpath = fpath + "/"
	}

	partialMatch := cloudy.HasPrefixOverlap(fpath, prefixFilter)
	fullMatch := strings.HasPrefix(fpath, prefixFilter)
	if !partialMatch {
		return nil
	}

	// Is a file
	if !entry.IsDir() {
		// Only add full matches for files
		if !fullMatch {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		*files = append(*files, &StoredObject{
			Key:  fpath,
			Size: info.Size(),
		})
		return nil
	}

	// Handle Directory
	// Only add full matches
	if fullMatch {
		*dirs = append(*dirs, &StoredPrefix{
			Key: fpath,
		})
	}

	root := filepath.Join(fso.rootDir, fpath)
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		err = fso.listinternal(entry, fpath, prefixFilter, files, dirs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (fso *FilesystemObjectStorage) UpdateMetadata(ctx context.Context, key string, tags map[string]string) error {
	return nil
}
