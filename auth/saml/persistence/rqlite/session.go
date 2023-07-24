package rqlite

import (
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/deepsourcecorp/runner/auth/saml"
	"github.com/rqlite/gorqlite"
	"golang.org/x/oauth2"
)

type SessionStore struct {
	db *gorqlite.Connection
}

func NewSessionStore(db *gorqlite.Connection) saml.SessionStore {
	return &SessionStore{db: db}
}

func (store *SessionStore) Create(s *saml.Session) error {
	builder := squirrel.Insert("saml_sessions").
		Columns("id",
			"access_code",
			"access_code_expiry",
			"backend_email",
			"backend_first_name",
			"backend_last_name",
			"backend_raw",
			"backend_expiry",
			"runner_access_token",
			"runner_refresh_token",
			"runner_token_expiry",
		).
		Values(s.ID,
			s.AccessCode,
			s.AccessCodeExpiry,
			s.BackendToken.Email,
			s.BackendToken.FirstName,
			s.BackendToken.LastName,
			s.BackendToken.Raw,
			s.BackendToken.Expiry,
			s.RunnerToken.AccessToken,
			s.RunnerToken.RefreshToken,
			s.RunnerToken.Expiry,
		)

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

func (store *SessionStore) Update(s *saml.Session) error {
	builder := squirrel.Update("saml_sessions").
		Set("access_code", s.AccessCode).
		Set("access_code_expiry", s.AccessCodeExpiry).
		Set("backend_email", s.BackendToken.Email).
		Set("backend_first_name", s.BackendToken.FirstName).
		Set("backend_last_name", s.BackendToken.LastName).
		Set("backend_raw", s.BackendToken.Raw).
		Set("backend_expiry", s.BackendToken.Expiry).
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
	return err
}

func (store *SessionStore) GetByID(id string) (*saml.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry",
		"backend_email", "backend_first_name", "backend_last_name", "backend_raw",
		"backend_expiry", "runner_access_token", "runner_refresh_token",
		"runner_token_expiry").
		From("saml_sessions").
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
	return parseRow(result)
}

func (store *SessionStore) GetByAccessCode(accessCode string) (*saml.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry",
		"backend_email", "backend_first_name", "backend_last_name", "backend_raw",
		"backend_expiry", "runner_access_token", "runner_refresh_token",
		"runner_token_expiry").
		From("saml_sessions").
		Where(squirrel.Eq{"access_code": accessCode})
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

func (store *SessionStore) GetByAccessToken(token string) (*saml.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry",
		"backend_email", "backend_first_name", "backend_last_name", "backend_raw",
		"backend_expiry", "runner_access_token", "runner_refresh_token",
		"runner_token_expiry").
		From("saml_sessions").
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

func (store *SessionStore) GetByRefreshToken(token string) (*saml.Session, error) {
	builder := squirrel.Select("id", "access_code", "access_code_expiry",
		"backend_email", "backend_first_name", "backend_last_name", "backend_raw",
		"backend_expiry", "runner_access_token", "runner_refresh_token",
		"runner_token_expiry").
		From("saml_sessions").
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

func (store *SessionStore) Delete(id string) error {
	builder := squirrel.Delete("saml_sessions").
		Where(squirrel.Eq{"id": id})
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

func parseRow(result gorqlite.QueryResult) (*saml.Session, error) {
	var (
		id                 string
		accessCode         string
		accessCodeExpiry   time.Time
		backendEmail       string
		backendFirstName   string
		backendLastName    string
		backendRaw         string
		backendExpiry      time.Time
		runnerAccessToken  string
		runnerRefreshToken string
		runnerTokenExpiry  time.Time
	)
	if result.NumRows() != 1 {
		return nil, nil
	}

	for result.Next() {
		err := result.Scan(&id,
			&accessCode,
			&accessCodeExpiry,
			&backendEmail,
			&backendFirstName,
			&backendLastName,
			&backendRaw,
			&backendExpiry,
			&runnerAccessToken,
			&runnerRefreshToken,
			&runnerTokenExpiry,
		)
		if err != nil {
			return nil, err
		}
	}
	return &saml.Session{
		ID:               id,
		AccessCode:       accessCode,
		AccessCodeExpiry: accessCodeExpiry,
		BackendToken: &saml.BackendToken{
			Email:     backendEmail,
			FirstName: backendFirstName,
			LastName:  backendLastName,
			Raw:       backendRaw,
			Expiry:    backendExpiry,
		},
		RunnerToken: &oauth2.Token{
			AccessToken:  runnerAccessToken,
			RefreshToken: runnerRefreshToken,
			Expiry:       runnerTokenExpiry,
		},
	}, nil
}
