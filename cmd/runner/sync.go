package main

import (
	"context"
	"net/http"

	"github.com/deepsourcecorp/runner/auth/cryptutil"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/sync"
)

var providers = map[string]string{
	"github": "gh",
}

func GetSyncer(_ context.Context, c *config.Config, client *http.Client) *sync.Syncer {
	deepsource := &sync.DeepSource{
		Host: c.DeepSource.Host,
	}
	runner := &sync.Runner{
		ID:            c.Runner.ID,
		Host:          c.Runner.Host,
		ClientID:      c.Runner.ClientID,
		ClientSecret:  c.Runner.ClientSecret,
		WebhookSecret: c.Runner.WebhookSecret,
	}

	apps := make([]sync.App, 0, len(c.Apps))
	for _, a := range c.Apps {
		apps = append(apps, sync.App{
			ID:       a.ID,
			Name:     a.Name,
			Provider: providers[a.Provider],
		})
	}

	signer := cryptutil.NewSigner(c.Runner.PrivateKey)
	return sync.New(deepsource, runner, apps, signer, client)
}
