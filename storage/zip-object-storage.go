package storage

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var _ ObjectStorage = (*ZipObjectStorage)(nil)

// ZipObjectStorage implements the ObjectStorage interface using a zip file
// as the backing store. It provides methods for managing files within the zip.
type ZipObjectStorage struct {
	zipFilePath string
	mu          sync.RWMutex // Protect concurrent access to the zip file
}

// Delete implements ObjectStorage.
func (z *ZipObjectStorage) Delete(ctx context.Context, key string) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return err
	}

	// Check if the zip file exists
	exists, err := CheckFileExists(z.zipFilePath)
	if !exists {
		return fmt.Errorf("%w: %s", ErrFileNotFound, key)
	}

	// Open the existing zip file
	existingZipReader, err := zip.OpenReader(z.zipFilePath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer existingZipReader.Close()

	// Create a temp file to write the new zip content
	tempZipFile, tempZipPath, err := CreateTempFile(z.zipFilePath, "temp_", ".zip")
	if err != nil {
		return fmt.Errorf("failed to create temp zip file: %w", err)
	}
	defer os.Remove(tempZipPath) // Clean up in case of failure

	// Create a new zip writer
	zipWriter := zip.NewWriter(tempZipFile)

	// Copy all files from the original zip except the one to delete
	for _, file := range existingZipReader.File {
		if file.Name == key {
			// Skip the file to delete
			continue
		}

		// Open the file from the original zip
		fileReader, err := file.Open()
		if err != nil {
			zipWriter.Close()
			tempZipFile.Close()
			return fmt.Errorf("failed to open file in original zip: %w", err)
		}

		// Create a new entry in the new zip
		writer, err := zipWriter.Create(file.Name)
		if err != nil {
			fileReader.Close()
			zipWriter.Close()
			tempZipFile.Close()
			return fmt.Errorf("failed to create file in new zip: %w", err)
		}

		// Copy the content with size limit to prevent decompression bombs
		// Limit size to 100MB per file (adjust as needed)
		const maxSize = 100 * 1024 * 1024
		limitedReader := io.LimitReader(fileReader, maxSize)
		written, err := io.Copy(writer, limitedReader)
		if err != nil {
			fileReader.Close()
			zipWriter.Close()
			tempZipFile.Close()
			return fmt.Errorf("failed to copy file content: %w", err)
		}
		if written >= maxSize {
			fileReader.Close()
			zipWriter.Close()
			tempZipFile.Close()
			return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
		}

		fileReader.Close()
	}

	// Finalize the zip file
	if err := zipWriter.Close(); err != nil {
		tempZipFile.Close()
		return fmt.Errorf("failed to close zip writer: %w", err)
	}
	tempZipFile.Close()

	// Replace the original zip with the new one using the utility function
	if err := SafeReplace(tempZipPath, z.zipFilePath); err != nil {
		return fmt.Errorf("failed to replace original zip file: %w", err)
	}

	return nil
}

// Download implements ObjectStorage.
func (z *ZipObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	z.mu.RLock()
	defer z.mu.RUnlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	// Check if the zip file exists
	exists, err := CheckFileExists(z.zipFilePath)
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, key)
	}

	// Open the zip file
	zipReader, err := zip.OpenReader(z.zipFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file: %w", err)
	}

	// Find the file inside the zip
	for _, file := range zipReader.File {
		if file.Name == key {
			// Open the file inside the zip
			fileReader, err := file.Open()
			if err != nil {
				zipReader.Close()
				return nil, fmt.Errorf("failed to open file in zip: %w", err)
			}

			// Read the content of the file with size limit
			content, err := ReadLimitedContent(fileReader)
			if err != nil {
				fileReader.Close()
				zipReader.Close()
				return nil, err
			}
			fileReader.Close()
			zipReader.Close()

			// Return a ReadCloser that wraps the content
			return io.NopCloser(bytes.NewReader(content)), nil
		}
	}

	zipReader.Close()
	return nil, fmt.Errorf("%w: %s", ErrFileNotFound, key)
}

// Exists implements ObjectStorage.
func (z *ZipObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	z.mu.RLock()
	defer z.mu.RUnlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	// Check if the zip file exists
	exists, err := CheckFileExists(z.zipFilePath)
	if !exists || err != nil {
		return false, nil
	}

	// Open the zip file
	zipReader, err := zip.OpenReader(z.zipFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	// Look for the file in the zip
	for _, file := range zipReader.File {
		if file.Name == key {
			return true, nil
		}
	}

	return false, nil
}

// List implements ObjectStorage.
func (z *ZipObjectStorage) List(ctx context.Context, prefix string) ([]*StoredObject, []*StoredPrefix, error) {
	z.mu.RLock()
	defer z.mu.RUnlock()

	var objects []*StoredObject
	var prefixes []*StoredPrefix

	// Check if the zip file exists
	exists, _ := CheckFileExists(z.zipFilePath)
	if !exists {
		// If zip doesn't exist, return empty results (not an error)
		return objects, prefixes, nil
	}

	// Open the zip file
	zipReader, err := zip.OpenReader(z.zipFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open zip file: %w", err)
	}
	defer zipReader.Close()

	// Iterate through all files in the zip
	for _, file := range zipReader.File {
		if !strings.HasPrefix(file.Name, prefix) {
			continue
		}

		// This file matches our prefix
		info := file.FileInfo()
		if info.IsDir() {
			// This is a directory
			prefixKey := file.Name
			if !strings.HasSuffix(file.Name, "/") {
				prefixKey += "/"
			}
			prefixes = append(prefixes, &StoredPrefix{Key: prefixKey})
		} else {
			// This is a file
			objects = append(objects, &StoredObject{
				Key:  file.Name,
				Size: info.Size(), // This is uint32, so safer to convert
				Tags: make(map[string]string),
			})
		}

	}

	return objects, prefixes, nil
}

// UpdateMetadata implements ObjectStorage.
func (z *ZipObjectStorage) UpdateMetadata(ctx context.Context, key string, tags map[string]string) error {
	// Validate the key
	if err := ValidateKey(key); err != nil {
		return err
	}

	// ZIP format doesn't natively support metadata/tags
	// We could implement this by storing a side JSON file inside the zip
	// with the metadata, but for simplicity we'll just check if the file exists

	exists, err := z.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check if file exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("%w: %s", ErrFileNotFound, key)
	}

	// File exists, but we can't update metadata in a standard ZIP file
	// so we'll just return success
	return nil
}

// Upload implements ObjectStorage.
func (z *ZipObjectStorage) Upload(ctx context.Context, key string, data io.Reader, tags map[string]string) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	// Validate the key
	if err := ValidateKey(key); err != nil {
		return err
	}

	// Read data with size limits to prevent decompression bombs
	content, err := ReadLimitedContent(data)
	if err != nil {
		return err
	}

	var existingZipReader *zip.ReadCloser
	zipExists := false

	// Check if the zip file already exists
	if _, err := os.Stat(z.zipFilePath); err == nil {
		zipExists = true
		// Open the existing zip file
		existingZipReader, err = zip.OpenReader(z.zipFilePath)
		if err != nil {
			return fmt.Errorf("failed to open existing zip file: %w", err)
		}
		defer existingZipReader.Close()
	}

	// Create a temp file for the new zip content
	tempZipFile, tempZipPath, err := CreateTempFile(z.zipFilePath, "temp_", ".zip")
	if err != nil {
		return fmt.Errorf("failed to create temp zip file: %w", err)
	}
	defer os.Remove(tempZipPath) // Clean up in case of failure

	// Create a new zip writer
	zipWriter := zip.NewWriter(tempZipFile)

	// If the zip existed, copy all files from the original except the one we're updating
	if zipExists {
		for _, file := range existingZipReader.File {
			if file.Name == key {
				// Skip, we'll add an updated version later
				continue
			}

			// Open the file from the original zip
			fileReader, err := file.Open()
			if err != nil {
				zipWriter.Close()
				tempZipFile.Close()
				return fmt.Errorf("failed to open file in original zip: %w", err)
			}

			// Create a new entry in the new zip
			writer, err := zipWriter.Create(file.Name)
			if err != nil {
				fileReader.Close()
				zipWriter.Close()
				tempZipFile.Close()
				return fmt.Errorf("failed to create file in new zip: %w", err)
			}

			// Copy the content with size limit
			const maxSize = 100 * 1024 * 1024 // 100MB per file
			limitedReader := io.LimitReader(fileReader, maxSize)
			written, err := io.Copy(writer, limitedReader)
			if err != nil {
				fileReader.Close()
				zipWriter.Close()
				tempZipFile.Close()
				return fmt.Errorf("failed to copy file content: %w", err)
			}
			if written >= maxSize {
				fileReader.Close()
				zipWriter.Close()
				tempZipFile.Close()
				return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
			}

			fileReader.Close()
		}
	}

	// Add the new/updated file
	writer, err := zipWriter.Create(key)
	if err != nil {
		zipWriter.Close()
		tempZipFile.Close()
		return fmt.Errorf("failed to create new file in zip: %w", err)
	}

	if _, err := writer.Write(content); err != nil {
		zipWriter.Close()
		tempZipFile.Close()
		return fmt.Errorf("failed to write content to zip: %w", err)
	}

	// Finalize the zip file
	if err := zipWriter.Close(); err != nil {
		tempZipFile.Close()
		return fmt.Errorf("failed to close zip writer: %w", err)
	}
	tempZipFile.Close()

	// Replace the original zip with the new one using the utility function
	if err := SafeReplace(tempZipPath, z.zipFilePath); err != nil {
		return fmt.Errorf("failed to replace original zip file: %w", err)
	}

	return nil
}

// NewZipObjectStorage creates a new object storage backend using a zip file
// as the storage medium. It treats the zip file as a virtual file system.
func NewZipObjectStorage(zipFilePath string) *ZipObjectStorage {
	// Make sure we have an absolute path
	absPath, err := filepath.Abs(zipFilePath)
	if err == nil {
		zipFilePath = absPath
	}

	return &ZipObjectStorage{
		zipFilePath: zipFilePath,
		mu:          sync.RWMutex{},
	}
}
