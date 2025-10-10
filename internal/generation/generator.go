package generation

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckExistingFiles scans for files that would be overwritten.
func CheckExistingFiles(descriptors []AssetFileDescriptor, baseDir string) OverwriteWarning {
	var existing []string

	for _, desc := range descriptors {
		fullPath := filepath.Join(baseDir, desc.Path)
		if _, err := os.Stat(fullPath); err == nil {
			existing = append(existing, fullPath)
		}
	}

	return OverwriteWarning{
		ExistingFiles:   existing,
		ConfirmedByUser: false,
	}
}

// GenerateAssetFiles orchestrates batch file generation.
func GenerateAssetFiles(descriptors []AssetFileDescriptor, baseDir string) GenerationReport {
	report := GenerationReport{
		TotalFiles: len(descriptors),
		Results:    make([]GenerationResult, 0, len(descriptors)),
	}

	for _, desc := range descriptors {
		fullPath := filepath.Join(baseDir, desc.Path)

		var result GenerationResult
		switch desc.Type {
		case AssetTypeSubagent:
			result = GenerateSubagentAssetFile(desc, fullPath)
		case AssetTypeHook:
			result = GenerateHookAssetFile(desc, fullPath)
		case AssetTypeSlashCommand:
			result = GenerateSlashCommandAssetFile(desc, fullPath)
		default:
			result = GenerationResult{
				FilePath: fullPath,
				Status:   StatusFailed,
				Error:    fmt.Errorf("unknown asset type: %v", desc.Type),
			}
		}

		report.Results = append(report.Results, result)

		switch result.Status {
		case StatusSuccess:
			report.Successful++
		case StatusPlaceholderGenerated:
			report.PlaceholdersGenerated++
		case StatusFailed:
			report.Failed++
			report.FailedDescriptors = append(report.FailedDescriptors, desc)
		}
	}

	return report
}

// RetryFailedGeneration retries only failed file generations.
func RetryFailedGeneration(report *GenerationReport, baseDir string) GenerationReport {
	return GenerateAssetFiles(report.FailedDescriptors, baseDir)
}
