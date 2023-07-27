package orchestrator

import (
	"context"
	"time"

	artifact "github.com/deepsourcelabs/artifacts/types"
)

const (
	transformerPublishPath = "/api/runner/transformer/results"
)

type TransformerTask struct {
	opts     *TaskOpts
	driver   Driver
	provider Provider
	signer   Signer
}

type TransformerRunRequest struct {
	Run            *artifact.TransformerRun
	AppID          string
	InstallationID string
}

func NewTransformerTask(opts *TaskOpts, driver Driver, provider Provider, signer Signer) *TransformerTask {
	return &TransformerTask{
		opts:     opts,
		driver:   driver,
		provider: provider,
		signer:   signer,
	}
}

func (t *TransformerTask) Run(ctx context.Context, req *TransformerRunRequest) error {
	remoteURL, err := t.provider.AuthenticatedRemoteURL(req.AppID, req.InstallationID, req.Run.VCSMeta.RemoteURL)
	if err != nil {
		return err
	}

	token, err := t.signer.GenerateToken([]string{ScopeTransform}, nil, 30*time.Minute)
	if err != nil {
		return err
	}

	req.Run.VCSMeta.RemoteURL = remoteURL
	job, err := NewTransformerJob(
		req.Run,
		&TransformerOpts{
			PublisherURL:   t.opts.RemoteHost + transformerPublishPath,
			PublisherToken: token,
			KubernetesOpts: t.opts.KubernetesOpts,
		},
	)
	if err != nil {
		return err
	}
	return t.driver.TriggerJob(ctx, job)
}
