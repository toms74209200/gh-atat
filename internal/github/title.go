package github

import (
	"encoding/json"

	"github.com/toms74209200/gh-atat/internal/todo"
)

// EventsFetcher is a function type that fetches timeline events for a GitHub issue
type EventsFetcher func(issueNumber uint64) ([]json.RawMessage, error)

// FindTitleMismatches returns issue numbers where the todo text doesn't match
// the GitHub issue title. Only open issues are considered.
func FindTitleMismatches(todoItems []todo.TodoItem, githubIssues []GitHubIssue) []uint64 {
	githubIssuesMap := make(map[uint64]GitHubIssue)
	for _, issue := range githubIssues {
		githubIssuesMap[issue.Number] = issue
	}

	var mismatches []uint64
	for _, todoItem := range todoItems {
		if todoItem.IssueNumber == nil {
			continue
		}
		githubIssue, exists := githubIssuesMap[*todoItem.IssueNumber]
		if !exists {
			continue
		}
		if githubIssue.State != IssueStateOpen {
			continue
		}
		if trimString(todoItem.Text) != trimString(githubIssue.Title) {
			mismatches = append(mismatches, githubIssue.Number)
		}
	}

	return mismatches
}

// ParsePastTitles extracts past titles from GitHub issue timeline events.
// It looks for "renamed" events and returns the "from" field of each rename.
func ParsePastTitles(eventsJSON []json.RawMessage) []string {
	var pastTitles []string
	for _, eventJSON := range eventsJSON {
		var raw map[string]interface{}
		if err := json.Unmarshal(eventJSON, &raw); err != nil {
			continue
		}
		eventType, ok := raw["event"].(string)
		if !ok || eventType != "renamed" {
			continue
		}
		rename, ok := raw["rename"].(map[string]interface{})
		if !ok {
			continue
		}
		from, ok := rename["from"].(string)
		if !ok {
			continue
		}
		pastTitles = append(pastTitles, from)
	}
	return pastTitles
}

// CollectPastTitles fetches past titles only for issues where the todo text
// doesn't match the current GitHub issue title.
func CollectPastTitles(todoItems []todo.TodoItem, githubIssues []GitHubIssue, eventsFetcher EventsFetcher) (map[uint64][]string, error) {
	pastTitles := make(map[uint64][]string)
	for _, issueNumber := range FindTitleMismatches(todoItems, githubIssues) {
		events, err := eventsFetcher(issueNumber)
		if err != nil {
			return nil, err
		}
		pastTitles[issueNumber] = ParsePastTitles(events)
	}
	return pastTitles, nil
}

// MatchesPastTitle checks if a text matches any past title for a given issue number.
// Comparison is done after trimming whitespace.
func MatchesPastTitle(pastTitles map[uint64][]string, issueNumber uint64, text string) bool {
	titles, exists := pastTitles[issueNumber]
	if !exists {
		return false
	}
	for _, title := range titles {
		if trimString(title) == trimString(text) {
			return true
		}
	}
	return false
}
