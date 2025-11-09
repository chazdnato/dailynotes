package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dailynotes/internal/tasks"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	templateFile := filepath.Join(tmpDir, "template.md")
	content := "# {{.Date}}\n\n## Custom Section"

	if err := os.WriteFile(templateFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	loaded, err := Load(templateFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded != content {
		t.Errorf("Expected content '%s', got '%s'", content, loaded)
	}
}

func TestLoadDefault(t *testing.T) {
	loaded, err := Load("")
	if err != nil {
		t.Fatalf("Load with empty path failed: %v", err)
	}

	if loaded != Default {
		t.Error("Expected default template when path is empty")
	}

	if !strings.Contains(loaded, "{{.Date}}") {
		t.Error("Default template should contain {{.Date}} placeholder")
	}
}

func TestMerge(t *testing.T) {
	tmpl := `# {{.Date}}

## Doing

- [ ] Default task

## Longer Term

- [ ] Future task
`

	taskDoc := &tasks.Document{
		Sections: []tasks.Section{
			{
				Title: "Doing",
				Level: 2,
				Tasks: []tasks.Task{
					{Content: "Review PRs", Checked: false},
				},
			},
			{
				Title: "Longer Term",
				Level: 2,
				Tasks: []tasks.Task{
					{Content: "Research framework", Checked: false},
				},
			},
		},
	}

	result := Merge(tmpl, taskDoc, "2024-11-09")

	// Check date was replaced
	if !strings.Contains(result, "# 2024-11-09") {
		t.Error("Date placeholder was not replaced")
	}

	// Check tasks were merged
	if !strings.Contains(result, "- [ ] Review PRs") {
		t.Error("Task 'Review PRs' not found in result")
	}

	if !strings.Contains(result, "- [ ] Research framework") {
		t.Error("Task 'Research framework' not found in result")
	}

	// Check sections are present
	if !strings.Contains(result, "## Doing") {
		t.Error("Doing section not found")
	}

	if !strings.Contains(result, "## Longer Term") {
		t.Error("Longer Term section not found")
	}
}

func TestMergeWithOrphanedSection(t *testing.T) {
	tmpl := `# {{.Date}}

## Doing

- [ ] Default task
`

	taskDoc := &tasks.Document{
		Sections: []tasks.Section{
			{
				Title: "Doing",
				Level: 2,
				Tasks: []tasks.Task{
					{Content: "Task in template", Checked: false},
				},
			},
			{
				Title: "Random Section",
				Level: 2,
				Tasks: []tasks.Task{
					{Content: "Orphaned task", Checked: false},
				},
			},
		},
	}

	result := Merge(tmpl, taskDoc, "2024-11-09")

	// Check that orphaned section was added
	if !strings.Contains(result, "## Random Section") {
		t.Error("Orphaned section was not added to result")
	}

	if !strings.Contains(result, "- [ ] Orphaned task") {
		t.Error("Orphaned task was not added to result")
	}
}
