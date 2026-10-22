package github

import (
	"testing"

	"github.com/toms74209200/gh-atat/internal/todo"
)

func TestSynchronizeWithGitHubIssuesUpdatesClosedIssues(t *testing.T) {
	issueNum123 := uint64(123)
	issueNum456 := uint64(456)
	todoItems := []todo.TodoItem{
		{
			Text:        "Fix bug",
			IsChecked:   false,
			IssueNumber: &issueNum123,
		},
		{
			Text:        "Add feature",
			IsChecked:   false,
			IssueNumber: &issueNum456,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Fix bug",
			State:  IssueStateClosed,
		},
		{
			Number: 456,
			Title:  "Add feature",
			State:  IssueStateOpen,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result[0].Text != "Fix bug" {
		t.Errorf("Expected text 'Fix bug', got '%s'", result[0].Text)
	}
	if result[0].IsChecked != true {
		t.Errorf("Expected is_checked true, got %v", result[0].IsChecked)
	}
	if result[0].IssueNumber == nil || *result[0].IssueNumber != 123 {
		t.Errorf("Expected issue_number 123, got %v", result[0].IssueNumber)
	}
	if result[1].Text != "Add feature" {
		t.Errorf("Expected text 'Add feature', got '%s'", result[1].Text)
	}
	if result[1].IsChecked != false {
		t.Errorf("Expected is_checked false, got %v", result[1].IsChecked)
	}
	if result[1].IssueNumber == nil || *result[1].IssueNumber != 456 {
		t.Errorf("Expected issue_number 456, got %v", result[1].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesAddsNewOpenIssues(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{
			Text:        "Existing task",
			IsChecked:   false,
			IssueNumber: &issueNum123,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Existing task",
			State:  IssueStateOpen,
		},
		{
			Number: 456,
			Title:  "New task",
			State:  IssueStateOpen,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result[0].Text != "Existing task" {
		t.Errorf("Expected text 'Existing task', got '%s'", result[0].Text)
	}
	if result[0].IssueNumber == nil || *result[0].IssueNumber != 123 {
		t.Errorf("Expected issue_number 123, got %v", result[0].IssueNumber)
	}
	if result[1].Text != "New task" {
		t.Errorf("Expected text 'New task', got '%s'", result[1].Text)
	}
	if result[1].IsChecked != false {
		t.Errorf("Expected is_checked false, got %v", result[1].IsChecked)
	}
	if result[1].IssueNumber == nil || *result[1].IssueNumber != 456 {
		t.Errorf("Expected issue_number 456, got %v", result[1].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesSkipsAlreadyChecked(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{
			Text:        "Completed task",
			IsChecked:   true,
			IssueNumber: &issueNum123,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Completed task",
			State:  IssueStateClosed,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0].Text != "Completed task" {
		t.Errorf("Expected text 'Completed task', got '%s'", result[0].Text)
	}
	if result[0].IsChecked != true {
		t.Errorf("Expected is_checked true, got %v", result[0].IsChecked)
	}
	if result[0].IssueNumber == nil || *result[0].IssueNumber != 123 {
		t.Errorf("Expected issue_number 123, got %v", result[0].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesIgnoresClosedIssuesForNewTodos(t *testing.T) {
	todoItems := []todo.TodoItem{}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Closed issue",
			State:  IssueStateClosed,
		},
		{
			Number: 456,
			Title:  "Open issue",
			State:  IssueStateOpen,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0].Text != "Open issue" {
		t.Errorf("Expected text 'Open issue', got '%s'", result[0].Text)
	}
	if result[0].IsChecked != false {
		t.Errorf("Expected is_checked false, got %v", result[0].IsChecked)
	}
	if result[0].IssueNumber == nil || *result[0].IssueNumber != 456 {
		t.Errorf("Expected issue_number 456, got %v", result[0].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesPreservesTodoWithoutIssueNumber(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{
			Text:        "Local task",
			IsChecked:   false,
			IssueNumber: nil,
		},
		{
			Text:        "Task with issue",
			IsChecked:   false,
			IssueNumber: &issueNum123,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Task with issue",
			State:  IssueStateClosed,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}
	if result[0].Text != "Local task" {
		t.Errorf("Expected text 'Local task', got '%s'", result[0].Text)
	}
	if result[0].IsChecked != false {
		t.Errorf("Expected is_checked false, got %v", result[0].IsChecked)
	}
	if result[0].IssueNumber != nil {
		t.Errorf("Expected issue_number nil, got %v", result[0].IssueNumber)
	}
	if result[1].Text != "Task with issue" {
		t.Errorf("Expected text 'Task with issue', got '%s'", result[1].Text)
	}
	if result[1].IsChecked != true {
		t.Errorf("Expected is_checked true, got %v", result[1].IsChecked)
	}
	if result[1].IssueNumber == nil || *result[1].IssueNumber != 123 {
		t.Errorf("Expected issue_number 123, got %v", result[1].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesAvoidsDuplicateByTitle(t *testing.T) {
	todoItems := []todo.TodoItem{
		{
			Text:        "Same title task",
			IsChecked:   false,
			IssueNumber: nil,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Same title task",
			State:  IssueStateOpen,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0].Text != "Same title task" {
		t.Errorf("Expected text 'Same title task', got '%s'", result[0].Text)
	}
	if result[0].IssueNumber != nil {
		t.Errorf("Expected issue_number nil, got %v", result[0].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesAvoidsDuplicateByTitleWithTrim(t *testing.T) {
	todoItems := []todo.TodoItem{
		{
			Text:        "  Task with spaces  ",
			IsChecked:   false,
			IssueNumber: nil,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Task with spaces",
			State:  IssueStateOpen,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0].Text != "  Task with spaces  " {
		t.Errorf("Expected text '  Task with spaces  ', got '%s'", result[0].Text)
	}
	if result[0].IssueNumber != nil {
		t.Errorf("Expected issue_number nil, got %v", result[0].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesNoMatchingIssue(t *testing.T) {
	issueNum999 := uint64(999)
	todoItems := []todo.TodoItem{
		{
			Text:        "Task without matching issue",
			IsChecked:   false,
			IssueNumber: &issueNum999,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 123,
			Title:  "Different issue",
			State:  IssueStateClosed,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result))
	}
	if result[0].Text != "Task without matching issue" {
		t.Errorf("Expected text 'Task without matching issue', got '%s'", result[0].Text)
	}
	if result[0].IsChecked != false {
		t.Errorf("Expected is_checked false, got %v", result[0].IsChecked)
	}
	if result[0].IssueNumber == nil || *result[0].IssueNumber != 999 {
		t.Errorf("Expected issue_number 999, got %v", result[0].IssueNumber)
	}
}

func TestSynchronizeWithGitHubIssuesEmptyInputs(t *testing.T) {
	todoItems := []todo.TodoItem{}
	githubIssues := []GitHubIssue{}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 0 {
		t.Errorf("Expected 0 items, got %d", len(result))
	}
}

func TestSynchronizeWithGitHubIssuesComplexScenario(t *testing.T) {
	issueNum100 := uint64(100)
	issueNum200 := uint64(200)
	todoItems := []todo.TodoItem{
		{
			Text:        "To be closed",
			IsChecked:   false,
			IssueNumber: &issueNum100,
		},
		{
			Text:        "Already closed",
			IsChecked:   true,
			IssueNumber: &issueNum200,
		},
		{
			Text:        "Local only task",
			IsChecked:   false,
			IssueNumber: nil,
		},
	}
	githubIssues := []GitHubIssue{
		{
			Number: 100,
			Title:  "To be closed",
			State:  IssueStateClosed,
		},
		{
			Number: 200,
			Title:  "Already closed",
			State:  IssueStateClosed,
		},
		{
			Number: 300,
			Title:  "New open issue",
			State:  IssueStateOpen,
		},
		{
			Number: 400,
			Title:  "Closed new issue",
			State:  IssueStateClosed,
		},
	}

	result := SynchronizeWithGitHubIssues(todoItems, githubIssues)

	if len(result) != 4 {
		t.Errorf("Expected 4 items, got %d", len(result))
	}
	if result[0].Text != "To be closed" {
		t.Errorf("Expected text 'To be closed', got '%s'", result[0].Text)
	}
	if result[0].IsChecked != true {
		t.Errorf("Expected is_checked true, got %v", result[0].IsChecked)
	}
	if result[1].Text != "Already closed" {
		t.Errorf("Expected text 'Already closed', got '%s'", result[1].Text)
	}
	if result[1].IsChecked != true {
		t.Errorf("Expected is_checked true, got %v", result[1].IsChecked)
	}
	if result[2].Text != "Local only task" {
		t.Errorf("Expected text 'Local only task', got '%s'", result[2].Text)
	}
	if result[2].IsChecked != false {
		t.Errorf("Expected is_checked false, got %v", result[2].IsChecked)
	}
	if result[2].IssueNumber != nil {
		t.Errorf("Expected issue_number nil, got %v", result[2].IssueNumber)
	}
	if result[3].Text != "New open issue" {
		t.Errorf("Expected text 'New open issue', got '%s'", result[3].Text)
	}
	if result[3].IsChecked != false {
		t.Errorf("Expected is_checked false, got %v", result[3].IsChecked)
	}
	if result[3].IssueNumber == nil || *result[3].IssueNumber != 300 {
		t.Errorf("Expected issue_number 300, got %v", result[3].IssueNumber)
	}
}

func TestSynchronizeTitlesUpdatesTextWhenRemoteRenamed(t *testing.T) {
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

	result := SynchronizeTitles(todoItems, githubIssues, pastTitles)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Text != "New title" {
		t.Errorf("expected 'New title', got '%s'", result.Items[0].Text)
	}
	if result.Items[0].IssueNumber == nil || *result.Items[0].IssueNumber != 123 {
		t.Errorf("expected issue number 123, got %v", result.Items[0].IssueNumber)
	}
	if len(result.LocallyEditedIssues) != 0 {
		t.Errorf("expected no locally edited issues, got %v", result.LocallyEditedIssues)
	}
}

func TestSynchronizeTitlesKeepsLocalEditAndReportsIt(t *testing.T) {
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

	result := SynchronizeTitles(todoItems, githubIssues, pastTitles)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Text != "Locally edited title" {
		t.Errorf("expected 'Locally edited title', got '%s'", result.Items[0].Text)
	}
	if len(result.LocallyEditedIssues) != 1 {
		t.Fatalf("expected 1 locally edited issue, got %d", len(result.LocallyEditedIssues))
	}
	if result.LocallyEditedIssues[0] != 123 {
		t.Errorf("expected issue 123, got %d", result.LocallyEditedIssues[0])
	}
}

func TestSynchronizeTitlesIgnoresClosedIssues(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "Old title", IsChecked: true, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "New title", State: IssueStateClosed},
	}
	pastTitles := map[uint64][]string{
		123: {"Old title"},
	}

	result := SynchronizeTitles(todoItems, githubIssues, pastTitles)

	if result.Items[0].Text != "Old title" {
		t.Errorf("expected 'Old title', got '%s'", result.Items[0].Text)
	}
	if len(result.LocallyEditedIssues) != 0 {
		t.Errorf("expected no locally edited issues, got %v", result.LocallyEditedIssues)
	}
}

func TestSynchronizeTitlesKeepsItemsInSyncUntouched(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "Same title", IsChecked: false, IssueNumber: &issueNum123},
		{Text: "Local task", IsChecked: false, IssueNumber: nil},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "Same title", State: IssueStateOpen},
	}
	pastTitles := map[uint64][]string{}

	result := SynchronizeTitles(todoItems, githubIssues, pastTitles)

	if result.Items[0].Text != "Same title" {
		t.Errorf("expected 'Same title', got '%s'", result.Items[0].Text)
	}
	if result.Items[1].Text != "Local task" {
		t.Errorf("expected 'Local task', got '%s'", result.Items[1].Text)
	}
	if len(result.LocallyEditedIssues) != 0 {
		t.Errorf("expected no locally edited issues, got %v", result.LocallyEditedIssues)
	}
}

func TestSynchronizeTitlesMatchesPastTitleWithTrim(t *testing.T) {
	issueNum123 := uint64(123)
	todoItems := []todo.TodoItem{
		{Text: "  Old title  ", IsChecked: false, IssueNumber: &issueNum123},
	}
	githubIssues := []GitHubIssue{
		{Number: 123, Title: "New title", State: IssueStateOpen},
	}
	pastTitles := map[uint64][]string{
		123: {"Old title"},
	}

	result := SynchronizeTitles(todoItems, githubIssues, pastTitles)

	if result.Items[0].Text != "New title" {
		t.Errorf("expected 'New title', got '%s'", result.Items[0].Text)
	}
}
