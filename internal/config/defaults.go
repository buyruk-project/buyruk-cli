package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Default returns a default config struct.
func Default() *Config {
	return &Config{
		DefaultFormat: DefaultFormatModern,
	}
}

// ResolveFormat resolves the format from flag > config > default.
// Priority: --format flag > config.default_format > "modern"
func ResolveFormat(cmd *cobra.Command) string {
	// Check flag first
	format, _ := cmd.Flags().GetString("format")
	if format != "" {
		return format
	}

	// Check config
	cfg, err := Get()
	if err == nil && cfg.DefaultFormat != "" {
		return cfg.DefaultFormat
	}

	// Return default
	return DefaultFormatModern
}

// ResolveProject resolves the project from flag > config > error.
// Priority: --project flag > config.default_project > error
func ResolveProject(cmd *cobra.Command) (string, error) {
	// Check flag first
	project, _ := cmd.Flags().GetString("project")
	if project != "" {
		return project, nil
	}

	// Check config
	cfg, err := Get()
	if err == nil && cfg.DefaultProject != "" {
		return cfg.DefaultProject, nil
	}

	// No project specified
	return "", fmt.Errorf("config: no project specified (use --project flag or set default_project in config)")
}
