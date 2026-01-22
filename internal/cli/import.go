package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/buyruk-project/buyruk-cli/internal/storage"
	"github.com/spf13/cobra"
)

// NewImportCmd creates and returns the import command.
func NewImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import a project",
		Long:  "Import a project from an export file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			return importProject(filePath, cmd)
		},
	}

	cmd.Flags().Bool("overwrite", false, "Overwrite existing project if it exists")

	return cmd
}

// importProject imports a project from an export file.
func importProject(filePath string, cmd *cobra.Command) error {
	// Read export file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("cli: failed to read export file: %w", err)
	}

	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		return fmt.Errorf("cli: failed to parse export file: %w", err)
	}

	// Validate export data
	if err := validateExportData(&exportData); err != nil {
		return fmt.Errorf("cli: invalid export file: %w", err)
	}

	projectKey := exportData.Project.ProjectKey

	// Check if project already exists
	projectDir, err := storage.ProjectDir(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve project directory: %w", err)
	}

	overwrite, _ := cmd.Flags().GetBool("overwrite")
	if _, err := os.Stat(projectDir); err == nil {
		if !overwrite {
			return fmt.Errorf("cli: project %q already exists (use --overwrite to replace)", projectKey)
		}

		// Remove existing project
		if err := os.RemoveAll(projectDir); err != nil {
			return fmt.Errorf("cli: failed to remove existing project: %w", err)
		}
	}

	// Create project directories
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

	// Write project index
	indexPath, err := storage.ProjectIndexPath(projectKey)
	if err != nil {
		return fmt.Errorf("cli: failed to resolve index path: %w", err)
	}

	if err := storage.WriteJSONAtomic(indexPath, exportData.Project); err != nil {
		return fmt.Errorf("cli: failed to write project index: %w", err)
	}

	// Write all issues
	for _, issue := range exportData.Issues {
		// Validate issue
		if err := issue.Validate(); err != nil {
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: skipping invalid issue %s: %v\n", issue.ID, err)
			continue
		}

		issuePath, err := storage.IssuePath(projectKey, issue.ID)
		if err != nil {
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: failed to resolve path for issue %s: %v\n", issue.ID, err)
			continue
		}

		if err := storage.WriteJSONAtomic(issuePath, issue); err != nil {
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: failed to write issue %s: %v\n", issue.ID, err)
			continue
		}
	}

	// Write all epics
	for _, epic := range exportData.Epics {
		// Validate epic
		if err := epic.Validate(); err != nil {
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: skipping invalid epic %s: %v\n", epic.ID, err)
			continue
		}

		epicPath, err := storage.EpicPath(projectKey, epic.ID)
		if err != nil {
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: failed to resolve path for epic %s: %v\n", epic.ID, err)
			continue
		}

		if err := storage.WriteJSONAtomic(epicPath, epic); err != nil {
			errOut := cmd.ErrOrStderr()
			fmt.Fprintf(errOut, "Warning: failed to write epic %s: %v\n", epic.ID, err)
			continue
		}
	}

	// Success message
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Imported project %q (%d issues, %d epics)\n",
		projectKey, len(exportData.Issues), len(exportData.Epics))

	return nil
}
