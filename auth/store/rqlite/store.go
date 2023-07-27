package rqlite

import (
	"encoding/json"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/store"

	"github.com/rqlite/gorqlite"
)

var tableName = "code"

type Store struct {
	db *gorqlite.Connection
}

func New(db *gorqlite.Connection) store.Store {
	return &Store{db: db}
}

func (s *Store) SetAccessCode(code string, user *model.User) error {
	builder := squirrel.Insert(tableName).
		Columns("code", "user").
		Values(code, user.String())
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
		return fmt.Errorf("persistence/rqlite: failed to insert code to rqlite: %w", err)
	}
	return nil
}

func (s *Store) VerifyAccessCode(code string) (*model.User, error) {
	builder := squirrel.Select("user").
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

	if rows.NumRows() == 0 {
		return nil, store.ErrEmpty
	}

	var val string
	for rows.Next() {
		err = rows.Scan(&val)
		if err != nil {
			return nil, fmt.Errorf("persistence/rqlite: failed to scan row: %w", err)
		}
	}

	user := &model.User{}
	if err := json.Unmarshal([]byte(val), user); err != nil {
		return nil, err
	}

	return user, nil
}
