package main

import (
	"context"
	"net/http"

	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/provider"
	"github.com/deepsourcecorp/runner/provider/common"
	"github.com/deepsourcecorp/runner/provider/github"
	"github.com/deepsourcecorp/runner/provider/gitlab"
)

func GetProvider(_ context.Context, c *config.Config, client *http.Client) *provider.Adapter {
	runner := &common.Runner{
		ID:            c.Runner.ID,
		WebhookSecret: c.Runner.WebhookSecret,
	}

	deepsource := &common.DeepSource{
		Host: c.DeepSource.Host,
	}

	githubService := github.NewService(&github.ServiceOpts{
		Runner:     runner,
		DeepSource: deepsource,
		Apps:       createGithubApps(c),
		Client:     client,
	})

	githubHandler := github.NewHandler(githubService, client)
	githubAuthenticator := github.NewAuthenticator(githubService)

	handlers := make(map[string]provider.Handler)
	authenticators := make(map[string]provider.Authenticator)

	// TODO: WARNING TokenResolver
	gitlabService := gitlab.NewService(&gitlab.ServiceOpts{
		Runner:     runner,
		DeepSource: deepsource,
		Apps:       createGitlabApps(c),
		Client:     client,
	})

	gitlabHandler := gitlab.NewHandler(gitlabService)
	gitlabAuthenticator := gitlab.NewAuthenticator(gitlabService)

	for _, v := range c.Apps {
		switch {
		case v.Provider == "github":
			handlers[v.ID] = githubHandler
			authenticators[v.ID] = githubAuthenticator
		case v.Provider == "gitlab":
			handlers[v.ID] = gitlabHandler
			authenticators[v.ID] = gitlabAuthenticator
		}
	}

	return provider.NewAdapter(handlers, authenticators)
}

func createGithubApps(c *config.Config) map[string]*github.App {
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
	return apps
}

func createGitlabApps(c *config.Config) map[string]*gitlab.App {
	apps := make(map[string]*gitlab.App)
	for _, v := range c.Apps {
		switch {
		case v.Provider == "gitlab":
			apps[v.ID] = &gitlab.App{
				ID:            v.ID,
				WebhookSecret: v.Gitlab.WebhookSecret,
				APIHost:       &v.Gitlab.APIHost,
			}
		}
	}
	return apps
}

func createProviderApps(c *config.Config) map[string]*provider.App {
	apps := make(map[string]*provider.App)
	for _, v := range c.Apps {
		switch {
		case v.Provider == "github":
			apps[v.ID] = &provider.App{
				Provider: "github",
			}
		}
	}
	return apps
}
