package github

import (
	"errors"
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
