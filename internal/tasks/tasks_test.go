package tasks

import (
	"os"
	"testing"
)

func TestExtract(t *testing.T) {
	markdown, err := os.ReadFile("testdata/extract_test.md")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	doc, err := Extract(markdown)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should have 3 sections: the H1 title + 2 H2 sections
	if len(doc.Sections) != 3 {
		t.Errorf("Expected 3 sections, got %d", len(doc.Sections))
	}

	// First section is the H1 title (no tasks)
	if doc.Sections[0].Title != "2024-11-08" {
		t.Errorf("Expected first section '2024-11-08', got '%s'", doc.Sections[0].Title)
	}

	// Check "Doing" section (index 1)
	if doc.Sections[1].Title != "Doing" {
		t.Errorf("Expected second section 'Doing', got '%s'", doc.Sections[1].Title)
	}
	if len(doc.Sections[1].Tasks) != 2 {
		t.Errorf("Expected 2 tasks in Doing, got %d", len(doc.Sections[1].Tasks))
	}

	// Check nested tasks
	if len(doc.Sections[1].Tasks[0].Children) != 2 {
		t.Errorf("Expected 2 child tasks, got %d", len(doc.Sections[1].Tasks[0].Children))
	}

	// Check task content
	if doc.Sections[1].Tasks[0].Content != "Review PRs" {
		t.Errorf("Expected 'Review PRs', got '%s'", doc.Sections[1].Tasks[0].Content)
	}

	// Check checked status
	if !doc.Sections[1].Tasks[1].Checked {
		t.Error("Expected second task to be checked")
	}
}

func TestFilterUncompleted(t *testing.T) {
	markdown, err := os.ReadFile("testdata/filter_uncompleted_test.md")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	doc, err := Extract(markdown)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	uncompleted := FilterUncompleted(doc)

	// Should only have 1 section (Doing), not the H1 title or Completed Section
	if len(uncompleted.Sections) != 1 {
		t.Errorf("Expected 1 section with uncompleted tasks, got %d", len(uncompleted.Sections))
	}

	if uncompleted.Sections[0].Title != "Doing" {
		t.Errorf("Expected 'Doing' section, got '%s'", uncompleted.Sections[0].Title)
	}

	// Should have 1 task (Review PRs with children)
	// "Write documentation" should be filtered out because it's checked
	if len(uncompleted.Sections[0].Tasks) != 1 {
		t.Errorf("Expected 1 uncompleted task, got %d", len(uncompleted.Sections[0].Tasks))
	}

	// Check that parent task with completed children is included
	task := uncompleted.Sections[0].Tasks[0]
	if task.Content != "Review PRs" {
		t.Errorf("Expected 'Review PRs', got '%s'", task.Content)
	}

	// Should have 1 child (PR #456), not PR #123 which is completed
	if len(task.Children) != 1 {
		t.Errorf("Expected 1 uncompleted child task, got %d", len(task.Children))
	}

	if task.Children[0].Content != "PR #456" {
		t.Errorf("Expected 'PR #456', got '%s'", task.Children[0].Content)
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		task     Task
		indent   int
		expected string
	}{
		{
			name: "simple unchecked task",
			task: Task{
				Content: "Do something",
				Checked: false,
			},
			indent:   0,
			expected: "- [ ] Do something",
		},
		{
			name: "simple checked task",
			task: Task{
				Content: "Done thing",
				Checked: true,
			},
			indent:   0,
			expected: "- [x] Done thing",
		},
		{
			name: "task with children",
			task: Task{
				Content: "Parent task",
				Checked: false,
				Children: []Task{
					{Content: "Child 1", Checked: false},
					{Content: "Child 2", Checked: true},
				},
			},
			indent:   0,
			expected: "- [ ] Parent task\n  - [ ] Child 1\n  - [x] Child 2",
		},
		{
			name: "indented task",
			task: Task{
				Content: "Indented",
				Checked: false,
			},
			indent:   2,
			expected: "    - [ ] Indented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Format(tt.task, tt.indent)
			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tasks := []Task{
		{
			Content: "Parent 1",
			Children: []Task{
				{Content: "Child 1"},
				{Content: "Child 2"},
			},
		},
		{
			Content: "Parent 2",
		},
	}

	count := Count(tasks)
	expected := 4 // 2 parents + 2 children
	if count != expected {
		t.Errorf("Expected count %d, got %d", expected, count)
	}
}

func TestCountWithCompleted(t *testing.T) {
	tasks := []Task{
		{
			Content: "Parent 1",
			Checked: false,
			Children: []Task{
				{Content: "Child 1", Checked: true},
				{Content: "Child 2", Checked: false},
			},
		},
		{
			Content: "Parent 2",
			Checked: true,
		},
	}

	total, completed := CountWithCompleted(tasks)

	if total != 4 {
		t.Errorf("Expected total 4, got %d", total)
	}

	if completed != 2 {
		t.Errorf("Expected completed 2, got %d", completed)
	}
}

func TestExtractWithLinks(t *testing.T) {
	markdown, err := os.ReadFile("testdata/links_test.md")
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	doc, err := Extract(markdown)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should have 2 sections (H1 + H2)
	if len(doc.Sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(doc.Sections))
	}

	section := doc.Sections[1] // The "Section" H2
	if len(section.Tasks) != 5 {
		t.Fatalf("Expected 5 tasks, got %d", len(section.Tasks))
	}

	// Check first task preserves link
	expected1 := "Task with [link text](https://example.com/path)"
	if section.Tasks[0].Content != expected1 {
		t.Errorf("Expected: %s\nGot: %s", expected1, section.Tasks[0].Content)
	}

	// Check second task preserves multiple links
	expected2 := "Task with multiple [link1](https://one.com) and [link2](https://two.com)"
	if section.Tasks[1].Content != expected2 {
		t.Errorf("Expected: %s\nGot: %s", expected2, section.Tasks[1].Content)
	}

	// Check completed task preserves link
	expected3 := "Completed [task](https://completed.com)"
	if section.Tasks[2].Content != expected3 {
		t.Errorf("Expected: %s\nGot: %s", expected3, section.Tasks[2].Content)
	}

	// Verify it's marked as checked
	if !section.Tasks[2].Checked {
		t.Error("Expected third task to be checked")
	}

	// Check task with backticks and formatting
	expected4 := "Task with `code in backticks` and **bold** and *italic*"
	if section.Tasks[3].Content != expected4 {
		t.Errorf("Expected: %s\nGot: %s", expected4, section.Tasks[3].Content)
	}

	// Check nested tasks with links and backticks
	expected5 := "fixing [a-problem](https://exemplar.com/example/12345678910)"
	if section.Tasks[4].Content != expected5 {
		t.Errorf("Expected: %s\nGot: %s", expected5, section.Tasks[4].Content)
	}

	if len(section.Tasks[4].Children) != 2 {
		t.Fatalf("Expected 2 children for task 5, got %d", len(section.Tasks[4].Children))
	}

	expectedChild := "once approved, `SOME_BACKTICKED_THING` should exist"
	if section.Tasks[4].Children[1].Content != expectedChild {
		t.Errorf("Expected: %s\nGot: %s", expectedChild, section.Tasks[4].Children[1].Content)
	}
}
