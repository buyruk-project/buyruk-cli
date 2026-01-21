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
// Uses atomic file creation (O_CREATE|O_EXCL) to prevent race conditions.
func AcquireLock(projectKey string) (func(), error) {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return nil, err
	}

	lockPath := filepath.Join(projectDir, ".buyruk.lock")

	// Try to create lock file atomically, waiting up to 5 seconds if it already exists
	pid := fmt.Sprintf("%d", os.Getpid())
	timeout := 5 * time.Second
	deadline := time.Now().Add(timeout)
	checkInterval := 100 * time.Millisecond

	for {
		// Use O_CREATE|O_EXCL for atomic test-and-set semantics
		// This ensures only one process can create the file
		f, err := os.OpenFile(lockPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err == nil {
			// Successfully created lock file
			_, writeErr := f.Write([]byte(pid))
			closeErr := f.Close()
			if writeErr != nil {
				os.Remove(lockPath)
				return nil, fmt.Errorf("storage: failed to write to lock file: %w", writeErr)
			}
			if closeErr != nil {
				os.Remove(lockPath)
				return nil, fmt.Errorf("storage: failed to close lock file: %w", closeErr)
			}
			// Return cleanup function
			return func() {
				os.Remove(lockPath)
			}, nil
		}

		// If file already exists, wait and retry
		if !os.IsExist(err) {
			// Some other error occurred
			return nil, fmt.Errorf("storage: failed to create lock file: %w", err)
		}

		// Check if we've exceeded the timeout
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("storage: lock timeout after %v", timeout)
		}

		// Wait before retrying
		time.Sleep(checkInterval)
	}
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
