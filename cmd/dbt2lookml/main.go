package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/cli"
)

const banner = `
    _ _   ___ _         _         _
  _| | |_|_  | |___ ___| |_ _____| |
 | . | . |  _| | . | . | '_|     | |
 |___|___|___|_|___|___|_,_|_|_|_|_|
    Convert your dbt models to LookML views
`

func main() {
	fmt.Print(banner)

	if err := cli.Execute(); err != nil {
		// Print error with visual emphasis
		fmt.Fprintf(os.Stderr, "\n%s\n", strings.Repeat("═", 80))
		fmt.Fprintf(os.Stderr, "\033[1;31m✗ ERROR:\033[0m %v\n", err)
		fmt.Fprintf(os.Stderr, "%s\n\n", strings.Repeat("═", 80))
		os.Exit(1)
	}
}
