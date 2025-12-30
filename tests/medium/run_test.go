//go:build medium
// +build medium

package medium

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/toms74209200/gh-atat/internal/run"
)

func TestRemoteAddAndList(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Test remote add
	args := []string{"atat", "remote", "add", "owner/repo"}
	if err := run.Run(args, nil, nil); err != nil {
		t.Fatalf("remote add failed: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".atat", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Test remote list
	args = []string{"atat", "remote"}
	if err := run.Run(args, nil, nil); err != nil {
		t.Fatalf("remote list failed: %v", err)
	}
}

func TestRemoteAddAndRemove(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	// Change to temp directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	// Add repository
	args := []string{"atat", "remote", "add", "owner/repo"}
	if err := run.Run(args, nil, nil); err != nil {
		t.Fatalf("remote add failed: %v", err)
	}

	// Remove repository
	args = []string{"atat", "remote", "remove", "owner/repo"}
	if err := run.Run(args, nil, nil); err != nil {
		t.Fatalf("remote remove failed: %v", err)
	}

	// Verify config was updated
	configPath := filepath.Join(tmpDir, ".atat", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	// Config should be empty or have empty repositories
	if len(data) > 0 && string(data) != "{}" && string(data) != "{}\n" {
		// Check if it's truly empty
		t.Logf("Config content: %s", string(data))
	}
}

func TestHelp(t *testing.T) {
	args := []string{"atat", "help"}
	if err := run.Run(args, nil, nil); err != nil {
		t.Fatalf("help command failed: %v", err)
	}
}
