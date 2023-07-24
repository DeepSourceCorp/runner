package rqlite

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/rqlite/gorqlite"
	"golang.org/x/oauth2"
)

var store oauth.SessionStore

func TestMain(m *testing.M) {
	db, err := gorqlite.Open(fmt.Sprintf("http://%s:%d/?disableClusterDiscovery=true", "localhost", 4001))
	if err != nil {
		panic(err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS oauth_sessions (
		id VARCHAR(27) PRIMARY KEY, 
		access_code VARCHAR(255), 
		access_code_expiry DATETIME,
		backend_access_token VARCHAR(255), 
		backend_refresh_token VARCHAR(255), 
		backend_token_expiry DATETIME, 
		runner_access_token VARCHAR(255), 
		runner_refresh_token VARCHAR(255), 
		runner_token_expiry DATETIME
		) WITHOUT ROWID;
		`
	_, err = db.Write([]string{createTable})
	if err != nil {
		log.Println("Error creating table:", err)
		os.Exit(0)
	}

	store = NewSessionStore(db)
	code := m.Run()

	_, err = db.WriteOne("DROP TABLE oauth_sessions")
	if err != nil {
		log.Println("Error dropping table:", err)
		os.Exit(0)
	}

	db.Close()
	os.Exit(code)
}

func TestCreate(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	input := &oauth.Session{
		ID:               "123",
		AccessCode:       "abc",
		AccessCodeExpiry: now,
		BackendToken: &oauth2.Token{
			AccessToken:  "backedn-access-token",
			RefreshToken: "backend-refresh-token",
			Expiry:       now,
		},
		RunnerToken: &oauth2.Token{
			AccessToken:  "runner-access-token",
			RefreshToken: "runner-refresh-token",
			Expiry:       now,
		},
	}

	err := store.Create(input)
	if err != nil {
		t.Error("rqlite.SessionStore.Create() failed:", err)
		return
	}

	output, err := store.GetByID(input.ID)
	if err != nil {
		t.Error("rqlite.SessionStore.Create() failed:", err)
		return
	}

	if !reflect.DeepEqual(input, output) {
		t.Errorf("rqlite.SessionStore.Create() failed: expected %v, got %v", input, output)
		return
	}
}

func TestUpdate(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	input := &oauth.Session{
		ID:               "234",
		AccessCode:       "abc",
		AccessCodeExpiry: now,
		BackendToken: &oauth2.Token{
			AccessToken:  "backedn-access-token",
			RefreshToken: "backend-refresh-token",
			Expiry:       now,
		},
		RunnerToken: &oauth2.Token{
			AccessToken:  "runner-access-token",
			RefreshToken: "runner-refresh-token",
			Expiry:       now,
		},
	}
	update := &oauth.Session{
		ID:               "234",
		AccessCode:       "xtz",
		AccessCodeExpiry: now,
		BackendToken: &oauth2.Token{
			AccessToken:  "updated-backedn-access-token",
			RefreshToken: "updated-backend-refresh-token",
			Expiry:       now,
		},
		RunnerToken: &oauth2.Token{
			AccessToken:  "updated-runner-access-token",
			RefreshToken: "updated-runner-refresh-token",
			Expiry:       now,
		},
	}

	err := store.Create(input)
	if err != nil {
		t.Error("rqlite.SessionStore.Update() failed:", err)
		return
	}

	err = store.Update(update)
	if err != nil {
		t.Error("rqlite.SessionStore.Update() failed:", err)
		return
	}

	output, err := store.GetByID(input.ID)
	if err != nil {
		t.Error("rqlite.SessionStore.Update() failed:", err)
		return
	}

	if !reflect.DeepEqual(update, output) {
		t.Errorf("rqlite.SessionStore.Update() failed: expected %v, got %v", update, output)
		return
	}
}

func TestDelete(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	input := &oauth.Session{
		ID:               "456",
		AccessCode:       "abc",
		AccessCodeExpiry: now,
		BackendToken: &oauth2.Token{
			AccessToken:  "backend-access-token",
			RefreshToken: "backend-refresh-token",
			Expiry:       now,
		},
		RunnerToken: &oauth2.Token{
			AccessToken:  "runner-access-token",
			RefreshToken: "runner-refresh-token",
			Expiry:       now,
		},
	}

	err := store.Create(input)
	if err != nil {
		t.Error("rqlite.SessionStore.Delete() failed:", err)
		return
	}

	err = store.Delete(input.ID)
	if err != nil {
		t.Error("rqlite.SessionStore.Delete() failed:", err)
		return
	}

	_, err = store.GetByID(input.ID)
	if err != oauth.ErrNoSession {
		t.Errorf("rqlite.SessionStore.Delete() failed: expected %v, got %v", oauth.ErrNoSession, err)
		return
	}
}
