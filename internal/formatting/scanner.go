package formatting

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScanMarkdownFiles walks the directory tree and finds all markdown files to process.
func ScanMarkdownFiles(cfg FormatConfig) ([]MarkdownFile, error) {
	var files []MarkdownFile

	// Validate root directory exists
	info, err := os.Stat(cfg.RootDir)
	if err != nil {
		return nil, fmt.Errorf("root directory error: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %s", cfg.RootDir)
	}

	err = filepath.Walk(cfg.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories (will recurse into them unless excluded)
		if info.IsDir() {
			// Check if directory matches exclusion pattern
			relPath, _ := filepath.Rel(cfg.RootDir, path)
			for _, pattern := range cfg.ExcludePatterns {
				// Pattern can match directory name with or without trailing slash
				cleanPattern := strings.TrimSuffix(pattern, "/")
				if strings.HasPrefix(relPath, cleanPattern) || relPath == cleanPattern {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Only process .md files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Calculate relative path for display
		relPath, _ := filepath.Rel(cfg.RootDir, path)

		// Create MarkdownFile entry
		files = append(files, MarkdownFile{
			Path:    path,
			RelPath: relPath,
			Size:    info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %w", err)
	}

	return files, nil
}
