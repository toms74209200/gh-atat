package github

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/toms74209200/gh-atat/internal/todo"
)

func TestParseGitHubIssuesWithValidIssues(t *testing.T) {
	issuesJSON := []json.RawMessage{
		json.RawMessage(`{
			"number": 123,
			"title": "Test issue",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": 456,
			"title": "Closed issue",
			"state": "closed",
			"pull_request": null
		}`),
	}

	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
	if issues[0].Title != "Test issue" {
		t.Errorf("Expected title 'Test issue', got '%s'", issues[0].Title)
	}
	if issues[0].State != IssueStateOpen {
		t.Errorf("Expected state Open, got %v", issues[0].State)
	}
	if issues[1].Number != 456 {
		t.Errorf("Expected number 456, got %d", issues[1].Number)
	}
	if issues[1].Title != "Closed issue" {
		t.Errorf("Expected title 'Closed issue', got '%s'", issues[1].Title)
	}
	if issues[1].State != IssueStateClosed {
		t.Errorf("Expected state Closed, got %v", issues[1].State)
	}
}

func TestParseGitHubIssuesFiltersPullRequests(t *testing.T) {
	issuesJSON := []json.RawMessage{
		json.RawMessage(`{
			"number": 123,
			"title": "Regular issue",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": 456,
			"title": "Pull request",
			"state": "open",
			"pull_request": {"url": "https://api.github.com/repos/user/repo/pulls/456"}
		}`),
	}

	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
	if issues[0].Title != "Regular issue" {
		t.Errorf("Expected title 'Regular issue', got '%s'", issues[0].Title)
	}
}

func TestParseGitHubIssuesIgnoresInvalidState(t *testing.T) {
	issuesJSON := []json.RawMessage{
		json.RawMessage(`{
			"number": 123,
			"title": "Valid issue",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": 456,
			"title": "Invalid state",
			"state": "unknown",
			"pull_request": null
		}`),
	}

	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
}

func TestParseGitHubIssuesIgnoresMissingFields(t *testing.T) {
	issuesJSON := []json.RawMessage{
		json.RawMessage(`{
			"number": 123,
			"title": "Valid issue",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"title": "Missing number",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": 456,
			"state": "open",
			"pull_request": null
		}`),
	}

	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
}

func TestFetchGitHubIssuesSinglePage(t *testing.T) {
	mockFetcher := func(repo string, token string, page int, perPage int) ([]json.RawMessage, error) {
		if page == 1 {
			return []json.RawMessage{
				json.RawMessage(`{
					"number": 123,
					"title": "Test issue",
					"state": "open",
					"pull_request": null
				}`),
			}, nil
		}
		return []json.RawMessage{}, nil
	}

	issues, err := FetchGitHubIssues("user/repo", "token", mockFetcher)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
}

func TestFetchGitHubIssuesMultiplePages(t *testing.T) {
	mockFetcher := func(repo string, token string, page int, perPage int) ([]json.RawMessage, error) {
		switch page {
		case 1:
			return []json.RawMessage{
				json.RawMessage(`{
					"number": 123,
					"title": "First issue",
					"state": "open",
					"pull_request": null
				}`),
			}, nil
		case 2:
			return []json.RawMessage{
				json.RawMessage(`{
					"number": 456,
					"title": "Second issue",
					"state": "closed",
					"pull_request": null
				}`),
			}, nil
		default:
			return []json.RawMessage{}, nil
		}
	}

	issues, err := FetchGitHubIssues("user/repo", "token", mockFetcher)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
	if issues[1].Number != 456 {
		t.Errorf("Expected number 456, got %d", issues[1].Number)
	}
}

func TestFetchGitHubIssuesEmptyResponse(t *testing.T) {
	mockFetcher := func(repo string, token string, page int, perPage int) ([]json.RawMessage, error) {
		return []json.RawMessage{}, nil
	}

	issues, err := FetchGitHubIssues("user/repo", "token", mockFetcher)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestFetchGitHubIssuesErrorHandling(t *testing.T) {
	mockFetcher := func(repo string, token string, page int, perPage int) ([]json.RawMessage, error) {
		return nil, fmt.Errorf("Network error")
	}

	issues, err := FetchGitHubIssues("user/repo", "token", mockFetcher)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "failed to fetch issues: Network error" {
		t.Errorf("Expected 'failed to fetch issues: Network error', got '%s'", err.Error())
	}
	if issues != nil {
		t.Errorf("Expected nil issues, got %v", issues)
	}
}

func TestFetchGitHubIssuesExceedsMaxPageLimit(t *testing.T) {
	mockFetcher := func(repo string, token string, page int, perPage int) ([]json.RawMessage, error) {
		// Always return non-empty result to simulate infinite pagination
		return []json.RawMessage{
			json.RawMessage(`{
				"number": 1,
				"title": "Test issue",
				"state": "open",
				"pull_request": null
			}`),
		}, nil
	}

	issues, err := FetchGitHubIssues("user/repo", "token", mockFetcher)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != nil && err.Error() != "exceeded maximum page limit" {
		t.Errorf("Expected 'exceeded maximum page limit', got '%s'", err.Error())
	}
	if issues != nil {
		t.Errorf("Expected nil issues, got %v", issues)
	}
}

func TestParseGitHubIssuesEmptyArray(t *testing.T) {
	issuesJSON := []json.RawMessage{}
	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestParseGitHubIssuesAllInvalid(t *testing.T) {
	issuesJSON := []json.RawMessage{
		json.RawMessage(`{}`),
		json.RawMessage(`{"invalid": "data"}`),
		json.RawMessage(`{"number": "not_a_number"}`),
	}
	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestParseGitHubIssuesPartialValid(t *testing.T) {
	issuesJSON := []json.RawMessage{
		json.RawMessage(`{
			"number": 123,
			"title": "Valid issue",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": 456,
			"title": "Invalid state issue",
			"state": "invalid",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": 789,
			"title": "Another valid issue",
			"state": "closed",
			"pull_request": null
		}`),
	}

	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
	if issues[0].State != IssueStateOpen {
		t.Errorf("Expected state Open, got %v", issues[0].State)
	}
	if issues[1].Number != 789 {
		t.Errorf("Expected number 789, got %d", issues[1].Number)
	}
	if issues[1].State != IssueStateClosed {
		t.Errorf("Expected state Closed, got %v", issues[1].State)
	}
}

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

func TestParseGitHubIssuesNumberTypeVariants(t *testing.T) {
	issuesJSON := []json.RawMessage{
		json.RawMessage(`{
			"number": 123,
			"title": "Valid u64 number",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": "456",
			"title": "String number should be ignored",
			"state": "open",
			"pull_request": null
		}`),
		json.RawMessage(`{
			"number": 789.5,
			"title": "Float number should be ignored",
			"state": "open",
			"pull_request": null
		}`),
	}

	issues := ParseGitHubIssues(issuesJSON)

	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].Number != 123 {
		t.Errorf("Expected number 123, got %d", issues[0].Number)
	}
}
