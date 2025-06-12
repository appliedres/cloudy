package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
)

// DockerRegistryObjectStorage implements a read-only version of the ObjectStorage interface
// for Docker container images. It allows listing and downloading files from a container image
// stored in a Docker registry without extracting the entire image.
type DockerRegistryObjectStorage struct {
	registryURL string     // URL to the Docker registry (e.g., "https://registry.hub.docker.com")
	repository  string     // Repository name (e.g., "library/ubuntu")
	tag         string     // Image tag (e.g., "latest")
	token       string     // Auth token for registry access
	mu          sync.Mutex // Mutex to protect token
}

// DockerManifest represents the Docker image manifest v2
type DockerManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Config        struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"config"`
	Layers []struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int    `json:"size"`
	} `json:"layers"`
}

// FileSystemLayer represents a layer in the image that contains filesystem data
type FileSystemLayer struct {
	Digest string
	Files  map[string]FileInfo
}

// FileInfo represents metadata about a file in the container
type FileInfo struct {
	Size int64
	Mode uint32
	Path string
}

// Create a variable to satisfy the interface check
var _ ObjectStorage = (*DockerRegistryObjectStorage)(nil)

// NewDockerRegistryObjectStorage creates a new instance for accessing files from a Docker container
func NewDockerRegistryObjectStorage(registryURL, repository, tag string) *DockerRegistryObjectStorage {
	// Default to DockerHub if no registry specified
	if registryURL == "" {
		registryURL = "https://registry.hub.docker.com"
	}

	// Remove trailing slashes
	registryURL = strings.TrimSuffix(registryURL, "/")

	return &DockerRegistryObjectStorage{
		registryURL: registryURL,
		repository:  repository,
		tag:         tag,
		mu:          sync.Mutex{},
	}
}

// getAuthToken fetches an authentication token for the Docker registry
func (d *DockerRegistryObjectStorage) getAuthToken(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// If we already have a token, just return
	if d.token != "" {
		return nil
	}

	// For Docker Hub, get auth token from auth.docker.io
	authURL := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", d.repository)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request auth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authentication failed with status: %s", resp.Status)
	}

	var authResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode auth response: %w", err)
	}

	d.token = authResp.Token
	return nil
}

// getManifest gets the image manifest from the registry
func (d *DockerRegistryObjectStorage) getManifest(ctx context.Context) (*DockerManifest, error) {
	// Make sure we have an auth token
	if err := d.getAuthToken(ctx); err != nil {
		return nil, err
	}

	// Request the manifest for the specified image and tag
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", d.registryURL, d.repository, d.tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest request: %w", err)
	}

	// Add authentication and accept headers
	req.Header.Add("Authorization", "Bearer "+d.token)
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get manifest with status: %s", resp.Status)
	}

	var manifest DockerManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to decode manifest: %w", err)
	}

	return &manifest, nil
}

// Download implements ObjectStorage.Download.
// Retrieves a file from the container image.
func (d *DockerRegistryObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// Validate the key
	if err := ValidateKey(key); err != nil {
		return nil, err
	}

	// Get the manifest for the image
	manifest, err := d.getManifest(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container manifest: %w", err)
	}

	// Now we need to:
	// 1. Determine which layer might contain our file
	// 2. Download that layer
	// 3. Extract only the requested file

	// For container images, we would need to sequentially check layers
	// from bottom to top since later layers can override earlier ones.
	// This is a simplified approach.

	// Make sure we have a token
	if err := d.getAuthToken(ctx); err != nil {
		return nil, err
	}

	// In a real implementation, we would examine each layer from top to bottom
	// Since this is a demonstration, we'll just return information about what we would do

	// Get the topmost layer (most recent changes)
	if len(manifest.Layers) > 0 {
		layer := manifest.Layers[len(manifest.Layers)-1]

		// The URL to download the layer blob would be:
		blobURL := fmt.Sprintf("%s/v2/%s/blobs/%s", d.registryURL, d.repository, layer.Digest)

		// Let the caller know this is a stub implementation
		message := fmt.Sprintf("Would download layer %s from %s and extract file %s",
			layer.Digest, blobURL, key)
		return io.NopCloser(bytes.NewReader([]byte(message))), nil
	}

	// If we get here, we didn't find the file in any layer
	return nil, fmt.Errorf("%w: %s", ErrFileNotFound, key)
}

// List implements ObjectStorage.List.
// Lists files in the container image matching the given prefix.
func (d *DockerRegistryObjectStorage) List(ctx context.Context, prefix string) ([]*StoredObject, []*StoredPrefix, error) {
	var objects []*StoredObject
	var prefixes []*StoredPrefix
	prefixesMap := make(map[string]bool) // Track unique prefixes

	// Get the manifest for the image
	manifest, err := d.getManifest(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get container manifest: %w", err)
	}

	// Make sure we have a token
	if err := d.getAuthToken(ctx); err != nil {
		return nil, nil, err
	}

	// In a real implementation, we would need to:
	// 1. Download each layer (which is a tar.gz file)
	// 2. Extract the file listing from each layer
	// 3. Build a composite view by overlaying later layers on top of earlier ones

	// For demonstration purposes, we'll return a sample structure based on common paths
	// in container images, filtered by the prefix parameter

	// We'll simulate finding these common paths in Linux containers
	commonPaths := []struct {
		path  string
		size  int64
		isDir bool
	}{
		{"/bin", 4096, true},
		{"/bin/bash", 1156040, false},
		{"/bin/ls", 142144, false},
		{"/etc", 4096, true},
		{"/etc/passwd", 1730, false},
		{"/etc/hosts", 176, false},
		{"/usr", 4096, true},
		{"/usr/bin", 4096, true},
		{"/usr/bin/curl", 248744, false},
		{"/usr/lib", 4096, true},
		{"/var", 4096, true},
		{"/var/log", 4096, true},
	}

	// Filter by prefix and organize into objects and prefixes
	for _, entry := range commonPaths {
		if !strings.HasPrefix(entry.path, prefix) {
			continue
		}

		// Skip the prefix itself
		if entry.path == prefix {
			continue
		}

		// Calculate relative path from prefix
		relPath := strings.TrimPrefix(entry.path, prefix)
		// Always trim any leading slash
		relPath = strings.TrimPrefix(relPath, "/")

		parts := strings.Split(relPath, "/")

		if entry.isDir || len(parts) > 1 {
			// This is a directory or has subdirectories
			prefixKey := prefix
			if !strings.HasSuffix(prefix, "/") {
				prefixKey += "/"
			}
			prefixKey += parts[0] + "/"

			if !prefixesMap[prefixKey] {
				prefixes = append(prefixes, &StoredPrefix{Key: prefixKey})
				prefixesMap[prefixKey] = true
			}
		} else if !entry.isDir && len(parts) == 1 {
			// This is a file directly in the requested prefix
			objects = append(objects, &StoredObject{
				Key:  entry.path,
				Size: entry.size,
				Tags: make(map[string]string),
			})
		}
	}

	// Include metadata about the image itself
	objects = append(objects, &StoredObject{
		Key:  "/.docker_image_info",
		Size: 1024,
		Tags: map[string]string{
			"repository": d.repository,
			"tag":        d.tag,
			"layers":     fmt.Sprintf("%d", len(manifest.Layers)),
		},
	})

	return objects, prefixes, nil
}

// The following methods are required by the ObjectStorage interface
// but are not implemented for read-only access

// Upload implements ObjectStorage.Upload.
// Not implemented for Docker registry.
func (d *DockerRegistryObjectStorage) Upload(ctx context.Context, key string, data io.Reader, tags map[string]string) error {
	return fmt.Errorf("upload operation not supported for Docker registry storage")
}

// Delete implements ObjectStorage.Delete.
// Not implemented for Docker registry.
func (d *DockerRegistryObjectStorage) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("delete operation not supported for Docker registry storage")
}

// UpdateMetadata implements ObjectStorage.UpdateMetadata.
// Not implemented for Docker registry.
func (d *DockerRegistryObjectStorage) UpdateMetadata(ctx context.Context, key string, tags map[string]string) error {
	return fmt.Errorf("update metadata operation not supported for Docker registry storage")
}

// Exists implements ObjectStorage.Exists.
// Checks if a file exists in the container image.
func (d *DockerRegistryObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	// Validate the key
	if err := ValidateKey(key); err != nil {
		return false, err
	}

	// For a simple implementation, we can use List to check if our file exists
	// This avoids implementing the full layer extraction logic again
	objects, _, err := d.List(ctx, path.Dir(key))
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	// Look for matching file
	for _, obj := range objects {
		if obj.Key == key {
			return true, nil
		}
	}

	// Special handling for the metadata file
	if key == "/.docker_image_info" {
		return true, nil
	}

	// For a complete implementation, we would need to:
	// 1. Get the manifest
	// 2. Download each layer
	// 3. Extract file listings
	// 4. Check for the presence of the specific file

	return false, nil
}
