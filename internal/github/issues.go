package github

// IssueState represents the state of a GitHub issue
type IssueState string

const (
	IssueStateOpen   IssueState = "open"
	IssueStateClosed IssueState = "closed"
)

// GitHubIssue represents a GitHub issue
type GitHubIssue struct {
	Number uint64
	Title  string
	State  IssueState
}
