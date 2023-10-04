package session

import (
	"fmt"
	"strings"
	"time"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/model"
	"github.com/segmentio/ksuid"
	"golang.org/x/oauth2"
)

const (
	ScopeCode    = "code"
	ScopeRefresh = "refresh"
	ScopeUser    = "user"

	ClaimSessionID = "session_id"
	ClaimScope     = "scope"

	TokenTypeBearer = "bearer"
)

var (
	RefreshTokenExpiry = time.Hour * 24 * 30
)

type Service struct {
	store Store

	runner *model.Runner
}

func NewService(runner *model.Runner, store Store) *Service {
	return &Service{
		runner: runner,
		store:  store,
	}
}

func (s *Service) SetBackendToken(session *Session, token interface{}, expiry time.Time) error {
	if err := session.SetBackendToken(token); err != nil {
		return fmt.Errorf("failed to set backend token, %w", err)
	}
	session.ExpiresAt = expiry.Unix()

	if err := s.store.Create(session); err != nil {
		return fmt.Errorf("failed to create session, %w", err)
	}
	return nil
}

// GenerateTokenPair generates a session token and a refresh token for the given
// session.
func (s *Service) GenerateSessionTokens(session *Session) (string, string, error) {
	expiry := time.Duration(session.ExpiresAt-time.Now().Unix()) * time.Second
	signer := s.runner.Signer()

	sessionToken, err := signer.GenerateToken(
		s.runner.ID, []string{ScopeCode}, map[string]interface{}{
			ClaimSessionID: session.ID,
		}, expiry)

	if err != nil {
		return "", "", fmt.Errorf("failed to generate session token, %w", err)
	}

	refreshToken, err := signer.GenerateToken(
		s.runner.ID, []string{ScopeCode}, map[string]interface{}{
			ClaimSessionID: session.ID,
		}, RefreshTokenExpiry)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token, %w", err)
	}

	return sessionToken, refreshToken, nil
}

func (s *Service) ParseJWT(token string, scope string) (*Session, error) {
	verifier := s.runner.Verifier()
	claims, err := verifier.Verify(token)
	if err != nil {
		err := fmt.Errorf("failed to verify token, %w", err)
		return nil, httperror.ErrUnauthorized(err)
	}

	sessionID, ok := claims[ClaimSessionID].(string)
	if !ok || sessionID == "" {
		err := fmt.Errorf("invalid token")
		return nil, httperror.ErrUnauthorized(err)
	}

	session, err := s.store.Filter(&Filter{ID: sessionID})
	if err != nil {
		err := fmt.Errorf("failed to get session, %w", err)
		return nil, httperror.ErrUnknown(err)
	}

	if scope != "" {
		scopeClaim := claims[ClaimScope].(string)
		if scopeClaim == "" {
			err := fmt.Errorf("invalid token")
			return nil, httperror.ErrUnauthorized(err)
		}

		scopes := strings.Split(scopeClaim, " ")
		for _, s := range scopes {
			if s == scope {
				return session, nil
			}
		}
		return nil, httperror.ErrUnauthorized(fmt.Errorf("invalid scope"))
	}
	return session, nil
}

func (s *Service) GenerateAccessCode(session *Session) (string, error) {
	code := ksuid.New().String()
	session.Code = code
	if err := s.store.Update(session); err != nil {
		err := fmt.Errorf("failed to update session, %w", err)
		return "", httperror.ErrUnknown(err)
	}
	return code, nil
}

func (s *Service) GetByAccessCode(code string) (*Session, error) {
	session, err := s.store.Filter(&Filter{Code: code})
	if err != nil {
		err := fmt.Errorf("failed to get session, %w", err)
		return nil, httperror.ErrUnknown(err)
	}
	return session, nil
}

// ------------------------------OAuthToken------------------------------------

func (s *Service) GenerateOauthToken(session *Session) (*oauth2.Token, error) {
	signer := s.runner.Signer()

	accessToken, err := signer.GenerateToken(
		s.runner.ID, []string{ScopeUser}, map[string]interface{}{
			ClaimSessionID: session.ID,
		}, time.Duration(session.ExpiresAt-time.Now().Unix())*time.Second)
	if err != nil {
		err := fmt.Errorf("failed to generate access token, %w", err)
		return nil, httperror.ErrUnknown(err)
	}

	refreshToken, err := signer.GenerateToken(
		s.runner.ID, []string{ScopeRefresh}, map[string]interface{}{
			ClaimSessionID: session.ID,
		}, RefreshTokenExpiry)
	if err != nil {
		err := fmt.Errorf("failed to generate refresh token, %w", err)
		return nil, httperror.ErrUnknown(err)
	}

	token := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    TokenTypeBearer,
		Expiry:       time.Unix(session.ExpiresAt, 0),
	}

	if err := session.SetBackendToken(token); err != nil {
		err := fmt.Errorf("failed to set backend token, %w", err)
		return nil, httperror.ErrUnknown(err)
	}

	if err := s.store.Update(session); err != nil {
		err := fmt.Errorf("failed to update session, %w", err)
		return nil, httperror.ErrUnknown(err)
	}
	return token, nil
}
