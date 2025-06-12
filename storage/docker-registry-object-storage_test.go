package storage

import (
	"context"
	"io"
	"testing"
)

// TestDockerRegistryObjectStorage_Usage demonstrates how to use the Docker registry storage
func TestDockerRegistryObjectStorage_Usage(t *testing.T) {
	// Skip the test during automated runs to avoid network dependencies
	t.Skip("This test requires network access to Docker Hub and is for demonstration only")

	// Create a new Docker Registry storage for the official Ubuntu image
	storage := NewDockerRegistryObjectStorage(
		"https://registry.hub.docker.com",
		"library/ubuntu",
		"latest")

	// Create a context for our operations
	ctx := context.Background()

	// Example 1: List files in the root directory
	objects, prefixes, err := storage.List(ctx, "/")
	if err != nil {
		t.Logf("Error listing root: %v", err)
		// Continue with other examples
	} else {
		t.Logf("Found %d objects and %d prefixes in root", len(objects), len(prefixes))
		for _, obj := range objects {
			t.Logf("Object: %s, Size: %d bytes", obj.Key, obj.Size)
		}
		for _, prefix := range prefixes {
			t.Logf("Prefix: %s", prefix.Key)
		}
	}

	// Example 2: Check if a file exists
	exists, err := storage.Exists(ctx, "/etc/passwd")
	if err != nil {
		t.Logf("Error checking if /etc/passwd exists: %v", err)
	} else {
		t.Logf("File /etc/passwd exists: %v", exists)
	}

	// Example 3: Download a file
	reader, err := storage.Download(ctx, "/etc/passwd")
	if err != nil {
		t.Logf("Error downloading /etc/passwd: %v", err)
	} else {
		defer reader.Close()
		content, err := io.ReadAll(reader)
		if err != nil {
			t.Logf("Error reading content: %v", err)
		} else {
			t.Logf("Downloaded content: %s", string(content))
		}
	}

	// Example 4: Try to use unsupported operations
	err = storage.Upload(ctx, "/test", nil, nil)
	if err != nil {
		t.Logf("Expected error for Upload: %v", err)
	}

	err = storage.Delete(ctx, "/test")
	if err != nil {
		t.Logf("Expected error for Delete: %v", err)
	}
}

// TestDockerRegistryObjectStorage_Errors tests various error conditions
func TestDockerRegistryObjectStorage_Errors(t *testing.T) {
	t.Skip("This test requires network access to Docker Hub and is for demonstration only")

	// Create storage with invalid repository
	invalidStorage := NewDockerRegistryObjectStorage(
		"https://registry.hub.docker.com",
		"non-existent/image",
		"invalid")

	ctx := context.Background()

	// This should fail with an authentication or not found error
	_, _, err := invalidStorage.List(ctx, "/")
	if err == nil {
		t.Error("Expected error for invalid repository, but got nil")
	} else {
		t.Logf("Got expected error for invalid repository: %v", err)
	}

	// Test with empty key
	_, err = invalidStorage.Download(ctx, "")
	if err == nil || err != ErrEmptyKey {
		t.Errorf("Expected ErrEmptyKey error, got: %v", err)
	}
}
