package formatting

import (
	"time"
)

// FormatConfig holds configuration for the markdown formatting operation.
type FormatConfig struct {
	RootDir         string
	ExcludePatterns []string
	DryRun          bool
	Verbose         bool
	Standard        string // Fixed to "GFM"
}

// MarkdownFile represents a single markdown file to process.
type MarkdownFile struct {
	Path             string
	RelPath          string
	Size             int64
	Content          []byte
	FormattedContent []byte
	ParseErrors      []error
}

// MarkdownFile state constants.
const (
	StateDiscovered = "discovered"
	StateLoaded     = "loaded"
	StateParsed     = "parsed"
	StateFormatted  = "formatted"
	StateWritten    = "written"
	StateSkipped    = "skipped"
	StateFailed     = "failed"
)

// FormattingRule represents a single formatting rule.
type FormattingRule struct {
	Name        string
	Description string
	Category    string
	FixCount    int
}

// FormattingRule categories.
const (
	CategoryHeading        = "heading"
	CategoryList           = "list"
	CategoryCode           = "code"
	CategoryTable          = "table"
	CategoryLink           = "link"
	CategoryEmphasis       = "emphasis"
	CategoryWhitespace     = "whitespace"
	CategoryHorizontalRule = "horizontal-rule"
)

// FormatResult represents the result of formatting a single file.
type FormatResult struct {
	File         MarkdownFile
	Status       string
	RulesApplied []FormattingRule
	LineChanges  []LineChange
	Error        error
	Duration     time.Duration
}

// FormatResult status values.
const (
	StatusModified  = "modified"
	StatusUnchanged = "unchanged"
	StatusExcluded  = "excluded"
	StatusSkipped   = "skipped"
	StatusError     = "error"
)

// LineChange represents a detailed change to a specific line.
type LineChange struct {
	LineNumber int
	RuleName   string
	Before     string
	After      string
}

// FormatReport aggregates results for the entire formatting operation.
type FormatReport struct {
	TotalFiles        int
	FilesModified     int
	FilesUnchanged    int
	FilesExcluded     int
	FilesErrored      int
	TotalFixesApplied int
	RuleStats         map[string]int
	Results           []FormatResult
	Duration          time.Duration
	Errors            []error
}
