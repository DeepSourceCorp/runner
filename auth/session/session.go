package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/segmentio/ksuid"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

const (
	ScopeCode    = "code"
	ScopeRefresh = "refresh"

	ExpiryRunnerAccessToken  = time.Minute * 5
	ExpiryRunnerRefreshToken = time.Hour * 24 * 30
)

type Session struct {
	ID           string
	AppID        string
	Code         string
	BackendToken string

	OAuthAccessToken  string
	OAuthRefreshToken string
	OAuthTokenExpiry  int64

	RunnerAccessToken  string
	RunnerRefreshToken string
	RunnerTokenExpiry  int64
}

func NewSession() *Session {
	return &Session{
		ID: ksuid.New().String(),
	}
}

func (s *Session) SetBackendToken(token interface{}) error {
	switch t := token.(type) {
	case *oauth2.Token:
		s.OAuthTokenExpiry = int64(time.Until(t.Expiry).Seconds())
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

func (s *Session) GetBackendToken(ctx context.Context, v interface{}) error {
	switch t := v.(type) {
	case *oauth2.Token:
		err := json.Unmarshal([]byte(s.BackendToken), t)
		if err != nil {
			common.Log(ctx, slog.LevelError, "failed to unmarshal token", slog.Any("err", err))
			return fmt.Errorf("failed to unmarshal token: %w", err)
		}
		return nil
	default:
		common.Log(ctx, slog.LevelError, "unknown backend type")
		return fmt.Errorf("unknown backend type: %s", t)
	}
}

func (s *Session) SetRunnerToken(r *common.Runner) error {
	accessToken, err := r.IssueToken(
		ScopeCode, map[string]interface{}{
			ClaimSessionID: s.ID,
		}, ExpiryRunnerAccessToken)
	if err != nil {
		return fmt.Errorf("failed to issue runner access token: %w", err)
	}
	refreshToken, err := r.IssueToken(
		ScopeRefresh, map[string]interface{}{
			ClaimSessionID: s.ID,
		}, ExpiryRunnerRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to issue runner refresh token: %w", err)
	}

	s.RunnerAccessToken = accessToken
	s.RunnerRefreshToken = refreshToken
	s.RunnerTokenExpiry = time.Now().Add(ExpiryRunnerAccessToken).Unix()
	return nil
}
