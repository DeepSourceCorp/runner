package main

import (
	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/deepsourcecorp/runner/auth/store/rqlite"
	"github.com/deepsourcecorp/runner/config"
	"github.com/rqlite/gorqlite"
)

func GetOAuth(c *config.Config, database interface{}) *oauth.OAuth {
	var store session.Store
	switch v := database.(type) {
	case *gorqlite.Connection:
		store = rqlite.NewSessionStore(v)
	}

	apps := make(map[string]*oauth.App)
	for _, app := range c.Apps {
		switch app.Provider {
		case oauth.ProviderGithub:
			apps[app.ID] = &oauth.App{
				ClientID:     app.Github.ClientID,
				ClientSecret: app.Github.ClientSecret,
				AuthHost:     &app.Github.Host,
				APIHost:      &app.Github.APIHost,
				Provider:     oauth.ProviderGithub,
				RedirectURL:  oauth.CallBackURL(app.ID, c.Runner.Host),
			}
		}

	}

	oauth := oauth.New(&oauth.Opts{
		Runner: &common.Runner{
			ID:           c.Runner.ID,
			ClientID:     c.Runner.ClientID,
			ClientSecret: c.Runner.ClientSecret,
			PrivateKey:   c.Runner.PrivateKey,
		},
		Deepsource: &common.DeepSource{
			BaseURL: &c.DeepSource.Host,
		},
		Apps:         apps,
		SessionStore: store,
	})

	return oauth
}
