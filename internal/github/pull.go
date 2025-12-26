package github

import (
	"encoding/json"
	"fmt"

	"github.com/toms74209200/gh-atat/internal/todo"
)

// IssueFetcher is a function type that fetches issues from GitHub API
// Parameters: repo, token, page, perPage
// Returns: JSON values and error
type IssueFetcher func(repo string, token string, page int, perPage int) ([]json.RawMessage, error)

// ParseGitHubIssues parses JSON data and returns a list of GitHubIssue
func ParseGitHubIssues(issuesJSON []json.RawMessage) []GitHubIssue {
	var issues []GitHubIssue

	for _, issueJSON := range issuesJSON {
		var raw map[string]interface{}
		if err := json.Unmarshal(issueJSON, &raw); err != nil {
			continue
		}

		// Extract fields
		number, numberOk := raw["number"].(float64)
		title, titleOk := raw["title"].(string)
		stateStr, stateOk := raw["state"].(string)

		// Validate that number is an integer (not a float with decimal part)
		if !numberOk || number != float64(uint64(number)) {
			continue
		}

		if !titleOk || !stateOk {
			continue
		}

		// Filter out pull requests
		if pullRequest, exists := raw["pull_request"]; exists && pullRequest != nil {
			continue
		}

		// Parse state
		var state IssueState
		switch stateStr {
		case "open":
			state = IssueStateOpen
		case "closed":
			state = IssueStateClosed
		default:
			continue
		}

		issues = append(issues, GitHubIssue{
			Number: uint64(number),
			Title:  title,
			State:  state,
		})
	}

	return issues
}

// FetchGitHubIssues fetches all issues from GitHub with pagination
func FetchGitHubIssues(repo string, token string, fetcher IssueFetcher) ([]GitHubIssue, error) {
	const maxPages = 1000
	var allIssues []GitHubIssue
	page := 1
	perPage := 100

	for {
		issuesJSON, err := fetcher(repo, token, page, perPage)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}

		if len(issuesJSON) == 0 {
			break
		}

		parsedIssues := ParseGitHubIssues(issuesJSON)
		allIssues = append(allIssues, parsedIssues...)
		page++

		if page > maxPages {
			return nil, fmt.Errorf("exceeded maximum page limit")
		}
	}

	return allIssues, nil
}

// SynchronizeWithGitHubIssues synchronizes TodoItems with GitHub issues
func SynchronizeWithGitHubIssues(todoItems []todo.TodoItem, githubIssues []GitHubIssue) []todo.TodoItem {
	// Create a map of GitHub issues by number
	githubIssuesMap := make(map[uint64]GitHubIssue)
	for _, issue := range githubIssues {
		githubIssuesMap[issue.Number] = issue
	}

	// Update existing todo items
	var updatedItems []todo.TodoItem
	for _, todoItem := range todoItems {
		updated := todoItem

		// If todo has an issue number, check if it's closed on GitHub
		if todoItem.IssueNumber != nil {
			if githubIssue, exists := githubIssuesMap[*todoItem.IssueNumber]; exists {
				if githubIssue.State == IssueStateClosed && !todoItem.IsChecked {
					updated.IsChecked = true
				}
			}
		}

		updatedItems = append(updatedItems, updated)
	}

	// Add new items from GitHub issues that don't exist in todo list
	for _, githubIssue := range githubIssues {
		// Only add open issues
		if githubIssue.State != IssueStateOpen {
			continue
		}

		// Check if this issue already exists in updated items
		exists := false
		for _, todoItem := range updatedItems {
			// Check by issue number or by title
			if todoItem.IssueNumber != nil && *todoItem.IssueNumber == githubIssue.Number {
				exists = true
				break
			}
			// Compare titles with trim to handle whitespace
			if trimString(todoItem.Text) == trimString(githubIssue.Title) {
				exists = true
				break
			}
		}

		if !exists {
			issueNum := githubIssue.Number
			newItem := todo.TodoItem{
				Text:        githubIssue.Title,
				IsChecked:   false,
				IssueNumber: &issueNum,
			}
			updatedItems = append(updatedItems, newItem)
		}
	}

	return updatedItems
}

// trimString removes leading and trailing whitespace
func trimString(s string) string {
	start := 0
	end := len(s)

	// Find first non-whitespace character
	for start < len(s) && isWhitespace(s[start]) {
		start++
	}

	// Find last non-whitespace character
	for end > start && isWhitespace(s[end-1]) {
		end--
	}

	return s[start:end]
}

// isWhitespace checks if a byte is a whitespace character
func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
