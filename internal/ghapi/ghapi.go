package ghapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Request interface {
	isRequest()
}

type ListIssues struct {
	Repo    string
	Page    int
	PerPage int
}

type ListIssueEvents struct {
	Repo   string
	Number uint64
}

type CreateIssue struct {
	Repo  string
	Title string
}

type CloseIssue struct {
	Repo   string
	Number uint64
}

type RenameIssue struct {
	Repo   string
	Number uint64
	Title  string
}

type GetRepository struct {
	Repo string
}

func (ListIssues) isRequest()      {}
func (ListIssueEvents) isRequest() {}
func (CreateIssue) isRequest()     {}
func (CloseIssue) isRequest()      {}
func (RenameIssue) isRequest()     {}
func (GetRepository) isRequest()   {}

var ErrNotFound = errors.New("not found")

func Execute(request Request) ([]byte, error) {
	var method, endpoint string
	var payload any
	switch r := request.(type) {
	case ListIssues:
		method, endpoint = "GET", fmt.Sprintf("repos/%s/issues?state=all&per_page=%d&page=%d", r.Repo, r.PerPage, r.Page)
	case ListIssueEvents:
		method, endpoint = "GET", fmt.Sprintf("repos/%s/issues/%d/events", r.Repo, r.Number)
	case CreateIssue:
		method, endpoint, payload = "POST", fmt.Sprintf("repos/%s/issues", r.Repo), map[string]string{"title": r.Title}
	case CloseIssue:
		method, endpoint, payload = "PATCH", fmt.Sprintf("repos/%s/issues/%d", r.Repo, r.Number), map[string]string{"state": "closed"}
	case RenameIssue:
		method, endpoint, payload = "PATCH", fmt.Sprintf("repos/%s/issues/%d", r.Repo, r.Number), map[string]string{"title": r.Title}
	case GetRepository:
		method, endpoint = "GET", fmt.Sprintf("repos/%s", r.Repo)
	default:
		return nil, fmt.Errorf("unknown request: %T", request)
	}

	cmd := exec.Command("gh", "api", "--include", "--method", method, endpoint)
	if payload != nil {
		input, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		cmd.Args = append(cmd.Args, "--input", "-")
		cmd.Stdin = bytes.NewReader(input)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	// gh api exits with 1 for every HTTP error status, so the status is read from the --include output instead of the exit code
	var status int
	_, scanErr := fmt.Sscanf(stdout.String(), "HTTP/%s %d", new(string), &status)
	detail := strings.TrimSpace(stderr.String())
	switch {
	case scanErr != nil && runErr != nil && detail != "":
		return nil, fmt.Errorf("gh api %s %s failed: %w: %s", method, endpoint, runErr, detail)
	case scanErr != nil && runErr != nil:
		return nil, fmt.Errorf("gh api %s %s failed: %w", method, endpoint, runErr)
	case scanErr != nil:
		return nil, fmt.Errorf("gh api %s %s: no HTTP response", method, endpoint)
	case status == 404:
		return nil, fmt.Errorf("gh api %s %s: %w: %s", method, endpoint, ErrNotFound, detail)
	case status < 200 || status >= 300:
		return nil, fmt.Errorf("gh api %s %s failed: %s", method, endpoint, detail)
	}

	_, body, _ := bytes.Cut(stdout.Bytes(), []byte("\r\n\r\n"))
	return body, nil
}
