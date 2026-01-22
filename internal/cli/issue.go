package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/buyruk-project/buyruk-cli/internal/config"
	"github.com/buyruk-project/buyruk-cli/internal/models"
	"github.com/buyruk-project/buyruk-cli/internal/storage"
	"github.com/spf13/cobra"
)

// NewIssueCmd creates and returns the issue command.
func NewIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage issues",
		Long:  "Create and manage buyruk issues",
	}

	cmd.AddCommand(NewIssueCreateCmd())

	return cmd
}

// NewIssueCreateCmd creates and returns the issue create command.
func NewIssueCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		Long:  "Create a new issue in the project. Only title is required.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createIssue(cmd)
		},
	}

	cmd.Flags().String("id", "", "Issue ID (optional, auto-generated if not provided)")
	cmd.Flags().String("type", "task", "Issue type (task or bug, default: task)")
	cmd.Flags().String("title", "", "Issue title (required)")
	cmd.Flags().String("status", "TODO", "Issue status (TODO, DOING, DONE, default: TODO)")
	cmd.Flags().String("priority", "", "Issue priority (LOW, MEDIUM, HIGH, CRITICAL)")
	cmd.Flags().String("description", "", "Issue description (Markdown)")
	cmd.Flags().String("epic", "", "Link to epic ID")

	return cmd
}

// createIssue creates a new issue in the project.
func createIssue(cmd *cobra.Command) error {
	// Resolve project (required for auto-increment)
	projectKey, err := config.ResolveProject(cmd)
	if err != nil {
		return err
	}

	// Verify project exists
	projectDir, err := storage.ProjectDir(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve project directory: %w", err)
	}

	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return fmt.Errorf("cli: project %q does not exist", projectKey)
	}

	// Get title (required)
	title, _ := cmd.Flags().GetString("title")
	if title == "" {
		return fmt.Errorf("cli: title is required")
	}

	// Get ID (optional, auto-generate if not provided)
	issueID, _ := cmd.Flags().GetString("id")
	if issueID == "" {
		nextSeq, err := getNextIssueSequence(projectKey)
		if err != nil {
			return fmt.Errorf("cli: failed to get next issue sequence: %w", err)
		}
		issueID = models.GenerateIssueID(projectKey, nextSeq)
	} else {
		// Validate provided ID matches project key
		parsedKey, _, err := models.ParseIssueID(issueID)
		if err != nil {
			return fmt.Errorf("cli: invalid issue ID format: %w", err)
		}
		if parsedKey != projectKey {
			return fmt.Errorf("cli: issue ID %q does not match project key %q", issueID, projectKey)
		}
	}

	// Get type (default: task)
	issueType, _ := cmd.Flags().GetString("type")
	if issueType == "" {
		issueType = models.TypeTask
	}

	// Get status (default: TODO)
	status, _ := cmd.Flags().GetString("status")
	if status == "" {
		status = models.StatusTODO
	}

	// Get optional fields
	priority, _ := cmd.Flags().GetString("priority")
	description, _ := cmd.Flags().GetString("description")
	epicID, _ := cmd.Flags().GetString("epic")

	// Create issue
	issue := &models.Issue{
		ID:          issueID,
		Type:        issueType,
		Title:       title,
		Status:      status,
		Priority:    priority,
		Description: description,
		EpicID:      epicID,
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	// Validate issue
	if err := issue.Validate(); err != nil {
		return fmt.Errorf("cli: invalid issue: %w", err)
	}

	// Write issue file atomically (fails if file already exists)
	issuePath, err := storage.IssuePath(projectKey, issueID)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve issue path: %w", err)
	}

	if err := storage.WriteJSONAtomicCreate(issuePath, issue); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("cli: issue %q already exists", issueID)
		}
		return fmt.Errorf("cli: failed to create issue file: %w", err)
	}

	// Update project index atomically (read-modify-write with locking)
	indexPath, err := storage.ProjectIndexPath(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve index path: %w", err)
	}

	var index models.ProjectIndex
	if err := storage.UpdateJSONAtomic(indexPath, &index, func(v interface{}) error {
		idx := v.(*models.ProjectIndex)
		idx.AddIssue(issue)
		idx.UpdatedAt = time.Now().Format(time.RFC3339)
		return nil
	}); err != nil {
		return fmt.Errorf("cli: failed to update project index: %w", err)
	}

	// Success message
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Created issue %q\n", issueID)

	return nil
}

// getNextIssueSequence returns the next sequence number for an issue in the project.
// It parses all existing issue IDs to find the highest sequence number and returns the next one.
func getNextIssueSequence(projectKey string) (int, error) {
	// Load project index
	indexPath, err := storage.ProjectIndexPath(projectKey)
	if err != nil {
		return 0, fmt.Errorf("cli: failed to resolve index path: %w", err)
	}

	var index models.ProjectIndex
	if err := storage.ReadJSON(indexPath, &index); err != nil {
		// If index doesn't exist, start from 1
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, fmt.Errorf("cli: failed to load project index: %w", err)
	}

	// Find the highest sequence number
	maxSeq := 0
	for _, entry := range index.Issues {
		_, seq, err := models.ParseIssueID(entry.ID)
		if err != nil {
			// Skip invalid IDs
			continue
		}
		if seq > maxSeq {
			maxSeq = seq
		}
	}

	// Return next sequence number
	return maxSeq + 1, nil
}
