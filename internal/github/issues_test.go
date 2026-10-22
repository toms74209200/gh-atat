package github

import (
	"crypto/rand"
	"fmt"
	mrand "math/rand/v2"
	"testing"
)

func TestWhenParseGitHubIssueWithIssueThenReturnsIssue(t *testing.T) {
	number := mrand.Uint64N(100000) + 1
	title := rand.Text()

	tests := []struct {
		name string
		data string
		want GitHubIssue
	}{
		{"open issue then returns open issue", fmt.Sprintf(`{"number":%d,"title":"%s","state":"open","pull_request":null}`, number, title), GitHubIssue{Number: number, Title: title, State: IssueStateOpen}},
		{"closed issue then returns closed issue", fmt.Sprintf(`{"number":%d,"title":"%s","state":"closed","pull_request":null}`, number, title), GitHubIssue{Number: number, Title: title, State: IssueStateClosed}},
		{"issue without pull_request field then returns issue", fmt.Sprintf(`{"number":%d,"title":"%s","state":"open"}`, number, title), GitHubIssue{Number: number, Title: title, State: IssueStateOpen}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGitHubIssue([]byte(tt.data))

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestWhenParseGitHubIssueWithNonIssueThenReturnsError(t *testing.T) {
	number := mrand.Uint64N(100000) + 1
	title := rand.Text()

	tests := []struct {
		name string
		data string
	}{
		{"pull request then returns error", fmt.Sprintf(`{"number":%d,"title":"%s","state":"open","pull_request":{"url":"%s"}}`, number, title, rand.Text())},
		{"unknown state then returns error", fmt.Sprintf(`{"number":%d,"title":"%s","state":"%s","pull_request":null}`, number, title, rand.Text())},
		{"missing number then returns error", fmt.Sprintf(`{"title":"%s","state":"open","pull_request":null}`, title)},
		{"missing title then returns error", fmt.Sprintf(`{"number":%d,"state":"open","pull_request":null}`, number)},
		{"missing state then returns error", fmt.Sprintf(`{"number":%d,"title":"%s","pull_request":null}`, number, title)},
		{"string number then returns error", fmt.Sprintf(`{"number":"%d","title":"%s","state":"open","pull_request":null}`, number, title)},
		{"float number then returns error", fmt.Sprintf(`{"number":%d.5,"title":"%s","state":"open","pull_request":null}`, number, title)},
		{"empty object then returns error", `{}`},
		{"object with unrelated field then returns error", fmt.Sprintf(`{"%s":"%s"}`, rand.Text(), rand.Text())},
		{"non-JSON then returns error", rand.Text()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGitHubIssue([]byte(tt.data))

			if err == nil {
				t.Errorf("expected error, got issue %+v", got)
			}
		})
	}
}
