package token

import (
	"errors"
	"strings"

	"time"

	"github.com/deepsourcecorp/runner/auth/cryptutil"
	"github.com/deepsourcecorp/runner/auth/model"
)

const (
	ScopeUser     = "user"
	ScopeCodeRead = "code:read"
	ScopeRefresh  = "refresh"
)

var (
	ExpiryAccessToken  = 15 * time.Minute
	ExpiryRefreshToken = 15 * 24 * time.Hour
)

type Service struct {
	signer   *cryptutil.Signer
	verifier *cryptutil.Verifier
}

func NewService(signer *cryptutil.Signer, verifier *cryptutil.Verifier) *Service {
	return &Service{
		signer:   signer,
		verifier: verifier,
	}
}

func (s *Service) GenerateToken(issuer string, scopes []string, user *model.User, expiry time.Duration) (string, error) {
	return s.signer.GenerateToken(issuer, scopes, user.Claims(), expiry)
}

func (s *Service) ReadToken(issuer string, scope string, token string) (*model.User, error) {
	claims, err := s.verifier.Verify(token)
	if err != nil {
		return nil, err
	}
	for _, v := range []string{"id", "name", "email", "login", "provider"} {
		if _, ok := claims[v]; !ok {
			return nil, errors.New("invalid claims")
		}
	}
	if claims["iss"] != issuer {
		return nil, errors.New("invalid issuer")
	}

	scopeParts := strings.Split(claims["scp"].(string), " ")
	scopeValid := false
	for _, s := range scopeParts {
		if s == scope {
			scopeValid = true
		}
	}
	if !scopeValid {
		return nil, errors.New("invalid scope")
	}

	return &model.User{
		ID:       claims["id"].(string),
		Name:     claims["name"].(string),
		Email:    claims["email"].(string),
		Login:    claims["login"].(string),
		Provider: claims["provider"].(string),
	}, nil
}
