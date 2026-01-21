package cli

import (
	"testing"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	if cmd == nil {
		t.Fatal("NewRootCmd() returned nil")
	}
	if cmd.Use != "buyruk" {
		t.Errorf("Expected Use to be 'buyruk', got '%s'", cmd.Use)
	}
}

func TestRootCmdFlags(t *testing.T) {
	cmd := NewRootCmd()
	
	// Test format flag
	formatFlag := cmd.PersistentFlags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}
	if formatFlag.DefValue != "modern" {
		t.Errorf("Expected format default to be 'modern', got '%s'", formatFlag.DefValue)
	}

	// Test project flag
	projectFlag := cmd.PersistentFlags().Lookup("project")
	if projectFlag == nil {
		t.Fatal("project flag not found")
	}
	if projectFlag.DefValue != "" {
		t.Errorf("Expected project default to be empty, got '%s'", projectFlag.DefValue)
	}
}

func TestRootCmdHasVersionSubcommand(t *testing.T) {
	cmd := NewRootCmd()
	versionCmd, _, err := cmd.Find([]string{"version"})
	if err != nil {
		t.Fatalf("version subcommand not found: %v", err)
	}
	if versionCmd == nil {
		t.Fatal("version subcommand is nil")
	}
}

func TestGetFormat(t *testing.T) {
	// Reset format flag
	formatFlag = "test-format"
	if GetFormat() != "test-format" {
		t.Errorf("Expected GetFormat() to return 'test-format', got '%s'", GetFormat())
	}
}

func TestGetProject(t *testing.T) {
	// Reset project flag
	projectFlag = "test-project"
	if GetProject() != "test-project" {
		t.Errorf("Expected GetProject() to return 'test-project', got '%s'", GetProject())
	}
}
