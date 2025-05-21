package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestMainIntegration tests the complete workflow of the tool
func TestMainIntegration(t *testing.T) {
	// Get sample images from sample-data directory
	sampleDataDir := filepath.Join(".", "sample-data")
	files, err := os.ReadDir(sampleDataDir)
	if err != nil {
		t.Skipf("Skipping integration test, sample-data directory not found: %v", err)
		return
	}

	// Find tar files in the sample-data directory
	var tarFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".tar" {
			tarFiles = append(tarFiles, filepath.Join(sampleDataDir, file.Name()))
		}
	}

	if len(tarFiles) == 0 {
		t.Skip("No sample Docker image files found in sample-data directory")
		return
	}

	// For each tar file, run the analyze command
	for _, tarFile := range tarFiles {
		t.Run(filepath.Base(tarFile), func(t *testing.T) {
			// Build the tool first
			buildCmd := exec.Command("go", "build", "-o", "pasgan")
			buildOutput, err := buildCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to build tool: %v\nOutput: %s", err, buildOutput)
			}
			defer os.Remove("pasgan") // Clean up the binary

			// Create a temporary file for the output
			outputFile := filepath.Join(os.TempDir(), "Dockerfile."+filepath.Base(tarFile))
			defer os.Remove(outputFile)

			// Run the analyze command
			analyzeCmd := exec.Command("./pasgan", "analyze", tarFile, "-o", outputFile)
			output, err := analyzeCmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Failed to analyze %s: %v\nOutput: %s", tarFile, err, output)
			}

			// Verify the output file exists and is not empty
			outputContent, err := os.ReadFile(outputFile)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			if len(outputContent) == 0 {
				t.Error("Output Dockerfile is empty")
			}

			t.Logf("Successfully analyzed %s, Dockerfile generated at %s", tarFile, outputFile)
		})
	}
}