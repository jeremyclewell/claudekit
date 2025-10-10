package generation

// AssetType represents the type of asset being generated.
type AssetType int

const (
	AssetTypeSubagent AssetType = iota
	AssetTypeHook
	AssetTypeSlashCommand
)

// GenerationStatus indicates the outcome of file generation.
type GenerationStatus int

const (
	StatusSuccess GenerationStatus = iota
	StatusPlaceholderGenerated
	StatusFailed
)

// ComponentModule interface for accessing module properties.
type ComponentModule interface {
	GetDescription() string
	GetCategory() string
}

// AssetFileDescriptor represents metadata about a file to be generated.
type AssetFileDescriptor struct {
	Name           string           // Base name (e.g., "code-reviewer")
	Type           AssetType        // Type of asset
	Path           string           // Target path relative to assets/
	SourceTemplate string           // Optional template path
	Module         ComponentModule  // Reference to module
}

// GenerationResult tracks the outcome of a single file generation.
type GenerationResult struct {
	FilePath      string
	Status        GenerationStatus
	Error         error
	BytesWritten  int
	IsPlaceholder bool
}

// GenerationReport summarizes batch file generation.
type GenerationReport struct {
	TotalFiles            int
	Successful            int
	PlaceholdersGenerated int
	Failed                int
	Results               []GenerationResult
	FailedDescriptors     []AssetFileDescriptor
}

// OverwriteWarning represents files that will be overwritten.
type OverwriteWarning struct {
	ExistingFiles   []string
	ConfirmedByUser bool
}

// HasFailures returns true if any files failed to generate.
func (r *GenerationReport) HasFailures() bool {
	return r.Failed > 0
}

// GetFailedFiles returns list of failed file paths.
func (r *GenerationReport) GetFailedFiles() []string {
	var failed []string
	for _, result := range r.Results {
		if result.Status == StatusFailed {
			failed = append(failed, result.FilePath)
		}
	}
	return failed
}

// ShouldPromptRetry returns true if there are failures to retry.
func (r *GenerationReport) ShouldPromptRetry() bool {
	return r.HasFailures()
}
