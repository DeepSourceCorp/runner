package saml

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/segmentio/ksuid"
	"golang.org/x/oauth2"
)

const (
	TokenTypeBearer = "Bearer"
)

type BackendToken struct {
	Expiry    time.Time
	Email     string
	FirstName string
	LastName  string
	Raw       string
}

type Session struct {
	ID               string
	AccessCode       string
	AccessCodeExpiry time.Time
	BackendToken     *BackendToken
	RunnerToken      *oauth2.Token
}

func NewSession() *Session {
	return &Session{
		ID: ksuid.New().String(),
	}
}

func (s *Session) SetBackendToken(raw string) error {
	parser := new(jwt.Parser)
	token, _, err := parser.ParseUnverified(raw, jwt.MapClaims{})
	if err != nil {
		return err
	}
	claims := token.Claims.(jwt.MapClaims)

	exp := time.Unix(int64(claims["exp"].(float64)), 0)

	attr, ok := claims["attr"].(map[string]interface{})
	if !ok {
		return errors.New("missing attr claim")
	}

	firstName, err := extractAttr(attr, "first_name")
	if err != nil {
		return err
	}

	lastName, err := extractAttr(attr, "last_name")
	if err != nil {
		return err
	}

	email, err := extractAttr(attr, "email")
	if err != nil {
		return err
	}

	s.BackendToken = &BackendToken{
		Expiry:    exp,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Raw:       raw,
	}
	return nil
}

func extractAttr(attr map[string]interface{}, key string) (string, error) {
	if attr[key] == nil {
		return "", fmt.Errorf("missing %s claim", key)
	}
	v, ok := attr[key].([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid %s claim", key)
	}
	if len(v) == 0 {
		return "", fmt.Errorf("invalid %s claim", key)
	}

	value, ok := v[0].(string)
	if !ok {
		return "", fmt.Errorf("invalid %s claim", key)
	}
	return value, nil
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
