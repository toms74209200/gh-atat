package markdown

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/toms74209200/gh-atat/internal/todo"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// issueNumberRegexp is a precompiled regexp to extract issue numbers from text like "Task (#123)"
var issueNumberRegexp = regexp.MustCompile(`\s+\(#(\d+)\)\s*$`)

// ParseTodoMarkdown parses markdown content and extracts todo items.
func ParseTodoMarkdown(content string) ([]todo.TodoItem, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)

	source := []byte(content)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var items []todo.TodoItem

	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		taskCheckBox, ok := node.(*extast.TaskCheckBox)
		if !ok {
			return ast.WalkContinue, nil
		}

		isChecked := taskCheckBox.IsChecked

		// Get the parent ListItem
		listItem := taskCheckBox.Parent()
		if listItem == nil {
			return ast.WalkContinue, nil
		}

		// Extract text from the list item
		extractedText, err := extractText(listItem, source)
		if err != nil {
			return ast.WalkStop, err
		}
		extractedText = strings.TrimSpace(extractedText)

		if extractedText == "" {
			return ast.WalkContinue, nil
		}

		// Extract issue number if present
		cleanText, issueNumber := extractIssueNumber(extractedText)

		items = append(items, todo.TodoItem{
			Text:        cleanText,
			IsChecked:   isChecked,
			IssueNumber: issueNumber,
		})

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	return items, nil
}

// extractText extracts plain text from an AST node, handling formatting.
func extractText(node ast.Node, source []byte) (string, error) {
	var text strings.Builder

	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch v := n.(type) {
		case *ast.Text:
			text.Write(v.Segment.Value(source))
		case *ast.CodeSpan:
			// Extract text from code spans
			child := v.FirstChild()
			if child != nil {
				if txtNode, ok := child.(*ast.Text); ok {
					text.Write(txtNode.Segment.Value(source))
				}
			}
			// Skip children as we already processed them
			return ast.WalkSkipChildren, nil
		case *extast.TaskCheckBox:
			// Skip the checkbox itself
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return "", err
	}

	return text.String(), nil
}

// extractIssueNumber extracts issue number from text like "Task (#123)"
func extractIssueNumber(text string) (string, *uint64) {
	// Match " (#digits)" at the end of the string using the precompiled regexp
	matches := issueNumberRegexp.FindStringSubmatch(text)

	if len(matches) > 1 {
		if num, err := strconv.ParseUint(matches[1], 10, 64); err == nil {
			cleanText := issueNumberRegexp.ReplaceAllString(text, "")
			cleanText = strings.TrimSpace(cleanText)
			return cleanText, &num
		}
	}

	return text, nil
}

// SerializeTodoMarkdown converts todo items to markdown format.
func SerializeTodoMarkdown(items []todo.TodoItem) string {
	var builder strings.Builder

	for _, item := range items {
		checkbox := "[ ]"
		if item.IsChecked {
			checkbox = "[x]"
		}

		text := item.Text
		if item.IssueNumber != nil {
			text = fmt.Sprintf("%s (#%d)", item.Text, *item.IssueNumber)
		}

		fmt.Fprintf(&builder, "- %s %s\n", checkbox, text)
	}

	return builder.String()
}
