package utils

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractTar extracts a tar file to a destination directory
func ExtractTar(tarPath, destDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)
	
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Get absolute path for destination directory to check for path traversal
	destDirAbs, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destination: %w", err)
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar: %w", err)
		}

		// Skip entries with absolute paths or path traversal attempts
		if filepath.IsAbs(header.Name) || strings.Contains(header.Name, "..") {
			continue
		}

		// Construct target path
		target := filepath.Join(destDir, header.Name)
		
		// Check for path traversal
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("error resolving path: %w", err)
		}
		if !strings.HasPrefix(targetAbs, destDirAbs) {
			continue // Skip this file for security
		}

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		case tar.TypeSymlink:
			// Create symlinks only if they point inside the destination
			linkTarget := filepath.Join(filepath.Dir(target), header.Linkname)
			linkTargetAbs, err := filepath.Abs(linkTarget)
			if err != nil {
				continue
			}
			if !strings.HasPrefix(linkTargetAbs, destDirAbs) {
				continue
			}
			
			// Create the symlink
			if err := os.Symlink(header.Linkname, target); err != nil {
				// Non-fatal, just continue
				continue
			}
		default:
			// Skip other types for simplicity
			continue
		}
	}

	return nil
}