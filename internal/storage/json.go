package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadJSON reads and unmarshals JSON from a file path.
// This is a read-only operation, so no locking is needed.
func ReadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("storage: file not found %s: %w", path, err)
		}
		return fmt.Errorf("storage: failed to read file %s: %w", path, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("storage: failed to unmarshal JSON from %s: %w", path, err)
	}

	return nil
}

// ReadJSONAtomic is an alias for ReadJSON since reads don't need locking.
// This function exists for API consistency.
func ReadJSONAtomic(path string, v interface{}) error {
	return ReadJSON(path, v)
}

// WriteJSON writes JSON to a file using the atomic write protocol.
// This is a convenience wrapper around WriteJSONAtomic.
func WriteJSON(path string, v interface{}) error {
	return WriteJSONAtomic(path, v)
}

// EnsureDir ensures that the directory containing the given file path exists.
// It creates all necessary parent directories with 0755 permissions.
func EnsureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("storage: failed to create directory %s: %w", dir, err)
	}
	return nil
}
