package main

import (
	"fmt"
	"os"

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
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
