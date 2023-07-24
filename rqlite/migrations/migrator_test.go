package migrations

import (
	"fmt"
	"testing"

	"github.com/rqlite/gorqlite"
)

func TestMigrator(t *testing.T) {
	db, err := gorqlite.Open(fmt.Sprintf("http://%s:%d/?disableClusterDiscovery=true", "localhost", 4001))
	if err != nil {
		t.Fatal("failed to connect to rqlite", err)
	}

	migrator, err := NewMigrator(db)
	if err != nil {
		t.Fatal("failed to initialize migrator", err)
	}

	err = migrator.Migrate()
	if err != nil {
		t.Error("failed to migrate", err)
	}
}
