package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ArchiveFormat represents the format of an archive file
type ArchiveFormat int

const (
	// AutoDetect will try to determine the archive format based on the file extension or content
	AutoDetect ArchiveFormat = iota
	// ZipFormat indicates a ZIP archive
	ZipFormat
	// TarFormat indicates a TAR archive (possibly gzipped)
	TarFormat
)

// ArchiveURLObjectStorage implements ObjectStorage by downloading a ZIP or TAR archive
// from a URL and then using the appropriate storage implementation to access its contents.
// The archive is downloaded once and cached for the lifecycle of this storage instance.
type ArchiveURLObjectStorage struct {
	archiveURL     string            // URL of the archive to download
	format         ArchiveFormat     // Format of the archive
	archiveStorage ObjectStorage     // The actual storage implementation (Zip or Tar)
	tempFilePath   string            // Path to the downloaded archive file
	mu             sync.RWMutex      // Protects the archiveStorage initialization
	initialized    bool              // Whether the archive has been downloaded and initialized
	headers        map[string]string // Custom headers for HTTP requests
}

// Create a variable to satisfy the interface check
var _ ObjectStorage = (*ArchiveURLObjectStorage)(nil)

// NewArchiveURLObjectStorage creates a new object storage implementation that
// downloads an archive file from a URL and provides access to its contents.
func NewArchiveURLObjectStorage(archiveURL string, format ArchiveFormat, headers map[string]string) *ArchiveURLObjectStorage {
	return &ArchiveURLObjectStorage{
		archiveURL: archiveURL,
		format:     format,
		headers:    headers,
		mu:         sync.RWMutex{},
	}
}

func (a *ArchiveURLObjectStorage) downloadArchive(ctx context.Context) (io.ReadCloser, error) {
	// Create a new HTTP client with the configured headers
	client := &http.Client{
		Timeout: 300 * time.Second, // Set a reasonable timeout for downloads
	}

	req, err := http.NewRequestWithContext(ctx, "GET", a.archiveURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers if provided
	for key, value := range a.headers {
		req.Header.Add(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download archive: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download archive: status code %d", resp.StatusCode)
	}

	// Generate a unique filename based on the URL
	fileName := fmt.Sprintf("archive_%x", a.archiveURL)
	file, err := os.CreateTemp("", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write archive to temp file: %w", err)
	}

	err = file.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close temp file: %w", err)
	}
	a.tempFilePath = file.Name()

	return os.Open(file.Name()) // Ensure the file is opened for reading
}

// initialize downloads the archive file and initializes the appropriate storage
func (a *ArchiveURLObjectStorage) initialize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// If already initialized, do nothing
	if a.initialized {
		return nil
	}

	// Download the archive file
	reader, err := a.downloadArchive(ctx)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer reader.Close()

	// Add extension based on format or detection
	format := a.format
	if format == AutoDetect {
		// Try to detect format from URL
		if strings.HasSuffix(strings.ToLower(a.tempFilePath), ".zip") {
			format = ZipFormat
		} else if strings.HasSuffix(strings.ToLower(a.tempFilePath), ".tar") {
			format = TarFormat
		} else if strings.HasSuffix(strings.ToLower(a.tempFilePath), ".tar.gz") ||
			strings.HasSuffix(strings.ToLower(a.tempFilePath), ".tgz") {
			format = TarFormat
		} else {
			// Default to zip if can't detect
			format = ZipFormat
		}
	}

	// Create the appropriate storage implementation
	if format == ZipFormat {
		a.archiveStorage = NewZipObjectStorage(a.tempFilePath)
	} else if format == TarFormat {
		a.archiveStorage = NewTarObjectStorage(a.tempFilePath)
	} else {
		// This should not happen, but handle it anyway
		return fmt.Errorf("unknown archive format")
	}

	a.initialized = true
	return nil
}

// ensureInitialized makes sure the archive is downloaded and storage is ready
func (a *ArchiveURLObjectStorage) ensureInitialized(ctx context.Context) error {
	if !a.initialized {
		return a.initialize(ctx)
	}
	return nil
}

// cleanup removes the temporary file when done
func (a *ArchiveURLObjectStorage) cleanup() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.tempFilePath != "" {
		if err := os.Remove(a.tempFilePath); err != nil {
			return fmt.Errorf("failed to remove temp file: %w", err)
		}
		a.tempFilePath = ""
	}

	a.initialized = false
	a.archiveStorage = nil

	return nil
}

// Close cleans up resources when done
func (a *ArchiveURLObjectStorage) Close() error {
	return a.cleanup()
}

// Download implements ObjectStorage.Download
func (a *ArchiveURLObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// Ensure archive is downloaded and initialized
	if err := a.ensureInitialized(ctx); err != nil {
		return nil, err
	}

	// Delegate to the archive storage
	return a.archiveStorage.Download(ctx, key)
}

// Exists implements ObjectStorage.Exists
func (a *ArchiveURLObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	// Ensure archive is downloaded and initialized
	if err := a.ensureInitialized(ctx); err != nil {
		return false, err
	}

	// Delegate to the archive storage
	return a.archiveStorage.Exists(ctx, key)
}

// List implements ObjectStorage.List
func (a *ArchiveURLObjectStorage) List(ctx context.Context, prefix string) ([]*StoredObject, []*StoredPrefix, error) {
	// Ensure archive is downloaded and initialized
	if err := a.ensureInitialized(ctx); err != nil {
		return nil, nil, err
	}

	// Delegate to the archive storage
	return a.archiveStorage.List(ctx, prefix)
}

// Upload implements ObjectStorage.Upload
// Not supported for remote archives (read-only)
func (a *ArchiveURLObjectStorage) Upload(ctx context.Context, key string, data io.Reader, tags map[string]string) error {
	return fmt.Errorf("upload operation not supported for remote archive storage (read-only)")
}

// Delete implements ObjectStorage.Delete
// Not supported for remote archives (read-only)
func (a *ArchiveURLObjectStorage) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("delete operation not supported for remote archive storage (read-only)")
}

// UpdateMetadata implements ObjectStorage.UpdateMetadata
// Not supported for remote archives (read-only)
func (a *ArchiveURLObjectStorage) UpdateMetadata(ctx context.Context, key string, tags map[string]string) error {
	return fmt.Errorf("update metadata operation not supported for remote archive storage (read-only)")
}
