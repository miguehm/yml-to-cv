// Command cvyml renders a YAML CV into a Harvard-style PDF.
//
// Install with:
//
//	go install github.com/miguehm/yml-to-cv/cmd/cvyml@latest
package main

import (
	"os"

	"github.com/miguehm/yml-to-cv/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
