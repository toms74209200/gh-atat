package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/toms74209200/gh-atat/internal/clean"
	"github.com/toms74209200/gh-atat/internal/cli"
	"github.com/toms74209200/gh-atat/internal/config"
	"github.com/toms74209200/gh-atat/internal/ghapi"
	"github.com/toms74209200/gh-atat/internal/github"
	"github.com/toms74209200/gh-atat/internal/markdown"
	"github.com/toms74209200/gh-atat/internal/storage"
	"github.com/toms74209200/gh-atat/internal/todo"
)

// Run executes the given command
func Run(args []string, version string) error {
	command := cli.ParseArgs(args)

	switch cmd := command.(type) {
	case cli.Push:
		return runPush()
	case cli.Pull:
		return runPull()
	case cli.Clean:
		return runClean(cmd.DryRun)
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
	case cli.Version:
		printVersion(version)
		return nil
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
	githubIssues, err := listIssues(repo)
	if err != nil {
		return err
	}

	// Calculate title updates with rename history
	pastTitles, err := collectPastTitles(repo, todoItems, githubIssues)
	if err != nil {
		return err
	}
	titleUpdates := github.CalculateTitleUpdates(todoItems, githubIssues, pastTitles)

	for _, issueNumber := range titleUpdates.StaleIssues {
		fmt.Printf("Warning: issue #%d was renamed on GitHub; run `gh atat pull` to update TODO.md\n", issueNumber)
	}

	// Calculate create/close operations
	operations := github.CalculateGitHubOperations(todoItems, githubIssues)

	// Combine title rename operations with create/close operations
	allOperations := append(titleUpdates.Operations, operations...)

	// Execute operations
	updatedTodoItems := make([]todo.TodoItem, len(todoItems))
	copy(updatedTodoItems, todoItems)

	for _, todoOp := range allOperations {
		switch op := todoOp.Operation.(type) {
		case github.CreateIssueOp:
			body, err := ghapi.Execute(ghapi.CreateIssue{Repo: repo, Title: op.Title})
			if err != nil {
				return err
			}
			issueNumber, err := github.ParseCreatedIssueNumber(body)
			if err != nil {
				return err
			}
			fmt.Printf("Created issue #%d: %s\n", issueNumber, todoOp.Todo.Text)

			// Update TODO item with issue number
			for j := range updatedTodoItems {
				if updatedTodoItems[j].Text == todoOp.Todo.Text &&
					updatedTodoItems[j].IsChecked == todoOp.Todo.IsChecked &&
					updatedTodoItems[j].IssueNumber == todoOp.Todo.IssueNumber {
					updatedTodoItems[j].IssueNumber = &issueNumber
					break
				}
			}
		case github.CloseIssueOp:
			if _, err := ghapi.Execute(ghapi.CloseIssue{Repo: repo, Number: op.Number}); err != nil {
				return err
			}
			fmt.Printf("Closed issue #%d\n", op.Number)
		case github.RenameIssueOp:
			if _, err := ghapi.Execute(ghapi.RenameIssue{Repo: repo, Number: op.Number, Title: op.Title}); err != nil {
				return err
			}
			fmt.Printf("Renamed issue #%d: %s\n", op.Number, op.Title)
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
	githubIssues, err := listIssues(repo)
	if err != nil {
		return err
	}

	// Synchronize titles with rename history
	pastTitles, err := collectPastTitles(repo, todoItems, githubIssues)
	if err != nil {
		return err
	}
	titleSync := github.SynchronizeTitles(todoItems, githubIssues, pastTitles)

	for _, issueNumber := range titleSync.LocallyEditedIssues {
		fmt.Printf("Warning: TODO.md text for issue #%d was changed locally; run `gh atat push` to update the issue title\n", issueNumber)
	}

	// Synchronize with GitHub issues
	updatedTodoItems := github.SynchronizeWithGitHubIssues(titleSync.Items, githubIssues)

	// Write updated TODO.md
	updatedContent := markdown.SerializeTodoMarkdown(updatedTodoItems)
	if err := os.WriteFile("TODO.md", []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write TODO.md: %w", err)
	}

	return nil
}

func runClean(dryRun bool) error {
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

	// Build clean candidates from checked items with issue numbers
	var candidates []clean.CleanCandidate
	for _, item := range todoItems {
		if candidate, ok := clean.NewCleanCandidate(item); ok {
			candidates = append(candidates, candidate)
		}
	}

	// Fetch GitHub issues
	githubIssues, err := listIssues(repo)
	if err != nil {
		return err
	}

	// Find removable items
	removable := clean.FindRemovableItems(candidates, githubIssues)

	if len(removable) == 0 {
		return nil
	}

	// Build a set of removable issue numbers for quick lookup
	removableSet := make(map[uint64]bool)
	for _, r := range removable {
		removableSet[r.IssueNumber] = true
		fmt.Printf("Removing: %s (#%d)\n", r.Text, r.IssueNumber)
	}

	if dryRun {
		return nil
	}

	// Filter out removable items from the todo list
	var remaining []todo.TodoItem
	for _, item := range todoItems {
		if item.IssueNumber != nil && removableSet[*item.IssueNumber] {
			continue
		}
		remaining = append(remaining, item)
	}

	// Write updated TODO.md
	updatedContent := markdown.SerializeTodoMarkdown(remaining)
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
	if _, err := ghapi.Execute(ghapi.GetRepository{Repo: repo}); err != nil {
		if errors.Is(err, ghapi.ErrNotFound) {
			return fmt.Errorf("repository %s not found or not accessible", repo)
		}
		return fmt.Errorf("failed to check repository %s: %w", repo, err)
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

func listIssues(repo string) ([]github.GitHubIssue, error) {
	const maxPages = 1000
	var issues []github.GitHubIssue
	for page := 1; ; page++ {
		body, err := ghapi.Execute(ghapi.ListIssues{Repo: repo, Page: page, PerPage: 100})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch issues: %w", err)
		}
		var issuesJSON []json.RawMessage
		if err := json.Unmarshal(body, &issuesJSON); err != nil {
			return nil, fmt.Errorf("failed to parse issues: %w", err)
		}
		if len(issuesJSON) == 0 {
			return issues, nil
		}
		if page >= maxPages {
			return nil, fmt.Errorf("exceeded maximum page limit")
		}
		for _, issueJSON := range issuesJSON {
			issue, err := github.ParseGitHubIssue(issueJSON)
			if err != nil {
				continue
			}
			issues = append(issues, issue)
		}
	}
}

func collectPastTitles(repo string, todoItems []todo.TodoItem, githubIssues []github.GitHubIssue) (map[uint64][]string, error) {
	pastTitles := make(map[uint64][]string)
	for _, issueNumber := range github.FindTitleMismatches(todoItems, githubIssues) {
		body, err := ghapi.Execute(ghapi.ListIssueEvents{Repo: repo, Number: issueNumber})
		if err != nil {
			return nil, err
		}
		var events []json.RawMessage
		if err := json.Unmarshal(body, &events); err != nil {
			return nil, err
		}
		pastTitles[issueNumber] = github.ParsePastTitles(events)
	}
	return pastTitles, nil
}

func printVersion(version string) {
	if version == "" {
		version = "dev"
	}
	fmt.Printf("gh-atat version %s\n", version)
}

func printHelp() {
	help := `gh-atat: Automatic TODO and Tracker

Usage:
  gh atat <command> [arguments]

Commands:
  push          Push TODO items to GitHub Issues
  pull          Pull GitHub Issues to TODO items
  clean         Remove completed TODO items with closed issues
  remote        List configured repositories
  remote add    Add a repository
  remote remove Remove a repository
  help          Show this help message

Examples:
  gh atat push
  gh atat pull
  gh atat clean
  gh atat clean --dry-run
  gh atat remote
  gh atat remote add owner/repo
  gh atat remote remove owner/repo
`
	fmt.Print(help)
}
