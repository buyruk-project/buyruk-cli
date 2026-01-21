package cli

import (
	"github.com/spf13/cobra"
)

var (
	formatFlag  string
	projectFlag string
)

// NewRootCmd creates and returns the root command for buyruk CLI.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "buyruk",
		Short: "A local-first project management tool",
		Long:  "Buyruk is a high-performance, local-first orchestration tool that treats the filesystem as a database.",
	}

	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "modern", "Output format (modern, json, lson)")
	rootCmd.PersistentFlags().StringVar(&projectFlag, "project", "", "Project key to operate on")

	// Add subcommands
	rootCmd.AddCommand(NewVersionCmd())

	return rootCmd
}

// GetFormat returns the current format flag value.
func GetFormat() string {
	return formatFlag
}

// GetProject returns the current project flag value.
func GetProject() string {
	return projectFlag
}
