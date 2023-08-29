package orchestrator

import (
	"context"
	"time"

	artifact "github.com/DeepSourceCorp/artifacts/types"
)

const (
	autofixPublishPath = "/api/runner/autofix/results"
)

type AutofixTask struct {
	runner   *Runner
	opts     *TaskOpts
	driver   Driver
	provider Provider
	signer   Signer
}

type AutofixRunRequest struct {
	Run            *artifact.AutofixRun
	AppID          string
	InstallationID string
}

func NewAutofixTask(runner *Runner, opts *TaskOpts, driver Driver, provider Provider, signer Signer) *AutofixTask {
	return &AutofixTask{
		opts:     opts,
		driver:   driver,
		signer:   signer,
		provider: provider,
		runner:   runner,
	}
}

func (t *AutofixTask) Run(ctx context.Context, req *AutofixRunRequest) error {
	remoteURL, err := t.provider.AuthenticatedRemoteURL(req.AppID, req.InstallationID, req.Run.VCSMeta.RemoteURL)
	if err != nil {
		return err
	}

	token, err := t.signer.GenerateToken(t.runner.ID, []string{ScopeAutofix}, nil, 30*time.Minute)
	if err != nil {
		return err
	}

	req.Run.VCSMeta.RemoteURL = remoteURL
	job, err := NewAutofixDriverJob(req.Run, &AutofixOpts{
		PublisherURL:         t.opts.RemoteHost + autofixPublishPath,
		PublisherToken:       token,
		SnippetStorageType:   t.opts.SnippetStorageType,
		SnippetStorageBucket: t.opts.SnippetStorageBucket,
		SentryDSN:            t.opts.SentryDSN,
		KubernetesOpts:       t.opts.KubernetesOpts,
	})
	if err != nil {
		return err
	}
	return t.driver.TriggerJob(ctx, job)
}
