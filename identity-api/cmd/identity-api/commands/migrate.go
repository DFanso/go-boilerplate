package commands

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	"github.com/vidwadeseram/go-boilerplate/identity-api/internal/config"
)

func newMigrateCmd() *cobra.Command {
	var dir string
	var action string

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			path := fmt.Sprintf("file://%s", filepath.Clean(dir))
			m, err := migrate.New(path, cfg.DatabaseURL)
			if err != nil {
				return err
			}

			defer func() {
				_, _ = m.Close()
			}()

			switch action {
			case "up":
				err = m.Up()
			case "down":
				err = m.Steps(-1)
			case "drop":
				err = m.Drop()
			default:
				return fmt.Errorf("unknown action %s", action)
			}
			if err != nil && err != migrate.ErrNoChange {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "migrations", "path to migration files")
	cmd.Flags().StringVar(&action, "action", "up", "migration action: up, down, drop")

	return cmd
}
