package github

import (
	"encoding/json"
	"fmt"

	"github.com/toms74209200/gh-atat/internal/todo"
)

// GitHubOperation represents an operation to perform on GitHub
type GitHubOperation interface {
	isGitHubOperation()
}

// CreateIssueOp represents creating a new GitHub issue
type CreateIssueOp struct {
	Title string
}

func (CreateIssueOp) isGitHubOperation() {}

// CloseIssueOp represents closing an existing GitHub issue
type CloseIssueOp struct {
	Number uint64
}

func (CloseIssueOp) isGitHubOperation() {}

// RenameIssueOp represents renaming an existing GitHub issue
type RenameIssueOp struct {
	Number uint64
	Title  string
}

func (RenameIssueOp) isGitHubOperation() {}

// TodoOperation represents a todo item with its associated GitHub operation
type TodoOperation struct {
	Todo      todo.TodoItem
	Operation GitHubOperation
}

// CalculateGitHubOperations determines what GitHub operations need to be performed
// based on the current state of todo items and GitHub issues
func CalculateGitHubOperations(todoItems []todo.TodoItem, githubIssues []GitHubIssue) []TodoOperation {
	var operations []TodoOperation

	for _, todoItem := range todoItems {
		var op GitHubOperation

		switch {
		// Unchecked todo without issue number -> create new issue
		case !todoItem.IsChecked && todoItem.IssueNumber == nil:
			op = CreateIssueOp{Title: todoItem.Text}

		// Checked todo with issue number -> close issue if it's open
		case todoItem.IsChecked && todoItem.IssueNumber != nil:
			issueNum := *todoItem.IssueNumber
			for _, issue := range githubIssues {
				if issue.Number == issueNum && issue.State == IssueStateOpen {
					op = CloseIssueOp{Number: issueNum}
					break
				}
			}

		// All other cases -> no operation
		default:
			continue
		}

		if op != nil {
			operations = append(operations, TodoOperation{
				Todo:      todoItem,
				Operation: op,
			})
		}
	}

	return operations
}

// TitleUpdates holds the result of title update calculation during push.
// Operations contains rename operations to perform, and StaleIssues contains
// issue numbers where the remote was renamed (local text is stale).
type TitleUpdates struct {
	Operations  []TodoOperation
	StaleIssues []uint64
}

// CalculateTitleUpdates determines title update operations based on rename history.
// If local text matches a past title, the issue is marked as stale (remote renamed).
// If local text doesn't match any past title, a RenameIssueOp is generated.
func CalculateTitleUpdates(todoItems []todo.TodoItem, githubIssues []GitHubIssue, pastTitles map[uint64][]string) TitleUpdates {
	githubIssuesMap := make(map[uint64]GitHubIssue)
	for _, issue := range githubIssues {
		githubIssuesMap[issue.Number] = issue
	}

	var operations []TodoOperation
	var staleIssues []uint64

	for _, todoItem := range todoItems {
		if todoItem.IssueNumber == nil {
			continue
		}
		ghIssue, exists := githubIssuesMap[*todoItem.IssueNumber]
		if !exists {
			continue
		}
		if ghIssue.State != IssueStateOpen {
			continue
		}
		if trimString(todoItem.Text) == trimString(ghIssue.Title) {
			continue
		}

		if MatchesPastTitle(pastTitles, ghIssue.Number, todoItem.Text) {
			staleIssues = append(staleIssues, ghIssue.Number)
		} else {
			operations = append(operations, TodoOperation{
				Todo: todoItem,
				Operation: RenameIssueOp{
					Number: ghIssue.Number,
					Title:  trimString(todoItem.Text),
				},
			})
		}
	}

	return TitleUpdates{
		Operations:  operations,
		StaleIssues: staleIssues,
	}
}

func ParseCreatedIssueNumber(body []byte) (uint64, error) {
	var issue struct {
		Number *uint64 `json:"number"`
	}
	if err := json.Unmarshal(body, &issue); err != nil {
		return 0, err
	}
	if issue.Number == nil {
		return 0, fmt.Errorf("created issue has no number")
	}
	return *issue.Number, nil
}
