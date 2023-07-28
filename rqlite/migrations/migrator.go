package migrations

import (
	"time"

	"github.com/rqlite/gorqlite"
	"golang.org/x/exp/slog"
)

type Migration struct {
	Name string
	Up   string
	Down string
}

type Migrator struct {
	db *gorqlite.Connection
}

var migrations = []Migration{
	{
		Name: "001",
		Up:   Up001,
		Down: Down001,
	},
}

func NewMigrator(db *gorqlite.Connection) (*Migrator, error) {
	m := &Migrator{db: db}
	if err := m.migrateRoot(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Migrator) Migrate() error {
	applied, err := m.appliedMigrations()
	if err != nil {
		return err
	}

	// Create a set to track applied migrations
	appliedSet := make(map[string]bool)
	for _, a := range applied {
		appliedSet[a] = true
	}

	for _, migration := range migrations {
		if appliedSet[migration.Name] {
			continue
		}
		slog.Info("applying migration", slog.Any("name", migration.Name))

		_, err := m.db.WriteOne(migration.Up)
		if err != nil {
			return err
		}

		err = m.insertAppliedMigration(migration.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Migrator) migrateRoot() error {
	_, err := m.db.WriteOne(Up000)
	return err
}

func (m *Migrator) appliedMigrations() ([]string, error) {
	query := "SELECT name FROM migrations;"
	result, err := m.db.QueryOne(query)
	if err != nil {
		return nil, err
	}
	var names []string
	for result.Next() {
		var name string
		err := result.Scan(&name)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (m *Migrator) insertAppliedMigration(name string) error {
	query := "INSERT INTO migrations (name, time) VALUES (?, ?);"
	_, err := m.db.WriteOneParameterized(gorqlite.ParameterizedStatement{
		Query:     query,
		Arguments: []interface{}{name, time.Now()},
	})
	return err
}
