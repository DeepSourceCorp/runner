package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deepsourcecorp/runner/auth"
	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/deepsourcecorp/runner/auth/saml"
	"github.com/deepsourcecorp/runner/auth/store"
	rqlitestore "github.com/deepsourcecorp/runner/auth/store/rqlite"

	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/rqlite"
)

func GetAuthentiacator(ctx context.Context, c *config.Config) (*auth.Facade, error) {
	apps := createOAuthApps(c)

	store, err := createRQLiteStore(c.RQLite)
	if err != nil {
		return nil, fmt.Errorf("error initialising auth: %w", err)
	}

	samlOpts := setupSAMLOptions(c)

	runner := &model.Runner{
		ID:           c.Runner.ID,
		ClientID:     c.Runner.ClientID,
		ClientSecret: c.Runner.ClientSecret,
		PrivateKey:   c.Runner.PrivateKey,
	}

	deepsource := &model.DeepSource{
		Host: c.DeepSource.Host,
	}

	opts := &auth.Opts{
		Runner:     runner,
		DeepSource: deepsource,
		Apps:       apps,
		Store:      store,
		SAML:       samlOpts,
	}

	app, err := auth.New(ctx, opts, http.DefaultClient)
	if err != nil {
		return nil, fmt.Errorf("error initalizing auth: %w", err)
	}

	return app, nil
}

func createOAuthApps(c *config.Config) map[string]*oauth.App {
	apps := make(map[string]*oauth.App)
	for _, v := range c.Apps {
		switch v.Provider { // skipcq: CRT-A0014
		case "github":
			apps[v.ID] = &oauth.App{
				ID:           v.ID,
				ClientID:     v.Github.ClientID,
				ClientSecret: v.Github.ClientSecret,
				AuthHost:     v.Github.Host,
				APIHost:      v.Github.APIHost,
				Provider:     oauth.ProviderGithub,
				RedirectURL:  *c.Runner.Host.JoinPath(oauth.CallbackURL(v.ID)),
			}
		}
	}
	return apps
}

func createRQLiteStore(c *config.RQLite) (store.Store, error) {
	db, err := rqlite.Connect(c.Host, c.Port)
	if err != nil {
		return nil, fmt.Errorf("error creating rqlite store: %w", err)
	}
	return rqlitestore.New(db), nil
}

func setupSAMLOptions(c *config.Config) *saml.Opts {
	if c.SAML != nil && !c.SAML.Enabled {
		return &saml.Opts{
			Certificate: c.SAML.Certificate,
			MetadataURL: c.SAML.MetadataURL,
			RootURL:     c.DeepSource.Host,
		}
	}
	return nil
}
