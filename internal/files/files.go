package files

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"dailynotes/internal/debug"
	"dailynotes/internal/tasks"
)

var dateFilePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\.md$`)

// FindMostRecent finds the most recent YYYY-MM-DD.md file in the given directory
func FindMostRecent(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", dir)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied reading directory: %s", dir)
		}
		return "", fmt.Errorf("error reading directory %s: %w", dir, err)
	}

	var dateFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && dateFilePattern.MatchString(entry.Name()) {
			dateFiles = append(dateFiles, entry.Name())
		}
	}

	if len(dateFiles) == 0 {
		return "", nil
	}

	sort.Sort(sort.Reverse(sort.StringSlice(dateFiles)))

	debug.Printf("Found %d date files, most recent: %s", len(dateFiles), dateFiles[0])
	return filepath.Join(dir, dateFiles[0]), nil
}

// TodayFilename returns today's filename in YYYY-MM-DD.md format
func TodayFilename() string {
	return time.Now().Format("2006-01-02") + ".md"
}

// Load reads a markdown file and returns its content
func Load(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", path)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied reading file: %s", path)
		}
		return "", fmt.Errorf("error reading file %s: %w", path, err)
	}
	return string(content), nil
}

// Write writes content to a file
func Write(path string, content string) error {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied writing to: %s", path)
		}
		return fmt.Errorf("error writing file %s: %w", path, err)
	}
	debug.Printf("Wrote file: %s", path)
	return nil
}

// ListInfo contains information about a daily note file
type ListInfo struct {
	Filename     string
	Date         string
	SizeKB       float64
	Total        int
	Completed    int
	IsMostRecent bool
}

// List returns information about all daily note files in the directory
func List(dir string) ([]ListInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory does not exist: %s", dir)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("permission denied reading directory: %s", dir)
		}
		return nil, fmt.Errorf("error reading directory: %w", err)
	}

	var dateFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && dateFilePattern.MatchString(entry.Name()) {
			dateFiles = append(dateFiles, entry.Name())
		}
	}

	if len(dateFiles) == 0 {
		return []ListInfo{}, nil
	}

	// Sort in reverse order (most recent first)
	sort.Sort(sort.Reverse(sort.StringSlice(dateFiles)))

	var results []ListInfo

	for i, filename := range dateFiles {
		path := filepath.Join(dir, filename)
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		listInfo := ListInfo{
			Filename:     filename,
			Date:         strings.TrimSuffix(filename, ".md"),
			SizeKB:       float64(info.Size()) / 1024.0,
			IsMostRecent: i == 0,
		}

		// Try to count tasks in the file
		content, err := Load(path)
		if err == nil {
			doc, err := tasks.Extract([]byte(content))
			if err == nil {
				total := 0
				completed := 0
				for _, section := range doc.Sections {
					t, c := tasks.CountWithCompleted(section.Tasks)
					total += t
					completed += c
				}
				listInfo.Total = total
				listInfo.Completed = completed
			}
		}

		results = append(results, listInfo)
	}

	return results, nil
}
