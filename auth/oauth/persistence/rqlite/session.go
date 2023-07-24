package rqlite

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/rqlite/gorqlite"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

const (
	tableName = "oauth_sessions"
)

type SessionStore struct {
	db *gorqlite.Connection
}

func NewSessionStore(db *gorqlite.Connection) *SessionStore {
	return &SessionStore{db: db}
}

func (store *SessionStore) Create(s *oauth.Session) error {
	if s.BackendToken == nil {
		s.BackendToken = &oauth2.Token{}
	}
	if s.RunnerToken == nil {
		s.RunnerToken = &oauth2.Token{}
	}
	builder := squirrel.Insert(tableName).
		Columns("id", "access_code", "access_code_expiry", "backend_access_token", "backend_refresh_token", "backend_token_expiry", "runner_access_token", "runner_refresh_token", "runner_token_expiry").
		Values(s.ID, s.AccessCode, s.AccessCodeExpiry, s.BackendToken.AccessToken, s.BackendToken.RefreshToken, s.BackendToken.Expiry, s.RunnerToken.AccessToken, s.RunnerToken.RefreshToken, s.RunnerToken.Expiry)
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = store.db.WriteOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		slog.Info("failed to insert session to rqlite", slog.Any("err", err))
	}
	return err
}

func (store *SessionStore) Update(s *oauth.Session) error {
	builder := squirrel.Update(tableName).
		Set("access_code", s.AccessCode).
		Set("access_code_expiry", s.AccessCodeExpiry).
		Set("backend_access_token", s.BackendToken.AccessToken).
		Set("backend_refresh_token", s.BackendToken.RefreshToken).
		Set("backend_token_expiry", s.BackendToken.Expiry).
		Set("runner_access_token", s.RunnerToken.AccessToken).
		Set("runner_refresh_token", s.RunnerToken.RefreshToken).
		Set("runner_token_expiry", s.RunnerToken.Expiry).
		Where(squirrel.Eq{"id": s.ID})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = store.db.WriteOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		slog.Info("failed to insert session to rqlite", slog.Any("err", err))
	}
	return err
}

func (store *SessionStore) GetByID(id string) (*oauth.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry", "backend_access_token", "backend_refresh_token", "backend_token_expiry", "runner_access_token", "runner_refresh_token", "runner_token_expiry").
		From("oauth_sessions").
		Where(squirrel.Eq{"id": id})
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := store.db.QueryOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		return nil, err
	}
	if err != nil {
		slog.Info("failed to insert session to rqlite", slog.Any("err", err))
	}
	return parseRow(result)
}

func (store *SessionStore) GetByAccessToken(token string) (*oauth.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry", "backend_access_token", "backend_refresh_token", "backend_token_expiry", "runner_access_token", "runner_refresh_token", "runner_token_expiry").
		From(tableName).
		Where(squirrel.Eq{"runner_access_token": token})
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := store.db.QueryOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	if err != nil {
		return nil, err
	}
	return parseRow(result)
}

func (store *SessionStore) GetByRefreshToken(token string) (*oauth.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry", "backend_access_token", "backend_refresh_token", "backend_token_expiry", "runner_access_token", "runner_refresh_token", "runner_token_expiry").
		From(tableName).
		Where(squirrel.Eq{"runner_refresh_token": token})
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := store.db.QueryOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)

	if err != nil {
		return nil, err
	}

	return parseRow(result)
}

func (store *SessionStore) GetByAccessCode(code string) (*oauth.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry", "backend_access_token", "backend_refresh_token", "backend_token_expiry", "runner_access_token", "runner_refresh_token", "runner_token_expiry").
		From(tableName).
		Where(squirrel.Eq{"access_code": code})
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := store.db.QueryOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)

	if err != nil {
		return nil, err
	}

	return parseRow(result)
}

func (store *SessionStore) Delete(id string) error {
	builder := squirrel.Delete(tableName).Where(squirrel.Eq{"id": id})
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = store.db.WriteOneParameterized(
		gorqlite.ParameterizedStatement{
			Query:     query,
			Arguments: args,
		},
	)
	return err
}

func (store *SessionStore) IsValidSession(id string) bool {
	_, err := store.GetByID(id)
	return err == nil
}

func parseRow(result gorqlite.QueryResult) (*oauth.Session, error) {
	var (
		id                  string
		accessCode          string
		accessCodeExpiry    time.Time
		backendAccessToken  string
		backendRefreshToken string
		backendTokenExpiry  time.Time
		runnerAccessToken   string
		runnerRefreshToken  string
		runnerTokenExpiry   time.Time
	)

	if result.NumRows() == 0 {
		return nil, oauth.ErrNoSession
	}

	for result.Next() {
		err := result.Scan(&id, &accessCode, &accessCodeExpiry, &backendAccessToken, &backendRefreshToken, &backendTokenExpiry, &runnerAccessToken, &runnerRefreshToken, &runnerTokenExpiry)
		if err != nil {
			return nil, err
		}
	}
	return &oauth.Session{
		ID:               id,
		AccessCode:       accessCode,
		AccessCodeExpiry: accessCodeExpiry,
		BackendToken: &oauth2.Token{
			AccessToken:  backendAccessToken,
			RefreshToken: backendRefreshToken,
			Expiry:       backendTokenExpiry,
		},
		RunnerToken: &oauth2.Token{
			AccessToken:  runnerAccessToken,
			RefreshToken: runnerRefreshToken,
			Expiry:       runnerTokenExpiry,
		},
	}, nil
}
