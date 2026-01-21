package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteAtomic writes data to a file atomically using the temp file and rename pattern.
// This function does NOT handle locking - it should be called from within a locked context.
func WriteAtomic(path string, data []byte) error {
	// Ensure parent directory exists
	if err := EnsureDir(path); err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("storage: failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on error
		os.Remove(tmpPath)
		return fmt.Errorf("storage: failed to rename temp file: %w", err)
	}

	return nil
}

// WriteJSONAtomic writes a JSON-serializable value to a file atomically.
// This function handles the full atomic protocol: lock, transaction, write, commit.
// It extracts the project key from the file path.
func WriteJSONAtomic(path string, v interface{}) error {
	// Extract project key from path
	// Path format: [ConfigDir]/projects/[projectKey]/...
	projectKey, err := extractProjectKeyFromPath(path)
	if err != nil {
		return fmt.Errorf("storage: failed to extract project key from path: %w", err)
	}

	// Step 1: Acquire lock
	cleanup, err := AcquireLock(projectKey)
	if err != nil {
		return err
	}
	defer cleanup()

	// Step 2: Begin transaction
	if err := BeginTransaction(projectKey, "write_json", map[string]interface{}{
		"file": path,
	}); err != nil {
		return err
	}
	defer RollbackTransaction(projectKey)

	// Step 3: Marshal JSON
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("storage: failed to marshal JSON: %w", err)
	}

	// Step 4: Write atomically
	if err := WriteAtomic(path, data); err != nil {
		return err
	}

	// Step 5: Commit transaction
	if err := CommitTransaction(projectKey); err != nil {
		return err
	}

	return nil
}

// extractProjectKeyFromPath extracts the project key from a file path.
// Expected path format: [ConfigDir]/projects/[projectKey]/...
func extractProjectKeyFromPath(path string) (string, error) {
	// Normalize the path
	cleanPath := filepath.Clean(path)

	// Split the path into components
	parts := strings.Split(cleanPath, string(filepath.Separator))

	// Find the "projects" directory index
	projectsIndex := -1
	for i, part := range parts {
		if part == "projects" {
			projectsIndex = i
			break
		}
	}

	if projectsIndex == -1 || projectsIndex+1 >= len(parts) {
		return "", fmt.Errorf("storage: invalid path format, expected projects/[projectKey]/...")
	}

	return parts[projectsIndex+1], nil
}
