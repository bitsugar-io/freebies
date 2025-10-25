package main

import (
	"os"

	"github.com/retr0h/freebie/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
