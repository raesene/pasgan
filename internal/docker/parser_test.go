package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser(t *testing.T) {
	// Get the root directory of the project
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate to the root of the project (two levels up from internal/docker)
	rootDir := filepath.Join(wd, "..", "..")

	// Define test cases for each sample image
	testCases := []struct {
		name      string
		imagePath string
		wantErr   bool
	}{
		{
			name:      "nginx image",
			imagePath: filepath.Join(rootDir, "sample-data", "nginx.tar"),
			wantErr:   false,
		},
		{
			name:      "busybox image",
			imagePath: filepath.Join(rootDir, "sample-data", "busybox.tar"),
			wantErr:   false,
		},
		{
			name:      "nodegoat image",
			imagePath: filepath.Join(rootDir, "sample-data", "nodegoat.tar"),
			wantErr:   false,
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
			parser, err := NewParser(tc.imagePath)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}
			defer parser.Cleanup()

			// Parse the image
			metadata, err := parser.Parse()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tc.wantErr)
			}

			if metadata == nil && !tc.wantErr {
				t.Error("Expected metadata to be non-nil")
				return
			}

			if !tc.wantErr {
				// Validate basic metadata
				if len(metadata.History) == 0 {
					t.Error("Expected history to be non-empty")
				}

				if len(metadata.Layers) == 0 {
					t.Error("Expected layers to be non-empty")
				}

				// Check if we have a valid config
				if metadata.Config.Cmd == nil && metadata.Config.Entrypoint == nil {
					t.Error("Expected at least CMD or ENTRYPOINT to be set")
				}
			}
		})
	}
}