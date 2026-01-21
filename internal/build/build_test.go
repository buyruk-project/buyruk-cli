package build

import "testing"

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Version != "0.0.1-dev" {
		t.Errorf("Expected version '0.0.1-dev', got '%s'", Version)
	}
}
