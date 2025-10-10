package formatting

import (
	"bytes"
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// FormatMarkdownFile formats a single markdown file according to GFM rules.
func FormatMarkdownFile(file *MarkdownFile, cfg FormatConfig) (*FormatResult, error) {
	startTime := time.Now()

	result := &FormatResult{
		File:   *file,
		Status: StatusUnchanged,
	}

	// Read file content if not already loaded
	if file.Content == nil {
		content, err := os.ReadFile(file.Path)
		if err != nil {
			result.Status = StatusError
			result.Error = fmt.Errorf("failed to read file: %w", err)
			result.Duration = time.Since(startTime)
			return result, result.Error
		}
		file.Content = content
	}

	// Validate UTF-8
	if !utf8.Valid(file.Content) {
		result.Status = StatusError
		result.Error = fmt.Errorf("file contains invalid UTF-8")
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// Parse markdown
	doc, ctx, err := ParseMarkdown(file.Content)
	if err != nil {
		result.Status = StatusError
		result.Error = fmt.Errorf("failed to parse markdown: %w", err)
		file.ParseErrors = append(file.ParseErrors, err)
		result.Duration = time.Since(startTime)
		return result, result.Error
	}

	// Apply formatting rules
	formatted, rulesApplied := ApplyFormattingRules(doc, ctx, file.Content, cfg)

	// Check if content changed
	if bytes.Equal(file.Content, formatted) {
		result.Status = StatusUnchanged
	} else {
		file.FormattedContent = formatted
		result.Status = StatusModified
		result.RulesApplied = rulesApplied

		// Write file if not dry-run
		if !cfg.DryRun {
			if err := AtomicWriteFile(file.Path, formatted); err != nil {
				result.Status = StatusError
				result.Error = fmt.Errorf("failed to write file: %w", err)
				result.Duration = time.Since(startTime)
				return result, result.Error
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// ApplyFormattingRules applies all formatting rules to the AST.
func ApplyFormattingRules(doc ast.Node, ctx parser.Context, source []byte, cfg FormatConfig) ([]byte, []FormattingRule) {
	var rulesApplied []FormattingRule

	// Transform AST: convert indented code blocks to fenced
	convertIndentedCodeToFenced(doc, source)

	// Walk AST and apply transformations
	applyHeadingRules(doc, source, &rulesApplied)
	applyListRules(doc, source, &rulesApplied)
	applyCodeBlockRules(doc, source, &rulesApplied)
	applyTableRules(doc, source, &rulesApplied)
	applyEmphasisRules(doc, source, &rulesApplied)
	applyWhitespaceRules(doc, source, &rulesApplied)
	applyHorizontalRuleRules(doc, source, &rulesApplied)

	// Re-render the AST to get formatted output
	// Goldmark's markdown renderer will normalize most formatting automatically
	formatted, err := RenderMarkdown(doc, source)
	if err != nil {
		// If rendering fails, return original
		return source, []FormattingRule{}
	}

	// Apply final whitespace cleanup
	formatted = NormalizeWhitespace(formatted)

	return formatted, rulesApplied
}
