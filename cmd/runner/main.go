package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/deepsourcecorp/runner/config"
	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

var (
	HideBanner bool
	RunnerPort int
	ConfigPath string
	Driver     string
	Debug      bool
)

func ParseFlags() {
	hideBanner := flag.Bool("hide-banner", false, "Hide the banner")
	runnerPort := flag.Int("port", 8080, "HTTP server port")
	configPath := flag.String("config", "/config/config.yaml", "Path to config file")
	driver := flag.String("driver", "kubernetes", "Driver to use for running jobs")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	HideBanner = *hideBanner
	RunnerPort = *runnerPort
	ConfigPath = *configPath
	Driver = *driver
	Debug = *debug
}

func LoadConfig() (*config.Config, error) {
	f, err := os.Open(ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	cfg, err := config.LoadConfig(f)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func SetLogLevel() {
	if Debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
		slog.Info("debug logging enabled")
	}
}

func main() {
	ParseFlags()
	SetLogLevel()
	ctx := context.Background()
	c, err := LoadConfig()
	if err != nil {
		slog.Error("failed to load config", slog.Any("err", err))
		os.Exit(1)
	}
	s := NewServer(c)
	if !HideBanner {
		s.PrintBanner()
	}

	err = Migrate(c.RQLite)
	if err != nil {
		sentry.CaptureException(err)
		slog.Error(err.Error())
		os.Exit(1)
	}

	syncer := GetSyncer(ctx, c, http.DefaultClient)
	err = syncer.Sync()
	if err != nil {
		sentry.CaptureException(err)
		slog.Error("failed to sync upstream", slog.Any("err", err))
		os.Exit(1)
	}

	r, err := s.Router()
	if err != nil {
		sentry.CaptureException(err)
		slog.Error("failed to initialize router", slog.Any("err", err))
		os.Exit(1)
	}

	db, err := GetDB(c)
	if err != nil {
		sentry.CaptureException(err)
		slog.Error("failed to initialize db", slog.Any("err", err))
		os.Exit(1)
	}

	auth := GetOAuth(c, db)
	auth.AddRoutes(r)

	provider, err := GetProvider(ctx, c, http.DefaultClient)
	if err != nil {
		sentry.CaptureException(err)
		slog.Error("failed to initialize provider", slog.Any("err", err))
		os.Exit(1)
	}
	provider.AddRoutes(r)

	m := DeepSourceMiddleware(c)
	orchestrator, err := GetOrchestrator(ctx, c, provider.Adapter, Driver)
	if err != nil {
		sentry.CaptureException(err)
		slog.Error("failed to initialize orchestrator", slog.Any("err", err))
		os.Exit(1)
	}
	orchestrator.AddRoutes(r, []echo.MiddlewareFunc{m}) // Add middleware

	artifacts, err := GetArtifacts(ctx, c)
	if err != nil {
		sentry.CaptureException(err)
		slog.Error("failed to initialize artifacts app", slog.Any("err", err))
		os.Exit(1)
	}
	artifacts.AddRoutes(r, []echo.MiddlewareFunc{}) // Add middleware

	go orchestrator.Cleaner.Start(ctx)

	r.Setup()
	err = s.Start()
	if err != nil {
		sentry.CaptureException(err)
		slog.Error("failed to start server", slog.Any("err", err))
		os.Exit(1)
	}
}
