package clean

import (
	"github.com/toms74209200/gh-atat/internal/github"
	"github.com/toms74209200/gh-atat/internal/todo"
)

// CleanCandidate represents a checked todo item with an associated issue number
// that is eligible for removal.
type CleanCandidate struct {
	Text        string
	IssueNumber uint64
}

// NewCleanCandidate converts a checked TodoItem with an issue number into a CleanCandidate.
// Returns the candidate and true if eligible, or a zero value and false otherwise.
func NewCleanCandidate(item todo.TodoItem) (CleanCandidate, bool) {
	if item.IsChecked && item.IssueNumber != nil {
		return CleanCandidate{
			Text:        item.Text,
			IssueNumber: *item.IssueNumber,
		}, true
	}
	return CleanCandidate{}, false
}

// FindRemovableItems filters candidates to only those whose associated GitHub issue is closed.
func FindRemovableItems(candidates []CleanCandidate, issues []github.GitHubIssue) []CleanCandidate {
	var removable []CleanCandidate
	for _, candidate := range candidates {
		for _, issue := range issues {
			if issue.Number == candidate.IssueNumber && issue.State == github.IssueStateClosed {
				removable = append(removable, candidate)
				break
			}
		}
	}
	return removable
}
