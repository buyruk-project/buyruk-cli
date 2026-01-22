package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates and returns the root command for buyruk CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "buyruk",
		Short: "A local-first project management tool",
		Long:  "Buyruk is a high-performance, local-first orchestration tool that treats the filesystem as a database.",
	}

	// Persistent flags
	rootCmd.PersistentFlags().String("format", "modern", "Output format (modern, json, lson)")
	rootCmd.PersistentFlags().String("project", "", "Project key to operate on")

	// Add subcommands
	rootCmd.AddCommand(NewVersionCmd())
	rootCmd.AddCommand(NewListCmd())
	rootCmd.AddCommand(NewViewCmd())
	rootCmd.AddCommand(NewProjectCmd())
	rootCmd.AddCommand(NewIssueCmd())
	rootCmd.AddCommand(NewConfigCmd())

	return rootCmd
}

// GetFormat returns the format flag value from the command.
func GetFormat(cmd *cobra.Command) string {
	format, _ := cmd.Flags().GetString("format")
	return format
}

// GetProject returns the project flag value from the command.
func GetProject(cmd *cobra.Command) string {
	project, _ := cmd.Flags().GetString("project")
	return project
}
