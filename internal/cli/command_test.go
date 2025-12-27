package cli

import (
	"testing"
)

func TestParseLoginCommand(t *testing.T) {
	args := []string{"program", "login"}
	result := ParseArgs(args)
	if _, ok := result.(Login); !ok {
		t.Errorf("Expected Login, got %T", result)
	}
}

func TestParseWhoamiCommand(t *testing.T) {
	args := []string{"program", "whoami"}
	result := ParseArgs(args)
	if _, ok := result.(Whoami); !ok {
		t.Errorf("Expected Whoami, got %T", result)
	}
}

func TestParsePushCommand(t *testing.T) {
	args := []string{"program", "push"}
	result := ParseArgs(args)
	if _, ok := result.(Push); !ok {
		t.Errorf("Expected Push, got %T", result)
	}
}

func TestParseRemoteListCommand(t *testing.T) {
	args := []string{"program", "remote"}
	result := ParseArgs(args)
	if _, ok := result.(RemoteList); !ok {
		t.Errorf("Expected RemoteList, got %T", result)
	}
}

func TestParseRemoteAddCommand(t *testing.T) {
	args := []string{"program", "remote", "add", "owner/repo"}
	result := ParseArgs(args)
	cmd, ok := result.(RemoteAdd)
	if !ok {
		t.Fatalf("Expected RemoteAdd, got %T", result)
	}
	if cmd.Repo != "owner/repo" {
		t.Errorf("Expected repo 'owner/repo', got '%s'", cmd.Repo)
	}
}

func TestParseRemoteAddMissingRepo(t *testing.T) {
	args := []string{"program", "remote", "add"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Missing repository argument. Usage: atat remote add <owner>/<repo>"
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteUnknownSubcommandWithTwoArgs(t *testing.T) {
	args := []string{"program", "remote", "unknown_sub"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "remote unknown_sub"
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseTooManyArgsForKnownCommand(t *testing.T) {
	args := []string{"program", "login", "extra_arg"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "login"
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteAddWithExtraArgs(t *testing.T) {
	args := []string{"program", "remote", "add", "owner/repo", "extra"}
	result := ParseArgs(args)
	cmd, ok := result.(RemoteAdd)
	if !ok {
		t.Fatalf("Expected RemoteAdd, got %T", result)
	}
	if cmd.Repo != "owner/repo" {
		t.Errorf("Expected repo 'owner/repo', got '%s'", cmd.Repo)
	}
}

func TestParseHelpCommand(t *testing.T) {
	args := []string{"program", "help"}
	result := ParseArgs(args)
	if _, ok := result.(Help); !ok {
		t.Errorf("Expected Help, got %T", result)
	}
}

func TestParseNoCommand(t *testing.T) {
	args := []string{"program"}
	result := ParseArgs(args)
	if _, ok := result.(Help); !ok {
		t.Errorf("Expected Help, got %T", result)
	}
}

func TestParseUnknownCommand(t *testing.T) {
	args := []string{"program", "unknown"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "unknown"
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteAddInvalidFormatNoSlash(t *testing.T) {
	args := []string{"program", "remote", "add", "ownerrepo"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteAddInvalidFormatEmptyOwner(t *testing.T) {
	args := []string{"program", "remote", "add", "/repo"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteAddInvalidFormatEmptyRepo(t *testing.T) {
	args := []string{"program", "remote", "add", "owner/"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteAddInvalidFormatTooManySlashes(t *testing.T) {
	args := []string{"program", "remote", "add", "owner/repo/extra"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteAddInvalidFormatOwnerContainsSlash(t *testing.T) {
	args := []string{"program", "remote", "add", "ow/ner/repo"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteRemoveCommand(t *testing.T) {
	args := []string{"program", "remote", "remove", "owner/repo"}
	result := ParseArgs(args)
	cmd, ok := result.(RemoteRemove)
	if !ok {
		t.Fatalf("Expected RemoteRemove, got %T", result)
	}
	if cmd.Repo != "owner/repo" {
		t.Errorf("Expected repo 'owner/repo', got '%s'", cmd.Repo)
	}
}

func TestParseRemoteRemoveMissingRepo(t *testing.T) {
	args := []string{"program", "remote", "remove"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Missing repository argument. Usage: atat remote remove <owner>/<repo>"
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteRemoveWithExtraArgs(t *testing.T) {
	args := []string{"program", "remote", "remove", "owner/repo", "extra"}
	result := ParseArgs(args)
	cmd, ok := result.(RemoteRemove)
	if !ok {
		t.Fatalf("Expected RemoteRemove, got %T", result)
	}
	if cmd.Repo != "owner/repo" {
		t.Errorf("Expected repo 'owner/repo', got '%s'", cmd.Repo)
	}
}

func TestParseRemoteRemoveInvalidFormatNoSlash(t *testing.T) {
	args := []string{"program", "remote", "remove", "ownerrepo"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteRemoveInvalidFormatEmptyOwner(t *testing.T) {
	args := []string{"program", "remote", "remove", "/repo"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteRemoveInvalidFormatEmptyRepo(t *testing.T) {
	args := []string{"program", "remote", "remove", "owner/"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteRemoveInvalidFormatTooManySlashes(t *testing.T) {
	args := []string{"program", "remote", "remove", "owner/repo/extra"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParseRemoteRemoveInvalidFormatOwnerContainsSlash(t *testing.T) {
	args := []string{"program", "remote", "remove", "ow/ner/repo"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "Invalid repository format. Please use <owner>/<repo>."
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}

func TestParsePullCommand(t *testing.T) {
	args := []string{"program", "pull"}
	result := ParseArgs(args)
	if _, ok := result.(Pull); !ok {
		t.Errorf("Expected Pull, got %T", result)
	}
}

func TestParseNonRemoteCommandWithFourOrMoreArgs(t *testing.T) {
	args := []string{"program", "push", "arg1", "arg2"}
	result := ParseArgs(args)
	cmd, ok := result.(Unknown)
	if !ok {
		t.Fatalf("Expected Unknown, got %T", result)
	}
	expected := "push arg1"
	if cmd.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, cmd.Message)
	}
}
