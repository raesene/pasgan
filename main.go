package main

import (
	"github.com/raesene/pasgan/cmd/pasgan"
)

// Build information. Populated at build-time by GoReleaser.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Pass version information to the cmd package
	pasgan.Execute(version, commit, date)
}