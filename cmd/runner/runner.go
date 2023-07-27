package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deepsourcecorp/runner/artifact"
	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/saml"
	store "github.com/deepsourcecorp/runner/auth/store/rqlite"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/orchestrator"
	"github.com/deepsourcecorp/runner/rqlite"
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
	*echo.Echo
	*config.Config
	*http.Client
}

func NewServer(c *config.Config) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	// e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339_nano} level=INFO method=${method}, uri=${uri}, status=${status}\n",
	}))
	return &Server{Echo: e, Config: c}
}

func (s *Server) Start() error {
	if !HideBanner {
		s.PrintBanner()
	}
	err := s.Echo.Start(fmt.Sprintf(":%d", RunnerPort))
	if err != nil {
		slog.Error("failed to start server", slog.Any("err", err))
		return err
	}
	return nil
}

func (s *Server) PrintBanner() {
	fmt.Println(fmt.Sprintf(Banner, Version))
}

func (s *Server) Router() (*Router, error) {
	auth, err := initializeAuth(s.Config)
	if err != nil {
		return nil, fmt.Errorf("router failed to initialize: %w", err)
	}

	orchestrator, err := initializeOrchestrator(s.Config, s.Client)
	if err != nil {
		return nil, fmt.Errorf("router failed to initialize: %w", err)
	}

	github, err := intitializeGithub(s.Config, s.Client)
	if err != nil {
		return nil, fmt.Errorf("router failed to initialize: %w", err)
	}

	artifacts, err := initializeArtifact(s.Config)
	if err != nil {
		return nil, fmt.Errorf("router failed to initialize: %w", err)
	}

	saml, err := SAMLHandler(s.Config)
	if err != nil {
		return nil, fmt.Errorf("router failed to initialize: %w", err)
	}

	corsMiddleware := artifact.CORSMiddleware(s.Config.DeepSource.Host.String())

	router := &Router{
		e: s.Echo,
		Routes: []Route{
			// Health check routes.
			{Method: http.MethodGet, Path: "/readyz", HandlerFunc: func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok"}) }},
			{Method: http.MethodPost, Path: "/refresh", HandlerFunc: auth.TokenHandlers.HandleRefresh},

			// OAuth routes.
			{Method: "GET", Path: "/apps/:app_id/auth/authorize", HandlerFunc: auth.OAuthHandlers.HandleAuthorize},
			{Method: "GET", Path: "/apps/:app_id/auth/callback", HandlerFunc: auth.OAuthHandlers.HandleCallback},
			{Method: "GET", Path: "/apps/:app_id/auth/session", HandlerFunc: auth.OAuthHandlers.HandleSession},
			{Method: "POST", Path: "/apps/:app_id/auth/token", HandlerFunc: auth.OAuthHandlers.HandleToken},
			{Method: "POST", Path: "/apps/:app_id/auth/refresh", HandlerFunc: auth.OAuthHandlers.HandleRefresh},
			{Method: "GET", Path: "/apps/:app_id/auth/user", HandlerFunc: auth.OAuthHandlers.HandleUser},

			// SAML routes.
			{Method: "*", Path: "/saml/*", HandlerFunc: saml.SAMLHandler()},
			{Method: http.MethodGet, Path: "/apps/saml/auth/authorize", HandlerFunc: saml.AuthorizationHandler()},
			{Method: http.MethodGet, Path: "/apps/saml/auth/session", HandlerFunc: saml.HandleSession},
			{Method: http.MethodPost, Path: "/apps/saml/auth/token", HandlerFunc: saml.HandleToken},
			{Method: http.MethodPost, Path: "/apps/saml/auth/refresh", HandlerFunc: saml.HandleRefresh},

			// Orchestrator routes.
			{Method: http.MethodPost, Path: "apps/:app_id/tasks/analysis", HandlerFunc: orchestrator.HandleAnalysis, Middleware: []echo.MiddlewareFunc{auth.TokenMiddleware}},
			{Method: http.MethodPost, Path: "apps/:app_id/tasks/autofix", HandlerFunc: orchestrator.HandleAutofix, Middleware: []echo.MiddlewareFunc{auth.TokenMiddleware}},
			{Method: http.MethodPost, Path: "apps/:app_id/tasks/transformer", HandlerFunc: orchestrator.HandleTransformer, Middleware: []echo.MiddlewareFunc{auth.TokenMiddleware}},
			{Method: http.MethodPost, Path: "apps/:app_id/tasks/cancelcheck", HandlerFunc: orchestrator.HandleCancelCheck, Middleware: []echo.MiddlewareFunc{auth.TokenMiddleware}},
			{Method: http.MethodPost, Path: "apps/:app_id/tasks/commit", HandlerFunc: orchestrator.HandlePatcher, Middleware: []echo.MiddlewareFunc{auth.TokenMiddleware}},

			// Github provider routes.
			{Method: "*", Path: "apps/:app_id/webhook", HandlerFunc: github.HandleWebhook},
			{Method: "*", Path: "apps/:app_id/api/*", HandlerFunc: github.HandleAPI},
			{Method: "*", Path: "apps/:app_id/installation/new", HandlerFunc: github.HandleInstallation},

			// Artifact routes.
			{Method: http.MethodOptions, Path: "apps/:app_id/artifacts/*", HandlerFunc: artifacts.HandleOptions, Middleware: []echo.MiddlewareFunc{corsMiddleware}},
			{Method: http.MethodPost, Path: "apps/:app_id/artifacts/analysis", HandlerFunc: artifacts.HandleAnalysis, Middleware: []echo.MiddlewareFunc{corsMiddleware, auth.SessionMiddleware}},
			{Method: http.MethodPost, Path: "apps/:app_id/artifacts/autofix", HandlerFunc: artifacts.HandleAutofix, Middleware: []echo.MiddlewareFunc{corsMiddleware, auth.SessionMiddleware}},
		},
	}
	return router, nil
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

	runner := &model.Runner{
		ID:           c.Runner.ID,
		ClientID:     c.Runner.ClientID,
		ClientSecret: c.Runner.ClientSecret,
		PrivateKey:   c.Runner.PrivateKey,
	}

	deepsource := &model.DeepSource{
		Host: c.DeepSource.Host,
	}

	db, err := rqlite.Connect(c.RQLite.Host, c.RQLite.Port)
	if err != nil {
		return nil, fmt.Errorf("error initalizing auth: %w", err)
	}
	store := store.New(db)

	return saml.NewHandler(runner, deepsource, middleware, store), nil
}

func StartCleanup(ctx context.Context, cfg *config.Config) error {
	driver, err := orchestrator.GetDriver(Driver)
	if err != nil {
		return fmt.Errorf("failed to initalize cleanup: %w", err)
	}
	// TODO: add configuration option for cleanup interval
	c := orchestrator.NewCleaner(driver, &orchestrator.CleanerOpts{
		Namespace: cfg.Kubernetes.Namespace,
	})
	go c.Start(ctx)
	return nil
}
