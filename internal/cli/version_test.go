package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/buyruk-project/buyruk-cli/internal/build"
)

func TestNewVersionCmd(t *testing.T) {
	cmd := NewVersionCmd()
	if cmd == nil {
		t.Fatal("NewVersionCmd() returned nil")
	}
	if cmd.Use != "version" {
		t.Errorf("Expected Use to be 'version', got '%s'", cmd.Use)
	}
}

func TestVersionCmdOutputModern(t *testing.T) {
	// Set format to modern
	formatFlag = "modern"
	
	cmd := NewVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	
	output := buf.String()
	expected := "buyruk version " + build.Version + "\n"
	if output != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, output)
	}
}

func TestVersionCmdOutputJSON(t *testing.T) {
	// Set format to json
	formatFlag = "json"
	
	cmd := NewVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	
	output := buf.String()
	expected := `{"version":"` + build.Version + `"}` + "\n"
	if output != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, output)
	}
}

func TestVersionCmdOutputLSON(t *testing.T) {
	// Set format to lson
	formatFlag = "lson"
	
	cmd := NewVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	
	output := buf.String()
	expected := "@VERSION: " + build.Version + "\n"
	if output != expected {
		t.Errorf("Expected output '%s', got '%s'", expected, output)
	}
}

func TestVersionCmdIntegration(t *testing.T) {
	// Test with actual root command
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"version", "--format", "modern"})
	
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("root command with version subcommand failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, build.Version) {
		t.Errorf("Output should contain version '%s', got '%s'", build.Version, output)
	}
}
