package main

import (
	"fmt"
	"os"

	"github.com/litents/litents/internal/core"
)

func main() {
	app := core.NewApp(os.Stdout, os.Stderr)
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
