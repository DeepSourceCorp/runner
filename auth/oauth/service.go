package oauth

import (
	"context"
	"net/url"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/deepsourcecorp/runner/auth/contract"
	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/deepsourcecorp/runner/httperror"
	"golang.org/x/exp/slog"
	"golang.org/x/oauth2"
)

type Service struct {
	apps                  Apps
	sessionService        *session.Service
	DeepSourceCallbackURL func(string, url.Values) string
}

func NewService(apps map[string]*App, sessionService *session.Service) *Service {
	return &Service{
		apps:           apps,
		sessionService: sessionService,

		DeepSourceCallbackURL: sessionService.DeepSourceCallBackURL,
	}
}

func (s *Service) GetAuthorizationURL(req *contract.AuthorizationRequest) (string, error) {
	if err := s.sessionService.ValidateClientID(req.ClientID); err != nil {
		return "", httperror.ErrBadRequest(err)
	}

	provider, err := s.apps.GetProvider(req.AppID)
	if err != nil {
		return "", err
	}

	return provider.AuthorizationURL(req.State, req.Scopes), nil
}

func (s *Service) CreateSession(req *CallbackRequest) (*session.Session, error) {
	provider, err := s.apps.GetProvider(req.AppID)
	if err != nil {
		return nil, err
	}

	token, err := provider.GetToken(req.Ctx, req.Code)
	if err != nil {
		return nil, httperror.ErrUnknown(err) // TODO: Handle upstream error types.
	}

	session, err := s.sessionService.CreateSession(token)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) GenerateAccessCode(req *SessionRequest) (*session.Session, error) {
	session, err := s.sessionService.FetchSessionByJWT(req.SessionToken, session.ScopeCode)
	if err != nil {
		return nil, err
	}

	session, err = s.sessionService.GenerateAccessCode(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) GenerateOAuthToken(req *TokenRequest) (*session.Session, error) {
	if err := s.sessionService.ValidateClient(req.ClientID, req.ClientSecret); err != nil {
		return nil, httperror.ErrUnauthorized(err)
	}
	session, err := s.sessionService.GenerateOAuthToken(req.Code)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) RefreshOAuthToken(req *contract.RefreshRequest) (*session.Session, error) {
	if err := s.sessionService.ValidateClient(req.ClientID, req.ClientSecret); err != nil {
		return nil, httperror.ErrUnauthorized(err)
	}

	session, err := s.sessionService.FetchSessionByJWT(req.RefreshToken, session.ScopeRefresh)
	if err != nil {
		return nil, err
	}

	backendToken, err := s.refreshBackendToken(req.Ctx, req.AppID, session)
	if err != nil {
		return nil, httperror.ErrUnknown(err)
	}

	if err := session.SetBackendToken(backendToken); err != nil {
		return nil, httperror.ErrUnknown(err)
	}

	session, err = s.sessionService.RefreshOAuthToken(session)
	if err != nil {
		return nil, httperror.ErrUnknown(err)
	}

	return session, nil
}

func (s *Service) GetUser(req *contract.UserRequest) (*common.User, error) {
	session, err := s.sessionService.FetchSessionByJWT(req.AccessToken, session.ScopeCode)
	if err != nil {
		slog.Error("failed to fetch session by jwt", slog.Any("err", err))
		return nil, err
	}

	token := new(oauth2.Token)
	if err = session.GetBackendToken(token); err != nil {
		slog.Error("failed to get backend token", slog.Any("err", err))
		return nil, err
	}

	provider, err := s.apps.GetProvider(session.AppID)
	if err != nil {
		slog.Error("failed to get provider", slog.Any("err", err))
		return nil, err
	}

	user, err := provider.GetUser(req.Ctx, token)
	if err != nil {
		slog.Error("failed to get user", slog.Any("err", err))
		return nil, httperror.ErrUnknown(err) // TODO: Handle upstream error types.
	}

	return user, nil
}

func (s *Service) refreshBackendToken(ctx context.Context, appID string, session *session.Session) (*oauth2.Token, error) {
	provider, err := s.apps.GetProvider(appID)
	if err != nil {
		return nil, err
	}

	token := new(oauth2.Token)
	if err := session.GetBackendToken(token); err != nil {
		return nil, err
	}

	token, err = provider.RefreshToken(ctx, token.RefreshToken)
	if err != nil {
		return nil, httperror.ErrUnknown(err) // TODO: Handle upstream error types.
	}
	return token, nil
}
