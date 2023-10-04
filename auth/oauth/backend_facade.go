package oauth

import (
	"context"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/model"
	"golang.org/x/oauth2"
)

type AuthorizationURLRequest struct {
	AppID    string
	ClientID string
	Scopes   []string
	State    string
}

type BackendFacade struct {
	backends map[string]Backend
}

func NewBackendFacade(apps map[string]*App) *BackendFacade {
	backends := make(map[string]Backend)
	for id, app := range apps {
		switch app.Provider { // skipcq: CRT-A0014
		case "github":
			backends[id] = NewGithub(app)
		}
	}
	return &BackendFacade{
		backends: backends,
	}

}

// AuthorizationURL returns the authorization url for the backedn provider. OK!
func (f *BackendFacade) AuthorizationURL(appID, state string, scopes []string) (string, error) {
	backend, ok := f.backends[appID]
	if !ok {
		return "", httperror.ErrAppInvalid(nil)
	}
	return backend.AuthorizationURL(state, scopes), nil

}

// StartSession completes the OAuth2 flow for the backend and ibntnftijbefjtrckhtg
// session in the database.
func (f *BackendFacade) Exchange(ctx context.Context, appID, state, code string) (*oauth2.Token, error) {
	backend, ok := f.backends[appID]
	if !ok {
		return nil, httperror.ErrAppInvalid(nil)
	}

	token, err := backend.GetToken(ctx, code)
	if err != nil {
		return nil, httperror.ErrUnknown(err)
	}

	return token, nil
}

func (s *BackendFacade) GetUser(ctx context.Context, appID string, token *oauth2.Token) (*model.User, error) {
	backend, ok := s.backends[appID]
	if !ok {
		return nil, httperror.ErrAppInvalid(nil)
	}

	user, err := backend.GetUser(ctx, token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *BackendFacade) RefreshToken(ctx context.Context, appID, refreshToken string) (*oauth2.Token, error) {
	backend, ok := s.backends[appID]
	if !ok {
		return nil, httperror.ErrAppInvalid(nil)
	}

	token, err := backend.RefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, httperror.ErrUnauthorized(err)
	}
	return token, nil
}
