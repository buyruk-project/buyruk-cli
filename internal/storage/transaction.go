package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TransactionLog represents a transaction log entry.
type TransactionLog struct {
	Operation string                 `json:"operation"`
	Timestamp string                 `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// BeginTransaction creates a transaction log entry before any file modification.
func BeginTransaction(projectKey, operation string, metadata map[string]interface{}) error {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return err
	}

	// Ensure project directory exists
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("storage: failed to create project directory: %w", err)
	}

	transactionPath := filepath.Join(projectDir, ".buyruk_pending")

	transaction := TransactionLog{
		Operation: operation,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metadata:  metadata,
	}

	data, err := json.MarshalIndent(transaction, "", "  ")
	if err != nil {
		return fmt.Errorf("storage: failed to marshal transaction log: %w", err)
	}

	// Use atomic write for the transaction log itself
	tmpPath := transactionPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("storage: failed to write transaction log: %w", err)
	}

	if err := os.Rename(tmpPath, transactionPath); err != nil {
		return fmt.Errorf("storage: failed to rename transaction log: %w", err)
	}

	return nil
}

// CommitTransaction removes the transaction log after successful operation.
func CommitTransaction(projectKey string) error {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return err
	}

	transactionPath := filepath.Join(projectDir, ".buyruk_pending")
	if err := os.Remove(transactionPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("storage: failed to remove transaction log: %w", err)
		}
		// If it doesn't exist, that's fine - transaction already committed
	}

	return nil
}

// RollbackTransaction removes the transaction log when an operation fails.
// This is the same as CommitTransaction but semantically different.
func RollbackTransaction(projectKey string) error {
	return CommitTransaction(projectKey)
}

// CheckPendingTransaction checks if there's a pending transaction and returns it.
// Returns true if a pending transaction exists, along with the transaction data.
func CheckPendingTransaction(projectKey string) (bool, map[string]interface{}, error) {
	projectDir, err := ProjectDir(projectKey)
	if err != nil {
		return false, nil, err
	}

	transactionPath := filepath.Join(projectDir, ".buyruk_pending")
	_, err = os.Stat(transactionPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("storage: failed to check transaction log: %w", err)
	}

	// Read and parse the transaction log
	data, err := os.ReadFile(transactionPath)
	if err != nil {
		return false, nil, fmt.Errorf("storage: failed to read transaction log: %w", err)
	}

	var transaction TransactionLog
	if err := json.Unmarshal(data, &transaction); err != nil {
		return false, nil, fmt.Errorf("storage: failed to unmarshal transaction log: %w", err)
	}

	return true, transaction.Metadata, nil
}
