package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/deepsourcecorp/runner/config"
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
	s := NewServer(c)
	s.Router().Setup()
	StartCleanup(context.Background(), c)
	err = s.Start()
	if err != nil {
		slog.Error("failed to start server", slog.Any("err", err))
		os.Exit(1)
	}
}
