package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTarObjectStorage_BasicOperations(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "tar_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test both regular and gzipped tar files
	tarTypes := []struct {
		name     string
		filename string
	}{
		{"regular", filepath.Join(tempDir, "test.tar")},
		{"gzipped", filepath.Join(tempDir, "test.tar.gz")},
	}

	for _, tt := range tarTypes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := NewTarObjectStorage(tt.filename)

			// Test file doesn't exist yet
			exists, err := storage.Exists(ctx, "test.txt")
			if err != nil {
				t.Fatalf("Error checking existence: %v", err)
			}
			if exists {
				t.Error("File should not exist yet")
			}

			// Test upload
			content := "Hello, World!"
			err = storage.Upload(ctx, "test.txt", bytes.NewReader([]byte(content)), nil)
			if err != nil {
				t.Fatalf("Error uploading: %v", err)
			}

			// Test exists
			exists, err = storage.Exists(ctx, "test.txt")
			if err != nil {
				t.Fatalf("Error checking existence: %v", err)
			}
			if !exists {
				t.Error("File should exist now")
			}

			// Test download
			reader, err := storage.Download(ctx, "test.txt")
			if err != nil {
				t.Fatalf("Error downloading: %v", err)
			}
			data, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("Error reading: %v", err)
			}
			reader.Close()

			if string(data) != content {
				t.Errorf("Expected %q but got %q", content, string(data))
			}

			// Test update content
			updatedContent := "Updated content"
			err = storage.Upload(ctx, "test.txt", bytes.NewReader([]byte(updatedContent)), nil)
			if err != nil {
				t.Fatalf("Error updating: %v", err)
			}

			// Verify update
			reader, err = storage.Download(ctx, "test.txt")
			if err != nil {
				t.Fatalf("Error downloading after update: %v", err)
			}
			data, err = io.ReadAll(reader)
			if err != nil {
				t.Fatalf("Error reading after update: %v", err)
			}
			reader.Close()

			if string(data) != updatedContent {
				t.Errorf("Expected %q but got %q", updatedContent, string(data))
			}

			// Upload a second file
			err = storage.Upload(ctx, "folder/nested.txt", bytes.NewReader([]byte("Nested content")), nil)
			if err != nil {
				t.Fatalf("Error uploading nested file: %v", err)
			}

			// Test list
			objects, prefixes, err := storage.List(ctx, "")
			if err != nil {
				t.Fatalf("Error listing: %v", err)
			}

			// Should have one object and one prefix
			if len(objects) != 1 {
				t.Errorf("Expected 1 object, got %d", len(objects))
			}
			if len(prefixes) != 1 {
				t.Errorf("Expected 1 prefix, got %d", len(prefixes))
			}

			if len(objects) > 0 && objects[0].Key != "test.txt" {
				t.Errorf("Expected object key 'test.txt', got %q", objects[0].Key)
			}
			if len(prefixes) > 0 && prefixes[0].Key != "folder/" {
				t.Errorf("Expected prefix key 'folder/', got %q", prefixes[0].Key)
			}

			// Test list with prefix
			objects, prefixes, err = storage.List(ctx, "folder/")
			if err != nil {
				t.Fatalf("Error listing with prefix: %v", err)
			}

			// Should have one object and no prefixes in the folder
			if len(objects) != 1 {
				t.Errorf("Expected 1 object in folder, got %d", len(objects))
			}
			if len(prefixes) != 0 {
				t.Errorf("Expected 0 prefixes in folder, got %d", len(prefixes))
			}

			if len(objects) > 0 && objects[0].Key != "folder/nested.txt" {
				t.Errorf("Expected object key 'folder/nested.txt', got %q", objects[0].Key)
			}

			// Test delete
			err = storage.Delete(ctx, "test.txt")
			if err != nil {
				t.Fatalf("Error deleting: %v", err)
			}

			// Verify delete
			exists, err = storage.Exists(ctx, "test.txt")
			if err != nil {
				t.Fatalf("Error checking existence after delete: %v", err)
			}
			if exists {
				t.Error("File should have been deleted")
			}

			// Test UpdateMetadata (which mostly just checks for existence)
			err = storage.UpdateMetadata(ctx, "folder/nested.txt", map[string]string{"key": "value"})
			if err != nil {
				t.Fatalf("Error updating metadata: %v", err)
			}

			// Try to update metadata for a non-existent file
			err = storage.UpdateMetadata(ctx, "nonexistent.txt", nil)
			if err == nil || !strings.Contains(err.Error(), "not found") {
				t.Errorf("Expected 'not found' error for UpdateMetadata on non-existent file, got: %v", err)
			}
		})
	}
}

func TestTarObjectStorage_EdgeCases(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "tar_edge_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tarPath := filepath.Join(tempDir, "edge.tar")
	ctx := context.Background()
	storage := NewTarObjectStorage(tarPath)

	// Test empty key
	_, err = storage.Download(ctx, "")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error for empty key in Download, got: %v", err)
	}

	_, err = storage.Exists(ctx, "")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error for empty key in Exists, got: %v", err)
	}

	err = storage.Delete(ctx, "")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error for empty key in Delete, got: %v", err)
	}

	err = storage.UpdateMetadata(ctx, "", nil)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected error for empty key in UpdateMetadata, got: %v", err)
	}

	// Test non-existent file operations
	_, err = storage.Download(ctx, "nonexistent.txt")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error for Download, got: %v", err)
	}

	err = storage.Delete(ctx, "nonexistent.txt")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error for Delete, got: %v", err)
	}

	// Test directory operations
	err = storage.Upload(ctx, "dir/", bytes.NewReader([]byte{}), nil)
	if err != nil {
		t.Fatalf("Error creating directory: %v", err)
	}

	exists, err := storage.Exists(ctx, "dir/")
	if err != nil {
		t.Fatalf("Error checking directory existence: %v", err)
	}
	if !exists {
		t.Error("Directory should exist")
	}

	// Upload a file in the directory
	err = storage.Upload(ctx, "dir/file.txt", bytes.NewReader([]byte("File in directory")), nil)
	if err != nil {
		t.Fatalf("Error uploading file to directory: %v", err)
	}

	// List the directory contents
	objects, _, err := storage.List(ctx, "dir/")
	if err != nil {
		t.Fatalf("Error listing directory: %v", err)
	}
	if len(objects) != 1 {
		t.Errorf("Expected 1 object in directory, got %d", len(objects))
	}
}

// Test for handling files with the same prefix but different paths
func TestTarObjectStorage_PrefixHandling(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "tar_prefix_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tarPath := filepath.Join(tempDir, "prefix.tar.gz")
	ctx := context.Background()
	storage := NewTarObjectStorage(tarPath)

	// Upload files with similar prefixes
	err = storage.Upload(ctx, "folder/file.txt", bytes.NewReader([]byte("File in folder")), nil)
	if err != nil {
		t.Fatalf("Error uploading file: %v", err)
	}

	err = storage.Upload(ctx, "folder-extra/file.txt", bytes.NewReader([]byte("File in folder-extra")), nil)
	if err != nil {
		t.Fatalf("Error uploading file: %v", err)
	}

	// List with empty prefix (should show both folders)
	objects, prefixes, err := storage.List(ctx, "")
	if err != nil {
		t.Fatalf("Error listing with empty prefix: %v", err)
	}
	if len(prefixes) != 2 {
		t.Errorf("Expected 2 prefixes, got %d", len(prefixes))
	}

	// List with folder/ prefix (should only show folder/file.txt)
	objects, prefixes, err = storage.List(ctx, "folder/")
	if err != nil {
		t.Fatalf("Error listing with folder/ prefix: %v", err)
	}
	if len(objects) != 1 {
		t.Errorf("Expected 1 object, got %d", len(objects))
	}
	if len(objects) > 0 && objects[0].Key != "folder/file.txt" {
		t.Errorf("Expected object key 'folder/file.txt', got %q", objects[0].Key)
	}

	// List with folder-extra/ prefix (should only show folder-extra/file.txt)
	objects, prefixes, err = storage.List(ctx, "folder-extra/")
	if err != nil {
		t.Fatalf("Error listing with folder-extra/ prefix: %v", err)
	}
	if len(objects) != 1 {
		t.Errorf("Expected 1 object, got %d", len(objects))
	}
	if len(objects) > 0 && objects[0].Key != "folder-extra/file.txt" {
		t.Errorf("Expected object key 'folder-extra/file.txt', got %q", objects[0].Key)
	}
}
