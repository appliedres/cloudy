package keyvalue

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/logging"
	"github.com/hashicorp/go-multierror"
)

var DefaultDir = "~/.arkloud/secrets"

// Compile time interface checks
var _ KeyValueStoreFactory = (*DirectoryKeyValueStoreFactory)(nil)

// Factory
type DirectoryKeyValueStoreFactory struct{}

func (f *DirectoryKeyValueStoreFactory) NewConfig() interface{} {
	return &DirectoryKeyValueStore{}
}

func (f *DirectoryKeyValueStoreFactory) New(ctx context.Context, config any) (KeyValueStore, error) {
	return NewDirectoryKeyValueStore(ctx, config)
}

// Store
type DirectoryKeyValueStore struct {
	lastError error
	Dir       string
	loaded    bool
}

func NewDirectoryKeyValueStore(ctx context.Context, config interface{}) (*DirectoryKeyValueStore, error) {
	dirKv := config.(*DirectoryKeyValueStore)
	if dirKv.Dir == "" {
		dirKv.Dir = DefaultDir
		slog.WarnContext(ctx,
			fmt.Sprintf("No dirname found, using default [%v]", DefaultDir),
		)
	}

	fixed, err := fixDir(dirKv.Dir)
	if err != nil {
		slog.Error(
			fmt.Sprintf("Error fixing dir [%v]", dirKv.Dir),
			logging.WithError(err),
		)
	}
	dirKv.Dir = fixed

	slog.InfoContext(ctx,
		fmt.Sprintf("File Store, dirname [%v]", dirKv.Dir),
	)

	dirKv.loadIfNeeded()
	if dirKv.lastError != nil {
		slog.Error(
			fmt.Sprintf("Error loading %v", dirKv.Dir),
			logging.WithError(err),
		)
	}

	return dirKv, dirKv.lastError
}

func NewDefaultDirectoryKeyValueStore() *DirectoryKeyValueStore {
	return NewDirectoryKeyValueStoreWFilename(DefaultDir)
}

func NewDirectoryKeyValueStoreWFilename(dirname string) *DirectoryKeyValueStore {
	fixed, err := fixDir(dirname)
	if err != nil {
		slog.Error(
			fmt.Sprintf("Error fixing dirname [%v]", dirname),
			logging.WithError(err),
		)
	}

	return &DirectoryKeyValueStore{
		Dir: fixed,
	}
}

func fixDir(dir string) (string, error) {
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return dir, err
		}
		dir = filepath.Join(home, dir[2:])
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return dir, err
	}

	exists, err := cloudy.Exists(dir)
	if err != nil {
		return dir, err
	}
	if !exists {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return dir, err
		}
	}
	return dir, nil
}

func (fs *DirectoryKeyValueStore) GetAll() (map[string]string, error) {
	var merr *multierror.Error
	fs.loadIfNeeded()

	all := make(map[string]string)
	items, err := os.ReadDir(fs.Dir)
	if err != nil {
		return all, err
	}

	for _, item := range items {
		if item.IsDir() {
			continue
		}
		normKey := NormalizeKey(item.Name())
		v, err := fs.Load(normKey)
		if err != nil {
			merr = multierror.Append(merr, err)
		} else {
			all[normKey] = v
		}
	}
	return all, nil
}

func (fs *DirectoryKeyValueStore) loadIfNeeded() {
	var err error
	if fs.Dir == "" {
		slog.Warn(fmt.Sprintf("Creating DirectoryKeyValueStore without a name. Defaulting to %v", DefaultDir))
		fs.Dir, err = fixDir(DefaultDir)
		fs.lastError = err
		if err != nil {
			slog.Error(
				fmt.Sprintf("Error fixing filename [%v]", DefaultDir),
				logging.WithError(err),
			)
		}
	}
	if fs.loaded {
		return
	}
}

func (fs *DirectoryKeyValueStore) Get(key string) (string, error) {
	fs.loadIfNeeded()
	return fs.Load(key)
}

func (fs *DirectoryKeyValueStore) Set(name string, value string) error {
	fs.loadIfNeeded()
	return fs.Save(name, value)
}

func (fs *DirectoryKeyValueStore) SetMany(many map[string]string) error {
	var merr *multierror.Error
	fs.loadIfNeeded()
	for k, value := range many {
		merr = multierror.Append(merr, fs.Save(k, value))
	}

	return merr.ErrorOrNil()
}

func (fs *DirectoryKeyValueStore) Delete(name string) error {
	fs.loadIfNeeded()
	key := NormalizeKey(name)
	filename := fs.Filename(key)

	return os.Remove(filename)
}

func (fs *DirectoryKeyValueStore) Load(key string) (string, error) {

	keyNorm := NormalizeKey(key)
	filename := fs.Filename(keyNorm)
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", nil
	}
	return string(data), nil
}

func (fs *DirectoryKeyValueStore) Filename(key string) string {
	filename := filepath.Join(fs.Dir, key)
	return filename
}

func (fs *DirectoryKeyValueStore) Save(key string, value string) error {
	keyNorm := NormalizeKey(key)
	filename := fs.Filename(keyNorm)
	return os.WriteFile(filename, []byte(value), 0600)
}
