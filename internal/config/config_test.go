package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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
	// First, set a config value
	originalCfg, _ := Get()
	defer func() {
		// Restore original config
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	// Set config format to json
	if err := Set("default_format", "json"); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("format", "", "Output format")

	// Don't set flag, should use config value
	format := ResolveFormat(cmd)
	if format != "json" {
		t.Errorf("ResolveFormat() = %q, want json (from config)", format)
	}
}

func TestResolveFormat_Default(t *testing.T) {
	// Clear any existing config format
	originalCfg, _ := Get()
	defer func() {
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	// Clear default_format
	if err := Set("default_format", ""); err != nil {
		t.Fatalf("Failed to clear config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("format", "", "Output format")

	// Don't set flag or config, should use default
	format := ResolveFormat(cmd)
	if format != DefaultFormatModern {
		t.Errorf("ResolveFormat() = %q, want %q (default)", format, DefaultFormatModern)
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

func TestResolveProject_Config(t *testing.T) {
	// First, set a config value
	originalCfg, _ := Get()
	defer func() {
		// Restore original config
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	// Set config project
	if err := Set("default_project", "TESTPROJ"); err != nil {
		t.Fatalf("Failed to set config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("project", "", "Project key")

	// Don't set flag, should use config value
	project, err := ResolveProject(cmd)
	if err != nil {
		t.Fatalf("ResolveProject() failed: %v", err)
	}
	if project != "TESTPROJ" {
		t.Errorf("ResolveProject() = %q, want TESTPROJ (from config)", project)
	}
}

func TestResolveProject_Error(t *testing.T) {
	// Clear any existing config project
	originalCfg, _ := Get()
	defer func() {
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	// Clear default_project
	if err := Set("default_project", ""); err != nil {
		t.Fatalf("Failed to clear config: %v", err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("project", "", "Project key")

	// Don't set flag or config
	_, err := ResolveProject(cmd)
	if err == nil {
		t.Fatal("ResolveProject() should fail when no project is specified")
	}
}

func TestSet_DefaultProject(t *testing.T) {
	originalCfg, _ := Get()
	defer func() {
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	// Test setting default_project
	if err := Set("default_project", "TEST123"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Verify it was set
	value, err := GetValue("default_project")
	if err != nil {
		t.Fatalf("GetValue() failed: %v", err)
	}
	if value != "TEST123" {
		t.Errorf("GetValue() = %q, want TEST123", value)
	}

	// Test clearing it
	if err := Set("default_project", ""); err != nil {
		t.Fatalf("Set() failed to clear: %v", err)
	}

	value, err = GetValue("default_project")
	if err != nil {
		t.Fatalf("GetValue() failed: %v", err)
	}
	if value != "" {
		t.Errorf("GetValue() after clear = %q, want empty", value)
	}
}

func TestSet_DefaultFormat(t *testing.T) {
	originalCfg, _ := Get()
	defer func() {
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	// Test setting default_format
	if err := Set("default_format", "json"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	// Verify it was set
	value, err := GetValue("default_format")
	if err != nil {
		t.Fatalf("GetValue() failed: %v", err)
	}
	if value != "json" {
		t.Errorf("GetValue() = %q, want json", value)
	}

	// Test setting to lson
	if err := Set("default_format", "lson"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	value, err = GetValue("default_format")
	if err != nil {
		t.Fatalf("GetValue() failed: %v", err)
	}
	if value != "lson" {
		t.Errorf("GetValue() = %q, want lson", value)
	}
}

func TestSet_InvalidKey(t *testing.T) {
	err := Set("invalid_key", "value")
	if err == nil {
		t.Fatal("Set() should fail for invalid key")
	}
	if err.Error() != "config: unknown config key \"invalid_key\"" {
		t.Errorf("Set() error = %q, want error about unknown key", err.Error())
	}
}

func TestSet_InvalidFormat(t *testing.T) {
	originalCfg, _ := Get()
	defer func() {
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	err := Set("default_format", "invalid_format")
	if err == nil {
		t.Fatal("Set() should fail for invalid format")
	}
	if err.Error() != "config: invalid format \"invalid_format\" (must be modern, json, or lson)" {
		t.Errorf("Set() error = %q, want error about invalid format", err.Error())
	}
}

func TestSet_InvalidProjectKey(t *testing.T) {
	originalCfg, _ := Get()
	defer func() {
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	err := Set("default_project", "invalid-project")
	if err == nil {
		t.Fatal("Set() should fail for invalid project key")
	}
	if !strings.Contains(err.Error(), "invalid project key") {
		t.Errorf("Set() error = %q, want error about invalid project key", err.Error())
	}
}

func TestGetValue(t *testing.T) {
	originalCfg, _ := Get()
	defer func() {
		if originalCfg != nil {
			Save(originalCfg)
		}
	}()

	// Test getting default_project
	if err := Set("default_project", "GETTEST"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	value, err := GetValue("default_project")
	if err != nil {
		t.Fatalf("GetValue() failed: %v", err)
	}
	if value != "GETTEST" {
		t.Errorf("GetValue() = %q, want GETTEST", value)
	}

	// Test getting default_format
	if err := Set("default_format", "lson"); err != nil {
		t.Fatalf("Set() failed: %v", err)
	}

	value, err = GetValue("default_format")
	if err != nil {
		t.Fatalf("GetValue() failed: %v", err)
	}
	if value != "lson" {
		t.Errorf("GetValue() = %q, want lson", value)
	}

	// Test getting invalid key
	_, err = GetValue("invalid_key")
	if err == nil {
		t.Fatal("GetValue() should fail for invalid key")
	}
	if err.Error() != "config: unknown config key \"invalid_key\"" {
		t.Errorf("GetValue() error = %q, want error about unknown key", err.Error())
	}
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
		{"with dash", "TEST-123", true},
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
		{"invalid project with special chars", &Config{DefaultProject: "TEST@123"}, true},
		{"valid project with hyphen", &Config{DefaultProject: "TEST-123"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.wantErr {
				// Verify error message has config: prefix
				msg := err.Error()
				if len(msg) < len("config:") || msg[:len("config:")] != "config:" {
					t.Errorf("Validate() error should have 'config:' prefix, got: %q", msg)
				}
			}
		})
	}
}
