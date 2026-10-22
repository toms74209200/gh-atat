package github

import (
	"crypto/rand"
	"fmt"
	mrand "math/rand/v2"
	"strings"
	"testing"

	"github.com/toms74209200/gh-atat/internal/todo"
)

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func TestCalculateGitHubOperations(t *testing.T) {
	tests := []struct {
		name            string
		todoItems       []todo.TodoItem
		githubIssues    []GitHubIssue
		expectedOpCount int
		expectedOpType  any
		validateOp      func(t *testing.T, op GitHubOperation)
	}{
		{
			name: "unchecked_no_issue_creates_issue",
			todoItems: []todo.TodoItem{
				{Text: "New task", IsChecked: false, IssueNumber: nil},
			},
			githubIssues:    []GitHubIssue{},
			expectedOpCount: 1,
			expectedOpType:  CreateIssueOp{},
			validateOp: func(t *testing.T, op GitHubOperation) {
				createOp, ok := op.(CreateIssueOp)
				if !ok {
					t.Fatalf("expected CreateIssueOp, got %T", op)
				}
				if createOp.Title != "New task" {
					t.Errorf("expected title 'New task', got '%s'", createOp.Title)
				}
			},
		},
		{
			name: "checked_with_open_issue_closes_issue",
			todoItems: []todo.TodoItem{
				{Text: "Completed task", IsChecked: true, IssueNumber: uint64Ptr(123)},
			},
			githubIssues: []GitHubIssue{
				{Number: 123, Title: "Completed task", State: IssueStateOpen},
			},
			expectedOpCount: 1,
			expectedOpType:  CloseIssueOp{},
			validateOp: func(t *testing.T, op GitHubOperation) {
				closeOp, ok := op.(CloseIssueOp)
				if !ok {
					t.Fatalf("expected CloseIssueOp, got %T", op)
				}
				if closeOp.Number != 123 {
					t.Errorf("expected issue number 123, got %d", closeOp.Number)
				}
			},
		},
		{
			name: "checked_with_closed_issue_no_operation",
			todoItems: []todo.TodoItem{
				{Text: "Already closed task", IsChecked: true, IssueNumber: uint64Ptr(123)},
			},
			githubIssues: []GitHubIssue{
				{Number: 123, Title: "Already closed task", State: IssueStateClosed},
			},
			expectedOpCount: 0,
		},
		{
			name: "checked_with_nonexistent_issue_no_operation",
			todoItems: []todo.TodoItem{
				{Text: "Task with missing issue", IsChecked: true, IssueNumber: uint64Ptr(999)},
			},
			githubIssues: []GitHubIssue{
				{Number: 123, Title: "Different issue", State: IssueStateOpen},
			},
			expectedOpCount: 0,
		},
		{
			name: "unchecked_with_existing_issue_no_operation",
			todoItems: []todo.TodoItem{
				{Text: "Unchecked with issue", IsChecked: false, IssueNumber: uint64Ptr(456)},
			},
			githubIssues: []GitHubIssue{
				{Number: 456, Title: "Existing issue", State: IssueStateOpen},
			},
			expectedOpCount: 0,
		},
		{
			name: "checked_without_issue_no_operation",
			todoItems: []todo.TodoItem{
				{Text: "Checked but no issue", IsChecked: true, IssueNumber: nil},
			},
			githubIssues:    []GitHubIssue{},
			expectedOpCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := CalculateGitHubOperations(tt.todoItems, tt.githubIssues)

			if len(operations) != tt.expectedOpCount {
				t.Fatalf("expected %d operations, got %d", tt.expectedOpCount, len(operations))
			}

			if tt.expectedOpCount > 0 && tt.validateOp != nil {
				tt.validateOp(t, operations[0].Operation)
				if operations[0].Todo.Text != tt.todoItems[0].Text {
					t.Errorf("expected todo text to match")
				}
			}
		})
	}
}

func TestCalculateTitleUpdatesRenamesLocallyEditedTitle(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "Locally edited title", IsChecked: false, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Original title", State: IssueStateOpen},
	}
	pastTitles := map[uint64][]string{
		123: {},
	}

	updates := CalculateTitleUpdates(todoItems, githubIssues, pastTitles)

	if len(updates.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(updates.Operations))
	}
	renameOp, ok := updates.Operations[0].Operation.(RenameIssueOp)
	if !ok {
		t.Fatalf("expected RenameIssueOp, got %T", updates.Operations[0].Operation)
	}
	if renameOp.Number != 123 {
		t.Errorf("expected issue number 123, got %d", renameOp.Number)
	}
	if renameOp.Title != "Locally edited title" {
		t.Errorf("expected 'Locally edited title', got '%s'", renameOp.Title)
	}
	if len(updates.StaleIssues) != 0 {
		t.Errorf("expected no stale issues, got %v", updates.StaleIssues)
	}
}

func TestCalculateTitleUpdatesSkipsStaleLocalText(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "Old title", IsChecked: false, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "New title", State: IssueStateOpen},
	}
	pastTitles := map[uint64][]string{
		123: {"Old title"},
	}

	updates := CalculateTitleUpdates(todoItems, githubIssues, pastTitles)

	if len(updates.Operations) != 0 {
		t.Errorf("expected 0 operations, got %d", len(updates.Operations))
	}
	if len(updates.StaleIssues) != 1 {
		t.Fatalf("expected 1 stale issue, got %d", len(updates.StaleIssues))
	}
	if updates.StaleIssues[0] != 123 {
		t.Errorf("expected stale issue 123, got %d", updates.StaleIssues[0])
	}
}

func TestCalculateTitleUpdatesIgnoresSyncedClosedAndUnnumberedItems(t *testing.T) {
	issueNum123 := uint64(123)
	issueNum456 := uint64(456)
	todoItems := []todo.TodoItem{
		{Text: "Same title", IsChecked: false, IssueNumber: &issueNum123},
		{Text: "Edited closed title", IsChecked: true, IssueNumber: &issueNum456},
		{Text: "Local task", IsChecked: false, IssueNumber: nil},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Same title", State: IssueStateOpen},
		{Number: 456, Title: "Closed title", State: IssueStateClosed},
	}
	pastTitles := map[uint64][]string{}

	updates := CalculateTitleUpdates(todoItems, githubIssues, pastTitles)

	if len(updates.Operations) != 0 {
		t.Errorf("expected 0 operations, got %d", len(updates.Operations))
	}
	if len(updates.StaleIssues) != 0 {
		t.Errorf("expected 0 stale issues, got %d", len(updates.StaleIssues))
	}
}

func TestCalculateTitleUpdatesTrimsTitleForRename(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "  Edited title  ", IsChecked: false, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Original title", State: IssueStateOpen},
	}
	pastTitles := map[uint64][]string{}

	updates := CalculateTitleUpdates(todoItems, githubIssues, pastTitles)

	if len(updates.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(updates.Operations))
	}
	renameOp := updates.Operations[0].Operation.(RenameIssueOp)
	if renameOp.Title != "Edited title" {
		t.Errorf("expected 'Edited title', got '%s'", renameOp.Title)
	}
}

func TestCalculateTitleUpdatesSkipsItemWithNonexistentIssue(t *testing.T) {
	issueNum999 := uint64(999)
	todoItems := []todo.TodoItem{
		{Text: "Task with missing issue", IsChecked: false, IssueNumber: &issueNum999},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Different issue", State: IssueStateOpen},
	}
	pastTitles := map[uint64][]string{}

	updates := CalculateTitleUpdates(todoItems, githubIssues, pastTitles)

	if len(updates.Operations) != 0 {
		t.Errorf("expected 0 operations, got %d", len(updates.Operations))
	}
	if len(updates.StaleIssues) != 0 {
		t.Errorf("expected 0 stale issues, got %d", len(updates.StaleIssues))
	}
}

func TestWhenParseCreatedIssueNumberWithNumberThenReturnsIssueNumber(t *testing.T) {
	number := mrand.Uint64N(100000) + 1

	got, err := ParseCreatedIssueNumber(fmt.Appendf(nil, `{"number":%d,"title":"%s"}`, number, rand.Text()))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != number {
		t.Errorf("expected issue number %d, got %d", number, got)
	}
}

func TestWhenParseCreatedIssueNumberWithBodyWithoutNumberThenReturnsError(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		wantErr string
	}{
		{"object without number then returns error", []byte(`{"title":"` + rand.Text() + `"}`), "created issue has no number"},
		{"non-JSON body then returns error", []byte(rand.Text()), "invalid character"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCreatedIssueNumber(tt.body)

			if err == nil || !strings.HasPrefix(err.Error(), tt.wantErr) {
				t.Errorf("expected error starting with %q, got %v", tt.wantErr, err)
			}
		})
	}
}
