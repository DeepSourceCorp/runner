package main

import (
	"net/http"

	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/provider/github"
)

func intitializeGithub(c *config.Config, client *http.Client) (*github.Handler, error) {
	apps := make(map[string]*github.App)
	for _, v := range c.Apps {
		switch {
		case v.Provider == "github":
			apps[v.ID] = &github.App{
				ID:            v.ID,
				AppID:         v.Github.AppID,
				WebhookSecret: v.Github.WebhookSecret,
				BaseHost:      v.Github.Host,
				APIHost:       v.Github.APIHost,
				AppSlug:       v.Github.Slug,
				PrivateKey:    v.Github.PrivateKey,
			}
		}
	}
	runner := &github.Runner{
		ID:            c.Runner.ID,
		WebhookSecret: c.Runner.WebhookSecret,
	}

	deepsource := &github.DeepSource{
		Host: c.DeepSource.Host,
	}

	api := github.NewAPIProxyFactory(apps, http.DefaultClient)
	webhook := github.NewWebhookProxyFactory(runner, deepsource, apps, client)

	return github.NewHandler(api, webhook)
}
