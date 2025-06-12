package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Common constants used across storage implementations
const (
	// MaxFileSize is the maximum allowed file size to prevent decompression bombs
	MaxFileSize = 100 * 1024 * 1024 // 100MB limit per file
)

// Common errors
var (
	ErrEmptyKey       = errors.New("key cannot be empty")
	ErrFileNotFound   = errors.New("file not found")
	ErrFileTooLarge   = errors.New("file size exceeds maximum allowed size")
	ErrInvalidStorage = errors.New("invalid storage configuration")
)

// FileEntry represents a generic file entry that can be used across storage implementations
type FileEntry struct {
	Name    string // Path of the file
	Size    int64  // Size of the file in bytes
	IsDir   bool   // Whether this entry is a directory
	Content []byte // File content
}

// ValidateKey checks if the key is valid (not empty)
func ValidateKey(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	return nil
}

// IsDirectory checks if a key represents a directory (ends with '/')
func IsDirectory(key string) bool {
	return strings.HasSuffix(key, "/")
}

// CreateTempFile creates a temporary file in the same directory as the specified path
// with the given prefix and extension
func CreateTempFile(basePath, prefix, ext string) (*os.File, string, error) {
	dir := filepath.Dir(basePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tempFile, err := os.CreateTemp(dir, prefix+ext)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp file: %w", err)
	}

	return tempFile, tempFile.Name(), nil
}

// SafeReplace safely replaces the destination file with the source file
// It ensures the directory exists and handles any potential errors
func SafeReplace(srcPath, destPath string) error {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Replace the original file with the temp file
	if err := os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("failed to replace file: %w", err)
	}

	return nil
}

// ReadLimitedContent reads content from a reader with a size limit
// Returns the content and an error if the content exceeds the size limit
func ReadLimitedContent(reader io.Reader) ([]byte, error) {
	limitedReader := io.LimitReader(reader, MaxFileSize)
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	if len(content) >= MaxFileSize {
		return nil, fmt.Errorf("%w: %d bytes", ErrFileTooLarge, MaxFileSize)
	}

	return content, nil
}

// SplitPrefixPath splits a path into components useful for listing operations
// Returns the filename and containing directory parts
func SplitPrefixPath(path, prefix string) (string, []string) {
	// Extract the portion after the prefix
	remainingPath := strings.TrimPrefix(path, prefix)
	if remainingPath == "" {
		return "", nil
	}

	// Split the remaining path by directories
	return remainingPath, strings.Split(remainingPath, "/")
}

// CheckFileExists checks if a file exists at the specified path
func CheckFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// AbsPath returns the absolute path for a given path
// If the conversion fails, it returns the original path
func AbsPath(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absPath
}
