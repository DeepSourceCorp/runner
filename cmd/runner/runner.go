package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/deepsourcecorp/runner/artifact"
	"github.com/deepsourcecorp/runner/auth"
	authmiddleware "github.com/deepsourcecorp/runner/auth/middleware"
	"github.com/deepsourcecorp/runner/auth/oauth"
	oauthsessionstore "github.com/deepsourcecorp/runner/auth/oauth/persistence/rqlite"
	"github.com/deepsourcecorp/runner/auth/saml"
	samlsessionstore "github.com/deepsourcecorp/runner/auth/saml/persistence/rqlite"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/orchestrator"
	"github.com/deepsourcecorp/runner/provider/facade"
	"github.com/deepsourcecorp/runner/provider/github"
	"github.com/deepsourcecorp/runner/rqlite"
	"github.com/deepsourcelabs/artifacts/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/exp/slog"
)

const (
	Banner = `________
___  __ \___  ____________________________
__  /_/ /  / / /_  __ \_  __ \  _ \_  ___/
_  _, _// /_/ /_  / / /  / / /  __/  /    
/_/ |_| \__,_/ /_/ /_//_/ /_/\___//_/     
------------------------------------------
By DeepSource | v%s
------------------------------------------`

	Version = "0.1.0-beta.1"
)

const (
	DriverKubernetes = "kubernetes"
	DriverPrinter    = "printer"
	CleanupInterval  = 5 * time.Minute
)

type Server struct {
	e *echo.Echo
	c *config.Config
}

func NewServer(c *config.Config) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339_nano} level=INFO method=${method}, uri=${uri}, status=${status}\n",
	}))
	return &Server{e: e, c: c}
}

func (s *Server) Start() error {
	if !HideBanner {
		s.PrintBanner()
	}
	err := s.e.Start(fmt.Sprintf(":%d", RunnerPort))
	if err != nil {
		slog.Error("failed to start server", slog.Any("err", err))
		return err
	}
	return nil
}

func (s *Server) PrintBanner() {
	fmt.Println(fmt.Sprintf(Banner, Version))
}

func (s *Server) Router() *Router {
	oauthHandler, err := OauthHandler(s.c)
	if err != nil {
		slog.Error("error initializing OAuth handler", slog.Any("err", err))
		os.Exit(1)
	}

	githubHandler, err := GithubHandler(s.c)
	if err != nil {
		slog.Error("error initializing Github provider handler", slog.Any("err", err))
		os.Exit(1)
	}

	orchestratorHandler, err := OrchestratorHandler(s.c)
	if err != nil {
		slog.Error("error initializing orchestrator handler", slog.Any("err", err))
		os.Exit(1)
	}

	artifactHandler, err := ArtifactHandler(s.c)
	if err != nil {
		slog.Error("error initializing artifact handler", slog.Any("err", err))
		os.Exit(1)
	}

	sessionMiddleware, err := SessionMiddleware(s.c)
	if err != nil {
		slog.Error("error initializing session middleware", slog.Any("err", err))
		os.Exit(1)
	}

	tokenMiddleware, err := TokenMiddleware(s.c)
	if err != nil {
		slog.Error("error initializing token middleware", slog.Any("err", err))
		os.Exit(1)
	}

	corsMiddleware := &artifact.CORSMiddleware{
		AllowedOrigin: s.c.DeepSource.Host.String(),
	}

	fmt.Println(s.c.DeepSource.Host.String())
	return &Router{
		e: s.e,
		Routes: []Route{
			// Health check routes.
			{
				Method:      http.MethodGet,
				Path:        "/readyz",
				HandlerFunc: func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok"}) },
			},

			// OAuth routes.
			{
				Method:      "GET",
				Path:        "/apps/:app_id/auth/authorize",
				HandlerFunc: oauthHandler.HandleAuthorize,
			},
			{
				Method:      "GET",
				Path:        "/apps/:app_id/auth/callback",
				HandlerFunc: oauthHandler.HandleCallback,
			},
			{
				Method:      "GET",
				Path:        "/apps/:app_id/auth/session",
				HandlerFunc: oauthHandler.HandleSession,
			},
			{
				Method:      "POST",
				Path:        "/apps/:app_id/auth/token",
				HandlerFunc: oauthHandler.HandleToken,
			},
			{
				Method:      "POST",
				Path:        "/apps/:app_id/auth/refresh",
				HandlerFunc: oauthHandler.HandleRefresh,
			},
			{
				Method:      "GET",
				Path:        "/apps/:app_id/auth/user",
				HandlerFunc: oauthHandler.HandleUser,
			},

			// Orchestrator routes.
			{
				Method:      http.MethodPost,
				Path:        "apps/:app_id/tasks/analysis",
				HandlerFunc: orchestratorHandler.HandleAnalysis,
				Middleware:  []echo.MiddlewareFunc{tokenMiddleware},
			},
			{
				Method:      http.MethodPost,
				Path:        "apps/:app_id/tasks/autofix",
				HandlerFunc: orchestratorHandler.HandleAutofix,
				Middleware:  []echo.MiddlewareFunc{tokenMiddleware},
			},
			{
				Method:      http.MethodPost,
				Path:        "apps/:app_id/tasks/transformer",
				HandlerFunc: orchestratorHandler.HandleTransformer,
				Middleware:  []echo.MiddlewareFunc{tokenMiddleware},
			},
			{
				Method:      http.MethodPost,
				Path:        "apps/:app_id/tasks/cancelcheck",
				HandlerFunc: orchestratorHandler.HandleCancelCheck,
				Middleware:  []echo.MiddlewareFunc{tokenMiddleware},
			},
			{
				Method:      http.MethodPost,
				Path:        "apps/:app_id/tasks/commit",
				HandlerFunc: orchestratorHandler.HandlePatcher,
				Middleware:  []echo.MiddlewareFunc{tokenMiddleware},
			},

			// Github provider routes.
			{
				Method:      "*",
				Path:        "apps/:app_id/webhook",
				HandlerFunc: githubHandler.HandleWebhook,
			},
			{
				Method:      "*",
				Path:        "apps/:app_id/api/*",
				HandlerFunc: githubHandler.HandleAPI,
			},
			{
				Method:      "*",
				Path:        "apps/:app_id/installation/new",
				HandlerFunc: githubHandler.HandleInstallation,
			},

			// Artifact routes.
			{
				Method: http.MethodOptions,
				Path:   "apps/:app_id/artifacts/*",
				HandlerFunc: func(c echo.Context) error {
					return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok"})
				},
				Middleware: []echo.MiddlewareFunc{
					corsMiddleware.Handle,
				},
			},
			{
				Method:      http.MethodPost,
				Path:        "apps/:app_id/artifacts/analysis",
				HandlerFunc: artifactHandler.HandleAnalysis,
				Middleware: []echo.MiddlewareFunc{
					corsMiddleware.Handle,
					sessionMiddleware,
				},
			},
			{
				Method:      http.MethodPost,
				Path:        "apps/:app_id/artifacts/autofix",
				HandlerFunc: artifactHandler.HandleAutofix,
				Middleware: []echo.MiddlewareFunc{
					corsMiddleware.Handle,
					sessionMiddleware,
				},
			},
		},
	}
}

// OauthHandler returns a new oauth handler.
func OauthHandler(c *config.Config) (*oauth.Handler, error) {
	apps := getOauthApps(c)
	factory := oauth.NewFactory(apps)
	runner := &auth.Runner{
		ClientID:     c.Runner.ClientID,
		ClientSecret: c.Runner.ClientSecret,
	}
	deepsource := &auth.DeepSource{
		Host: c.DeepSource.Host,
	}
	store, err := getOauthStore(c)
	if err != nil {
		slog.Error("error initalizing OAuth handler", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}
	return oauth.NewHandler(runner, deepsource, store, factory), nil
}

func getOauthStore(cfg *config.Config) (*oauthsessionstore.SessionStore, error) {
	db, err := rqlite.Connect(cfg.RQLite.Host, cfg.RQLite.Port)
	if err != nil {
		slog.Error("error connecting to rqlite", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}
	return oauthsessionstore.NewSessionStore(db), nil
}

func getOauthApps(cfg *config.Config) map[string]*oauth.App {
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
	return apps
}

func SessionMiddleware(c *config.Config) (echo.MiddlewareFunc, error) {
	oauthStore, err := getOauthStore(c)
	if err != nil {
		slog.Error("error initializing session middleware", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}
	return authmiddleware.NewSessionMiddleware(oauthStore).Middleware, nil
}

func TokenMiddleware(c *config.Config) (echo.MiddlewareFunc, error) {
	verfifier := auth.NewVerifier(c.DeepSource.PublicKey)
	return authmiddleware.NewTokenMiddleware(c.Runner.ID, verfifier).Middleware, nil
}

// OrchestratorHandler returns a new orchestrator handler.  This uses the standard
// K8s driver by default.  However, if you are developing locally, this might
// error out.
func OrchestratorHandler(c *config.Config) (*orchestrator.Handler, error) {
	provider := getOrchestratorProvider(c)

	driver, err := getOrchestratorDriver()
	if err != nil {
		slog.Error("error initializing k8s driver", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}
	signer := auth.NewSigner(c.Runner.ID, c.Runner.PrivateKey)
	return orchestrator.NewHandler(&orchestrator.TaskOpts{
		RemoteHost:           c.DeepSource.Host.String(),
		SnippetStorageType:   c.ObjectStorage.Backend,
		SnippetStorageBucket: c.ObjectStorage.Bucket,
		KubernetesOpts: &orchestrator.KubernetesOpts{
			Namespace:        c.Kubernetes.Namespace,
			NodeSelector:     c.Kubernetes.NodeSelector,
			ImageURL:         c.Kubernetes.ImageRegistry.RegistryUrl,
			ImagePullSecrets: []string{c.Kubernetes.ImageRegistry.PullSecretName},
		},
	}, driver, provider, signer), nil
}

func getOrchestratorDriver() (orchestrator.Driver, error) {
	if Driver == DriverPrinter {
		return orchestrator.NewK8sPrinterDriver(), nil
	}
	return orchestrator.NewK8sDriver("")
}

// getOrchestratorProvider returns a new orchestrator provider.  This
// initliazes the provider facade.  Right now, this only supports Github.
// TODO: Add support for other VCS providers.  Ideally, the facade should
// abstract away the VCS provider.
func getOrchestratorProvider(cfg *config.Config) orchestrator.Provider {
	apps := getGithubProviderApps(cfg)
	githubAPIFactory := github.NewAPIProxyFactory(apps, http.DefaultClient)
	return facade.NewProviderFacade(githubAPIFactory)
}

// GithubHandler returns a new Github provider handler.
// TODO: Change the default client to a custom client with timeouts.
func GithubHandler(c *config.Config) (*github.Handler, error) {
	apps := getGithubProviderApps(c)
	runner := getGithubProviderRunner(c)
	deepsource := getGithubProviderDeepSource(c)
	apiProxyFactory := github.NewAPIProxyFactory(apps, http.DefaultClient)
	webhookProxyFactory := github.NewWebhookProxyFactory(runner, deepsource, apps, http.DefaultClient)
	return github.NewHandler(apiProxyFactory, webhookProxyFactory)
}

// getGithubProviderApps generates the apps for the provider module
// from the config.  We do not want the config dependency to leak into
// the provider modules.
func getGithubProviderApps(cfg *config.Config) map[string]*github.App {
	apps := make(map[string]*github.App)
	for _, v := range cfg.Apps {
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

func getGithubProviderRunner(cfg *config.Config) *github.Runner {
	return &github.Runner{
		ID:            cfg.Runner.ID,
		WebhookSecret: cfg.Runner.WebhookSecret,
	}
}

func getGithubProviderDeepSource(cfg *config.Config) *github.DeepSource {
	return &github.DeepSource{
		Host: cfg.DeepSource.Host,
	}
}

func ArtifactHandler(c *config.Config) (*artifact.Handler, error) {
	storage, err := storage.NewGoogleCloudStorageClient(context.Background(), []byte(c.ObjectStorage.Credential))
	if err != nil {
		slog.Error("error initializing GCS client", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}
	return artifact.NewHandler(storage, c.ObjectStorage.Bucket), nil
}

func SAMLHandler(c *config.Config) (*saml.Handler, error) {
	samlOpts := &saml.SAMLOpts{
		Certificate: c.SAML.Certificate,
		MetadataURL: c.SAML.MetadataURL,
		RootURL:     c.Runner.Host,
	}

	middleware, err := saml.NewSAMLMiddleware(context.Background(), samlOpts, http.DefaultClient)
	if err != nil {
		slog.Error("error initializing SAML middleware", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}

	runner := &auth.Runner{
		ClientID:     c.Runner.ClientID,
		ClientSecret: c.Runner.ClientSecret,
	}

	deepsource := &auth.DeepSource{
		Host: c.DeepSource.Host,
	}

	store, err := getSAMLStore(c)
	if err != nil {
		slog.Error("error initializing SAML store", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}

	return saml.NewHandler(runner, deepsource, middleware, store), nil
}

func getSAMLStore(cfg *config.Config) (saml.SessionStore, error) {
	db, err := rqlite.Connect(cfg.RQLite.Host, cfg.RQLite.Port)
	if err != nil {
		slog.Error("error initializing rqlite client", slog.Any("error", err), slog.Any("component", "main"))
		return nil, err
	}
	return samlsessionstore.NewSessionStore(db), nil
}

func StartCleanup(ctx context.Context, cfg *config.Config) error {
	d, err := getOrchestratorDriver()
	if err != nil {
		return err
	}
	// TODO: add configuration option for cleanup interval
	c := orchestrator.NewCleaner(d, &orchestrator.CleanerOpts{
		Namespace: cfg.Kubernetes.Namespace,
	})
	go c.Start(ctx)
	return nil
}
