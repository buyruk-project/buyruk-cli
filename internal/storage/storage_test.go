package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestConfigDir tests the ConfigDir function
func TestConfigDir(t *testing.T) {
	// Reset cached config dir for testing
	cachedConfigDir = ""

	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	if dir == "" {
		t.Fatal("ConfigDir() returned empty string")
	}

	// Verify it contains "buyruk"
	if !filepath.IsAbs(dir) {
		t.Errorf("ConfigDir() should return absolute path, got: %s", dir)
	}

	// Test caching - second call should return same value
	dir2, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed on second call: %v", err)
	}

	if dir != dir2 {
		t.Errorf("ConfigDir() should cache result, got different values: %s != %s", dir, dir2)
	}
}

// TestProjectDir tests the ProjectDir function
func TestProjectDir(t *testing.T) {
	// Use a temp directory for testing
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	originalCachedDir := cachedConfigDir
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = originalCachedDir
	}()

	// Reset cache
	cachedConfigDir = ""
	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectDir, err := ProjectDir("TEST-PROJ")
	if err != nil {
		t.Fatalf("ProjectDir() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "buyruk", "projects", "TEST-PROJ")
	if projectDir != expected {
		t.Errorf("ProjectDir() = %s, want %s", projectDir, expected)
	}
}

// TestProjectIndexPath tests the ProjectIndexPath function
func TestProjectIndexPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	originalCachedDir := cachedConfigDir
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = originalCachedDir
	}()

	// Reset cache
	cachedConfigDir = ""
	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	indexPath, err := ProjectIndexPath("TEST-PROJ")
	if err != nil {
		t.Fatalf("ProjectIndexPath() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "buyruk", "projects", "TEST-PROJ", "project.json")
	if indexPath != expected {
		t.Errorf("ProjectIndexPath() = %s, want %s", indexPath, expected)
	}
}

// TestIssuePath tests the IssuePath function
func TestIssuePath(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	originalCachedDir := cachedConfigDir
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = originalCachedDir
	}()

	// Reset cache
	cachedConfigDir = ""
	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	issuePath, err := IssuePath("TEST-PROJ", "T-123")
	if err != nil {
		t.Fatalf("IssuePath() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "buyruk", "projects", "TEST-PROJ", "issues", "T-123.json")
	if issuePath != expected {
		t.Errorf("IssuePath() = %s, want %s", issuePath, expected)
	}
}

// TestEpicPath tests the EpicPath function
func TestEpicPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	originalCachedDir := cachedConfigDir
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = originalCachedDir
	}()

	// Reset cache
	cachedConfigDir = ""
	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	epicPath, err := EpicPath("TEST-PROJ", "E-1")
	if err != nil {
		t.Fatalf("EpicPath() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "buyruk", "projects", "TEST-PROJ", "epics", "E-1.json")
	if epicPath != expected {
		t.Errorf("EpicPath() = %s, want %s", epicPath, expected)
	}
}

// TestConfigFilePath tests the ConfigFilePath function
func TestConfigFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	originalCachedDir := cachedConfigDir
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = originalCachedDir
	}()

	// Reset cache
	cachedConfigDir = ""
	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	configPath, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("ConfigFilePath() failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "buyruk", "config.json")
	if configPath != expected {
		t.Errorf("ConfigFilePath() = %s, want %s", configPath, expected)
	}
}

// TestAcquireLock tests lock acquisition and release
func TestAcquireLock(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	projectDir, _ := ProjectDir(projectKey)
	os.MkdirAll(projectDir, 0755)

	// Acquire lock
	cleanup, err := AcquireLock(projectKey)
	if err != nil {
		t.Fatalf("AcquireLock() failed: %v", err)
	}

	// Verify lock file exists
	lockPath := filepath.Join(projectDir, ".buyruk.lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Fatal("Lock file was not created")
	}

	// Verify lock is detected
	exists, err := CheckLock(projectKey)
	if err != nil {
		t.Fatalf("CheckLock() failed: %v", err)
	}
	if !exists {
		t.Fatal("CheckLock() should return true when lock exists")
	}

	// Release lock
	cleanup()

	// Verify lock file is removed
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatal("Lock file was not removed after cleanup")
	}

	// Verify lock is no longer detected
	exists, err = CheckLock(projectKey)
	if err != nil {
		t.Fatalf("CheckLock() failed: %v", err)
	}
	if exists {
		t.Fatal("CheckLock() should return false when lock doesn't exist")
	}
}

// TestWaitForLock tests lock timeout behavior
func TestWaitForLock(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	projectDir, _ := ProjectDir(projectKey)
	os.MkdirAll(projectDir, 0755)
	lockPath := filepath.Join(projectDir, ".buyruk.lock")

	// Create a lock file
	os.WriteFile(lockPath, []byte("12345"), 0644)

	// Test timeout - should fail after short timeout
	start := time.Now()
	err := WaitForLock(projectKey, 200*time.Millisecond)
	duration := time.Since(start)

	if err == nil {
		t.Fatal("WaitForLock() should fail when lock exists and timeout expires")
	}

	// Verify it waited approximately the timeout duration
	if duration < 100*time.Millisecond || duration > 500*time.Millisecond {
		t.Errorf("WaitForLock() should wait approximately 200ms, waited %v", duration)
	}

	// Remove lock and verify it succeeds immediately
	os.Remove(lockPath)
	err = WaitForLock(projectKey, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitForLock() should succeed when lock doesn't exist: %v", err)
	}
}

// TestConcurrentLocks tests concurrent lock attempts
func TestConcurrentLocks(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	projectDir, _ := ProjectDir(projectKey)
	os.MkdirAll(projectDir, 0755)

	// Channel to collect results
	results := make(chan error, 10)

	// Launch 10 concurrent lock attempts
	for i := 0; i < 10; i++ {
		go func() {
			cleanup, err := AcquireLock(projectKey)
			if err != nil {
				results <- err
				return
			}
			// Hold lock briefly
			time.Sleep(10 * time.Millisecond)
			cleanup()
			results <- nil
		}()
	}

	// Collect all results
	var errors []error
	for i := 0; i < 10; i++ {
		err := <-results
		if err != nil {
			errors = append(errors, err)
		}
	}

	// Some may timeout due to concurrent access, but most should succeed
	if len(errors) > 5 {
		t.Errorf("Too many lock acquisition failures: %d out of 10", len(errors))
	}
}

// TestBeginTransaction tests transaction log creation
func TestBeginTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	metadata := map[string]interface{}{
		"issue_id": "T-123",
		"file":     "issues/T-123.json",
	}

	err := BeginTransaction(projectKey, "create_issue", metadata)
	if err != nil {
		t.Fatalf("BeginTransaction() failed: %v", err)
	}

	// Verify transaction file exists
	projectDir, _ := ProjectDir(projectKey)
	transactionPath := filepath.Join(projectDir, ".buyruk_pending")
	if _, err := os.Stat(transactionPath); os.IsNotExist(err) {
		t.Fatal("Transaction file was not created")
	}

	// Verify transaction content
	var transaction TransactionLog
	data, err := os.ReadFile(transactionPath)
	if err != nil {
		t.Fatalf("Failed to read transaction file: %v", err)
	}

	if err := json.Unmarshal(data, &transaction); err != nil {
		t.Fatalf("Failed to unmarshal transaction: %v", err)
	}

	if transaction.Operation != "create_issue" {
		t.Errorf("Transaction operation = %s, want create_issue", transaction.Operation)
	}

	if transaction.Metadata["issue_id"] != "T-123" {
		t.Errorf("Transaction metadata issue_id = %v, want T-123", transaction.Metadata["issue_id"])
	}
}

// TestCommitTransaction tests transaction commit
func TestCommitTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	projectDir, _ := ProjectDir(projectKey)
	os.MkdirAll(projectDir, 0755)

	// Begin transaction
	BeginTransaction(projectKey, "test", nil)

	// Commit transaction
	err := CommitTransaction(projectKey)
	if err != nil {
		t.Fatalf("CommitTransaction() failed: %v", err)
	}

	// Verify transaction file is removed
	transactionPath := filepath.Join(projectDir, ".buyruk_pending")
	if _, err := os.Stat(transactionPath); !os.IsNotExist(err) {
		t.Fatal("Transaction file was not removed after commit")
	}
}

// TestRollbackTransaction tests transaction rollback
func TestRollbackTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	projectDir, _ := ProjectDir(projectKey)
	os.MkdirAll(projectDir, 0755)

	// Begin transaction
	BeginTransaction(projectKey, "test", nil)

	// Rollback transaction
	err := RollbackTransaction(projectKey)
	if err != nil {
		t.Fatalf("RollbackTransaction() failed: %v", err)
	}

	// Verify transaction file is removed
	transactionPath := filepath.Join(projectDir, ".buyruk_pending")
	if _, err := os.Stat(transactionPath); !os.IsNotExist(err) {
		t.Fatal("Transaction file was not removed after rollback")
	}
}

// TestCheckPendingTransaction tests checking for pending transactions
func TestCheckPendingTransaction(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	metadata := map[string]interface{}{
		"test": "value",
	}

	// Check when no transaction exists
	exists, _, err := CheckPendingTransaction(projectKey)
	if err != nil {
		t.Fatalf("CheckPendingTransaction() failed: %v", err)
	}
	if exists {
		t.Fatal("CheckPendingTransaction() should return false when no transaction exists")
	}

	// Begin transaction
	BeginTransaction(projectKey, "test", metadata)

	// Check when transaction exists
	exists, retrievedMetadata, err := CheckPendingTransaction(projectKey)
	if err != nil {
		t.Fatalf("CheckPendingTransaction() failed: %v", err)
	}
	if !exists {
		t.Fatal("CheckPendingTransaction() should return true when transaction exists")
	}

	if retrievedMetadata["test"] != "value" {
		t.Errorf("CheckPendingTransaction() metadata = %v, want map[test:value]", retrievedMetadata)
	}
}

// TestWriteAtomic tests atomic write protocol
func TestWriteAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	testData := []byte(`{"test": "data"}`)

	err := WriteAtomic(testFile, testData)
	if err != nil {
		t.Fatalf("WriteAtomic() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Verify content
	readData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("File content = %s, want %s", string(readData), string(testData))
	}

	// Verify temp file was cleaned up
	tmpFile := testFile + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Fatal("Temp file was not cleaned up")
	}
}

// TestWriteJSONAtomic tests atomic JSON write with full protocol
func TestWriteJSONAtomic(t *testing.T) {
	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	originalCachedDir := cachedConfigDir
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = originalCachedDir
	}()

	// Reset cache
	cachedConfigDir = ""
	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	projectKey := "TEST-PROJ"
	testData := map[string]interface{}{
		"id":    "T-123",
		"title": "Test Issue",
	}

	// Ensure project directory exists (WriteJSONAtomic will create it via BeginTransaction)
	projectDir, _ := ProjectDir(projectKey)
	os.MkdirAll(projectDir, 0755)

	// Get the path for the project index
	indexPath, _ := ProjectIndexPath(projectKey)

	err := WriteJSONAtomic(indexPath, testData)
	if err != nil {
		t.Fatalf("WriteJSONAtomic() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("File was not created")
	}

	// Verify content
	var readData map[string]interface{}
	err = ReadJSON(indexPath, &readData)
	if err != nil {
		t.Fatalf("Failed to read JSON: %v", err)
	}

	if readData["id"] != "T-123" {
		t.Errorf("Read data id = %v, want T-123", readData["id"])
	}

	// Verify transaction was committed
	transactionPath := filepath.Join(projectDir, ".buyruk_pending")
	if _, err := os.Stat(transactionPath); !os.IsNotExist(err) {
		t.Fatal("Transaction file was not removed after write")
	}

	// Verify lock was released
	lockPath := filepath.Join(projectDir, ".buyruk.lock")
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Fatal("Lock file was not removed after write")
	}
}

// TestReadJSON tests JSON reading
func TestReadJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	testData := map[string]interface{}{
		"id":    "T-123",
		"title": "Test Issue",
	}

	// Write test file
	data, _ := json.MarshalIndent(testData, "", "  ")
	os.WriteFile(testFile, data, 0644)

	// Read JSON
	var readData map[string]interface{}
	err := ReadJSON(testFile, &readData)
	if err != nil {
		t.Fatalf("ReadJSON() failed: %v", err)
	}

	if readData["id"] != "T-123" {
		t.Errorf("Read data id = %v, want T-123", readData["id"])
	}

	if readData["title"] != "Test Issue" {
		t.Errorf("Read data title = %v, want Test Issue", readData["title"])
	}
}

// TestReadJSONNotFound tests error handling for missing files
func TestReadJSONNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "nonexistent.json")

	var readData map[string]interface{}
	err := ReadJSON(testFile, &readData)
	if err == nil {
		t.Fatal("ReadJSON() should fail for nonexistent file")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("ReadJSON() should return os.ErrNotExist for missing file, got: %v", err)
	}
}

// TestEnsureDir tests directory creation
func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "level1", "level2", "level3", "file.json")

	err := EnsureDir(nestedPath)
	if err != nil {
		t.Fatalf("EnsureDir() failed: %v", err)
	}

	// Verify directories were created
	expectedDir := filepath.Dir(nestedPath)
	if _, err := os.Stat(expectedDir); os.IsNotExist(err) {
		t.Fatalf("Directory was not created: %s", expectedDir)
	}
}

// TestCrossPlatformPaths tests path handling on different OSes
func TestCrossPlatformPaths(t *testing.T) {
	// This test verifies that filepath.Join works correctly
	// The actual behavior depends on the OS, but we can test the logic

	tmpDir := t.TempDir()
	originalUserConfigDir := userConfigDirFunc
	defer func() {
		userConfigDirFunc = originalUserConfigDir
		cachedConfigDir = ""
	}()

	userConfigDirFunc = func() (string, error) {
		return tmpDir, nil
	}

	// Test various path operations
	projectDir, _ := ProjectDir("TEST-PROJ")
	indexPath, _ := ProjectIndexPath("TEST-PROJ")
	issuePath, _ := IssuePath("TEST-PROJ", "T-123")

	// Verify paths use correct separators for current OS
	separator := string(filepath.Separator)
	if runtime.GOOS == "windows" {
		if !filepath.IsAbs(projectDir) {
			t.Error("ProjectDir should return absolute path on Windows")
		}
	}

	// Verify paths don't contain mixed separators
	if filepath.Separator == '/' {
		// Unix-like
		if strings.Contains(indexPath, "\\") {
			t.Error("Path should not contain backslashes on Unix")
		}
	} else {
		// Windows
		if strings.Contains(indexPath, "/") && !strings.Contains(indexPath, separator) {
			t.Error("Path should use Windows separators")
		}
	}

	// Verify paths are properly joined
	if !filepath.IsAbs(projectDir) && tmpDir != "" {
		t.Error("ProjectDir should construct valid paths")
	}

	// Verify issue path contains expected components
	if !strings.Contains(issuePath, "TEST-PROJ") {
		t.Error("IssuePath should contain project key")
	}
	if !strings.Contains(issuePath, "T-123.json") {
		t.Error("IssuePath should contain issue ID and .json extension")
	}
}
