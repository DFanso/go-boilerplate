package main

import (
	"os"

	"github.com/vidwadeseram/go-boilerplate/dummy-api/cmd/dummy-api/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
