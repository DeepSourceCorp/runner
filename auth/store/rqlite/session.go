package rqlite

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/rqlite/gorqlite"
)

var tableName = "session"

type SessionStore struct {
	db *gorqlite.Connection
}

func NewSessionStore(db *gorqlite.Connection) session.Store {
	return &SessionStore{db: db}
}

func (s *SessionStore) Create(session *session.Session) error {
	builder := squirrel.Insert(tableName).
		Columns("id", "app_id", "code", "backend_token", "runner_access_token", "runner_token_expiry", "runner_refresh_token").
		Values(session.ID, session.AppID, session.Code, session.BackendToken, session.RunnerAccessToken, session.RunnerTokenExpiry, session.RunnerRefreshToken)
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("persistence/rqlite: failed to build query for insert: %w", err)
	}
	_, err = s.db.WriteOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		return fmt.Errorf("persistence/rqlite: failed to insert session to rqlite: %w", err)
	}
	return nil
}

func (s *SessionStore) Update(session *session.Session) error {
	builder := squirrel.Update(tableName).
		Set("app_id", session.AppID).
		Set("code", session.Code).
		Set("backend_token", session.BackendToken).
		Set("runner_access_token", session.RunnerAccessToken).
		Set("runner_access_token_expires_at", session.RunnerTokenExpiry).
		Set("runner_refresh_token", session.RunnerRefreshToken).
		Where(squirrel.Eq{"id": session.ID})
	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("persistence/rqlite: failed to build query for update: %w", err)
	}
	_, err = s.db.WriteOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		return fmt.Errorf("persistence/rqlite: failed to update code to rqlite: %w", err)
	}
	return nil
}

func (s *SessionStore) Get(id string) (*session.Session, error) {
	builder := squirrel.Select("id", "app_id", "code", "backend_token", "runner_access_token", "runner_token_expiry", "runner_refresh_token").
		From(tableName).
		Where(squirrel.Eq{"id": id})
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("persistence/rqlite: failed to build query for select: %w", err)
	}

	rows, err := s.db.QueryOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("persistence/rqlite: failed to query rqlite: %w", err)
	}
	return scan(rows)
}

func (s *SessionStore) GetByCode(code string) (*session.Session, error) {
	builder := squirrel.Select("id", "app_id", "code", "backend_token", "runner_access_token", "runner_token_expiry", "runner_refresh_token").
		From(tableName).
		Where(squirrel.Eq{"code": code})
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("persistence/rqlite: failed to build query for select: %w", err)
	}

	rows, err := s.db.QueryOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("persistence/rqlite: failed to query rqlite: %w", err)
	}
	return scan(rows)
}

func scan(rows gorqlite.QueryResult) (*session.Session, error) {
	var id string
	var appID string
	var code string
	var backendToken string
	var runnerAccessToken string
	var runnerTokenExpiry int64
	var runnerRefreshToken string

	for rows.Next() {
		err := rows.Scan(&id, &appID, &code, &backendToken, &runnerAccessToken, &runnerTokenExpiry, &runnerRefreshToken)
		if err != nil {
			return nil, fmt.Errorf("persistence/rqlite: failed to scan row: %w", err)
		}
	}
	session := &session.Session{
		ID:                 id,
		AppID:              appID,
		Code:               code,
		BackendToken:       backendToken,
		RunnerAccessToken:  runnerAccessToken,
		RunnerTokenExpiry:  runnerTokenExpiry,
		RunnerRefreshToken: runnerRefreshToken,
	}
	return session, nil
}
