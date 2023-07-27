package main

import (
	"fmt"

	"github.com/deepsourcecorp/runner/auth"
	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/oauth"
	store "github.com/deepsourcecorp/runner/auth/store/rqlite"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/rqlite"
)

func initializeAuth(cfg *config.Config) (*auth.Authentication, error) {
	apps := make(map[string]*oauth.App)
	for _, v := range cfg.Apps {
		switch {
		case v.Provider == "github":
			apps[v.ID] = &oauth.App{
				ID:           v.ID,
				ClientID:     v.Github.ClientID,
				ClientSecret: v.Github.ClientSecret,
				AuthHost:     v.Github.Host,
				APIHost:      v.Github.APIHost,
				Provider:     oauth.ProviderGithub,
				RedirectURL:  *cfg.Runner.Host.JoinPath(oauth.CallbackURL(v.ID)),
			}
		}
	}

	db, err := rqlite.Connect(cfg.RQLite.Host, cfg.RQLite.Port)
	if err != nil {
		return nil, fmt.Errorf("error initalizing auth: %w", err)
	}
	store := store.New(db)

	return auth.NewAuthentication(
		&model.Runner{
			ID:           cfg.Runner.ID,
			ClientID:     cfg.Runner.ClientID,
			ClientSecret: cfg.Runner.ClientSecret,
			PrivateKey:   cfg.Runner.PrivateKey,
		},
		&model.DeepSource{
			Host: cfg.DeepSource.Host,
		},
		apps,
		store,
	), nil
}
