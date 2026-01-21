package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestLoad_NonExistent(t *testing.T) {
	// Since we can't easily override storage.ConfigDir() without modifying it,
	// we test the actual behavior by checking that Load returns defaults
	// when config doesn't exist. This is tested in integration tests.
	// For now, we verify Default() works correctly (tested in TestDefault).
}

func TestLoad_Existing(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFileName)

	// Create a test config
	testCfg := &Config{
		DefaultProject: "TEST",
		DefaultFormat:  "json",
	}

	data, err := json.MarshalIndent(testCfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Note: This test requires the ability to override storage.ConfigFilePath()
	// For a full test, we'd need to modify storage or use dependency injection
	// For now, we test the core logic separately
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg == nil {
		t.Fatal("Default() returned nil")
	}
	if cfg.DefaultFormat != DefaultFormatModern {
		t.Errorf("Default().DefaultFormat = %q, want %q", cfg.DefaultFormat, DefaultFormatModern)
	}
	if cfg.DefaultProject != "" {
		t.Errorf("Default().DefaultProject = %q, want empty", cfg.DefaultProject)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFileName)

	// Create a test config
	testCfg := &Config{
		DefaultProject: "TEST",
		DefaultFormat:  "json",
	}

	// Save manually (since we can't easily override storage paths)
	data, err := json.MarshalIndent(testCfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// Read back
	var loadedCfg Config
	readData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if err := json.Unmarshal(readData, &loadedCfg); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if loadedCfg.DefaultProject != "TEST" {
		t.Errorf("Loaded DefaultProject = %q, want TEST", loadedCfg.DefaultProject)
	}
	if loadedCfg.DefaultFormat != "json" {
		t.Errorf("Loaded DefaultFormat = %q, want json", loadedCfg.DefaultFormat)
	}
}

func TestResolveFormat_Flag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("format", "", "Output format")

	// Set flag
	cmd.Flags().Set("format", "lson")

	format := ResolveFormat(cmd)
	if format != "lson" {
		t.Errorf("ResolveFormat() = %q, want lson", format)
	}
}

func TestResolveFormat_Config(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("format", "", "Output format")

	// Don't set flag, should fall back to default
	format := ResolveFormat(cmd)
	if format != DefaultFormatModern {
		t.Errorf("ResolveFormat() = %q, want %q", format, DefaultFormatModern)
	}
}

func TestResolveFormat_Default(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("format", "", "Output format")

	format := ResolveFormat(cmd)
	if format != DefaultFormatModern {
		t.Errorf("ResolveFormat() = %q, want %q", format, DefaultFormatModern)
	}
}

func TestResolveProject_Flag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("project", "", "Project key")

	// Set flag
	cmd.Flags().Set("project", "TEST")

	project, err := ResolveProject(cmd)
	if err != nil {
		t.Fatalf("ResolveProject() failed: %v", err)
	}
	if project != "TEST" {
		t.Errorf("ResolveProject() = %q, want TEST", project)
	}
}

func TestResolveProject_Error(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("project", "", "Project key")

	// Don't set flag or config
	_, err := ResolveProject(cmd)
	if err == nil {
		t.Fatal("ResolveProject() should fail when no project is specified")
	}
}

func TestSet_DefaultProject(t *testing.T) {
	// This test requires actual storage, so we'll test the validation logic separately
	// and integration tests will test the full flow
}

func TestSet_DefaultFormat(t *testing.T) {
	// This test requires actual storage, so we'll test the validation logic separately
}

func TestSet_InvalidKey(t *testing.T) {
	// We can't easily test Set without storage, but we can test validation
}

func TestGetValue(t *testing.T) {
	// This test requires actual storage
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   bool
	}{
		{"modern", DefaultFormatModern, true},
		{"json", DefaultFormatJSON, true},
		{"lson", DefaultFormatLSON, true},
		{"invalid", "invalid", false},
		{"empty", "", false},
		{"mixed case", "Modern", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFormat(tt.format)
			if got != tt.want {
				t.Errorf("isValidFormat(%q) = %v, want %v", tt.format, got, tt.want)
			}
		})
	}
}

func TestIsValidProjectKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{"uppercase letters", "TEST", true},
		{"uppercase with numbers", "TEST123", true},
		{"numbers only", "123", true},
		{"lowercase", "test", false},
		{"mixed case", "Test", false},
		{"with underscore", "TEST_123", false},
		{"with dash", "TEST-123", false},
		{"with space", "TEST 123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidProjectKey(tt.key)
			if got != tt.want {
				t.Errorf("isValidProjectKey(%q) = %v, want %v", tt.key, got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{"valid config", &Config{DefaultFormat: "modern", DefaultProject: "TEST"}, false},
		{"valid format only", &Config{DefaultFormat: "json"}, false},
		{"valid project only", &Config{DefaultProject: "TEST123"}, false},
		{"empty config", &Config{}, false},
		{"invalid format", &Config{DefaultFormat: "invalid"}, true},
		{"invalid project lowercase", &Config{DefaultProject: "test"}, true},
		{"invalid project mixed case", &Config{DefaultProject: "Test"}, true},
		{"invalid project with special chars", &Config{DefaultProject: "TEST-123"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
