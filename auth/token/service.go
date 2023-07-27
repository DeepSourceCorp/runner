package token

import (
	"crypto/rsa"
	"errors"

	"time"

	"github.com/deepsourcecorp/runner/auth/jwtutil"
	"github.com/deepsourcecorp/runner/auth/model"
)

const (
	ScopeCodeRead = "code:read"
	ScopeRefresh  = "refresh"
)

var (
	ExpiryAccessToken = 15 * time.Minute
)

type Service struct {
	signer   *jwtutil.Signer
	verifier *jwtutil.Verifier
}

func NewService(runnerID string, privateKey *rsa.PrivateKey) *Service {
	return &Service{
		signer:   jwtutil.NewSigner(runnerID, privateKey),
		verifier: jwtutil.NewVerifier(runnerID, &privateKey.PublicKey),
	}
}

func (s *Service) GetAccessToken(user *model.User) (string, error) {
	claims := user.ToMap()
	return s.signer.GenerateToken([]string{ScopeCodeRead}, claims, ExpiryAccessToken)
}

func (s *Service) GetRefreshToken(user *model.User) (string, error) {
	claims := user.ToMap()
	return s.signer.GenerateToken([]string{ScopeRefresh}, claims, ExpiryAccessToken)
}

func (s *Service) ReadAccessToken(token string) (*model.User, error) {
	claims, err := s.verifier.Verify(token)
	if err != nil {
		return nil, err
	}
	for _, v := range []string{"id", "name", "email", "login", "provider"} {
		if _, ok := claims[v]; !ok {
			return nil, errors.New("invalid claims")
		}
	}

	if claims["iss"] != s.signer.Issuer {
		return nil, errors.New("invalid issuer")
	}

	if claims["scp"] != ScopeCodeRead {
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

func (s *Service) ReadRefreshToken(token string) (*model.User, error) {
	claims, err := s.verifier.Verify(token)
	if err != nil {
		return nil, err
	}
	for _, v := range []string{"id", "name", "email", "login", "provider"} {
		if _, ok := claims[v]; !ok {
			return nil, errors.New("invalid claims")
		}
	}
	if claims["iss"] != s.signer.Issuer {
		return nil, errors.New("invalid issuer")
	}
	if claims["scp"] != ScopeRefresh {
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
