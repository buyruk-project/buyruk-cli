package cli

import (
	"bytes"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/buyruk-project/buyruk-cli/internal/config"
	"github.com/buyruk-project/buyruk-cli/internal/models"
	"github.com/buyruk-project/buyruk-cli/internal/storage"
)

func TestNewIssueCmd(t *testing.T) {
	cmd := NewIssueCmd()
	if cmd == nil {
		t.Fatal("NewIssueCmd() returned nil")
	}
	if cmd.Use != "issue" {
		t.Errorf("Expected Use to be 'issue', got '%s'", cmd.Use)
	}

	// Check subcommands
	createCmd, _, err := cmd.Find([]string{"create"})
	if err != nil {
		t.Fatalf("create subcommand not found: %v", err)
	}
	if createCmd == nil {
		t.Fatal("create subcommand is nil")
	}
}

func TestNewIssueCreateCmd(t *testing.T) {
	cmd := NewIssueCreateCmd()
	if cmd == nil {
		t.Fatal("NewIssueCreateCmd() returned nil")
	}
	if cmd.Use != "create" {
		t.Errorf("Expected Use to be 'create', got '%s'", cmd.Use)
	}
}

func TestCreateIssue_Minimal(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create issue with only title
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{"issue", "create", "--project", projectKey, "--title", "Test Issue"})

	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)
	rootCmd2.SetOut(buf)
	rootCmd2.SetErr(errBuf)

	err := rootCmd2.Execute()
	if err != nil {
		t.Fatalf("issue create command failed: %v\nStderr: %s", err, errBuf.String())
	}

	output := buf.String()
	expectedID := projectKey + "-1"
	if !strings.Contains(output, expectedID) {
		t.Errorf("Expected output to contain issue ID %q, got: %s", expectedID, output)
	}

	// Verify issue was created
	issuePath, err := storage.IssuePath(projectKey, expectedID)
	if err != nil {
		t.Fatalf("Failed to resolve issue path: %v", err)
	}

	var issue models.Issue
	if err := storage.ReadJSON(issuePath, &issue); err != nil {
		t.Fatalf("Failed to read issue: %v", err)
	}

	if issue.ID != expectedID {
		t.Errorf("Issue ID = %q, want %q", issue.ID, expectedID)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Issue Title = %q, want 'Test Issue'", issue.Title)
	}
	if issue.Type != models.TypeTask {
		t.Errorf("Issue Type = %q, want %q (default)", issue.Type, models.TypeTask)
	}
	if issue.Status != models.StatusTODO {
		t.Errorf("Issue Status = %q, want %q (default)", issue.Status, models.StatusTODO)
	}
}

func TestCreateIssue_WithAllFields(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create issue with all fields
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--id", projectKey + "-10",
		"--type", "bug",
		"--title", "Bug Report",
		"--status", "DOING",
		"--priority", "HIGH",
		"--description", "This is a bug",
		"--epic", "E-1",
	})

	buf := new(bytes.Buffer)
	rootCmd2.SetOut(buf)

	err := rootCmd2.Execute()
	if err != nil {
		t.Fatalf("issue create command failed: %v", err)
	}

	// Verify issue was created with correct values
	issuePath, err := storage.IssuePath(projectKey, projectKey+"-10")
	if err != nil {
		t.Fatalf("Failed to resolve issue path: %v", err)
	}

	var issue models.Issue
	if err := storage.ReadJSON(issuePath, &issue); err != nil {
		t.Fatalf("Failed to read issue: %v", err)
	}

	if issue.ID != projectKey+"-10" {
		t.Errorf("Issue ID = %q, want %q", issue.ID, projectKey+"-10")
	}
	if issue.Type != models.TypeBug {
		t.Errorf("Issue Type = %q, want %q", issue.Type, models.TypeBug)
	}
	if issue.Title != "Bug Report" {
		t.Errorf("Issue Title = %q, want 'Bug Report'", issue.Title)
	}
	if issue.Status != models.StatusDOING {
		t.Errorf("Issue Status = %q, want %q", issue.Status, models.StatusDOING)
	}
	if issue.Priority != "HIGH" {
		t.Errorf("Issue Priority = %q, want HIGH", issue.Priority)
	}
	if issue.Description != "This is a bug" {
		t.Errorf("Issue Description = %q, want 'This is a bug'", issue.Description)
	}
	if issue.EpicID != "E-1" {
		t.Errorf("Issue EpicID = %q, want E-1", issue.EpicID)
	}
}

func TestCreateIssue_AutoIncrement(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create first issue (should be -1)
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{"issue", "create", "--project", projectKey, "--title", "First Issue"})
	rootCmd2.SetOut(new(bytes.Buffer))
	if err := rootCmd2.Execute(); err != nil {
		t.Fatalf("Failed to create first issue: %v", err)
	}

	// Create second issue (should be -2)
	rootCmd3 := NewRootCmd()
	rootCmd3.SetArgs([]string{"issue", "create", "--project", projectKey, "--title", "Second Issue"})
	rootCmd3.SetOut(new(bytes.Buffer))
	if err := rootCmd3.Execute(); err != nil {
		t.Fatalf("Failed to create second issue: %v", err)
	}

	// Verify both issues exist with correct IDs
	issue1Path, _ := storage.IssuePath(projectKey, projectKey+"-1")
	issue2Path, _ := storage.IssuePath(projectKey, projectKey+"-2")

	var issue1, issue2 models.Issue
	if err := storage.ReadJSON(issue1Path, &issue1); err != nil {
		t.Fatalf("Failed to read first issue: %v", err)
	}
	if err := storage.ReadJSON(issue2Path, &issue2); err != nil {
		t.Fatalf("Failed to read second issue: %v", err)
	}

	if issue1.ID != projectKey+"-1" {
		t.Errorf("First issue ID = %q, want %q", issue1.ID, projectKey+"-1")
	}
	if issue2.ID != projectKey+"-2" {
		t.Errorf("Second issue ID = %q, want %q", issue2.ID, projectKey+"-2")
	}
}

func TestCreateIssue_MissingTitle(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create issue without title
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{"issue", "create", "--project", projectKey})

	errBuf := new(bytes.Buffer)
	rootCmd2.SetErr(errBuf)

	err := rootCmd2.Execute()
	if err == nil {
		t.Fatal("issue create should fail without title")
	}

	if !strings.Contains(err.Error(), "title is required") {
		t.Errorf("Expected error about title being required, got: %v", err)
	}
}

func TestCreateIssue_NoProject(t *testing.T) {
	// Clear any existing config project
	originalCfg, _ := config.Get()
	defer func() {
		if originalCfg != nil {
			config.Save(originalCfg)
		}
	}()

	// Clear default_project
	if err := config.Set("default_project", ""); err != nil {
		t.Fatalf("Failed to clear config: %v", err)
	}

	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"issue", "create", "--title", "Test Issue"})

	errBuf := new(bytes.Buffer)
	rootCmd.SetErr(errBuf)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("issue create should fail when no project is specified")
	}

	if !strings.Contains(err.Error(), "no project specified") {
		t.Errorf("Expected error about no project specified, got: %v", err)
	}
}

func TestCreateIssue_InvalidProject(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"issue", "create", "--project", "MISSING", "--title", "Test Issue"})

	errBuf := new(bytes.Buffer)
	rootCmd.SetErr(errBuf)

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("issue create should fail when project does not exist")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error about project not existing, got: %v", err)
	}
}

func TestCreateIssue_InvalidID(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create issue with invalid ID (wrong project key)
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--id", "OTHER-1",
		"--title", "Test Issue",
	})

	errBuf := new(bytes.Buffer)
	rootCmd2.SetErr(errBuf)

	err := rootCmd2.Execute()
	if err == nil {
		t.Fatal("issue create should fail with invalid ID")
	}

	if !strings.Contains(err.Error(), "does not match project key") {
		t.Errorf("Expected error about ID not matching project key, got: %v", err)
	}
}

func TestCreateIssue_DuplicateID(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create first issue with specific ID
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--id", projectKey + "-5",
		"--title", "First Issue",
	})
	rootCmd2.SetOut(new(bytes.Buffer))
	if err := rootCmd2.Execute(); err != nil {
		t.Fatalf("Failed to create first issue: %v", err)
	}

	// Try to create second issue with same ID
	rootCmd3 := NewRootCmd()
	rootCmd3.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--id", projectKey + "-5",
		"--title", "Second Issue",
	})

	errBuf := new(bytes.Buffer)
	rootCmd3.SetErr(errBuf)

	err := rootCmd3.Execute()
	if err == nil {
		t.Fatal("issue create should fail with duplicate ID")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected error about issue already existing, got: %v", err)
	}
}

func TestGetNextIssueSequence(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// First sequence should be 1
	seq, err := getNextIssueSequence(projectKey)
	if err != nil {
		t.Fatalf("getNextIssueSequence() failed: %v", err)
	}
	if seq != 1 {
		t.Errorf("First sequence = %d, want 1", seq)
	}

	// Create an issue with ID ending in -5
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--id", projectKey + "-5",
		"--title", "Test Issue",
	})
	rootCmd2.SetOut(new(bytes.Buffer))
	if err := rootCmd2.Execute(); err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	// Next sequence should be 6
	seq, err = getNextIssueSequence(projectKey)
	if err != nil {
		t.Fatalf("getNextIssueSequence() failed: %v", err)
	}
	if seq != 6 {
		t.Errorf("Next sequence after -5 = %d, want 6", seq)
	}
}

func TestCreateIssue_ConcurrentSameID(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create the same issue concurrently
	issueID := projectKey + "-1"
	numGoroutines := 5
	var successCount int64
	var errorCount int64
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			rootCmd := NewRootCmd()
			rootCmd.SetArgs([]string{
				"issue", "create",
				"--project", projectKey,
				"--id", issueID,
				"--title", "Concurrent Issue",
			})
			rootCmd.SetOut(new(bytes.Buffer))
			rootCmd.SetErr(new(bytes.Buffer))

			err := rootCmd.Execute()
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&errorCount, 1)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Only one should succeed, others should fail with "already exists"
	finalSuccessCount := atomic.LoadInt64(&successCount)
	finalErrorCount := atomic.LoadInt64(&errorCount)
	if finalSuccessCount != 1 {
		t.Errorf("Expected exactly 1 successful creation, got %d", finalSuccessCount)
	}
	if finalErrorCount != int64(numGoroutines-1) {
		t.Errorf("Expected %d failures, got %d", numGoroutines-1, finalErrorCount)
	}

	// Verify only one issue file exists
	issuePath, _ := storage.IssuePath(projectKey, issueID)
	if _, err := os.Stat(issuePath); os.IsNotExist(err) {
		t.Fatal("Issue file was not created")
	}

	// Verify issue content is valid
	var issue models.Issue
	if err := storage.ReadJSON(issuePath, &issue); err != nil {
		t.Fatalf("Failed to read issue: %v", err)
	}

	if issue.ID != issueID {
		t.Errorf("Issue ID = %q, want %q", issue.ID, issueID)
	}
}

func TestCreateIssue_InvalidType(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create issue with invalid type
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--title", "Test Issue",
		"--type", "invalid",
	})

	errBuf := new(bytes.Buffer)
	rootCmd2.SetErr(errBuf)

	err := rootCmd2.Execute()
	if err == nil {
		t.Fatal("issue create should fail with invalid type")
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Expected error about invalid type, got: %v", err)
	}
}

func TestCreateIssue_InvalidStatus(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create issue with invalid status
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--title", "Test Issue",
		"--status", "INVALID",
	})

	errBuf := new(bytes.Buffer)
	rootCmd2.SetErr(errBuf)

	err := rootCmd2.Execute()
	if err == nil {
		t.Fatal("issue create should fail with invalid status")
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Expected error about invalid status, got: %v", err)
	}
}

func TestCreateIssue_InvalidPriority(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create issue with invalid priority
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--title", "Test Issue",
		"--priority", "INVALID",
	})

	errBuf := new(bytes.Buffer)
	rootCmd2.SetErr(errBuf)

	err := rootCmd2.Execute()
	if err == nil {
		t.Fatal("issue create should fail with invalid priority")
	}

	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("Expected error about invalid priority, got: %v", err)
	}
}

func TestCreateIssue_InvalidIDFormat(t *testing.T) {
	// Use unique project key to avoid conflicts
	projectKey := sanitizeTestName("TEST" + t.Name())
	// Clean up after test
	defer func() {
		projectDir, _ := storage.ProjectDir(projectKey)
		os.RemoveAll(projectDir)
	}()

	// Create project first
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"project", "create", projectKey})
	rootCmd.SetOut(new(bytes.Buffer))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Try to create issue with invalid ID format (no hyphen)
	rootCmd2 := NewRootCmd()
	rootCmd2.SetArgs([]string{
		"issue", "create",
		"--project", projectKey,
		"--id", "INVALIDID",
		"--title", "Test Issue",
	})

	errBuf := new(bytes.Buffer)
	rootCmd2.SetErr(errBuf)

	err := rootCmd2.Execute()
	if err == nil {
		t.Fatal("issue create should fail with invalid ID format")
	}

	if !strings.Contains(err.Error(), "invalid issue ID format") {
		t.Errorf("Expected error about invalid ID format, got: %v", err)
	}
}
