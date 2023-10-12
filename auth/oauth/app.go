package oauth

import (
	"errors"
	"net/url"
)

type App struct {
	ClientID     string
	ClientSecret string
	RedirectURL  *url.URL
	AuthHost     *url.URL
	APIHost      *url.URL
	Provider     string
}

func (a *App) GetProvider() (Provider, error) {
	switch a.Provider {
	case ProviderGithub:
		return NewGithub(a), nil
	}
	return nil, errors.New("invalid provider")
}

type Apps map[string]*App

func (a Apps) GetProvider(appID string) (Provider, error) {
	app, ok := a[appID]
	if !ok {
		return nil, errors.New("invalid app id")
	}
	return app.GetProvider()
}
