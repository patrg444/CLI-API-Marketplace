package main

import (
	"fmt"
	"os"

	"github.com/api-direct/cli/cmd"
	"github.com/fatih/color"
)

func main() {
	// Set up color output
	color.NoColor = false

	// Execute the root command
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
