package github

import (
	"encoding/json"
	"fmt"
)

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

func ParseGitHubIssue(data []byte) (GitHubIssue, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return GitHubIssue{}, err
	}

	number, numberOk := raw["number"].(float64)
	title, titleOk := raw["title"].(string)
	stateStr, stateOk := raw["state"].(string)

	if !numberOk || number != float64(uint64(number)) {
		return GitHubIssue{}, fmt.Errorf("issue has no integer number")
	}

	if !titleOk || !stateOk {
		return GitHubIssue{}, fmt.Errorf("issue has no title or state")
	}

	if pullRequest, exists := raw["pull_request"]; exists && pullRequest != nil {
		return GitHubIssue{}, fmt.Errorf("issue is a pull request")
	}

	var state IssueState
	switch stateStr {
	case "open":
		state = IssueStateOpen
	case "closed":
		state = IssueStateClosed
	default:
		return GitHubIssue{}, fmt.Errorf("unknown issue state: %s", stateStr)
	}

	return GitHubIssue{
		Number: uint64(number),
		Title:  title,
		State:  state,
	}, nil
}
