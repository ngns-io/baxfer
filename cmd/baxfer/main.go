package main

import (
	"fmt"
	"os"

	"github.com/ngns-io/baxfer/internal/cli"
)

func main() {
	app := cli.NewApp()

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
