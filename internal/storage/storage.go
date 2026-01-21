package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	// cachedConfigDir stores the resolved config directory
	// It's safe to cache this as it's read-only after initialization
	cachedConfigDir string

	// userConfigDirFunc is a variable that holds the function to get user config directory.
	// This allows us to swap it in tests. Defaults to os.UserConfigDir.
	userConfigDirFunc = os.UserConfigDir
)

// ConfigDir returns the base config directory using os.UserConfigDir().
// The result is cached after the first call.
func ConfigDir() (string, error) {
	if cachedConfigDir != "" {
		return cachedConfigDir, nil
	}

	baseDir, err := userConfigDirFunc()
	if err != nil {
		return "", fmt.Errorf("storage: failed to get user config dir: %w", err)
	}

	cachedConfigDir = filepath.Join(baseDir, "buyruk")
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
	return filepath.Join(issuesDir, cleanID+".json"), nil
}

// EpicPath returns the individual epic file path for the given project key and epic ID.
func EpicPath(projectKey, epicID string) (string, error) {
	epicsDir, err := EpicsDir(projectKey)
	if err != nil {
		return "", err
	}

	// Clean the epic ID to prevent path traversal
	cleanID := filepath.Clean(epicID)
	return filepath.Join(epicsDir, cleanID+".json"), nil
}

// ConfigFilePath returns the config.json path.
func ConfigFilePath() (string, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "config.json"), nil
}
