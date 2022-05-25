package datastore

import (
	"context"
	"io"
	iofs "io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/appliedres/cloudy"
)

func init() {
	BinaryDataStoreProviders.Register(FileSytemBinaryStoreID, func(cfg interface{}) (BinaryDataStore, error) {
		fsConfig := cfg.(*FilesystemStoreConfig)
		if fsConfig == nil {
			return nil, InvalidConfiguration
		}
		return &FilesystemStore{
			Dir:   fsConfig.Dir,
			Ext:   fsConfig.Ext,
			Perms: fsConfig.Perms,
		}, nil
	})
}

const FileSytemBinaryStoreID = "file-system"

var RootFSDir = ""

type FilesystemStoreConfig struct {
	Dir   string
	Ext   string
	Perms os.FileMode
}

type FilesystemStore struct {
	Dir   string
	Ext   string
	Perms os.FileMode
}

func NewFilesystemStore(ext string, dir ...string) *FilesystemStore {
	fs := new(FilesystemStore)
	fs.Perms = 0660
	localdir := filepath.Join(dir...)
	fs.Dir = filepath.Join(RootFSDir, localdir)
	fs.Ext = ext
	return fs
}

func (fs *FilesystemStore) Open(ctx context.Context, config interface{}) error {
	err := fs.Init()
	return err
}

func (fs *FilesystemStore) Close(ctx context.Context) error {
	return nil
}

func (fs *FilesystemStore) Save(ctx context.Context, data []byte, key string) error {
	ierr := fs.Init()
	if ierr != nil {
		return ierr
	}

	// Assuming that key is the path
	fullpath := filepath.Join(fs.Dir, key+fs.Ext)

	// Write the file
	err := ioutil.WriteFile(fullpath, data, fs.Perms)

	return err
}

func (fs *FilesystemStore) SaveStream(ctx context.Context, data io.ReadCloser, key string) (int64, error) {
	ierr := fs.Init()
	if ierr != nil {
		return 0, ierr
	}

	out, err := os.Create(key)
	if err != nil {
		return 0, err
	}
	defer cloudy.DeferableClose(ctx, out)

	written, err := io.Copy(out, data)
	if err != nil {
		return 0, err
	}
	defer cloudy.DeferableClose(ctx, data)

	return written, err
}

func (fs *FilesystemStore) Get(ctx context.Context, key string) ([]byte, error) {
	ierr := fs.Init()
	if ierr != nil {
		return nil, ierr
	}

	// Assuming that key is the path
	fullpath := filepath.Clean(filepath.Join(fs.Dir, key+fs.Ext))

	// Read the file
	data, err := ioutil.ReadFile(fullpath)
	if err != nil && !isPathError(err) {
		return nil, err
	}
	return data, err
}

func (fs *FilesystemStore) Delete(ctx context.Context, key string) error {
	ierr := fs.Init()
	if ierr != nil {
		return ierr
	}

	fullpath := filepath.Join(fs.Dir, key+fs.Ext)

	err := os.Remove(fullpath)

	return err
}

func (fs *FilesystemStore) Exists(ctx context.Context, key string) (bool, error) {
	ierr := fs.Init()
	if ierr != nil {
		return false, ierr
	}

	fullpath := filepath.Join(fs.Dir, key+fs.Ext)
	_, err := os.Stat(fullpath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

var inited = &sync.Once{}
var initError error

func (fs *FilesystemStore) Init() error {
	inited.Do(func() {
		// Check if exists
		_, err := os.Stat(fs.Dir)
		if err != nil {
			if !isPathError(err) {
				initError = err
				return
			}

			// Create
			err = os.MkdirAll(fs.Dir, fs.Perms)
			initError = err
			return
		}
		initError = nil
	})
	return initError
}

func isPathError(err error) bool {
	if err == nil {
		return false
	}

	per := err.(*iofs.PathError)
	return per != nil
}
