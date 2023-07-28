package rqlite

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/store"
	"github.com/rqlite/gorqlite"
)

var s store.Store

func TestMain(m *testing.M) {
	if os.Getenv("TEST_ENV") != "integration" {
		os.Exit(0)
	}
	tableName = "testcode"
	db, err := gorqlite.Open("http://localhost:4001/?disableClusterDiscovery=true")
	if err != nil {
		fmt.Printf("failed to initialize tests for persistence/rqlite: %v", err)
		os.Exit(1)
	}
	createTable := `CREATE TABLE IF NOT EXISTS testcode (code TEXT PRIMARY KEY, user TEXT) WITHOUT ROWID;`
	_, err = db.Write([]string{createTable})
	if err != nil {
		fmt.Printf("failed to initialize tests for persistence/rqlite: %v", err)
		os.Exit(1)
	}

	s = New(db)
	code := m.Run()

	_, err = db.WriteOne("DROP TABLE testcode")
	if err != nil {
		fmt.Printf("failed to cleanup after tests for persistence/rqlite: %v", err)
		os.Exit(1)
	}
	db.Close()
	os.Exit(code)
}

func TestStore_SetAccessCode(t *testing.T) {
	user := &model.User{
		ID:       "test-id",
		Name:     "test-name",
		Email:    "abc@xyz.com",
		Login:    "test-login",
		Provider: "test-provider",
	}

	err := s.SetAccessCode("test-code", user)
	if err != nil {
		t.Error("failed to set access code", err)
	}

	got, err := s.VerifyAccessCode("test-code")
	if err != nil {
		t.Error("failed to verify access code", err)
	}

	if !reflect.DeepEqual(got, user) {
		t.Errorf("got %v, want %v", got, user)
	}
}
