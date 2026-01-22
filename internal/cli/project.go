package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/buyruk-project/buyruk-cli/internal/config"
	"github.com/buyruk-project/buyruk-cli/internal/models"
	"github.com/buyruk-project/buyruk-cli/internal/storage"
	"github.com/spf13/cobra"
)

// NewProjectCmd creates and returns the project command.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
		Long:  "Create and manage buyruk projects",
	}

	cmd.AddCommand(NewProjectCreateCmd())
	cmd.AddCommand(NewProjectRepairCmd())

	return cmd
}

// NewProjectCreateCmd creates and returns the project create command.
func NewProjectCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <key>",
		Short: "Create a new project",
		Long:  "Create a new buyruk project with the specified key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectKey := args[0]
			return createProject(projectKey, cmd)
		},
	}

	cmd.Flags().String("name", "", "Project name (optional)")

	return cmd
}

// NewProjectRepairCmd creates and returns the project repair command.
func NewProjectRepairCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repair <key>",
		Short: "Repair project index",
		Long:  "Rebuild project.json index from issues directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectKey := args[0]
			return repairProject(projectKey, cmd)
		},
	}

	return cmd
}

// createProject creates a new project with the given key.
func createProject(projectKey string, cmd *cobra.Command) error {
	// Validate project key format
	if !isValidProjectKey(projectKey) {
		return fmt.Errorf("cli: invalid project key %q (must be uppercase alphanumeric)", projectKey)
	}

	// Check if project already exists
	projectDir, err := storage.ProjectDir(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve project directory: %w", err)
	}

	if _, err := os.Stat(projectDir); err == nil {
		return fmt.Errorf("cli: project %q already exists", projectKey)
	}

	// Get project name from flag or use key
	projectName, _ := cmd.Flags().GetString("name")
	if projectName == "" {
		projectName = projectKey
	}

	// Create project structure
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("cli: failed to create project directory: %w", err)
	}

	issuesDir, err := storage.IssuesDir(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve issues directory: %w", err)
	}

	if err := os.MkdirAll(issuesDir, 0755); err != nil {
		return fmt.Errorf("cli: failed to create issues directory: %w", err)
	}

	epicsDir, err := storage.EpicsDir(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve epics directory: %w", err)
	}

	if err := os.MkdirAll(epicsDir, 0755); err != nil {
		return fmt.Errorf("cli: failed to create epics directory: %w", err)
	}

	// Create initial project index
	index := &models.ProjectIndex{
		ProjectKey:  projectKey,
		ProjectName: projectName,
		Issues:      []models.IndexEntry{},
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	indexPath, err := storage.ProjectIndexPath(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve index path: %w", err)
	}

	if err := storage.WriteJSONAtomic(indexPath, index); err != nil {
		return fmt.Errorf("cli: failed to create project index: %w", err)
	}

	// Success message
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Created project %q\n", projectKey)

	return nil
}

// repairProject repairs a project index by rebuilding it from the issues directory.
func repairProject(projectKey string, cmd *cobra.Command) error {
	// Check if project exists
	projectDir, err := storage.ProjectDir(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve project directory: %w", err)
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("cli: project %q does not exist", projectKey)
	}

	// Check for pending transaction
	pendingPath := filepath.Join(projectDir, ".buyruk_pending")
	if _, err := os.Stat(pendingPath); err == nil {
		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Warning: Found pending transaction. This may indicate a previous crash.\n")
	}

	// Read all issue files from issues directory
	issuesDir, err := storage.IssuesDir(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve issues directory: %w", err)
	}

	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return fmt.Errorf("cli: failed to read issues directory: %w", err)
	}

	// Rebuild index from issue files
	indexEntries := []models.IndexEntry{}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		issuePath := filepath.Join(issuesDir, entry.Name())
		var issue models.Issue

		if err := storage.ReadJSON(issuePath, &issue); err != nil {
			// Log error but continue
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: failed to read issue file %s: %v\n", entry.Name(), err)
			continue
		}

		// Validate issue
		if err := issue.Validate(); err != nil {
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: invalid issue in %s: %v\n", entry.Name(), err)
			continue
		}

		// Add to index
		indexEntries = append(indexEntries, models.IndexEntry{
			ID:     issue.ID,
			Title:  issue.Title,
			Status: issue.Status,
			Type:   issue.Type,
			EpicID: issue.EpicID,
		})
	}

	// Load existing index to preserve metadata
	indexPath, err := storage.ProjectIndexPath(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve index path: %w", err)
	}

	var index models.ProjectIndex
	if _, err := os.Stat(indexPath); err == nil {
		if err := storage.ReadJSON(indexPath, &index); err != nil {
			return fmt.Errorf("cli: failed to read existing index: %w", err)
		}
	} else {
		// Create new index if it doesn't exist
		index = models.ProjectIndex{
			ProjectKey: projectKey,
			Issues:     []models.IndexEntry{},
		}
	}

	// Update index with rebuilt entries
	index.Issues = indexEntries
	index.UpdatedAt = time.Now().Format(time.RFC3339)

	// Write index atomically
	if err := storage.WriteJSONAtomic(indexPath, &index); err != nil {
		return fmt.Errorf("cli: failed to write repaired index: %w", err)
	}

	// Success message
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Repaired project %q: %d issues indexed\n", projectKey, len(indexEntries))

	return nil
}

// isValidProjectKey validates that the project key is uppercase alphanumeric or hyphen.
func isValidProjectKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	for _, r := range key {
		if !((r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	return true
}

// ResolveProjectKey resolves the project key from the command.
// This is a convenience wrapper around config.ResolveProject.
func ResolveProjectKey(cmd *cobra.Command) (string, error) {
	return config.ResolveProject(cmd)
}
