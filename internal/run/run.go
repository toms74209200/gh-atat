package run

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/toms74209200/gh-atat/internal/cli"
	"github.com/toms74209200/gh-atat/internal/config"
	"github.com/toms74209200/gh-atat/internal/github"
	"github.com/toms74209200/gh-atat/internal/markdown"
	"github.com/toms74209200/gh-atat/internal/storage"
	"github.com/toms74209200/gh-atat/internal/todo"
)

// Run executes the given command
func Run(args []string) error {
	command := cli.ParseArgs(args)

	switch cmd := command.(type) {
	case cli.Push:
		return runPush()
	case cli.Pull:
		return runPull()
	case cli.RemoteList:
		return runRemoteList()
	case cli.RemoteAdd:
		return runRemoteAdd(cmd.Repo)
	case cli.RemoteRemove:
		return runRemoteRemove(cmd.Repo)
	case cli.Login:
		return fmt.Errorf("login command is not needed for gh extension. Authentication is handled by gh CLI")
	case cli.Whoami:
		return fmt.Errorf("whoami command is not needed for gh extension. Use 'gh auth status' instead")
	case cli.Help:
		printHelp()
		return nil
	case cli.Unknown:
		return fmt.Errorf("%s", cmd.Message)
	default:
		return fmt.Errorf("invalid command or arguments. Use --help for usage")
	}
}

func runPush() error {
	// Load configuration
	configStorage, err := storage.NewLocalConfigStorage()
	if err != nil {
		return fmt.Errorf("failed to read project configuration: %w", err)
	}

	configMap, err := configStorage.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading project config: %w", err)
	}

	// Get repository
	repo, err := getFirstRepository(configMap)
	if err != nil {
		return err
	}

	// Read TODO.md
	todoContent, err := os.ReadFile("TODO.md")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("TODO.md file not found")
		}
		return fmt.Errorf("failed to read TODO.md: %w", err)
	}

	todoItems, err := markdown.ParseTodoMarkdown(string(todoContent))
	if err != nil {
		return err
	}

	// Fetch GitHub issues
	githubIssues, err := fetchGitHubIssues(repo)
	if err != nil {
		return err
	}

	// Calculate operations
	operations := github.CalculateGitHubOperations(todoItems, githubIssues)

	// Execute operations
	updatedTodoItems := make([]todo.TodoItem, len(todoItems))
	copy(updatedTodoItems, todoItems)

	for _, todoOp := range operations {
		switch op := todoOp.Operation.(type) {
		case github.CreateIssueOp:
			issueNumber, err := createGitHubIssue(repo, op.Title)
			if err != nil {
				return err
			}
			fmt.Printf("Created issue #%d: %s\n", issueNumber, todoOp.Todo.Text)

			// Update TODO item with issue number
			issueNum := uint64(issueNumber)
			for j := range updatedTodoItems {
				if updatedTodoItems[j].Text == todoOp.Todo.Text &&
					updatedTodoItems[j].IsChecked == todoOp.Todo.IsChecked &&
					updatedTodoItems[j].IssueNumber == todoOp.Todo.IssueNumber {
					updatedTodoItems[j].IssueNumber = &issueNum
					break
				}
			}
		case github.CloseIssueOp:
			err := closeGitHubIssue(repo, int(op.Number))
			if err != nil {
				return err
			}
			fmt.Printf("Closed issue #%d\n", op.Number)
		}
	}

	// Write updated TODO.md
	updatedContent := markdown.SerializeTodoMarkdown(updatedTodoItems)
	if err := os.WriteFile("TODO.md", []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write TODO.md: %w", err)
	}

	return nil
}

func runPull() error {
	// Load configuration
	configStorage, err := storage.NewLocalConfigStorage()
	if err != nil {
		return fmt.Errorf("failed to read project configuration: %w", err)
	}

	configMap, err := configStorage.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading project config: %w", err)
	}

	// Get repository
	repo, err := getFirstRepository(configMap)
	if err != nil {
		return err
	}

	// Read TODO.md
	todoContent, err := os.ReadFile("TODO.md")
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("TODO.md file not found")
		}
		return fmt.Errorf("failed to read TODO.md: %w", err)
	}

	todoItems, err := markdown.ParseTodoMarkdown(string(todoContent))
	if err != nil {
		return err
	}

	// Fetch GitHub issues
	githubIssues, err := fetchGitHubIssues(repo)
	if err != nil {
		return err
	}

	// Synchronize with GitHub issues
	updatedTodoItems := github.SynchronizeWithGitHubIssues(todoItems, githubIssues)

	// Write updated TODO.md
	updatedContent := markdown.SerializeTodoMarkdown(updatedTodoItems)
	if err := os.WriteFile("TODO.md", []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write TODO.md: %w", err)
	}

	return nil
}

func runRemoteList() error {
	configStorage, err := storage.NewLocalConfigStorage()
	if err != nil {
		return fmt.Errorf("failed to read project configuration: %w", err)
	}

	configMap, err := configStorage.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading project config: %w", err)
	}

	if reposValue, ok := configMap[config.Repositories]; ok {
		if reposArray, ok := reposValue.([]interface{}); ok {
			for _, repoVal := range reposArray {
				if repoStr, ok := repoVal.(string); ok {
					fmt.Println(repoStr)
				}
			}
		}
	}

	return nil
}

func runRemoteAdd(repo string) error {
	configStorage, err := storage.NewLocalConfigStorage()
	if err != nil {
		return fmt.Errorf("error initializing config storage: %w", err)
	}

	configMap, _ := configStorage.LoadConfig()
	if configMap == nil {
		configMap = make(map[config.ConfigKey]any)
	}

	// Get or create repositories array
	var reposArray []interface{}
	if reposValue, ok := configMap[config.Repositories]; ok {
		if arr, ok := reposValue.([]interface{}); ok {
			reposArray = arr
		}
	}

	// Check if repository already exists
	for _, repoVal := range reposArray {
		if repoStr, ok := repoVal.(string); ok && repoStr == repo {
			// Already exists, nothing to do
			return nil
		}
	}

	// Check if repository exists on GitHub
	exists, err := checkRepoExists(repo)
	if err != nil {
		return fmt.Errorf("failed to check repository %s: %w", repo, err)
	}
	if !exists {
		return fmt.Errorf("repository %s not found or not accessible", repo)
	}

	// Add repository
	reposArray = append(reposArray, repo)
	configMap[config.Repositories] = reposArray

	// Save configuration
	if err := configStorage.SaveConfig(configMap); err != nil {
		return fmt.Errorf("error saving project config: %w", err)
	}

	return nil
}

func runRemoteRemove(repo string) error {
	configStorage, err := storage.NewLocalConfigStorage()
	if err != nil {
		return fmt.Errorf("error initializing config storage: %w", err)
	}

	configMap, _ := configStorage.LoadConfig()
	if configMap == nil {
		configMap = make(map[config.ConfigKey]any)
	}

	// Get repositories array
	var reposArray []interface{}
	if reposValue, ok := configMap[config.Repositories]; ok {
		if arr, ok := reposValue.([]interface{}); ok {
			reposArray = arr
		}
	}

	// Filter out the repository
	var filteredRepos []interface{}
	for _, repoVal := range reposArray {
		if repoStr, ok := repoVal.(string); ok && repoStr != repo {
			filteredRepos = append(filteredRepos, repoStr)
		}
	}

	// Update configuration
	if len(filteredRepos) == 0 {
		delete(configMap, config.Repositories)
	} else {
		configMap[config.Repositories] = filteredRepos
	}

	// Save configuration
	if err := configStorage.SaveConfig(configMap); err != nil {
		return fmt.Errorf("error saving project config: %w", err)
	}

	return nil
}

func getFirstRepository(configMap map[config.ConfigKey]any) (string, error) {
	reposValue, ok := configMap[config.Repositories]
	if !ok {
		return "", fmt.Errorf("no repository configured")
	}

	reposArray, ok := reposValue.([]interface{})
	if !ok || len(reposArray) == 0 {
		return "", fmt.Errorf("no repository configured")
	}

	repo, ok := reposArray[0].(string)
	if !ok {
		return "", fmt.Errorf("invalid repository configuration")
	}

	return repo, nil
}

func fetchGitHubIssues(repo string) ([]github.GitHubIssue, error) {
	fetchFunc := func(repo string, token string, page int, perPage int) ([]json.RawMessage, error) {
		endpoint := fmt.Sprintf("repos/%s/issues?state=all&per_page=%d&page=%d", repo, perPage, page)
		data, err := ghAPI(endpoint)
		if err != nil {
			return nil, err
		}

		var issues []json.RawMessage
		if err := json.Unmarshal(data, &issues); err != nil {
			return nil, err
		}
		return issues, nil
	}

	return github.FetchGitHubIssues(repo, "", fetchFunc)
}

func createGitHubIssue(repo, title string) (int, error) {
	body := map[string]string{"title": title}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	output, err := ghAPIPost(fmt.Sprintf("repos/%s/issues", repo), string(bodyJSON))
	if err != nil {
		return 0, err
	}

	var issue github.GitHubIssue
	if err := json.Unmarshal(output, &issue); err != nil {
		return 0, err
	}

	return int(issue.Number), nil
}

func closeGitHubIssue(repo string, number int) error {
	body := map[string]string{"state": "closed"}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return err
	}

	_, err = ghAPIPatch(fmt.Sprintf("repos/%s/issues/%d", repo, number), string(bodyJSON))
	return err
}

func checkRepoExists(repo string) (bool, error) {
	_, err := ghAPI(fmt.Sprintf("repos/%s", repo))
	if err != nil {
		// If error contains "404", repo doesn't exist
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func ghAPI(endpoint string) ([]byte, error) {
	cmd := exec.Command("gh", "api", endpoint)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh api failed: %w: %s", err, string(output))
	}
	return output, nil
}

func ghAPIPost(endpoint, body string) ([]byte, error) {
	cmd := exec.Command("gh", "api", endpoint, "-X", "POST", "--input", "-")
	cmd.Stdin = strings.NewReader(body)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh api POST failed: %w: %s", err, string(output))
	}
	return output, nil
}

func ghAPIPatch(endpoint, body string) ([]byte, error) {
	cmd := exec.Command("gh", "api", endpoint, "-X", "PATCH", "--input", "-")
	cmd.Stdin = strings.NewReader(body)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("gh api PATCH failed: %w: %s", err, string(output))
	}
	return output, nil
}

func printHelp() {
	help := `gh-atat: Automatic TODO and Tracker

Usage:
  gh atat <command> [arguments]

Commands:
  push          Push TODO items to GitHub Issues
  pull          Pull GitHub Issues to TODO items
  remote        List configured repositories
  remote add    Add a repository
  remote remove Remove a repository
  help          Show this help message

Examples:
  gh atat push
  gh atat pull
  gh atat remote
  gh atat remote add owner/repo
  gh atat remote remove owner/repo
`
	fmt.Print(help)
}
