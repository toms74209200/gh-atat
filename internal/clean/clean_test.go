package clean

import (
	"testing"

	"github.com/toms74209200/gh-atat/internal/github"
	"github.com/toms74209200/gh-atat/internal/todo"
)

func checked(text string, issueNumber uint64) todo.TodoItem {
	return todo.TodoItem{
		Text:        text,
		IsChecked:   true,
		IssueNumber: &issueNumber,
	}
}

func unchecked(text string, issueNumber uint64) todo.TodoItem {
	return todo.TodoItem{
		Text:        text,
		IsChecked:   false,
		IssueNumber: &issueNumber,
	}
}

func checkedNoIssue(text string) todo.TodoItem {
	return todo.TodoItem{
		Text:      text,
		IsChecked: true,
	}
}

func openIssue(number uint64) github.GitHubIssue {
	return github.GitHubIssue{
		Number: number,
		Title:  "",
		State:  github.IssueStateOpen,
	}
}

func closedIssue(number uint64) github.GitHubIssue {
	return github.GitHubIssue{
		Number: number,
		Title:  "",
		State:  github.IssueStateClosed,
	}
}

func TestNewCleanCandidateCheckedWithIssue(t *testing.T) {
	item := checked("Task", 42)
	candidate, ok := NewCleanCandidate(item)
	if !ok {
		t.Fatal("Expected ok to be true")
	}
	if candidate.Text != "Task" {
		t.Errorf("Expected text 'Task', got '%s'", candidate.Text)
	}
	if candidate.IssueNumber != 42 {
		t.Errorf("Expected issue number 42, got %d", candidate.IssueNumber)
	}
}

func TestNewCleanCandidateUncheckedIsErr(t *testing.T) {
	item := unchecked("Task", 42)
	_, ok := NewCleanCandidate(item)
	if ok {
		t.Error("Expected ok to be false for unchecked item")
	}
}

func TestNewCleanCandidateCheckedWithoutIssue(t *testing.T) {
	item := checkedNoIssue("Task")
	_, ok := NewCleanCandidate(item)
	if ok {
		t.Error("Expected ok to be false for item without issue number")
	}
}

func TestFindRemovableRemovesCheckedWithClosedIssue(t *testing.T) {
	candidates := []CleanCandidate{
		{Text: "Done", IssueNumber: 1},
	}
	issues := []github.GitHubIssue{closedIssue(1)}
	result := FindRemovableItems(candidates, issues)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	if result[0].IssueNumber != 1 {
		t.Errorf("Expected issue number 1, got %d", result[0].IssueNumber)
	}
}

func TestFindRemovableKeepsCandidateWithOpenIssue(t *testing.T) {
	candidates := []CleanCandidate{
		{Text: "Still open", IssueNumber: 1},
	}
	issues := []github.GitHubIssue{openIssue(1)}
	result := FindRemovableItems(candidates, issues)
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}

func TestFindRemovableKeepsCandidateWithNoMatchingIssue(t *testing.T) {
	candidates := []CleanCandidate{
		{Text: "Unknown", IssueNumber: 99},
	}
	issues := []github.GitHubIssue{closedIssue(1)}
	result := FindRemovableItems(candidates, issues)
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}

func TestFindRemovableMixedCandidates(t *testing.T) {
	candidates := []CleanCandidate{
		{Text: "Remove me", IssueNumber: 1},
		{Text: "Keep me", IssueNumber: 2},
	}
	issues := []github.GitHubIssue{closedIssue(1), openIssue(2)}
	result := FindRemovableItems(candidates, issues)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	if result[0].IssueNumber != 1 {
		t.Errorf("Expected issue number 1, got %d", result[0].IssueNumber)
	}
}

func TestFindRemovableEmptyInputs(t *testing.T) {
	result := FindRemovableItems([]CleanCandidate{}, []github.GitHubIssue{})
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}
