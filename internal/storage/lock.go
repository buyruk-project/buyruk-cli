package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AcquireLock acquires a lock for the given project key.
// It returns a cleanup function that must be called to release the lock.
// The function will wait up to 5 seconds for an existing lock to be released.
func AcquireLock(projectKey string) (func(), error) {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return nil, err
	}

	lockPath := filepath.Join(projectDir, ".buyruk.lock")

	// Check if lock exists and wait if needed
	if err := WaitForLock(projectKey, 5*time.Second); err != nil {
		return nil, fmt.Errorf("storage: failed to acquire lock: %w", err)
	}

	// Create lock file with process ID
	pid := fmt.Sprintf("%d", os.Getpid())
	if err := os.WriteFile(lockPath, []byte(pid), 0644); err != nil {
		return nil, fmt.Errorf("storage: failed to create lock file: %w", err)
	}

	// Return cleanup function
	return func() {
		os.Remove(lockPath)
	}, nil
}

// CheckLock checks if a lock exists for the given project key.
// Returns true if lock exists, false otherwise.
func CheckLock(projectKey string) (bool, error) {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return false, err
	}

	lockPath := filepath.Join(projectDir, ".buyruk.lock")
	_, err = os.Stat(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("storage: failed to check lock: %w", err)
	}

	return true, nil
}

// WaitForLock waits for a lock to be released, checking at 100ms intervals.
// Returns an error if the lock still exists after the timeout duration.
func WaitForLock(projectKey string, timeout time.Duration) error {
	checkInterval := 100 * time.Millisecond
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		exists, err := CheckLock(projectKey)
		if err != nil {
			return err
		}

		if !exists {
			return nil
		}

		time.Sleep(checkInterval)
	}

	// Lock still exists after timeout
	return fmt.Errorf("storage: lock timeout after %v", timeout)
}
