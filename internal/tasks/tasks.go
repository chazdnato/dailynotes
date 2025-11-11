package tasks

import (
	"bytes"
	"fmt"
	"strings"

	"dailynotes/internal/debug"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	gast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

// Task represents a checkbox item
type Task struct {
	Content  string
	Checked  bool
	Indent   int
	Children []Task
}

// Section represents a markdown section with its tasks
type Section struct {
	Title string
	Level int
	Tasks []Task
}

// Document represents the structure of a markdown document
type Document struct {
	Sections []Section
}

// Extract parses markdown and extracts all tasks organized by section
func Extract(markdown []byte) (*Document, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.TaskList),
	)

	reader := text.NewReader(markdown)
	doc := md.Parser().Parse(reader)

	result := &Document{
		Sections: []Section{},
	}

	var currentSection *Section

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			title := extractText(node, markdown)
			section := Section{
				Title: title,
				Level: node.Level,
				Tasks: []Task{},
			}
			result.Sections = append(result.Sections, section)
			currentSection = &result.Sections[len(result.Sections)-1]
			debug.Printf("Found section: %s (level %d)", title, node.Level)

		case *ast.List:
			taskList := extractTasksFromList(node, markdown, 0)
			if len(taskList) > 0 && currentSection != nil {
				currentSection.Tasks = append(currentSection.Tasks, taskList...)
				debug.Printf("Added %d tasks to section '%s'", len(taskList), currentSection.Title)
			}
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// extractTasksFromList recursively extracts tasks from a list
func extractTasksFromList(list ast.Node, source []byte, indent int) []Task {
	var tasks []Task

	for child := list.FirstChild(); child != nil; child = child.NextSibling() {
		if listItem, ok := child.(*ast.ListItem); ok {
			var hasCheckbox bool
			var isChecked bool

			for itemChild := listItem.FirstChild(); itemChild != nil; itemChild = itemChild.NextSibling() {
				if tb, ok := itemChild.(*ast.TextBlock); ok {
					for tbChild := tb.FirstChild(); tbChild != nil; tbChild = tbChild.NextSibling() {
						if checkbox, ok := tbChild.(*gast.TaskCheckBox); ok {
							hasCheckbox = true
							isChecked = checkbox.IsChecked
							debug.Printf("Found checkbox: checked=%v", isChecked)
							break
						}
					}
				}
				if hasCheckbox {
					break
				}
			}

			if hasCheckbox {
				content := extractText(listItem, source)

				task := Task{
					Content: content,
					Checked: isChecked,
					Indent:  indent,
				}

				for itemChild := listItem.FirstChild(); itemChild != nil; itemChild = itemChild.NextSibling() {
					if sublist, ok := itemChild.(*ast.List); ok {
						task.Children = extractTasksFromList(sublist, source, indent+1)
					}
				}

				tasks = append(tasks, task)
			} else {
				for itemChild := listItem.FirstChild(); itemChild != nil; itemChild = itemChild.NextSibling() {
					if sublist, ok := itemChild.(*ast.List); ok {
						tasks = append(tasks, extractTasksFromList(sublist, source, indent)...)
					}
				}
			}
		}
	}

	return tasks
}

// extractText gets the text content from a list item, preserving inline markdown
func extractText(node ast.Node, source []byte) string {
	var buf bytes.Buffer

	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		// Skip nested lists - they're handled as children
		if _, ok := child.(*ast.List); ok {
			continue
		}

		// Recursively collect text from this child
		collectText(child, source, &buf)
	}

	return strings.TrimSpace(buf.String())
}

// collectText recursively collects formatted text from AST nodes
func collectText(node ast.Node, source []byte, buf *bytes.Buffer) {
	switch n := node.(type) {
	case *gast.TaskCheckBox:
		// Skip the checkbox itself - we don't want it in the content
		return

	case *ast.Text:
		// Plain text - just get the segment
		buf.Write(n.Segment.Value(source))

	case *ast.CodeSpan:
		// Code span - add backticks
		buf.WriteString("`")
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, buf)
		}
		buf.WriteString("`")

	case *ast.Emphasis:
		// Emphasis/strong - add markers
		marker := "*"
		if n.Level == 2 {
			marker = "**"
		}
		buf.WriteString(marker)
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, buf)
		}
		buf.WriteString(marker)

	case *ast.Link:
		// Link - reconstruct markdown syntax
		buf.WriteString("[")
		for child := n.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, buf)
		}
		buf.WriteString("](")
		buf.Write(n.Destination)
		buf.WriteString(")")

	default:
		// For any other node, recurse into children
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			collectText(child, source, buf)
		}
	}
}

// FilterUncompleted returns a new document with only sections that have uncompleted tasks
func FilterUncompleted(doc *Document) *Document {
	result := &Document{
		Sections: []Section{},
	}

	for _, section := range doc.Sections {
		uncompletedTasks := filterUncompletedTasks(section.Tasks)
		debug.Printf("Section '%s': %d total tasks, %d uncompleted", section.Title, len(section.Tasks), len(uncompletedTasks))
		if len(uncompletedTasks) > 0 {
			newSection := Section{
				Title: section.Title,
				Level: section.Level,
				Tasks: uncompletedTasks,
			}
			result.Sections = append(result.Sections, newSection)
		}
	}

	debug.Printf("Filtered result: %d sections with uncompleted tasks", len(result.Sections))
	return result
}

// filterUncompletedTasks recursively filters for uncompleted tasks
func filterUncompletedTasks(tasks []Task) []Task {
	var result []Task

	for _, task := range tasks {
		filteredChildren := filterUncompletedTasks(task.Children)

		if !task.Checked || len(filteredChildren) > 0 {
			newTask := task
			newTask.Children = filteredChildren
			result = append(result, newTask)
		}
	}

	return result
}

// Format converts a task back to markdown
func Format(task Task, baseIndent int) string {
	indent := strings.Repeat("  ", baseIndent)
	checkbox := "- [ ]"
	if task.Checked {
		checkbox = "- [x]"
	}

	result := fmt.Sprintf("%s%s %s", indent, checkbox, task.Content)

	for _, child := range task.Children {
		result += "\n" + Format(child, baseIndent+1)
	}

	return result
}

// Count recursively counts all tasks including children
func Count(tasks []Task) int {
	count := 0
	for _, task := range tasks {
		count++
		count += Count(task.Children)
	}
	return count
}

// CountWithCompleted recursively counts tasks and returns (total, completed)
func CountWithCompleted(tasks []Task) (int, int) {
	total := 0
	completed := 0
	for _, task := range tasks {
		total++
		if task.Checked {
			completed++
		}
		childTotal, childCompleted := CountWithCompleted(task.Children)
		total += childTotal
		completed += childCompleted
	}
	return total, completed
}
