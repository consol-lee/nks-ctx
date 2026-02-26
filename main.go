package main

import (
	"os"

	"github.com/consol-lee/nks-ctx/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
