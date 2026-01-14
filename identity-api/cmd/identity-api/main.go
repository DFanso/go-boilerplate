package main

import (
	"os"

	"github.com/vidwadeseram/go-boilerplate/identity-api/cmd/identity-api/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
