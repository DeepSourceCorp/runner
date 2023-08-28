package main

import (
	"context"
	"fmt"

	"github.com/DeepSourceCorp/artifacts/storage"
	"github.com/deepsourcecorp/runner/artifact"
	"github.com/deepsourcecorp/runner/config"
)

func GetArtifacts(ctx context.Context, c *config.Config) (*artifact.Facade, error) {
	storage, err := storage.NewStorageClient(ctx, c.ObjectStorage.Provider, []byte(c.ObjectStorage.Credential))
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
