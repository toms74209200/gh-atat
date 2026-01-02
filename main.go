package main

import (
	"fmt"
	"os"

	"github.com/toms74209200/gh-atat/internal/run"
)

func main() {
	if err := run.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
