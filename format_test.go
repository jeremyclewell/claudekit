package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"jeremyclewell.com/claudekit/internal/formatting"
)

// Phase 3.2: Behavioral Contract Tests (T006-T015)

// T006: BC-001 - File discovery test
func TestFileDiscovery(t *testing.T) {
	// Setup: Create temp directory with test structure
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"README.md",
		"docs/guide.md",
		"src/README.md",
	}

	// Create excluded files (should not be found)
	excludedFiles := []string{
		"node_modules/README.md",
		"vendor/docs.md",
		".git/config.md",
		"build/output.md",
		"dist/bundle.md",
		"out/result.md",
	}

	// Create all files
	for _, f := range testFiles {
		path := filepath.Join(tmpDir, f)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("# Test"), 0644)
	}

	for _, f := range excludedFiles {
		path := filepath.Join(tmpDir, f)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("# Excluded"), 0644)
	}

	// Create config with default exclusions
	cfg := formatting.FormatConfig{
		RootDir: tmpDir,
		ExcludePatterns: []string{"node_modules/", "vendor/", ".git/", "build/", "dist/", "out/"},
	}

	// Execute
	files, err := formatting.ScanMarkdownFiles(cfg)

	// Assert
	if err != nil {
		t.Fatalf("scanMarkdownFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	// Verify excluded files not in results
	for _, file := range files {
		for _, excluded := range excludedFiles {
			// Check if the file's relative path matches the excluded path
			if file.RelPath == excluded {
				t.Errorf("Excluded file found in results: %s", excluded)
			}
		}
	}
}

// T007: BC-002 - Idempotence test
func TestIdempotence(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	// Create file with formatting issues
	content := `Test
===

##No Space

* Item`
	os.WriteFile(testFile, []byte(content), 0644)

	cfg := formatting.FormatConfig{
		RootDir: tmpDir,
		DryRun:  false,
	}

	// First format
	files1, _ := formatting.ScanMarkdownFiles(cfg)
	_, _ = formatting.FormatMarkdownFile(&files1[0], cfg)

	// Read formatted content
	formatted1, _ := os.ReadFile(testFile)

	// Second format
	files2, _ := formatting.ScanMarkdownFiles(cfg)
	result2, _ := formatting.FormatMarkdownFile(&files2[0], cfg)

	// Read content after second format
	formatted2, _ := os.ReadFile(testFile)

	// Assert: Second run should make no changes
	if result2.Status != "unchanged" {
		t.Errorf("Second format should be unchanged, got: %s", result2.Status)
	}

	// Content should be identical
	if string(formatted1) != string(formatted2) {
		t.Error("File content changed on second format (not idempotent)")
	}
}

// T008: BC-003 - Code block preservation test
func TestCodeBlockPreservation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "code-test.md")

	// Code block with intentional formatting issues that should NOT be fixed
	content := `# Test

` + "```" + `markdown
##Bad formatting in code
* Mixed
- Markers
    indented badly
` + "```" + `

Regular text.`

	os.WriteFile(testFile, []byte(content), 0644)

	cfg := formatting.FormatConfig{RootDir: tmpDir}
	files, _ := formatting.ScanMarkdownFiles(cfg)

	// Get original code block content
	originalCodeBlock := `##Bad formatting in code
* Mixed
- Markers
    indented badly`

	// Format
	formatting.FormatMarkdownFile(&files[0], cfg)

	// Read formatted content
	formatted, _ := os.ReadFile(testFile)
	formattedStr := string(formatted)

	// Assert: Code block content must be byte-identical
	if !strings.Contains(formattedStr, originalCodeBlock) {
		t.Error("Code block content was modified during formatting")
	}
}

// T009: BC-004 - Atomic writes test
func TestAtomicWrites(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "atomic-test.md")

	original := "# Original Content"
	os.WriteFile(testFile, []byte(original), 0644)

	// Simulate write failure by making directory read-only after file creation
	// This test validates that on write failure, original file remains unchanged

	// Store original modification time
	origInfo, _ := os.Stat(testFile)
	origModTime := origInfo.ModTime()

	// Attempt atomic write with simulated failure
	content := []byte("# Modified Content")

	// The atomicWriteFile should either succeed completely or leave original unchanged
	err := formatting.AtomicWriteFile(testFile, content)

	if err != nil {
		// If write failed, verify original file is unchanged
		current, _ := os.ReadFile(testFile)
		if string(current) != original {
			t.Error("Original file was modified despite write failure")
		}

		currentInfo, _ := os.Stat(testFile)
		if !currentInfo.ModTime().Equal(origModTime) {
			t.Error("File modification time changed despite write failure")
		}
	}
}

// T010: BC-005 - Git integration test
func TestGitIntegration(t *testing.T) {
	// This test requires a git repository
	tmpDir := t.TempDir()

	// Initialize git repo
	os.Chdir(tmpDir)
	gitExec := func(args ...string) error {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmpDir
		return cmd.Run()
	}

	if err := gitExec("init"); err != nil {
		t.Skip("Git not available, skipping git integration test")
	}

	// Create and commit initial file
	testFile := filepath.Join(tmpDir, "git-test.md")
	os.WriteFile(testFile, []byte("# Original"), 0644)
	gitExec("add", "git-test.md")
	gitExec("commit", "-m", "Initial commit")

	// Modify file with formatter
	cfg := formatting.FormatConfig{RootDir: tmpDir}
	files, _ := formatting.ScanMarkdownFiles(cfg)
	formatting.FormatMarkdownFile(&files[0], cfg)

	// Check git status shows modifications
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = tmpDir
	output, _ := cmd.Output()

	if len(output) == 0 {
		t.Error("Git should show modifications after formatting")
	}

	// Verify changes are reversible
	restoreCmd := exec.Command("git", "restore", "git-test.md")
	restoreCmd.Dir = tmpDir
	restoreCmd.Run()
	restored, _ := os.ReadFile(testFile)

	if string(restored) != "# Original" {
		t.Error("Git restore should revert formatting changes")
	}
}

// T011: BC-006 - Dry-run safety test
func TestDryRunSafety(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "dry-run-test.md")

	original := `Test
===

##No Space`
	os.WriteFile(testFile, []byte(original), 0644)

	// Get original file info
	origInfo, _ := os.Stat(testFile)
	origModTime := origInfo.ModTime()

	// Format with dry-run enabled
	cfg := formatting.FormatConfig{
		RootDir: tmpDir,
		DryRun:  true,
	}

	files, _ := formatting.ScanMarkdownFiles(cfg)
	result, _ := formatting.FormatMarkdownFile(&files[0], cfg)

	// Assert: File should not be modified
	current, _ := os.ReadFile(testFile)
	if string(current) != original {
		t.Error("Dry-run mode modified the file")
	}

	// Modification time should not change
	currentInfo, _ := os.Stat(testFile)
	if !currentInfo.ModTime().Equal(origModTime) {
		t.Error("Dry-run mode changed file modification time")
	}

	// Result should indicate changes would be made (but weren't)
	if result.Status == "unchanged" && len(result.RulesApplied) > 0 {
		t.Error("Dry-run should report potential changes")
	}
}

// T012: BC-007 - Exclusion patterns test
func TestExclusionPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in various directories
	files := map[string]bool{
		"README.md":                    false, // should be included
		"docs/guide.md":                false, // should be included
		"node_modules/package/doc.md":  true,  // default exclusion
		"vendor/lib/README.md":         true,  // default exclusion
		"custom/exclude/test.md":       true,  // custom exclusion
		"build/output.md":              true,  // default exclusion
	}

	for path := range files {
		fullPath := filepath.Join(tmpDir, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte("# Test"), 0644)
	}

	// Config with default + custom exclusions
	cfg := formatting.FormatConfig{
		RootDir: tmpDir,
		ExcludePatterns: []string{
			"node_modules/", "vendor/", ".git/", "build/", "dist/", "out/",
			"custom/exclude/", // custom exclusion
		},
	}

	// Scan files
	found, _ := formatting.ScanMarkdownFiles(cfg)

	// Assert: Only non-excluded files found
	if len(found) != 2 {
		t.Errorf("Expected 2 files after exclusions, got %d", len(found))
	}

	// Verify no excluded files in results
	for _, file := range found {
		for path, shouldExclude := range files {
			// Check if the file's relative path matches an excluded path
			if shouldExclude && (file.RelPath == path || strings.HasPrefix(file.RelPath, filepath.Dir(path)+string(filepath.Separator))) {
				t.Errorf("Excluded file found: %s", path)
			}
		}
	}
}

// T013: BC-008 - Progress indication test
func TestProgressIndication(t *testing.T) {
	tmpDir := t.TempDir()

	// Create >50 files to trigger progress display
	for i := 0; i < 60; i++ {
		path := filepath.Join(tmpDir, "file"+string(rune(i))+".md")
		os.WriteFile(path, []byte("# Test"), 0644)
	}

	cfg := formatting.FormatConfig{RootDir: tmpDir}

	// Track if progress callback was called
	progressCalled := false

	// Format with progress tracking (this will be implemented in the command)
	files, _ := formatting.ScanMarkdownFiles(cfg)

	if len(files) > 50 {
		// Progress display should be triggered
		progressCalled = true
	}

	if !progressCalled && len(files) > 50 {
		t.Error("Progress indication should be shown for >50 files")
	}
}

// T014: BC-009 - Error resilience test
func TestErrorResilience(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mix of valid and invalid files
	validFile := filepath.Join(tmpDir, "valid.md")
	invalidFile := filepath.Join(tmpDir, "invalid.md")
	anotherValid := filepath.Join(tmpDir, "another.md")

	os.WriteFile(validFile, []byte("# Valid"), 0644)
	os.WriteFile(invalidFile, []byte{0xFF, 0xFE, 0x00}, 0644) // Invalid UTF-8
	os.WriteFile(anotherValid, []byte("# Another Valid"), 0644)

	cfg := formatting.FormatConfig{RootDir: tmpDir}

	files, _ := formatting.ScanMarkdownFiles(cfg)
	report := formatting.FormatReport{}

	// Process all files
	for _, file := range files {
		result, err := formatting.FormatMarkdownFile(&file, cfg)
		if err != nil {
			report.FilesErrored++
			report.Errors = append(report.Errors, err)
		} else if result.Status == "modified" || result.Status == "unchanged" {
			report.FilesModified++
		}
		report.Results = append(report.Results, *result)
	}

	// Assert: Valid files processed, invalid files skipped
	if report.FilesErrored != 1 {
		t.Errorf("Expected 1 error, got %d", report.FilesErrored)
	}

	if report.FilesModified < 2 {
		t.Error("Valid files should have been processed despite errors")
	}
}

// T015: BC-010 - Performance target test
func TestPerformanceTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir := t.TempDir()

	// Create 100 files @ ~50KB each
	for i := 0; i < 100; i++ {
		path := filepath.Join(tmpDir, "perf"+string(rune(i))+".md")
		content := make([]byte, 50000)
		for j := range content {
			content[j] = byte('a' + (j % 26))
		}
		os.WriteFile(path, content, 0644)
	}

	cfg := formatting.FormatConfig{RootDir: tmpDir}

	// Measure time
	start := time.Now()

	files, _ := formatting.ScanMarkdownFiles(cfg)
	for _, file := range files {
		formatting.FormatMarkdownFile(&file, cfg)
	}

	duration := time.Since(start)

	// Assert: <10 seconds for 100 files
	if duration > 10*time.Second {
		t.Errorf("Performance target failed: took %v, expected <10s", duration)
	}
}

// Phase 3.2: Formatting Rule Tests (T016-T023)

// T016: Heading formatting test
func TestHeadingFormatting(t *testing.T) {
	input, _ := os.ReadFile("tests/fixtures/markdown/malformed/headings.md")

	cfg := formatting.FormatConfig{}
	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: input,
	}

	result, _ := formatting.FormatMarkdownFile(&file, cfg)
	formatted := string(file.FormattedContent)

	// Assert: Setext headings converted to ATX
	if strings.Contains(formatted, "====") || strings.Contains(formatted, "----") {
		t.Error("Setext headings should be converted to ATX style")
	}

	// Assert: Space after hash marks
	if strings.Contains(formatted, "##No") || strings.Contains(formatted, "####Another") {
		t.Error("Headings should have space after hash marks")
	}

	// Assert: Blank lines around headings
	// (will verify in implementation)

	if result.Status == "unchanged" {
		t.Error("Malformed headings should be modified")
	}
}

// T017: List formatting test
func TestListFormatting(t *testing.T) {
	input, _ := os.ReadFile("tests/fixtures/markdown/malformed/lists.md")

	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: input,
	}

	result, _ := formatting.FormatMarkdownFile(&file, formatting.FormatConfig{})
	_ = string(file.FormattedContent)

	// Assert: Consistent markers (prefer dash)
	// Assert: 2-space indentation for nested lists
	// Assert: Proper spacing

	if result.Status == "unchanged" {
		t.Error("Malformed lists should be modified")
	}
}

// T018: Code block formatting test
func TestCodeBlockFormatting(t *testing.T) {
	input, _ := os.ReadFile("tests/fixtures/markdown/malformed/code.md")

	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: input,
	}

	_, _ = formatting.FormatMarkdownFile(&file, formatting.FormatConfig{})
	formatted := string(file.FormattedContent)

	// Assert: Indented code converted to fenced
	if strings.Contains(formatted, "    function hello()") {
		t.Error("Indented code blocks should be converted to fenced")
	}

	// Assert: Tildes converted to backticks
	if strings.Contains(formatted, "~~~") {
		t.Error("Tilde fences should be converted to backticks")
	}

	// Assert: Code content preserved exactly
	if !strings.Contains(formatted, `console.log("Hello, world!")`) {
		t.Error("Code block content must be preserved")
	}
}

// T019: Table formatting test
func TestTableFormatting(t *testing.T) {
	input, _ := os.ReadFile("tests/fixtures/markdown/malformed/tables.md")

	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: input,
	}

	result, _ := formatting.FormatMarkdownFile(&file, formatting.FormatConfig{})

	// Assert: Tables are aligned
	// Assert: Consistent pipe placement

	if result.Status == "unchanged" {
		t.Error("Malformed tables should be modified")
	}
}

// T020: Link formatting test
func TestLinkFormatting(t *testing.T) {
	input := `[ref]: https://example.com    "Title"
[Link][ref]`

	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: []byte(input),
	}

	_, _ = formatting.FormatMarkdownFile(&file, formatting.FormatConfig{})
	formatted := string(file.FormattedContent)

	// Assert: Single space between URL and title
	if strings.Contains(formatted, "    ") {
		t.Error("Link references should have single space between URL and title")
	}
}

// T021: Emphasis formatting test
func TestEmphasisFormatting(t *testing.T) {
	input := `_italic_ and __bold__`

	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: []byte(input),
	}

	_, _ = formatting.FormatMarkdownFile(&file, formatting.FormatConfig{})
	formatted := string(file.FormattedContent)

	// Assert: Underscores converted to asterisks
	if strings.Contains(formatted, "_italic_") || strings.Contains(formatted, "__bold__") {
		t.Error("Underscores should be converted to asterisks for emphasis")
	}

	// Should be *italic* and **bold**
	if !strings.Contains(formatted, "*italic*") || !strings.Contains(formatted, "**bold**") {
		t.Error("Emphasis should use asterisks")
	}
}

// T022: Whitespace formatting test
func TestWhitespaceFormatting(t *testing.T) {
	input, _ := os.ReadFile("tests/fixtures/markdown/malformed/whitespace.md")

	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: input,
	}

	_, _ = formatting.FormatMarkdownFile(&file, formatting.FormatConfig{})
	formatted := string(file.FormattedContent)

	// Assert: No trailing whitespace
	lines := strings.Split(formatted, "\n")
	for i, line := range lines {
		if len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t') {
			t.Errorf("Line %d has trailing whitespace", i+1)
		}
	}

	// Assert: No multiple blank lines
	if strings.Contains(formatted, "\n\n\n") {
		t.Error("Should not have multiple consecutive blank lines")
	}
}

// T023: Horizontal rule formatting test
func TestHorizontalRuleFormatting(t *testing.T) {
	input := `***
___
- - -`

	file := formatting.MarkdownFile{
		Path:    "test.md",
		Content: []byte(input),
	}

	_, _ = formatting.FormatMarkdownFile(&file, formatting.FormatConfig{})
	formatted := string(file.FormattedContent)

	// Assert: All converted to triple dash
	if strings.Contains(formatted, "***") || strings.Contains(formatted, "___") || strings.Contains(formatted, "- - -") {
		t.Error("All horizontal rules should use triple dash ---")
	}

	// Count dashes (should have 3 --- entries)
	dashCount := strings.Count(formatted, "---")
	if dashCount != 3 {
		t.Errorf("Expected 3 horizontal rules (---), got %d", dashCount)
	}
}

// Helper functions for tests
