package main

import (
	"fmt"

	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/rqlite"
	"github.com/deepsourcecorp/runner/rqlite/migrations"
)

func Migrate(c *config.RQLite) error {
	db, err := rqlite.Connect(c.Host, c.Port)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	migrator, err := migrations.NewMigrator(db)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	err = migrator.Migrate()
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
