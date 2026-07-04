package github

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/toms74209200/gh-atat/internal/todo"
)

func TestParsePastTitlesExtractsRenamedEvents(t *testing.T) {
	eventsJSON := []json.RawMessage{
		json.RawMessage(`{
			"event": "renamed",
			"rename": {"from": "First title", "to": "Second title"}
		}`),
		json.RawMessage(`{
			"event": "labeled",
			"label": {"name": "bug"}
		}`),
		json.RawMessage(`{
			"event": "renamed",
			"rename": {"from": "Second title", "to": "Third title"}
		}`),
	}

	pastTitles := ParsePastTitles(eventsJSON)

	if len(pastTitles) != 2 {
		t.Fatalf("expected 2 past titles, got %d", len(pastTitles))
	}
	if pastTitles[0] != "First title" {
		t.Errorf("expected 'First title', got '%s'", pastTitles[0])
	}
	if pastTitles[1] != "Second title" {
		t.Errorf("expected 'Second title', got '%s'", pastTitles[1])
	}
}

func TestParsePastTitlesEmptyEvents(t *testing.T) {
	pastTitles := ParsePastTitles([]json.RawMessage{})

	if len(pastTitles) != 0 {
		t.Errorf("expected 0 past titles, got %d", len(pastTitles))
	}
}

func TestParsePastTitlesIgnoresMalformedRename(t *testing.T) {
	eventsJSON := []json.RawMessage{
		json.RawMessage(`{"event": "renamed"}`),
		json.RawMessage(`{"event": "renamed", "rename": {"to": "No from"}}`),
	}

	pastTitles := ParsePastTitles(eventsJSON)

	if len(pastTitles) != 0 {
		t.Errorf("expected 0 past titles, got %d", len(pastTitles))
	}
}

func TestFindTitleMismatchesDetectsOpenIssueWithDifferentTitle(t *testing.T) {
	issueNum123 := uint64(123)
	issueNum456 := uint64(456)
	todoItems := []todo.TodoItem{
		{Text: "Old title", IsChecked: false, IssueNumber: &issueNum123},
		{Text: "Same title", IsChecked: false, IssueNumber: &issueNum456},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "New title", State: IssueStateOpen},
		{Number: 456, Title: "Same title", State: IssueStateOpen},
	}

	mismatches := FindTitleMismatches(todoItems, githubIssues)

	if len(mismatches) != 1 {
		t.Fatalf("expected 1 mismatch, got %d", len(mismatches))
	}
	if mismatches[0] != 123 {
		t.Errorf("expected issue number 123, got %d", mismatches[0])
	}
}

func TestFindTitleMismatchesIgnoresClosedIssues(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "Old title", IsChecked: false, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "New title", State: IssueStateClosed},
	}

	mismatches := FindTitleMismatches(todoItems, githubIssues)

	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches, got %d", len(mismatches))
	}
}

func TestFindTitleMismatchesIgnoresItemsWithoutIssueNumber(t *testing.T) {
	todoItems := []todo.TodoItem{
		{Text: "Local task", IsChecked: false, IssueNumber: nil},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Unrelated issue", State: IssueStateOpen},
	}

	mismatches := FindTitleMismatches(todoItems, githubIssues)

	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches, got %d", len(mismatches))
	}
}

func TestFindTitleMismatchesComparesTrimmed(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "  Same title  ", IsChecked: false, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Same title", State: IssueStateOpen},
	}

	mismatches := FindTitleMismatches(todoItems, githubIssues)

	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches, got %d", len(mismatches))
	}
}

func TestMatchesPastTitleWithTrim(t *testing.T) {
	pastTitles := map[uint64][]string{
		123: {"Old title"},
	}

	if !MatchesPastTitle(pastTitles, 123, "  Old title  ") {
		t.Error("expected match with trimmed text")
	}
	if MatchesPastTitle(pastTitles, 123, "Other title") {
		t.Error("expected no match with different text")
	}
	if MatchesPastTitle(pastTitles, 456, "Old title") {
		t.Error("expected no match with different issue number")
	}
}

func TestCollectPastTitlesFetchesOnlyMismatchedIssues(t *testing.T) {
	issueNum123 := uint64(123)
	issueNum456 := uint64(456)
	todoItems := []todo.TodoItem{
		{Text: "Old title", IsChecked: false, IssueNumber: &issueNum123},
		{Text: "Same title", IsChecked: false, IssueNumber: &issueNum456},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "New title", State: IssueStateOpen},
		{Number: 456, Title: "Same title", State: IssueStateOpen},
	}

	eventsFetcher := func(issueNumber uint64) ([]json.RawMessage, error) {
		if issueNumber == 123 {
			return []json.RawMessage{
				json.RawMessage(`{
					"event": "renamed",
					"rename": {"from": "Old title", "to": "New title"}
				}`),
			}, nil
		}
		return nil, fmt.Errorf("history should not be fetched for issue #%d", issueNumber)
	}

	result, err := CollectPastTitles(todoItems, githubIssues, eventsFetcher)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	titles, exists := result[123]
	if !exists {
		t.Fatal("expected entry for issue 123")
	}
	if len(titles) != 1 {
		t.Fatalf("expected 1 past title, got %d", len(titles))
	}
	if titles[0] != "Old title" {
		t.Errorf("expected 'Old title', got '%s'", titles[0])
	}
}

func TestFindTitleMismatchesSkipsNonexistentIssue(t *testing.T) {
	issueNum999 := uint64(999)
	todoItems := []todo.TodoItem{
		{Text: "Some task", IsChecked: false, IssueNumber: &issueNum999},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Different issue", State: IssueStateOpen},
	}

	mismatches := FindTitleMismatches(todoItems, githubIssues)

	if len(mismatches) != 0 {
		t.Errorf("expected 0 mismatches, got %d", len(mismatches))
	}
}

func TestParsePastTitlesIgnoresInvalidJSON(t *testing.T) {
	eventsJSON := []json.RawMessage{
		json.RawMessage(`invalid json`),
		json.RawMessage(`{"event": "renamed", "rename": {"from": "Valid title", "to": "New title"}}`),
	}

	pastTitles := ParsePastTitles(eventsJSON)

	if len(pastTitles) != 1 {
		t.Fatalf("expected 1 past title, got %d", len(pastTitles))
	}
	if pastTitles[0] != "Valid title" {
		t.Errorf("expected 'Valid title', got '%s'", pastTitles[0])
	}
}

func TestCollectPastTitlesPropagatesFetcherError(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "Old title", IsChecked: false, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "New title", State: IssueStateOpen},
	}

	eventsFetcher := func(issueNumber uint64) ([]json.RawMessage, error) {
		return nil, fmt.Errorf("Network error")
	}

	result, err := CollectPastTitles(todoItems, githubIssues, eventsFetcher)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "Network error" {
		t.Errorf("expected 'Network error', got '%s'", err.Error())
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}
