package commands

import (
	"context"

	"github.com/spf13/cobra"
)

// Execute runs the CLI.
func Execute() error {
	root := newRootCmd()
	root.SetContext(context.Background())
	return root.Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dummy-api",
		Short: "Dummy CRUD microservice",
	}

	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newMigrateCmd())

	return cmd
}
