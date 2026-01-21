package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	// cachedConfigDir stores the resolved config directory
	// It's safe to cache this as it's read-only after initialization
	cachedConfigDir string

	// configDirOnce ensures thread-safe initialization of cachedConfigDir
	// Using a pointer allows resetting in tests
	configDirOnce = &sync.Once{}

	// userConfigDirFunc is a variable that holds the function to get user config directory.
	// This allows us to swap it in tests. Defaults to os.UserConfigDir.
	userConfigDirFunc = os.UserConfigDir
)

// resetConfigDirCache resets the config directory cache and sync.Once.
// This is only used for testing purposes.
func resetConfigDirCache() {
	cachedConfigDir = ""
	configDirOnce = &sync.Once{}
}

// ConfigDir returns the base config directory using os.UserConfigDir().
// The result is cached after the first call in a thread-safe manner.
func ConfigDir() (string, error) {
	var initErr error
	configDirOnce.Do(func() {
		baseDir, err := userConfigDirFunc()
		if err != nil {
			initErr = fmt.Errorf("storage: failed to get user config dir: %w", err)
			return
		}
		cachedConfigDir = filepath.Join(baseDir, "buyruk")
	})

	if initErr != nil {
		return "", initErr
	}

	return cachedConfigDir, nil
}

// ProjectDir returns the project directory path for the given project key.
func ProjectDir(projectKey string) (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	// Clean the project key to prevent path traversal
	cleanKey := filepath.Clean(projectKey)
	return filepath.Join(configDir, "projects", cleanKey), nil
}

// ProjectIndexPath returns the project.json path for the given project key.
func ProjectIndexPath(projectKey string) (string, error) {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return "", err
	}

	return filepath.Join(projectDir, "project.json"), nil
}

// IssuesDir returns the issues/ directory path for the given project key.
func IssuesDir(projectKey string) (string, error) {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return "", err
	}

	return filepath.Join(projectDir, "issues"), nil
}

// EpicsDir returns the epics/ directory path for the given project key.
func EpicsDir(projectKey string) (string, error) {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return "", err
	}

	return filepath.Join(projectDir, "epics"), nil
}

// IssuePath returns the individual issue file path for the given project key and issue ID.
func IssuePath(projectKey, issueID string) (string, error) {
	issuesDir, err := IssuesDir(projectKey)
	if err != nil {
		return "", err
	}

	// Clean the issue ID to prevent path traversal
	cleanID := filepath.Clean(issueID)

	// Validate that the cleaned ID doesn't contain path separators (prevents traversal)
	if cleanID != issueID || filepath.IsAbs(cleanID) {
		return "", fmt.Errorf("storage: invalid issue ID: contains path separators or is absolute")
	}

	// Build the full path and validate it's within the issues directory
	fullPath := filepath.Join(issuesDir, cleanID+".json")

	// Use filepath.Rel to ensure the path is within the issues directory
	relPath, err := filepath.Rel(issuesDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("storage: failed to validate issue path: %w", err)
	}

	// Check if the relative path tries to escape the directory
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("storage: invalid issue ID: path traversal detected")
	}

	return fullPath, nil
}

// EpicPath returns the individual epic file path for the given project key and epic ID.
func EpicPath(projectKey, epicID string) (string, error) {
	epicsDir, err := EpicsDir(projectKey)
	if err != nil {
		return "", err
	}

	// Clean the epic ID to prevent path traversal
	cleanID := filepath.Clean(epicID)

	// Validate that the cleaned ID doesn't contain path separators (prevents traversal)
	if cleanID != epicID || filepath.IsAbs(cleanID) {
		return "", fmt.Errorf("storage: invalid epic ID: contains path separators or is absolute")
	}

	// Build the full path and validate it's within the epics directory
	fullPath := filepath.Join(epicsDir, cleanID+".json")

	// Use filepath.Rel to ensure the path is within the epics directory
	relPath, err := filepath.Rel(epicsDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("storage: failed to validate epic path: %w", err)
	}

	// Check if the relative path tries to escape the directory
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("storage: invalid epic ID: path traversal detected")
	}

	return fullPath, nil
}

// ConfigFilePath returns the config.json path.
func ConfigFilePath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}
