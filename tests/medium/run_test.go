//go:build medium
// +build medium

package medium

import (
	"testing"

	"github.com/toms74209200/gh-atat/internal/run"
)

func TestHelp(t *testing.T) {
	args := []string{"atat", "help"}
	if err := run.Run(args); err != nil {
		t.Fatalf("help command failed: %v", err)
	}
}
