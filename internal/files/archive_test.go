package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mustWriteFile writes a file under root, creating parent directories as needed.
func mustWriteFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir parent for %s: %v", rel, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", rel, err)
	}
}

// mustExist asserts that a file exists at the given relative path under root.
func mustExist(t *testing.T, root, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, rel)); err != nil {
		t.Errorf("expected file %s to exist: %v", rel, err)
	}
}

// mustNotExist asserts that a file does not exist at the given relative path under root.
func mustNotExist(t *testing.T, root, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, rel)); !os.IsNotExist(err) {
		t.Errorf("expected file %s to not exist, but stat returned err=%v", rel, err)
	}
}

func TestArchiveHappyPath(t *testing.T) {
	root := t.TempDir()
	// Now: 2026-04-19 (month in progress)
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	// Current-month file: should remain
	mustWriteFile(t, root, "2026-04-17.md", "apr content")
	// Prior month: should move
	mustWriteFile(t, root, "2026-03-15.md", "mar content")
	// Prior year: should move
	mustWriteFile(t, root, "2025-11-22.md", "nov content")

	result, err := Archive(root, now, false)
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Current-month stays put
	mustExist(t, root, "2026-04-17.md")
	// Prior-month file moved under YYYY/MM/
	mustNotExist(t, root, "2026-03-15.md")
	mustExist(t, root, "2026/03/2026-03-15.md")
	// Prior-year file moved
	mustNotExist(t, root, "2025-11-22.md")
	mustExist(t, root, "2025/11/2025-11-22.md")

	if len(result.Moved) != 2 {
		t.Errorf("expected 2 files moved, got %d: %v", len(result.Moved), result.Moved)
	}
	if len(result.Skipped) != 0 {
		t.Errorf("expected no skips, got %d: %+v", len(result.Skipped), result.Skipped)
	}
}

func TestArchiveCurrentMonthBoundary(t *testing.T) {
	root := t.TempDir()
	// Now: 2026-04-01 00:00:00 UTC - the very first day of April
	now := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	// Last day of March: prior month, should move
	mustWriteFile(t, root, "2026-03-31.md", "")
	// First day of April: current month, should stay
	mustWriteFile(t, root, "2026-04-01.md", "")

	if _, err := Archive(root, now, false); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	mustExist(t, root, "2026/03/2026-03-31.md")
	mustNotExist(t, root, "2026-03-31.md")
	mustExist(t, root, "2026-04-01.md")
	mustNotExist(t, root, "2026/04/2026-04-01.md")
}

func TestArchiveIdempotent(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	mustWriteFile(t, root, "2026-03-01.md", "a")
	mustWriteFile(t, root, "2025-12-31.md", "b")

	// First run moves everything.
	first, err := Archive(root, now, false)
	if err != nil {
		t.Fatalf("first Archive failed: %v", err)
	}
	if len(first.Moved) != 2 {
		t.Errorf("first run: expected 2 moved, got %d", len(first.Moved))
	}

	// Second run is a no-op.
	second, err := Archive(root, now, false)
	if err != nil {
		t.Fatalf("second Archive failed: %v", err)
	}
	if len(second.Moved) != 0 || len(second.Skipped) != 0 {
		t.Errorf("second run should be a no-op; got moved=%v skipped=%+v", second.Moved, second.Skipped)
	}
}

func TestArchiveDestinationExists(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	// Source file at root
	mustWriteFile(t, root, "2026-03-15.md", "source content")
	// Destination already exists with DIFFERENT content
	mustWriteFile(t, root, "2026/03/2026-03-15.md", "archived content")

	result, err := Archive(root, now, false)
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Source must still exist; destination must still exist
	mustExist(t, root, "2026-03-15.md")
	mustExist(t, root, "2026/03/2026-03-15.md")

	// Must be reported as skipped
	if len(result.Moved) != 0 {
		t.Errorf("expected 0 moved, got %d: %v", len(result.Moved), result.Moved)
	}
	if len(result.Skipped) != 1 {
		t.Fatalf("expected 1 skip, got %d: %+v", len(result.Skipped), result.Skipped)
	}
	if !strings.Contains(result.Skipped[0].Reason, "destination") {
		t.Errorf("expected skip reason to mention destination, got: %s", result.Skipped[0].Reason)
	}

	// Contents unchanged
	srcContent, _ := os.ReadFile(filepath.Join(root, "2026-03-15.md"))
	destContent, _ := os.ReadFile(filepath.Join(root, "2026/03/2026-03-15.md"))
	if string(srcContent) != "source content" {
		t.Errorf("source content modified: %s", srcContent)
	}
	if string(destContent) != "archived content" {
		t.Errorf("destination content modified: %s", destContent)
	}
}

func TestArchiveIgnoresNonDateFiles(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	mustWriteFile(t, root, "README.md", "readme")
	mustWriteFile(t, root, "notes.md", "notes")
	mustWriteFile(t, root, "2026-03-15.md.backup", "backup") // not matched: has .backup suffix
	mustWriteFile(t, root, "not-a-date.md", "other")
	mustWriteFile(t, root, "2026-3-15.md", "single-digit month, not 2026-03-15") // not matched: regex needs \d{2}-\d{2}

	if _, err := Archive(root, now, false); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	// Non-matching files must remain at root
	mustExist(t, root, "README.md")
	mustExist(t, root, "notes.md")
	mustExist(t, root, "2026-03-15.md.backup")
	mustExist(t, root, "not-a-date.md")
	mustExist(t, root, "2026-3-15.md")

	// Nothing should have been moved (no valid date-matching prior-month files)
	if _, err := os.Stat(filepath.Join(root, "2026")); !os.IsNotExist(err) {
		t.Errorf("expected 2026/ subdir to not exist; stat err=%v", err)
	}
}

func TestArchiveDryRun(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	mustWriteFile(t, root, "2026-03-15.md", "mar")
	mustWriteFile(t, root, "2026-04-17.md", "apr")

	result, err := Archive(root, now, true)
	if err != nil {
		t.Fatalf("Archive dry-run failed: %v", err)
	}

	// Nothing should actually move
	mustExist(t, root, "2026-03-15.md")
	mustExist(t, root, "2026-04-17.md")
	mustNotExist(t, root, "2026/03/2026-03-15.md")
	if _, err := os.Stat(filepath.Join(root, "2026")); !os.IsNotExist(err) {
		t.Errorf("expected 2026/ subdir to not exist in dry-run mode, stat err=%v", err)
	}

	// But the result reflects what WOULD have moved
	if len(result.Moved) != 1 || result.Moved[0] != "2026-03-15.md" {
		t.Errorf("expected dry-run to report [2026-03-15.md] moved, got %v", result.Moved)
	}
}

func TestArchiveEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	result, err := Archive(root, now, false)
	if err != nil {
		t.Fatalf("Archive on empty dir failed: %v", err)
	}

	if len(result.Moved) != 0 || len(result.Skipped) != 0 {
		t.Errorf("expected empty result on empty dir, got moved=%v skipped=%v", result.Moved, result.Skipped)
	}
}

func TestArchiveNonexistentDirectory(t *testing.T) {
	_, err := Archive("/nonexistent/dailynotes/path", time.Now(), false)
	if err == nil {
		t.Fatal("expected error for nonexistent directory")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' error, got: %v", err)
	}
}

func TestArchiveAlreadyArchivedNotReArchived(t *testing.T) {
	// If files are already in YYYY/MM/ subdirs, Archive() should NOT
	// descend into them or re-process them (it only looks at root-level).
	root := t.TempDir()
	now := time.Date(2026, 4, 19, 12, 0, 0, 0, time.UTC)

	mustWriteFile(t, root, "2026/03/2026-03-15.md", "archived")
	mustWriteFile(t, root, "2025/11/2025-11-10.md", "archived older")

	result, err := Archive(root, now, false)
	if err != nil {
		t.Fatalf("Archive failed: %v", err)
	}

	if len(result.Moved) != 0 || len(result.Skipped) != 0 {
		t.Errorf("expected nothing to move for already-archived; got moved=%v skipped=%v", result.Moved, result.Skipped)
	}

	// Verify files still where they were
	mustExist(t, root, "2026/03/2026-03-15.md")
	mustExist(t, root, "2025/11/2025-11-10.md")
}
