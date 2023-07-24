package oauth

import (
	"errors"
	"time"

	"github.com/segmentio/ksuid"
	"golang.org/x/oauth2"
)

const (
	TokenTypeBearer = "Bearer"
)

var ErrNoSession = errors.New("no session found")

type Session struct {
	ID               string
	AccessCode       string
	AccessCodeExpiry time.Time
	BackendToken     *oauth2.Token
	RunnerToken      *oauth2.Token
}

func NewSession() *Session {
	return &Session{
		ID: ksuid.New().String(),
	}
}

func (s *Session) GenerateRunnerToken(expiry time.Time) {
	s.RunnerToken = &oauth2.Token{
		AccessToken:  ksuid.New().String(),
		RefreshToken: ksuid.New().String(),
		Expiry:       expiry,
		TokenType:    TokenTypeBearer,
	}
}

func (s *Session) SetAccessCode(code string) {
	s.AccessCode = code
	s.AccessCodeExpiry = time.Now().Add(5 * time.Minute)
}

func (s *Session) UnsetAccessCode() {
	s.AccessCode = ""
	s.AccessCodeExpiry = time.Now()
}

type SessionStore interface {
	Create(*Session) error
	Update(*Session) error
	Delete(string) error
	GetByID(string) (*Session, error)
	GetByAccessToken(string) (*Session, error)
	GetByRefreshToken(string) (*Session, error)
	GetByAccessCode(string) (*Session, error)
}
