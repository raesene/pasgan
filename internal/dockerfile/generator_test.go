package dockerfile

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/raesene/pasgan/internal/docker"
)

func TestGenerator(t *testing.T) {
	// Get the root directory of the project
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to the root of the project (two levels up from internal/dockerfile)
	rootDir := filepath.Join(wd, "..", "..")

	// Define test cases for each sample image
	testCases := []struct {
		name         string
		imagePath    string
		wantErr      bool
		expectStrings []string  // Strings we expect to find in the generated Dockerfile
	}{
		{
			name:         "nginx image",
			imagePath:    filepath.Join(rootDir, "sample-data", "nginx.tar"),
			wantErr:      false,
			expectStrings: []string{
				"FROM",
				"EXPOSE 80",
				"ENTRYPOINT",
				"CMD",
			},
		},
		{
			name:         "busybox image",
			imagePath:    filepath.Join(rootDir, "sample-data", "busybox.tar"),
			wantErr:      false,
			expectStrings: []string{
				"FROM",
				"CMD",
			},
		},
		{
			name:         "nodegoat image",
			imagePath:    filepath.Join(rootDir, "sample-data", "nodegoat.tar"),
			wantErr:      false,
			expectStrings: []string{
				"FROM",
				"WORKDIR",
				"COPY",
				"RUN",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Verify sample image exists
			if _, err := os.Stat(tc.imagePath); os.IsNotExist(err) {
				t.Skipf("Sample image not found: %s", tc.imagePath)
				return
			}

			// Create a new parser
			parser, err := docker.NewParser(tc.imagePath)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}
			defer parser.Cleanup()

			// Parse the image
			metadata, err := parser.Parse()
			if err != nil {
				t.Fatalf("Failed to parse image: %v", err)
				return
			}

			// Create a generator
			generator := NewGenerator(metadata)

			// Generate the Dockerfile
			var buf bytes.Buffer
			err = generator.Generate(&buf)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Generate() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr {
				// Check the generated content
				result := buf.String()
				t.Logf("Generated Dockerfile:\n%s", result)

				// Check if all expected strings are present
				for _, expectStr := range tc.expectStrings {
					if !strings.Contains(result, expectStr) {
						t.Errorf("Expected '%s' in the generated Dockerfile, but it was not found", expectStr)
					}
				}

				// Ensure the Dockerfile is not empty
				if len(result) == 0 {
					t.Error("Generated Dockerfile is empty")
				}
			}
		})
	}
}