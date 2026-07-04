//go:build medium
// +build medium

package medium

import (
	"os"
	"strings"
	"testing"

	"github.com/toms74209200/gh-atat/internal/run"
)

func TestHelp(t *testing.T) {
	args := []string{"atat", "help"}
	if err := run.Run(args, ""); err != nil {
		t.Fatalf("help command failed: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = origStdout

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

func TestVersionWithVersionString(t *testing.T) {
	args := []string{"atat", "--version"}
	output := captureStdout(t, func() {
		if err := run.Run(args, "v1.2.3"); err != nil {
			t.Fatalf("version command failed: %v", err)
		}
	})

	expected := "gh-atat version v1.2.3"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain '%s', got '%s'", expected, output)
	}
}

func TestVersionWithoutVersionString(t *testing.T) {
	args := []string{"atat", "--version"}
	output := captureStdout(t, func() {
		if err := run.Run(args, ""); err != nil {
			t.Fatalf("version command failed: %v", err)
		}
	})

	expected := "gh-atat version dev"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain '%s', got '%s'", expected, output)
	}
}
