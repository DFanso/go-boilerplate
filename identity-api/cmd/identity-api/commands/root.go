package commands

import (
	"context"

	"github.com/spf13/cobra"
)

// Execute is the CLI entry point.
func Execute() error {
	root := newRootCmd()
	root.SetContext(context.Background())
	return root.Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity-api",
		Short: "Identity microservice",
	}

	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newMigrateCmd())

	return cmd
}
