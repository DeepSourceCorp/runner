package main

import (
	"net/http"

	"github.com/deepsourcecorp/runner/auth/jwtutil"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/orchestrator"
	"github.com/deepsourcecorp/runner/provider/facade"
	"github.com/deepsourcecorp/runner/provider/github"
)

func initializeOrchestrator(c *config.Config, client *http.Client) (*orchestrator.Handler, error) {
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
	factory := github.NewAPIProxyFactory(apps, client)
	provider := facade.NewProviderFacade(factory)

	driver, err := orchestrator.GetDriver(Driver)
	if err != nil {
		return nil, err
	}

	opts := &orchestrator.TaskOpts{
		RemoteHost:           c.DeepSource.Host.String(),
		SnippetStorageType:   c.ObjectStorage.Backend,
		SnippetStorageBucket: c.ObjectStorage.Bucket,
		KubernetesOpts: &orchestrator.KubernetesOpts{
			Namespace:        c.Kubernetes.Namespace,
			NodeSelector:     c.Kubernetes.NodeSelector,
			ImageURL:         c.Kubernetes.ImageRegistry.RegistryUrl,
			ImagePullSecrets: []string{c.Kubernetes.ImageRegistry.PullSecretName},
		},
	}
	signer := jwtutil.NewSigner(c.Runner.ID, c.Runner.PrivateKey)
	return orchestrator.NewHandler(opts, driver, provider, signer), nil
}
