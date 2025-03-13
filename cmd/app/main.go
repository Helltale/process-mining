package main

import (
	"os"

	"github.com/Helltale/process-mining/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
