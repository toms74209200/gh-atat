package markdown

import (
	"testing"

	"github.com/toms74209200/gh-atat/internal/todo"
)

func TestParseTodoMarkdown(t *testing.T) {
	num123 := uint64(123)
	num456 := uint64(456)

	tests := []struct {
		name     string
		input    string
		expected []todo.TodoItem
	}{
		{
			name: "simple checklist exact text",
			input: `- [ ] Task 1
- [x] Task 2
- [X] Task 3`,
			expected: []todo.TodoItem{
				{Text: "Task 1", IsChecked: false, IssueNumber: nil},
				{Text: "Task 2", IsChecked: true, IssueNumber: nil},
				{Text: "Task 3", IsChecked: true, IssueNumber: nil},
			},
		},
		{
			name: "text formatting exact match",
			input: `- [ ] **bold** text
- [x] *italic* text
- [ ] ` + "`code`" + ` text
- [x] [link](url) text
- [ ] ~~strikethrough~~ text`,
			expected: []todo.TodoItem{
				{Text: "bold text", IsChecked: false, IssueNumber: nil},
				{Text: "italic text", IsChecked: true, IssueNumber: nil},
				{Text: "code text", IsChecked: false, IssueNumber: nil},
				{Text: "link text", IsChecked: true, IssueNumber: nil},
				{Text: "strikethrough text", IsChecked: false, IssueNumber: nil},
			},
		},
		{
			name: "issue numbers exact text",
			input: `- [ ] Task with issue (#123)
- [x] Another task (#456)
- [ ] Task without issue`,
			expected: []todo.TodoItem{
				{Text: "Task with issue", IsChecked: false, IssueNumber: &num123},
				{Text: "Another task", IsChecked: true, IssueNumber: &num456},
				{Text: "Task without issue", IsChecked: false, IssueNumber: nil},
			},
		},
		{
			name: "nested checklist flat structure",
			input: `- [ ] Main task
  - [ ] Sub task 1
  - [x] Sub task 2
    - [ ] Sub sub task
- [x] Another main task`,
			expected: []todo.TodoItem{
				{Text: "Main task", IsChecked: false, IssueNumber: nil},
				{Text: "Sub task 1", IsChecked: false, IssueNumber: nil},
				{Text: "Sub task 2", IsChecked: true, IssueNumber: nil},
				{Text: "Sub sub task", IsChecked: false, IssueNumber: nil},
				{Text: "Another main task", IsChecked: true, IssueNumber: nil},
			},
		},
		{
			name: "sections with checklist",
			input: `# Section 1

- [x] Completed task
- [ ] Pending task

## Subsection

- [x] Another completed
- [ ] Another pending`,
			expected: []todo.TodoItem{
				{Text: "Completed task", IsChecked: true, IssueNumber: nil},
				{Text: "Pending task", IsChecked: false, IssueNumber: nil},
				{Text: "Another completed", IsChecked: true, IssueNumber: nil},
				{Text: "Another pending", IsChecked: false, IssueNumber: nil},
			},
		},
		{
			name: "mixed content ignore non-checklist",
			input: `# Title

Regular text.

- Regular bullet
- Another bullet

- [ ] Checklist item 1
- [x] Checklist item 2

` + "```\ncode block\n```" + `

- [ ] Checklist item 3`,
			expected: []todo.TodoItem{
				{Text: "Checklist item 1", IsChecked: false, IssueNumber: nil},
				{Text: "Checklist item 2", IsChecked: true, IssueNumber: nil},
				{Text: "Checklist item 3", IsChecked: false, IssueNumber: nil},
			},
		},
		{
			name:     "empty content",
			input:    "",
			expected: []todo.TodoItem{},
		},
		{
			name: "whitespace only",
			input: `


  `,
			expected: []todo.TodoItem{},
		},
		{
			name: "invalid issue format",
			input: `- [ ] Task (#invalid)
- [x] Task (#)
- [ ] Task (# 123)
- [x] Valid task (#456)`,
			expected: []todo.TodoItem{
				{Text: "Task (#invalid)", IsChecked: false, IssueNumber: nil},
				{Text: "Task (#)", IsChecked: true, IssueNumber: nil},
				{Text: "Task (# 123)", IsChecked: false, IssueNumber: nil},
				{Text: "Valid task", IsChecked: true, IssueNumber: &num456},
			},
		},
		{
			name: "special characters in text",
			input: `- [ ] Task with emoji 🚀
- [x] Task with symbols !@#$%
- [ ] Task with Japanese 日本語
- [x] Task with numbers 123`,
			expected: []todo.TodoItem{
				{Text: "Task with emoji 🚀", IsChecked: false, IssueNumber: nil},
				{Text: "Task with symbols !@#$%", IsChecked: true, IssueNumber: nil},
				{Text: "Task with Japanese 日本語", IsChecked: false, IssueNumber: nil},
				{Text: "Task with numbers 123", IsChecked: true, IssueNumber: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := ParseTodoMarkdown(tt.input)
			if err != nil {
				t.Fatalf("ParseTodoMarkdown failed: %v", err)
			}

			if len(items) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d", len(tt.expected), len(items))
			}

			for i, expected := range tt.expected {
				actual := items[i]

				if actual.Text != expected.Text {
					t.Errorf("item[%d].Text: expected %q, got %q", i, expected.Text, actual.Text)
				}

				if actual.IsChecked != expected.IsChecked {
					t.Errorf("item[%d].IsChecked: expected %v, got %v", i, expected.IsChecked, actual.IsChecked)
				}

				if (actual.IssueNumber == nil) != (expected.IssueNumber == nil) {
					t.Errorf("item[%d].IssueNumber: expected %v, got %v", i, expected.IssueNumber, actual.IssueNumber)
				} else if actual.IssueNumber != nil && *actual.IssueNumber != *expected.IssueNumber {
					t.Errorf("item[%d].IssueNumber: expected %d, got %d", i, *expected.IssueNumber, *actual.IssueNumber)
				}
			}
		})
	}
}

func TestSerializeTodoMarkdown(t *testing.T) {
	num123 := uint64(123)
	num456 := uint64(456)

	tests := []struct {
		name     string
		input    []todo.TodoItem
		expected string
	}{
		{
			name: "serialize todo markdown",
			input: []todo.TodoItem{
				{Text: "Unchecked task", IsChecked: false, IssueNumber: nil},
				{Text: "Checked task", IsChecked: true, IssueNumber: nil},
				{Text: "Task with issue", IsChecked: false, IssueNumber: &num123},
				{Text: "Checked task with issue", IsChecked: true, IssueNumber: &num456},
			},
			expected: "- [ ] Unchecked task\n- [x] Checked task\n- [ ] Task with issue (#123)\n- [x] Checked task with issue (#456)\n",
		},
		{
			name:     "serialize empty list",
			input:    []todo.TodoItem{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := SerializeTodoMarkdown(tt.input)
			if actual != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, actual)
			}
		})
	}
}

func TestSerializeRoundtrip(t *testing.T) {
	originalContent := "- [ ] Task 1\n- [x] Task 2 (#123)\n- [ ] Task 3\n"
	parsedItems, err := ParseTodoMarkdown(originalContent)
	if err != nil {
		t.Fatalf("ParseTodoMarkdown failed: %v", err)
	}

	serialized := SerializeTodoMarkdown(parsedItems)

	if serialized != originalContent {
		t.Errorf("roundtrip failed:\noriginal:\n%s\nserialized:\n%s", originalContent, serialized)
	}
}
