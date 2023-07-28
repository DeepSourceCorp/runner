package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/deepsourcecorp/runner/auth/jwtutil"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcecorp/runner/orchestrator"
)

var (
	CleanerInterval = 30 * time.Minute
)

func GetOrchestrator(ctx context.Context, c *config.Config, provider orchestrator.Provider, driverType string) (*orchestrator.Facade, error) {
	driver, err := createDriver(driverType)
	if err != nil {
		return nil, fmt.Errorf("error initializing orchestrator: %w", err)
	}

	signer := jwtutil.NewSigner(c.Runner.PrivateKey)

	if c.Kubernetes == nil {
		return nil, errors.New("error initializing orchestrator: kubernetes config is empty")
	}

	kubernetesOpts := &orchestrator.KubernetesOpts{
		Namespace:        c.Kubernetes.Namespace,
		NodeSelector:     c.Kubernetes.NodeSelector,
		ImageURL:         c.Kubernetes.ImageRegistry.RegistryUrl,
		ImagePullSecrets: []string{c.Kubernetes.ImageRegistry.PullSecretName},
	}

	taskOpts := &orchestrator.TaskOpts{
		RemoteHost:           c.DeepSource.Host.String(),
		SnippetStorageType:   c.ObjectStorage.Backend,
		SnippetStorageBucket: c.ObjectStorage.Bucket,
		KubernetesOpts:       kubernetesOpts,
	}

	cleanerOpts := &orchestrator.CleanerOpts{
		Namespace: c.Kubernetes.Namespace,
		Interval:  &CleanerInterval,
	}

	opts := &orchestrator.Opts{
		TaskOpts:    taskOpts,
		CleanerOpts: cleanerOpts,
		Driver:      driver,
		Provider:    provider,
		Signer:      signer,
	}

	return orchestrator.New(opts)
}

func createDriver(driver string) (orchestrator.Driver, error) {
	switch driver {
	case orchestrator.DriverPrinter:
		return orchestrator.NewK8sPrinterDriver(), nil
	default:
		return orchestrator.NewK8sDriver("")
	}
}
