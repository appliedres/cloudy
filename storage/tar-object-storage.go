package storage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var _ ObjectStorage = (*TarObjectStorage)(nil)

// TarObjectStorage implements the ObjectStorage interface using a tar file
// as the backing store. It provides methods for managing files within the tar archive.
// It can handle both uncompressed .tar files and gzip-compressed .tar.gz files.
//
// Thread safety is achieved through a read-write mutex that protects all operations.
// All operations have protection against decompression bombs by enforcing a 100MB
// file size limit for each file in the archive.
type TarObjectStorage struct {
	tarFilePath string       // The path to the tar file
	isGzipped   bool         // Whether the tar file is gzipped
	mu          sync.RWMutex // Protect concurrent access to the tar file
}

// TarEntry represents an entry in the tar file
// This struct holds both metadata and content for each file in the tar archive
type TarEntry struct {
	Name    string    // Path of the file within the tar
	Size    int64     // Size of the file in bytes
	IsDir   bool      // Whether this entry is a directory
	ModTime time.Time // Modification time
	Content []byte    // File content (nil for directories)
}

// NewTarObjectStorage creates a new object storage backend using a tar file
// as the storage medium. It treats the tar file as a virtual file system.
// It automatically detects if the file should be gzip-compressed based on
// the file extension (.tar.gz or .tgz).
//
// Parameters:
//   - tarFilePath: The path to the tar file to use as storage
//
// Returns:
//   - A new TarObjectStorage instance
func NewTarObjectStorage(tarFilePath string) *TarObjectStorage {
	// Make sure we have an absolute path
	absPath, err := filepath.Abs(tarFilePath)
	if err == nil {
		tarFilePath = absPath
	}

	// Check if the path indicates a gzipped tar file
	// Accept both .tar.gz and .tgz extensions
	isGzipped := strings.HasSuffix(tarFilePath, ".tar.gz") ||
		strings.HasSuffix(tarFilePath, ".tgz") ||
		strings.ToLower(filepath.Ext(tarFilePath)) == ".gz"

	return &TarObjectStorage{
		tarFilePath: tarFilePath,
		isGzipped:   isGzipped,
		mu:          sync.RWMutex{},
	}
}

// readTarEntries reads all entries from the tar archive
// This is the core function that performs the reading logic for tar files
// It handles both regular tar files and gzip-compressed tar files
func (t *TarObjectStorage) readTarEntries(ctx context.Context) ([]TarEntry, error) {
	var entries []TarEntry

	// Check if the file exists
	_, err := os.Stat(t.tarFilePath)
	if os.IsNotExist(err) {
		// File doesn't exist yet - return empty entries
		return entries, nil
	}

	// Open the file
	file, err := os.Open(t.tarFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	var tarReader *tar.Reader

	if t.isGzipped {
		// Create a gzip reader for compressed tar files
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			// If we get an error here, the file might be corrupt or not actually gzipped
			return nil, fmt.Errorf("failed to create gzip reader (file may not be in gzip format): %w", err)
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else {
		tarReader = tar.NewReader(file)
	}

	// Read all entries from the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			// End of archive
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading tar entry: %w", err)
		}

		// Create entry with basic information from the header
		entry := TarEntry{
			Name:    header.Name,
			Size:    header.Size,
			IsDir:   header.Typeflag == tar.TypeDir,
			ModTime: header.ModTime,
		}

		// Only read content for regular files (not dirs or special files)
		if header.Typeflag == tar.TypeReg {
			// Limit file size to prevent decompression bombs
			const maxSize = 100 * 1024 * 1024 // 100MB limit
			if header.Size > maxSize {
				return nil, fmt.Errorf("file %s exceeds maximum allowed size of %d bytes", header.Name, maxSize)
			}

			// Use LimitReader to safely read the file content up to maxSize
			limitedReader := io.LimitReader(tarReader, header.Size)
			content, err := io.ReadAll(limitedReader)
			if err != nil {
				return nil, fmt.Errorf("error reading file content for %s: %w", header.Name, err)
			}

			// Store the content in the entry
			entry.Content = content

			// Update the size in case the actual content is smaller than the header size
			entry.Size = int64(len(content))
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// writeTarEntries writes all entries to a new tar file
// This method takes the entire list of entries and writes them to a new tar file
// It handles both regular tar files and gzip-compressed tar files based on the isGzipped flag
func (t *TarObjectStorage) writeTarEntries(entries []TarEntry) error {
	// Ensure the parent directory exists
	if err := os.MkdirAll(filepath.Dir(t.tarFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create a temp file with appropriate extension (.tar or .tar.gz)
	ext := filepath.Ext(t.tarFilePath)
	tempFile, err := os.CreateTemp(filepath.Dir(t.tarFilePath), "temp_*"+ext)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFilePath := tempFile.Name()
	defer os.Remove(tempFilePath) // Clean up in case of failure

	var tarWriter *tar.Writer
	var gzWriter *gzip.Writer

	if t.isGzipped {
		// Create a gzip writer
		gzWriter = gzip.NewWriter(tempFile)
		tarWriter = tar.NewWriter(gzWriter)
	} else {
		tarWriter = tar.NewWriter(tempFile)
	}

	// Write all entries
	for _, entry := range entries {
		header := &tar.Header{
			Name:    entry.Name,
			Size:    entry.Size,
			ModTime: entry.ModTime,
		}

		if entry.IsDir {
			header.Typeflag = tar.TypeDir
			header.Mode = 0755
		} else {
			header.Typeflag = tar.TypeReg
			header.Mode = 0644
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			if gzWriter != nil {
				gzWriter.Close()
			}
			tarWriter.Close()
			tempFile.Close()
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		if !entry.IsDir && entry.Content != nil {
			if _, err := tarWriter.Write(entry.Content); err != nil {
				if gzWriter != nil {
					gzWriter.Close()
				}
				tarWriter.Close()
				tempFile.Close()
				return fmt.Errorf("failed to write file content: %w", err)
			}
		}
	}

	// Close the writers
	if err := tarWriter.Close(); err != nil {
		if gzWriter != nil {
			gzWriter.Close()
		}
		tempFile.Close()
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	if gzWriter != nil {
		if err := gzWriter.Close(); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Replace the original file with the temp file
	if err := os.Rename(tempFilePath, t.tarFilePath); err != nil {
		return fmt.Errorf("failed to replace original tar file: %w", err)
	}

	return nil
}

// Download implements ObjectStorage.
// Retrieves a file from the tar archive and returns it as an io.ReadCloser
func (t *TarObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	// Handle cases where the tar file doesn't exist yet
	exists, _ := CheckFileExists(t.tarFilePath)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, key)
	}

	entries, err := t.readTarEntries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read tar entries: %w", err)
	}

	// Find the entry with the specified key
	for _, entry := range entries {
		if entry.Name == key && !entry.IsDir {
			// Return empty content if needed
			if len(entry.Content) == 0 {
				return io.NopCloser(bytes.NewReader([]byte{})), nil
			}
			return io.NopCloser(bytes.NewReader(entry.Content)), nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrFileNotFound, key)
}

// Exists implements ObjectStorage.
// Checks if a file exists in the tar archive
func (t *TarObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	// Handle cases where the tar file doesn't exist yet
	exists, err := CheckFileExists(t.tarFilePath)
	if !exists || err != nil {
		return false, nil
	}

	entries, err := t.readTarEntries(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to read tar entries: %w", err)
	}

	// Check if any entry matches the key
	for _, entry := range entries {
		if entry.Name == key {
			return true, nil
		}
	}

	return false, nil
}

// List implements ObjectStorage.
// Returns the objects and prefixes in the tar archive that match the given prefix
func (t *TarObjectStorage) List(ctx context.Context, prefix string) ([]*StoredObject, []*StoredPrefix, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var objects []*StoredObject
	var prefixes []*StoredPrefix

	// Read all entries from the tar file
	entries, err := t.readTarEntries(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read tar entries: %w", err)
	}

	// All entries in a tar file start with "./"
	prefix = "./" + prefix

	// Filter entries based on the prefix and organize them into objects and prefixes
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name, prefix) {
			// Skip entries that do not match the prefix
			continue
		}

		if entry.IsDir {
			// This is a directory
			prefixKey := entry.Name
			if !strings.HasSuffix(entry.Name, "/") {
				prefixKey += "/"
			}
			prefixes = append(prefixes, &StoredPrefix{Key: prefixKey})
			continue
		}

		objects = append(objects, &StoredObject{
			Key:  entry.Name,
			Size: entry.Size, // This is uint32, so safer to convert
			Tags: make(map[string]string),
		})
	}

	return objects, prefixes, nil
}

// Delete implements ObjectStorage.
// Removes a file from the tar archive
func (t *TarObjectStorage) Delete(ctx context.Context, key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return err
	}

	// Read all entries from the tar file
	entries, err := t.readTarEntries(ctx)
	if err != nil {
		return fmt.Errorf("failed to read tar entries: %w", err)
	}

	// Filter out the entry with the specified key
	found := false
	var newEntries []TarEntry
	for _, entry := range entries {
		if entry.Name != key {
			newEntries = append(newEntries, entry)
		} else {
			found = true
		}
	}

	// Return an error if the file wasn't found
	if !found {
		return fmt.Errorf("%w: %s", ErrFileNotFound, key)
	}

	// Write the filtered entries back to the tar file
	return t.writeTarEntries(newEntries)
}

// UpdateMetadata implements ObjectStorage.
// TAR format doesn't have built-in support for metadata, so this implementation
// just checks if the file exists and returns success.
// A more sophisticated implementation could store metadata in a special file
// within the tar archive.
func (t *TarObjectStorage) UpdateMetadata(ctx context.Context, key string, tags map[string]string) error {
	// Validate the key
	if err := ValidateKey(key); err != nil {
		return err
	}

	// Check if the file exists
	exists, err := t.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("%w: %s", ErrFileNotFound, key)
	}

	// File exists, but we can't update metadata in a standard TAR file
	// A better implementation could store metadata in a special file in the archive
	// For now, just return success
	return nil
}

// Upload implements ObjectStorage.
// This method adds or updates a file in the tar archive
func (t *TarObjectStorage) Upload(ctx context.Context, key string, data io.Reader, tags map[string]string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return err
	}

	// If the key ends with '/', treat it as a directory
	isDir := IsDirectory(key)

	var content []byte
	var err error

	// Only read content for regular files, not directories
	if !isDir {
		content, err = ReadLimitedContent(data)
		if err != nil {
			return err
		}
	}

	// Read all existing entries
	entries, err := t.readTarEntries(ctx)
	if err != nil {
		// Only return error if it's not just that the file doesn't exist yet
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read existing tar entries: %w", err)
		}
	}

	// Find and update or add the entry
	found := false
	for i, entry := range entries {
		if entry.Name == key {
			entries[i].Content = content
			entries[i].Size = int64(len(content))
			entries[i].ModTime = time.Now()
			entries[i].IsDir = isDir
			found = true
			break
		}
	}

	if !found {
		// Add a new entry
		entries = append(entries, TarEntry{
			Name:    key,
			Size:    int64(len(content)),
			IsDir:   isDir,
			ModTime: time.Now(),
			Content: content,
		})
	}

	// Write the entries back to the file
	return t.writeTarEntries(entries)
}
