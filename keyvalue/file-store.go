package keyvalue

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Jeffail/gabs/v2"
	"github.com/appliedres/cloudy"
	"github.com/appliedres/cloudy/logging"
	"github.com/hashicorp/go-multierror"
)

var DefaultFilename = "~/.arkloud/config.json"

// Compile time interface checks
var _ WritableKeyValueStore = (*FileKeyValueStore)(nil)
var _ KeyValueStoreFactory = (*FileKeyValueStoreFactory)(nil)

// Factory
type FileKeyValueStoreFactory struct{}

func (f *FileKeyValueStoreFactory) NewConfig() interface{} {
	return &FileKeyValueStore{}
}

func (f *FileKeyValueStoreFactory) New(ctx context.Context, config any) (KeyValueStore, error) {
	return NewFileKeyValueStore(ctx, config)
}

// Store
type FileKeyValueStore struct {
	lastError error
	Filename  string
	loaded    bool
	lock      sync.RWMutex
	data      map[string]string
}

func NewFileKeyValueStore(ctx context.Context, config interface{}) (*FileKeyValueStore, error) {
	fileKv := config.(*FileKeyValueStore)
	if fileKv.Filename == "" {
		fileKv.Filename = DefaultFilename
		slog.WarnContext(ctx,
			fmt.Sprintf("No filename found, using default [%v]", fileKv.Filename),
		)
	}

	fixed, err := fixFile(fileKv.Filename)
	if err != nil {
		slog.Error(
			fmt.Sprintf("Error fixing filename [%v]", fileKv.Filename),
			logging.WithError(err),
		)
	}
	fileKv.Filename = fixed

	slog.InfoContext(ctx,
		fmt.Sprintf("File Store, filename [%v]", fileKv.Filename),
	)

	fileKv.loadIfNeeded()
	if fileKv.lastError != nil {
		slog.Error(
			fmt.Sprintf("Error loading %v", fileKv.Filename),
			logging.WithError(err),
		)
	}

	return fileKv, fileKv.lastError
}

func NewDefaultFileKeyValueStore() *FileKeyValueStore {
	return NewFileKeyValueStoreWFilename(DefaultFilename)
}

func NewFileKeyValueStoreWFilename(filename string) *FileKeyValueStore {
	fixed, err := fixFile(filename)
	if err != nil {
		slog.Error(
			fmt.Sprintf("Error fixing filename [%v]", filename),
			logging.WithError(err),
		)
	}

	return &FileKeyValueStore{
		Filename: fixed,
		data:     make(map[string]string),
	}
}

func fixFile(filename string) (string, error) {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return filename, err
		}
		dir = filepath.Join(home, dir[2:])
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return filename, err
	}

	exists, err := cloudy.Exists(dir)
	if err != nil {
		return filename, err
	}
	if !exists {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return filename, err
		}
	}
	return filepath.Join(dir, base), nil
}

func (fs *FileKeyValueStore) GetAll() (map[string]string, error) {
	fs.loadIfNeeded()

	return fs.data, nil
}

func (fs *FileKeyValueStore) loadIfNeeded() {
	var err error
	if fs.Filename == "" {
		slog.Warn(fmt.Sprintf("Creating FileKeyValueStore without a name. Defaulting to %v", DefaultFilename))
		fs.Filename, err = fixFile(DefaultFilename)
		if err != nil {
			slog.Error(
				fmt.Sprintf("Error fixing filename [%v]", DefaultFilename),
				logging.WithError(err),
			)
		}
	}
	if fs.loaded {
		return
	}
	fs.lastError = fs.Load(false)
}

func (fs *FileKeyValueStore) Get(name string) (string, error) {
	fs.loadIfNeeded()

	fs.lock.Lock()
	defer fs.lock.Unlock()
	key := NormalizeKey(name)
	rtn := fs.data[key]
	if rtn != "" {
		return rtn, nil
	}

	rtn = fs.data[name]
	return rtn, nil
}

func (fs *FileKeyValueStore) Set(name string, value string) error {
	fs.loadIfNeeded()

	fs.lock.Lock()
	defer fs.lock.Unlock()

	key := NormalizeKey(name)
	fs.data[key] = value

	return fs.Save()
}

func (fs *FileKeyValueStore) SetMany(many map[string]string) error {
	fs.loadIfNeeded()

	fs.lock.Lock()
	defer fs.lock.Unlock()

	for k, value := range many {
		key := NormalizeKey(k)
		fs.data[key] = value
	}

	return fs.Save()
}

func (fs *FileKeyValueStore) Delete(name string) error {
	fs.loadIfNeeded()

	fs.lock.Lock()
	defer fs.lock.Unlock()

	key := NormalizeKey(name)
	delete(fs.data, key)

	return fs.Save()
}

func (fs *FileKeyValueStore) Load(mustExist bool) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	fs.loaded = true
	exists, err := cloudy.Exists(fs.Filename)
	if err != nil && mustExist {
		return err
	}
	if !exists {
		if mustExist {
			return fmt.Errorf("%v does not exist", fs.Filename)
		}
		return nil
	}

	if strings.HasSuffix(fs.Filename, ".json") {
		return fs.loadFromJson()
	}
	return fs.loadFromEnv()
}

func (fs *FileKeyValueStore) Save() error {

	if strings.HasSuffix(fs.Filename, ".json") {
		return fs.saveAsJson()
	}
	return fs.saveAsEnv()
}

func LoadEnvFromString(data string) (map[string]string, error) {
	rtn := make(map[string]string)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Comment
			continue
		}
		if len(trimmed) == 0 {
			continue
		}

		index := strings.Index(trimmed, "=")
		if index > 0 {
			k := trimmed[0:index]
			v := trimmed[index+1:]

			name := NormalizeKey(k)
			rtn[name] = v
		}

	}
	return rtn, nil
}

func (fs *FileKeyValueStore) saveAsJson() error {
	c := gabs.New()

	var merr *multierror.Error
	for k, v := range fs.data {
		_, err := c.SetP(v, k)
		merr = multierror.Append(merr, err)
	}
	err := merr.ErrorOrNil()
	if err != nil {
		return err
	}
	return os.WriteFile(fs.Filename, c.BytesIndent("", "  "), 0600)
}

func (fs *FileKeyValueStore) saveAsEnv() error {
	var sb strings.Builder
	for k, v := range fs.data {
		key := cloudy.ToEnvName(k, "")
		sb.WriteString(fmt.Sprintf("%v=%v\n", key, v))
	}
	return os.WriteFile(fs.Filename, []byte(sb.String()), 0600)
}

func (fs *FileKeyValueStore) loadFromJson() error {
	container, err := gabs.ParseJSONFile(fs.Filename)
	if err != nil {
		return err
	}

	src, err := container.Flatten()
	if err != nil {
		return err
	}

	mapStr := make(map[string]string)
	for k, v := range src {
		if v == nil {
			continue
		}
		val := fmt.Sprintf("%v", v)
		normalKey := NormalizeKey(k)
		mapStr[normalKey] = val
	}
	fs.data = mapStr
	return nil
}

func (fs *FileKeyValueStore) loadFromEnv() error {
	data, err := os.ReadFile(fs.Filename)
	if err != nil {
		return err
	}

	m, err := LoadEnvFromString(string(data))
	if err != nil {
		return err
	}
	fs.data = m
	return nil
}

type ReadOnlyFileStore struct {
	fs KeyValueStore
}

func (rfs *ReadOnlyFileStore) Get(name string) (string, error) {
	return rfs.fs.Get(name)
}

func (rfs *ReadOnlyFileStore) GetAll() (map[string]string, error) {
	return rfs.fs.GetAll()
}

func NewReadOnly(s KeyValueStore) *ReadOnlyFileStore {
	return &ReadOnlyFileStore{
		fs: s,
	}
}
