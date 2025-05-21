package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/raesene/pasgan/internal/docker"
	"github.com/raesene/pasgan/internal/dockerfile"
	"github.com/spf13/cobra"
)

var (
	outputFile   string
	outputFormat string
	verbose      bool
)

// Initialize all commands
func initCommands() {
	// Add version command
	rootCmd.AddCommand(createVersionCmd())
	
	// Add analyze command
	rootCmd.AddCommand(createAnalyzeCmd())
}

// Create the version command
func createVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version, build date, and commit hash.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("pasgan version: %s\n", version)
			fmt.Printf("commit: %s\n", commit)
			fmt.Printf("built: %s\n", date)
		},
	}
}

// Create the analyze command
func createAnalyzeCmd() *cobra.Command {
	analyzeCmd := &cobra.Command{
		Use:   "analyze [image_tar]",
		Short: "Analyze a Docker image and generate a Dockerfile",
		Long: `Analyze takes a saved Docker image (.tar file) and analyzes its structure
to reconstruct a Dockerfile that could have been used to create it.

Example:
  pasgan analyze nginx.tar`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			imagePath := args[0]
			
			// Ensure the file exists
			if _, err := os.Stat(imagePath); os.IsNotExist(err) {
				return fmt.Errorf("image file not found: %s", imagePath)
			}
			
			// Get the absolute path
			absPath, err := filepath.Abs(imagePath)
			if err != nil {
				return fmt.Errorf("failed to get absolute path: %w", err)
			}
			
			fmt.Printf("Analyzing Docker image: %s\n", absPath)
			
			// Create a parser for the image
			parser, err := docker.NewParser(absPath)
			if err != nil {
				return fmt.Errorf("failed to create parser: %w", err)
			}
			defer parser.Cleanup()
			
			// Parse the image
			metadata, err := parser.Parse()
			if err != nil {
				return fmt.Errorf("failed to parse image: %w", err)
			}
			
			// Print image info if verbose
			if verbose {
				printImageInfo(metadata)
			}
			
			// Determine where to write the output
			var out *os.File
			if outputFile == "" {
				out = os.Stdout
			} else {
				out, err = os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("failed to create output file: %w", err)
				}
				defer out.Close()
			}
			
			// Generate the Dockerfile
			if strings.ToLower(outputFormat) == "dockerfile" {
				// Create a Dockerfile generator
				generator := dockerfile.NewGenerator(metadata)
				
				// Generate the Dockerfile
				if err := generator.Generate(out); err != nil {
					return fmt.Errorf("failed to generate Dockerfile: %w", err)
				}
				
				if outputFile != "" {
					fmt.Printf("Dockerfile written to: %s\n", outputFile)
				}
			} else if strings.ToLower(outputFormat) == "json" {
				// Output as JSON (for debugging or further processing)
				encoder := json.NewEncoder(out)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(metadata); err != nil {
					return fmt.Errorf("failed to encode metadata as JSON: %w", err)
				}
				
				if outputFile != "" {
					fmt.Printf("JSON metadata written to: %s\n", outputFile)
				}
			} else {
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}
			
			return nil
		},
	}
	
	// Add flags to the analyze command
	analyzeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file for the Dockerfile (default: stdout)")
	analyzeCmd.Flags().StringVarP(&outputFormat, "format", "f", "dockerfile", "Output format (dockerfile, json)")
	analyzeCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	
	return analyzeCmd
}

// printImageInfo prints basic information about the parsed image
func printImageInfo(metadata *docker.ImageMetadata) {
	fmt.Println("Image Information:")
	fmt.Println("==================")
	
	// Print repo tags
	if len(metadata.RepoTags) > 0 {
		fmt.Printf("Repository Tags: %s\n", strings.Join(metadata.RepoTags, ", "))
	}
	
	// Print image creation date
	fmt.Printf("Created: %s\n", metadata.Created.Format(time.RFC3339))
	
	// Print architecture and OS
	fmt.Printf("Architecture: %s, OS: %s\n", metadata.Architecture, metadata.OS)
	
	// Print exposed ports
	if len(metadata.Config.ExposedPorts) > 0 {
		var ports []string
		for port := range metadata.Config.ExposedPorts {
			ports = append(ports, port)
		}
		fmt.Printf("Exposed Ports: %s\n", strings.Join(ports, ", "))
	}
	
	// Print environment variables
	if len(metadata.Config.Env) > 0 {
		fmt.Println("\nEnvironment Variables:")
		for _, env := range metadata.Config.Env {
			fmt.Printf("  %s\n", env)
		}
	}
	
	// Print layers count
	fmt.Printf("\nLayers: %d\n", len(metadata.Layers))
	
	// Print history count
	fmt.Printf("History Entries: %d\n", len(metadata.History))
	fmt.Println("==================")
}