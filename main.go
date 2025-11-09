package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dailynotes/internal/debug"
	"dailynotes/internal/files"
	"dailynotes/internal/tasks"
	"dailynotes/internal/template"
)

// CLI flags
var (
	debugFlag    = flag.Bool("debug", false, "Enable debug output")
	dryRunFlag   = flag.Bool("dry-run", false, "Show what would be created without writing files")
	dirFlag      = flag.String("dir", ".", "Directory containing daily note files")
	templateFlag = flag.String("template", "", "Path to template file (uses default if not specified)")
	forceFlag    = flag.Bool("force", false, "Overwrite existing file without prompting")
	listFlag     = flag.Bool("list", false, "List all daily note files in the directory")
)

// promptYesNo asks the user a yes/no question and returns true for yes
func promptYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s (y/n): ", question)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// listDailyNotes lists all daily note files in the directory
func listDailyNotes() error {
	results, err := files.List(*dirFlag)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Printf("No daily notes found in: %s\n", *dirFlag)
		fmt.Println("Daily notes use the format: YYYY-MM-DD.md")
		return nil
	}

	fmt.Printf("Found %d daily notes in: %s\n\n", len(results), *dirFlag)

	for _, info := range results {
		marker := "  "
		if info.IsMostRecent {
			marker = "→ "
		}

		taskStats := "?"
		if info.Total > 0 {
			taskStats = fmt.Sprintf("%d/%d", info.Completed, info.Total)
		} else {
			taskStats = "0"
		}

		fmt.Printf("%s%s  (%.1f KB, %s tasks)\n", marker, info.Date, info.SizeKB, taskStats)
	}

	return nil
}

// createDailyNote is the main workflow
func createDailyNote() error {
	// Validate directory exists
	if _, err := os.Stat(*dirFlag); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s\n  Create it with: mkdir -p %s", *dirFlag, *dirFlag)
		}
		return fmt.Errorf("cannot access directory %s: %w", *dirFlag, err)
	}

	// Load template
	tmpl, err := template.Load(*templateFlag)
	if err != nil {
		if *templateFlag != "" {
			return fmt.Errorf("failed to load template file: %w\n  Check that the file exists and is readable", err)
		}
		return err
	}

	// Validate template has date placeholder
	if !strings.Contains(tmpl, "{{.Date}}") {
		fmt.Println("Warning: Template does not contain {{.Date}} placeholder")
	}

	// Get today's filename
	todayFilename := files.TodayFilename()
	todayPath := filepath.Join(*dirFlag, todayFilename)

	// Check if file already exists
	if _, err := os.Stat(todayPath); err == nil {
		if !*forceFlag {
			if !promptYesNo(fmt.Sprintf("File %s already exists. Overwrite?", todayFilename)) {
				return fmt.Errorf("aborted: file already exists at %s", todayPath)
			}
		}
	}

	// Find most recent file
	recentFile, err := files.FindMostRecent(*dirFlag)
	if err != nil {
		return err
	}

	var mergedContent string

	if recentFile == "" {
		// No previous files, just use template
		fmt.Println("No previous daily notes found, creating from template")
		mergedContent = strings.ReplaceAll(tmpl, "{{.Date}}", time.Now().Format("2006-01-02"))
	} else {
		fmt.Printf("Loading tasks from: %s\n", filepath.Base(recentFile))

		// Load and parse previous file
		content, err := files.Load(recentFile)
		if err != nil {
			return err
		}

		doc, err := tasks.Extract([]byte(content))
		if err != nil {
			return fmt.Errorf("failed to parse markdown in %s: %w\n  The file may have invalid markdown syntax", filepath.Base(recentFile), err)
		}

		// Get uncompleted tasks
		uncompleted := tasks.FilterUncompleted(doc)

		taskCount := 0
		for _, section := range uncompleted.Sections {
			taskCount += tasks.Count(section.Tasks)
		}

		if taskCount == 0 {
			fmt.Println("No uncompleted tasks found to carry over")
		} else {
			fmt.Printf("Carrying over %d uncompleted task(s) from %d section(s)\n", taskCount, len(uncompleted.Sections))
		}

		// Merge with template
		mergedContent = template.Merge(tmpl, uncompleted, time.Now().Format("2006-01-02"))
	}

	// Write or display
	if *dryRunFlag {
		fmt.Println("\n=== Dry run: would create ===")
		fmt.Printf("File: %s\n\n", todayPath)
		fmt.Println(mergedContent)
	} else {
		err := files.Write(todayPath, mergedContent)
		if err != nil {
			return err
		}
		fmt.Printf("✓ Created: %s\n", todayPath)
	}

	return nil
}

func main() {
	flag.Parse()

	if *debugFlag {
		debug.Enabled = true
	}

	if *listFlag {
		err := listDailyNotes()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	err := createDailyNote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
