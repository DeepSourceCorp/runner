package main

import (
	"context"
	"fmt"

	"github.com/deepsourcecorp/runner/artifact"
	"github.com/deepsourcecorp/runner/config"
	"github.com/deepsourcelabs/artifacts/storage"
)

func GetArtifacts(ctx context.Context, c *config.Config) (*artifact.Facade, error) {
	storage, err := storage.NewGoogleCloudStorageClient(ctx, []byte(c.ObjectStorage.Credential))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize artifacts: %w", err)
	}

	opts := &artifact.Opts{
		Storage:       storage,
		Bucket:        c.ObjectStorage.Bucket,
		AllowedOrigin: c.DeepSource.Host.String(),
	}

	return artifact.New(ctx, opts)
}
