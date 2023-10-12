package session

import (
	"fmt"
	"net/url"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/jwtutil"
	"github.com/segmentio/ksuid"
	"golang.org/x/exp/slog"
)

const (
	ClaimSessionID = "session_id"
	ClaimScope     = "scp"

	// BifrostCallbackURLFmt is the callback path for DeepSource Cloud.
	BifrostCallbackURLFmt = "/accounts/runner/apps/%s/login/callback/bifrost/"
)

type Service struct {
	*common.Runner
	*common.DeepSource
	sessionStore Store
}

func NewService(runner *common.Runner, deepsource *common.DeepSource, sessionStore Store) *Service {
	return &Service{
		Runner:       runner,
		DeepSource:   deepsource,
		sessionStore: sessionStore,
	}
}

func (s *Service) CreateSession(token interface{}) (*Session, error) {
	session := NewSession()
	if err := session.SetBackendToken(token); err != nil {
		return nil, fmt.Errorf("failed to set backend token, %w", err)
	}

	if err := session.SetRunnerToken(s.Runner); err != nil {
		return nil, fmt.Errorf("failed to set runner token, %w", err)
	}

	if err := s.sessionStore.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session, %w", err)
	}

	return session, nil
}

func (s *Service) GenerateAccessCode(session *Session) (*Session, error) {
	session.Code = ksuid.New().String()
	if err := s.sessionStore.Update(session); err != nil {
		err := fmt.Errorf("failed to update session, %w", err)
		return nil, httperror.ErrUnknown(err)
	}
	return session, nil
}

func (s *Service) GenerateOAuthToken(code string) (*Session, error) {
	session, err := s.sessionStore.GetByCode(code)
	if err != nil {
		return nil, err
	}
	if err := session.SetRunnerToken(s.Runner); err != nil {
		return nil, err
	}

	if err := s.sessionStore.Update(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) RefreshOAuthToken(session *Session) (*Session, error) {
	if err := session.SetRunnerToken(s.Runner); err != nil {
		return nil, err
	}

	if err := s.sessionStore.Update(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Service) DeepSourceCallBackURL(appID string, q url.Values) string {
	path := fmt.Sprintf(BifrostCallbackURLFmt, appID)
	u := (*s.DeepSource.BaseURL).JoinPath(path)
	u.RawQuery = q.Encode()
	return u.String()
}

// FetchSessionByJWT parses the JWT token and retrieves the session from the
// underlying store.  It errors if the scope is invalid.
func (s *Service) FetchSessionByJWT(token string, expectedScope string) (*Session, error) {
	claims, err := s.ParseToken(token)
	if err != nil {
		return nil, httperror.ErrUnauthorized(err)
	}

	if !jwtutil.IsValidScope(claims, expectedScope) {
		err := fmt.Errorf("invalid scope: %s", expectedScope)
		return nil, httperror.ErrUnauthorized(err)
	}

	sessionID, ok := claims[ClaimSessionID].(string)
	if !ok || sessionID == "" {
		err := fmt.Errorf("session id missing in jwt claim")
		return nil, httperror.ErrUnauthorized(err)
	}

	session, err := s.sessionStore.Get(sessionID)
	if err != nil {
		slog.Error("failed to fetch session", slog.Any("err", err))
		return nil, httperror.ErrUnknown(err)
	}

	return session, nil
}
