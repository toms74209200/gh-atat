package github

import (
	"github.com/toms74209200/gh-atat/internal/todo"
)

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

// TitleSynchronization holds the result of title sync during pull.
// Items contains the updated todo items, and LocallyEditedIssues contains
// issue numbers where the local text was changed (not matching any past title).
type TitleSynchronization struct {
	Items               []todo.TodoItem
	LocallyEditedIssues []uint64
}

// SynchronizeTitles resolves title mismatches between todo items and GitHub issues
// using the rename history. If local text matches a past title, the text is updated
// to the current GitHub title. Otherwise, the item is kept as-is and the issue number
// is added to the locally edited list.
func SynchronizeTitles(todoItems []todo.TodoItem, githubIssues []GitHubIssue, pastTitles map[uint64][]string) TitleSynchronization {
	githubIssuesMap := make(map[uint64]GitHubIssue)
	for _, issue := range githubIssues {
		githubIssuesMap[issue.Number] = issue
	}

	var localEdits []uint64
	updatedItems := make([]todo.TodoItem, 0, len(todoItems))

	for _, todoItem := range todoItems {
		var renamedIssue *GitHubIssue

		if todoItem.IssueNumber != nil {
			if ghIssue, exists := githubIssuesMap[*todoItem.IssueNumber]; exists {
				if ghIssue.State == IssueStateOpen && trimString(todoItem.Text) != trimString(ghIssue.Title) {
					renamedIssue = &ghIssue
				}
			}
		}

		if renamedIssue == nil {
			updatedItems = append(updatedItems, todoItem)
			continue
		}

		if MatchesPastTitle(pastTitles, renamedIssue.Number, todoItem.Text) {
			updatedItems = append(updatedItems, todo.TodoItem{
				Text:        renamedIssue.Title,
				IsChecked:   todoItem.IsChecked,
				IssueNumber: todoItem.IssueNumber,
			})
		} else {
			localEdits = append(localEdits, renamedIssue.Number)
			updatedItems = append(updatedItems, todoItem)
		}
	}

	return TitleSynchronization{
		Items:               updatedItems,
		LocallyEditedIssues: localEdits,
	}
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
