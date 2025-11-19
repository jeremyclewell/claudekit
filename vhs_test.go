package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestVHSVisualScenarios runs VHS-based visual tests if VHS is installed
// These tests generate screenshots for manual visual validation
func TestVHSVisualScenarios(t *testing.T) {
	// Skip if VHS is not installed
	if !isVHSInstalled() {
		t.Skip("VHS not installed - skipping visual tests. Install with: brew install vhs")
	}

	// Skip if running in CI without explicit opt-in
	if os.Getenv("CI") == "true" && os.Getenv("RUN_VHS_TESTS") != "true" {
		t.Skip("Skipping VHS tests in CI (set RUN_VHS_TESTS=true to enable)")
	}

	vhsTestDir := "specs/002-lets-make-the/vhs-tests"

	// Ensure the binary is built
	t.Log("Building claudekit binary...")
	buildCmd := exec.Command("go", "build", ".")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build claudekit: %v\n%s", err, output)
	}

	// Create output directory
	outputDir := filepath.Join(vhsTestDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Define test scenarios
	scenarios := []struct {
		name     string
		tapeFile string
		testID   string
	}{
		{"Wide Terminal ASCII Art", "scenario-1-wide-terminal.tape", "T019"},
		{"Gradient Foreground Coloring", "scenario-2-gradient-foreground.tape", "T020"},
		{"Narrow Terminal Fallback", "scenario-3-fallback.tape", "T021"},
		{"60-Column Boundary", "scenario-4-boundary.tape", "T022"},
		{"256-Color Terminal", "scenario-6-256color.tape", "T024"},
		{"8-Color Terminal", "scenario-7-8color.tape", "T025"},
	}

	// Run each VHS scenario
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			tapePath := filepath.Join(vhsTestDir, scenario.tapeFile)

			// Check if tape file exists
			if _, err := os.Stat(tapePath); os.IsNotExist(err) {
				t.Skipf("Tape file not found: %s", tapePath)
			}

			t.Logf("Running VHS scenario: %s (%s)", scenario.name, scenario.testID)

			// Run VHS
			cmd := exec.Command("vhs", tapePath)
			cmd.Dir = "." // Run from repo root
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("VHS scenario failed: %s\nOutput: %s", scenario.name, output)
			} else {
				t.Logf("✓ Screenshot generated successfully")
			}
		})
	}

	// Print summary
	t.Log("\n" + string('=') + " VHS Visual Testing Complete " + string('='))
	t.Logf("Screenshots generated in: %s", outputDir)
	t.Log("\nNext steps:")
	t.Log("1. Review screenshots in output/ directory")
	t.Log("2. Use specs/002-lets-make-the/vhs-tests/VALIDATION-CHECKLIST.md")
	t.Log("3. Validate visual quality manually")
}

// isVHSInstalled checks if VHS command is available
func isVHSInstalled() bool {
	_, err := exec.LookPath("vhs")
	return err == nil
}

// TestVHSInstallation provides installation guidance if VHS is not found
func TestVHSInstallation(t *testing.T) {
	if !isVHSInstalled() {
		t.Log("VHS is not installed")
		t.Log("\nTo install VHS for visual testing:")
		t.Log("  brew install vhs")
		t.Log("\nOr via Go:")
		t.Log("  go install github.com/charmbracelet/vhs@latest")
		t.Log("\nVHS enables automated screenshot generation for visual validation")
		t.Log("See: https://github.com/charmbracelet/vhs")
	} else {
		// Get VHS version
		cmd := exec.Command("vhs", "--version")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Logf("✓ VHS is installed: %s", string(output))
		} else {
			t.Logf("✓ VHS is installed")
		}
	}
}
