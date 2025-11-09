package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTodayFilename(t *testing.T) {
	filename := TodayFilename()
	expected := time.Now().Format("2006-01-02") + ".md"

	if filename != expected {
		t.Errorf("Expected %s, got %s", expected, filename)
	}
}

func TestFindMostRecent(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create some test files
	files := []string{
		"2024-11-05.md",
		"2024-11-08.md",
		"2024-11-07.md",
		"not-a-date.md",
		"2024-11-06.md",
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	recent, err := FindMostRecent(tmpDir)
	if err != nil {
		t.Fatalf("FindMostRecent failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "2024-11-08.md")
	if recent != expected {
		t.Errorf("Expected %s, got %s", expected, recent)
	}
}

func TestFindMostRecentEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	recent, err := FindMostRecent(tmpDir)
	if err != nil {
		t.Fatalf("FindMostRecent failed: %v", err)
	}

	if recent != "" {
		t.Errorf("Expected empty string for directory with no date files, got %s", recent)
	}
}

func TestFindMostRecentNonexistent(t *testing.T) {
	_, err := FindMostRecent("/nonexistent/directory")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test Content\n\n- [ ] Task"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loaded, err := Load(testFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded != content {
		t.Errorf("Expected content '%s', got '%s'", content, loaded)
	}
}

func TestLoadNonexistent(t *testing.T) {
	_, err := Load("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Test\n\n- [ ] Task"

	err := Write(testFile, content)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}

	if string(data) != content {
		t.Errorf("Expected content '%s', got '%s'", content, string(data))
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	testFiles := map[string]string{
		"2024-11-05.md": "# 2024-11-05\n\n## Tasks\n\n- [x] Done\n- [ ] Todo",
		"2024-11-08.md": "# 2024-11-08\n\n## Tasks\n\n- [ ] Task 1\n- [ ] Task 2",
		"not-a-date.md": "random content",
	}

	for filename, content := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	results, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Should only find the 2 date-formatted files
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// First result should be most recent (2024-11-08)
	if results[0].Date != "2024-11-08" {
		t.Errorf("Expected first result to be 2024-11-08, got %s", results[0].Date)
	}

	if !results[0].IsMostRecent {
		t.Error("Expected first result to be marked as most recent")
	}

	// Check task counts
	if results[0].Total != 2 {
		t.Errorf("Expected 2 total tasks for 2024-11-08, got %d", results[0].Total)
	}

	if results[0].Completed != 0 {
		t.Errorf("Expected 0 completed tasks for 2024-11-08, got %d", results[0].Completed)
	}

	// Second result should be 2024-11-05
	if results[1].Date != "2024-11-05" {
		t.Errorf("Expected second result to be 2024-11-05, got %s", results[1].Date)
	}

	if results[1].Total != 2 {
		t.Errorf("Expected 2 total tasks for 2024-11-05, got %d", results[1].Total)
	}

	if results[1].Completed != 1 {
		t.Errorf("Expected 1 completed task for 2024-11-05, got %d", results[1].Completed)
	}
}

func TestListEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	results, err := List(tmpDir)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results for empty directory, got %d", len(results))
	}
}
