package session

import (
	"encoding/json"
	"fmt"

	"github.com/segmentio/ksuid"
	"golang.org/x/oauth2"
)

const (
	BackendTypeOauth2 = "oauth2"
	BackendTypeSAML   = "saml"
)

type Session struct {
	ID           string
	Code         string
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	BackendToken string
}

func New() *Session {
	return &Session{
		ID: ksuid.New().String(),
	}
}

func (s *Session) SetBackendToken(token interface{}) error {
	switch t := token.(type) {
	case *oauth2.Token:
		raw, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("failed to marshal token: %w", err)
		}
		s.BackendToken = string(raw)
	default:
		return fmt.Errorf("unknown backend type: %s", t)
	}
	return nil
}

func (s *Session) DeserializeBackendToken(v interface{}) error {
	switch t := v.(type) {
	case *oauth2.Token:
		return json.Unmarshal([]byte(s.BackendToken), t)
	default:
		return fmt.Errorf("unknown backend type: %s", t)
	}
}
