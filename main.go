package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Build information. Populated at build-time by GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "pasgan",
	Short: "Pasgan - Docker image to Dockerfile generator",
	Long: `Pasgan (Scots Gaelic for "package") is a tool that analyzes saved Docker images
and generates a Dockerfile that could have been used to create that image.

It works by analyzing the JSON metadata in Docker images and reconstructing the
build instructions.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	// Initialize commands
	initCommands()
	
	// Execute the root command
	execute()
}