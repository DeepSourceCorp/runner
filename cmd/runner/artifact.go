package main

import (
	"context"
	"fmt"

	"github.com/deepsourcecorp/runner/artifact"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcelabs/artifacts/storage"
)

func initializeArtifact(c *config.Config) (*artifact.Handler, error) {
	storage, err := storage.NewGoogleCloudStorageClient(context.Background(), []byte(c.ObjectStorage.Credential))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize artifacts: %w", err)
	}
	return artifact.NewHandler(storage, c.ObjectStorage.Bucket), nil
}
