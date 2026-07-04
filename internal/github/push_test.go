package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/toms74209200/gh-atat/internal/todo"
)

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func TestCalculateGitHubOperations(t *testing.T) {
	tests := []struct {
		name             string
		todoItems        []todo.TodoItem
		githubIssues     []GitHubIssue
		expectedOpCount  int
		expectedOpType   any
		validateOp       func(t *testing.T, op GitHubOperation)
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

func TestCreateIssueOperationCallsCreator(t *testing.T) {
	todoItem := todo.TodoItem{
		Text:        "New task",
		IsChecked:   false,
		IssueNumber: nil,
	}
	githubOperations := []TodoOperation{
		{
			Todo: todoItem,
			Operation: CreateIssueOp{
				Title: "New task",
			},
		},
	}

	mockCreator := func(title string) (uint64, error) {
		if title != "New task" {
			t.Errorf("expected title 'New task', got '%s'", title)
		}
		return 789, nil
	}
	mockCloser := func(number uint64) error {
		return nil
	}

	updates, err := CalculateTodoUpdates(githubOperations, mockCreator, mockCloser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}

	if updates[0].Todo.Text != "New task" {
		t.Errorf("expected todo text 'New task', got '%s'", updates[0].Todo.Text)
	}

	if updates[0].IssueNumber == nil {
		t.Fatal("expected issue number to be set")
	}

	if *updates[0].IssueNumber != 789 {
		t.Errorf("expected issue number 789, got %d", *updates[0].IssueNumber)
	}
}

func TestCloseIssueOperationCallsCloser(t *testing.T) {
	issueNum := uint64(123)
	todoItem := todo.TodoItem{
		Text:        "Completed task",
		IsChecked:   true,
		IssueNumber: &issueNum,
	}
	githubOperations := []TodoOperation{
		{
			Todo: todoItem,
			Operation: CloseIssueOp{
				Number: 123,
			},
		},
	}

	mockCreator := func(title string) (uint64, error) {
		return 0, errors.New("should not be called")
	}
	mockCloser := func(number uint64) error {
		if number != 123 {
			t.Errorf("expected issue number 123, got %d", number)
		}
		return nil
	}

	updates, err := CalculateTodoUpdates(githubOperations, mockCreator, mockCloser)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}

	if updates[0].Todo.Text != "Completed task" {
		t.Errorf("expected todo text 'Completed task', got '%s'", updates[0].Todo.Text)
	}

	if updates[0].IssueNumber != nil {
		t.Errorf("expected issue number to be nil, got %v", *updates[0].IssueNumber)
	}
}

func TestCreateIssueOperationPropagatesError(t *testing.T) {
	todoItem := todo.TodoItem{
		Text:        "New task",
		IsChecked:   false,
		IssueNumber: nil,
	}
	githubOperations := []TodoOperation{
		{
			Todo: todoItem,
			Operation: CreateIssueOp{
				Title: "New task",
			},
		},
	}

	expectedErr := errors.New("failed to create issue")
	mockCreator := func(title string) (uint64, error) {
		return 0, expectedErr
	}
	mockCloser := func(number uint64) error {
		t.Error("closer should not be called")
		return nil
	}

	updates, err := CalculateTodoUpdates(githubOperations, mockCreator, mockCloser)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	if updates != nil {
		t.Errorf("expected nil updates on error, got %v", updates)
	}
}

func TestCloseIssueOperationPropagatesError(t *testing.T) {
	issueNum := uint64(123)
	todoItem := todo.TodoItem{
		Text:        "Completed task",
		IsChecked:   true,
		IssueNumber: &issueNum,
	}
	githubOperations := []TodoOperation{
		{
			Todo: todoItem,
			Operation: CloseIssueOp{
				Number: 123,
			},
		},
	}

	expectedErr := errors.New("failed to close issue")
	mockCreator := func(title string) (uint64, error) {
		t.Error("creator should not be called")
		return 0, nil
	}
	mockCloser := func(number uint64) error {
		return expectedErr
	}

	updates, err := CalculateTodoUpdates(githubOperations, mockCreator, mockCloser)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != expectedErr {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}

	if updates != nil {
		t.Errorf("expected nil updates on error, got %v", updates)
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

func TestCalculateTitleUpdatesWithHistoryResolvesByRenameHistory(t *testing.T) {
	issueNum123 := uint64(123)
	issueNum456 := uint64(456)
	todoItems := []todo.TodoItem{
		{Text: "Locally edited title", IsChecked: false, IssueNumber: &issueNum123},
		{Text: "Old title", IsChecked: false, IssueNumber: &issueNum456},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Original title", State: IssueStateOpen},
		{Number: 456, Title: "New title", State: IssueStateOpen},
	}
	eventsFetcher := func(issueNumber uint64) ([]json.RawMessage, error) {
		switch issueNumber {
		case 123:
			return []json.RawMessage{}, nil
		case 456:
			return []json.RawMessage{
				json.RawMessage(`{"event": "renamed", "rename": {"from": "Old title", "to": "New title"}}`),
			}, nil
		default:
			return nil, fmt.Errorf("history should not be fetched for issue #%d", issueNumber)
		}
	}

	result, err := CalculateTitleUpdatesWithHistory(todoItems, githubIssues, eventsFetcher)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(result.Operations))
	}
	renameOp := result.Operations[0].Operation.(RenameIssueOp)
	if renameOp.Number != 123 {
		t.Errorf("expected issue 123, got %d", renameOp.Number)
	}
	if renameOp.Title != "Locally edited title" {
		t.Errorf("expected 'Locally edited title', got '%s'", renameOp.Title)
	}
	if len(result.StaleIssues) != 1 {
		t.Fatalf("expected 1 stale issue, got %d", len(result.StaleIssues))
	}
	if result.StaleIssues[0] != 456 {
		t.Errorf("expected stale issue 456, got %d", result.StaleIssues[0])
	}
}

func TestRenameIssueOperationUpdatesNothingInTodo(t *testing.T) {
	issueNum123 := uint64(123)
	todoItem := todo.TodoItem{
		Text:        "Edited title",
		IsChecked:   false,
		IssueNumber: &issueNum123,
	}
	githubOperations := []TodoOperation{
		{
			Todo: todoItem,
			Operation: RenameIssueOp{
				Number: 123,
				Title:  "Edited title",
			},
		},
	}

	mockCreator := func(title string) (uint64, error) {
		t.Fatal("creator should not be called")
		return 0, nil
	}
	mockCloser := func(number uint64) error {
		t.Fatal("closer should not be called")
		return nil
	}

	updates, err := CalculateTodoUpdates(githubOperations, mockCreator, mockCloser)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].IssueNumber != nil {
		t.Errorf("expected nil issue number, got %v", updates[0].IssueNumber)
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

func TestCalculateTitleUpdatesWithHistoryPropagatesFetcherError(t *testing.T) {
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

	result, err := CalculateTitleUpdatesWithHistory(todoItems, githubIssues, eventsFetcher)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "Network error" {
		t.Errorf("expected 'Network error', got '%s'", err.Error())
	}
	if len(result.Operations) != 0 {
		t.Errorf("expected empty operations, got %v", result.Operations)
	}
}
