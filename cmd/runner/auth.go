package main

import (
	"context"
	"fmt"

	"github.com/deepsourcecorp/runner/auth"
	"github.com/deepsourcecorp/runner/auth/v2/model"
	"github.com/deepsourcecorp/runner/auth/v2/oauth"
	"github.com/deepsourcecorp/runner/auth/v2/saml"

	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/rqlite"
)

func GetAuthentiacator(ctx context.Context, c *config.Config) (*auth.Auth, error) {

	runner := &model.Runner{
		ID:           c.Runner.ID,
		ClientID:     c.Runner.ClientID,
		ClientSecret: c.Runner.ClientSecret,
		PrivateKey:   c.Runner.PrivateKey,
	}

	deepsource := &model.DeepSource{
		Host:      c.DeepSource.Host,
		PublicKey: c.DeepSource.PublicKey,
	}

	apps := createOAuthApps(c)

	database, err := getDatabase(c)
	if err != nil {
		return nil, fmt.Errorf("error initialising auth: %w", err)
	}

	opts := &auth.Opts{
		Runner:     runner,
		DeepSource: deepsource,
		Apps:       apps,

		Database: database,
	}

	app, err := auth.New(opts)
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

func getDatabase(c *config.Config) (interface{}, error) {
	if c.RQLite != nil {
		db, err := rqlite.Connect(c.RQLite.Host, c.RQLite.Port)
		if err != nil {
			return nil, fmt.Errorf("error creating rqlite store: %w", err)
		}
		return db, nil
	}
	return nil, nil
}

func setupSAMLOptions(c *config.Config) *saml.Opts {
	if c.SAML != nil && c.SAML.Enabled {
		return &saml.Opts{
			Certificate: c.SAML.Certificate,
			MetadataURL: c.SAML.MetadataURL,
			RootURL:     c.DeepSource.Host,
		}
	}
	return nil
}
