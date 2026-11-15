package main

import (
	"os"

	"github.com/eliel-elie/ai-commits/internal/cli"
	"github.com/eliel-elie/ai-commits/internal/ui"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		ui.PrintError(err.Error())
		os.Exit(1)
	}
}
