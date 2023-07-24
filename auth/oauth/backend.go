package oauth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

const ()

type IBackend interface {
	AuthorizationURL(state string, scopes []string) (string, error)
	GetToken(ctx context.Context, code string) (*oauth2.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error)
	GetUser(ctx context.Context, token *oauth2.Token) (*User, error)
}

type Factory struct {
	apps map[string]*App
}

func NewFactory(apps map[string]*App) *Factory {
	return &Factory{
		apps: apps,
	}
}

func (f *Factory) GetBackend(appID string) (IBackend, error) {
	app := f.apps[appID]
	if app == nil {
		return nil, fmt.Errorf("no configuration found for app %s", appID)
	}
	switch app.Provider {
	case "github":
		return NewGithub(app)
	}
	return nil, fmt.Errorf("unknown provider %s", app.Provider)
}
