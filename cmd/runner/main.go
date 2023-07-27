package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/rqlite"
	"github.com/deepsourcecorp/runner/rqlite/migrations"
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
	slog.Info("config loaded successfully")
	return cfg, nil
}

func SetLogLevel() {
	if Debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}
}

func main() {
	ParseFlags()
	SetLogLevel()
	c, err := LoadConfig()
	if err != nil {
		slog.Error("failed to load config", slog.Any("err", err))
		os.Exit(1)
	}
	db, err := rqlite.Connect(c.RQLite.Host, c.RQLite.Port)
	if err != nil {
		slog.Error("failed to connect to rqlite", slog.Any("err", err))
		os.Exit(1)
	}
	migrator, err := migrations.NewMigrator(db)
	if err != nil {
		slog.Error("failed to initialize migrator", slog.Any("err", err))
		os.Exit(1)
	}
	err = migrator.Migrate()
	if err != nil {
		slog.Error("failed to migrate database", slog.Any("err", err))
		os.Exit(1)
	}

	s := NewServer(c)
	r, err := s.Router()
	if err != nil {
		slog.Error("failed to initialize router", slog.Any("err", err))
		os.Exit(1)
	}
	r.Setup()
	StartCleanup(context.Background(), c)
	err = s.Start()
	if err != nil {
		slog.Error("failed to start server", slog.Any("err", err))
		os.Exit(1)
	}
}
