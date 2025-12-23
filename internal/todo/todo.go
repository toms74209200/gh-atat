package todo

// TodoItem represents a single todo item from a markdown checklist.
type TodoItem struct {
	Text        string
	IsChecked   bool
	IssueNumber *uint64
}
