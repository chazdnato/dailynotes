package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/chazdnato/dailynotes/internal/debug"
)

// ArchiveResult describes the outcome of an Archive run.
type ArchiveResult struct {
	Moved   []string // files moved, relative to root (source path)
	Skipped []Skip   // files skipped, with reason
}

// Skip records a file that was not moved, and why.
type Skip struct {
	Source string // relative path of the source file in root
	Reason string // human-readable reason (e.g. "destination exists")
}

// Archive moves any YYYY-MM-DD.* file in root (flat, not recursive) into
// root/YYYY/MM/ subdirectories, except files whose date falls in the current
// month (based on `now`).
//
// Behavior:
//   - Idempotent: running multiple times has no effect after the first run.
//   - Current-month files are left alone.
//   - Non-date filenames are ignored.
//   - If a destination file already exists, the source is left in place and
//     the skip is reported in the result.
//   - When dryRun is true, no files are actually moved or directories created.
//
// The returned ArchiveResult is always non-nil. Errors from the filesystem
// are returned as the error return; partial progress may be reflected in
// the result when such an error occurs.
func Archive(root string, now time.Time, dryRun bool) (*ArchiveResult, error) {
	result := &ArchiveResult{}

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return result, fmt.Errorf("directory does not exist: %s", root)
		}
		if os.IsPermission(err) {
			return result, fmt.Errorf("permission denied reading directory: %s", root)
		}
		return result, fmt.Errorf("error reading directory %s: %w", root, err)
	}

	currentMonth := now.Format("2006-01")

	// Collect candidates first (entries matching YYYY-MM-DD.md at root level),
	// then sort for deterministic output.
	type candidate struct {
		filename string
		year     string
		month    string
	}
	var candidates []candidate

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !dateFilePattern.MatchString(name) {
			continue
		}
		// Extract YYYY and MM from the filename prefix.
		year := name[0:4]
		month := name[5:7]
		if year+"-"+month == currentMonth {
			// Current-month file: never move.
			debug.Printf("archive: skipping current-month file %s", name)
			continue
		}
		candidates = append(candidates, candidate{filename: name, year: year, month: month})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].filename < candidates[j].filename
	})

	for _, c := range candidates {
		src := filepath.Join(root, c.filename)
		destDir := filepath.Join(root, c.year, c.month)
		dest := filepath.Join(destDir, c.filename)

		// If destination already exists, skip with warning. We do not compare
		// contents or overwrite.
		if _, err := os.Stat(dest); err == nil {
			result.Skipped = append(result.Skipped, Skip{
				Source: c.filename,
				Reason: fmt.Sprintf("destination %s already exists", filepath.Join(c.year, c.month, c.filename)),
			})
			debug.Printf("archive: skipping %s, destination exists", c.filename)
			continue
		} else if !os.IsNotExist(err) {
			// Unexpected stat error: halt with context.
			return result, fmt.Errorf("error checking destination %s: %w", dest, err)
		}

		if dryRun {
			result.Moved = append(result.Moved, c.filename)
			debug.Printf("archive: would move %s -> %s", c.filename, filepath.Join(c.year, c.month, c.filename))
			continue
		}

		if err := os.MkdirAll(destDir, 0755); err != nil {
			return result, fmt.Errorf("error creating archive directory %s: %w", destDir, err)
		}

		if err := os.Rename(src, dest); err != nil {
			return result, fmt.Errorf("error moving %s to %s: %w", src, dest, err)
		}

		result.Moved = append(result.Moved, c.filename)
		debug.Printf("archive: moved %s -> %s", c.filename, filepath.Join(c.year, c.month, c.filename))
	}

	return result, nil
}

// ArchiveCmd is the CLI wrapper around Archive. It runs the archive, emits
// status output to the given writers, and returns any error from the archive
// itself.
//
// Parameters:
//
//   - root:    directory to archive within
//   - now:     current time, used to determine "current month" (injected for testability)
//   - dryRun:  if true, report what would move without moving anything
//   - silent:  if true, do not print status when no files were moved and no
//     skips occurred. Warnings about skips and errors are still emitted. When
//     silent is false (explicit `archive` subcommand), dry-run output is more
//     verbose and a "nothing to move" line is printed in the no-op case.
//   - stdout:  writer for normal status messages
//   - stderr:  writer for warnings about skipped files
func ArchiveCmd(root string, now time.Time, dryRun, silent bool, stdout, stderr io.Writer) error {
	result, err := Archive(root, now, dryRun)
	if err != nil {
		return err
	}

	if dryRun {
		if len(result.Moved) == 0 && len(result.Skipped) == 0 {
			if !silent {
				_, _ = fmt.Fprintln(stdout, "Archive dry-run: nothing to move.")
			}
			return nil
		}
		_, _ = fmt.Fprintf(stdout, "Archive dry-run: would move %d file(s)\n", len(result.Moved))
		for _, f := range result.Moved {
			_, _ = fmt.Fprintf(stdout, "  %s -> %s/%s/%s\n", f, f[0:4], f[5:7], f)
		}
		for _, s := range result.Skipped {
			_, _ = fmt.Fprintf(stdout, "  skip %s (%s)\n", s.Source, s.Reason)
		}
		return nil
	}

	if len(result.Moved) > 0 {
		_, _ = fmt.Fprintf(stdout, "Archived %d file(s) into YYYY/MM/ subdirectories\n", len(result.Moved))
	}
	for _, s := range result.Skipped {
		_, _ = fmt.Fprintf(stderr, "Warning: skipped %s (%s)\n", s.Source, s.Reason)
	}
	return nil
}
