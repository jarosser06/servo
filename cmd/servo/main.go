package main

import (
	"fmt"
	"os"

	"github.com/servo/servo/internal/cli"
)

var version = "dev" // Set via ldflags during build

func main() {
	app, err := cli.NewApp(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing servo: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
