package store

import (
	"errors"

	"github.com/deepsourcecorp/runner/auth/model"
)

var ErrEmpty = errors.New("store: empty")

type Store interface {
	SetAccessCode(code string, user *model.User) error
	VerifyAccessCode(code string) (*model.User, error)
}
