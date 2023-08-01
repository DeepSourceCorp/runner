package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/deepsourcecorp/runner/config"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
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
	cors echo.MiddlewareFunc
}

func NewServer(c *config.Config) *Server {
	if c.Sentry != nil && c.Sentry.DSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: c.Sentry.DSN,
		}); err != nil {
			slog.Error("failed to initialize sentry", slog.Any("err", err))
		}
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(sentryecho.New(sentryecho.Options{}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time=${time_rfc3339_nano} level=INFO method=${method}, uri=${uri}, status=${status}\n",
	}))
	cors := CorsMiddleware(c.DeepSource.Host.String())
	return &Server{Echo: e, Config: c, cors: cors}
}

func (s *Server) Start() error {
	err := s.Echo.Start(fmt.Sprintf(":%d", RunnerPort))
	if err != nil {
		slog.Error("failed to start server", slog.Any("err", err))
		return err
	}
	return nil
}

func (*Server) PrintBanner() {
	fmt.Println(fmt.Sprintf(Banner, Version))
}

func (s *Server) Router() (*Router, error) {
	router := &Router{
		e: s.Echo,
		Routes: []Route{
			{
				Method: http.MethodGet, Path: "/readyz", HandlerFunc: func(c echo.Context) error { return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok"}) },
			},
			{
				Method: http.MethodOptions, Path: "/*", HandlerFunc: func(c echo.Context) error { return c.NoContent(http.StatusOK) }, Middleware: []echo.MiddlewareFunc{s.cors},
			},
		},
	}
	return router, nil
}
