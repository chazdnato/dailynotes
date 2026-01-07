package template

import (
	"fmt"
	"strings"

	"dailynotes/internal/debug"
	"dailynotes/internal/files"
	"dailynotes/internal/tasks"
)

const Default = `# {{.Date}}

## Useful Links

## Read, Attend or Watch

## Doing

- [ ]

## Longer Term

- [ ]
`

// Load loads a template from file or returns the default
func Load(templatePath string) (string, error) {
	if templatePath == "" {
		debug.Printf("Using default template")
		return Default, nil
	}

	content, err := files.Load(templatePath)
	if err != nil {
		return "", fmt.Errorf("loading template: %w", err)
	}

	debug.Printf("Loaded template from %s", templatePath)
	return content, nil
}

// Merge takes uncompleted tasks and merges them into the template
func Merge(tmpl string, taskDoc *tasks.Document, date string) string {
	result := strings.ReplaceAll(tmpl, "{{.Date}}", date)

	templateDoc, err := tasks.Extract([]byte(result))
	if err != nil {
		debug.Printf("Warning: failed to parse template as markdown: %v", err)
		templateDoc = &tasks.Document{Sections: []tasks.Section{}}
	}

	tasksBySection := make(map[string][]tasks.Task)
	for _, section := range taskDoc.Sections {
		tasksBySection[section.Title] = section.Tasks
	}

	var output strings.Builder
	for _, section := range templateDoc.Sections {
		hashes := strings.Repeat("#", section.Level)
		output.WriteString(fmt.Sprintf("%s %s\n\n", hashes, section.Title))

		if sectionTasks, found := tasksBySection[section.Title]; found {
			for _, task := range sectionTasks {
				output.WriteString(tasks.Format(task, 0) + "\n")
			}
			output.WriteString("\n")
			delete(tasksBySection, section.Title)
		}
	}

	// Add any remaining sections not in template
	for sectionTitle, sectionTasks := range tasksBySection {
		output.WriteString(fmt.Sprintf("## %s\n\n", sectionTitle))
		for _, task := range sectionTasks {
			output.WriteString(tasks.Format(task, 0) + "\n")
		}
	}

	// Remove trailing newlines to ensure file ends with single newline
	return strings.TrimRight(output.String(), "\n") + "\n"
}
