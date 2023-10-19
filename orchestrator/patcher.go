package orchestrator

import (
	"context"
	"time"

	artifact "github.com/DeepSourceCorp/artifacts/types"
)

const (
	patcherPublishPath = "/api/runner/autofix/committer/results"
)

// PatcherTask represents the patcher job task structure.
type PatcherTask struct {
	runner   *Runner
	driver   Driver
	provider Provider
	signer   Signer
	opts     *TaskOpts
}

// PatcherRunRequest represents the data corresponding to the patcher run including the
// AppID and InstallationID of the client.
type PatcherRunRequest struct {
	Run            *artifact.PatcherRun
	AppID          string
	InstallationID string
}

// NewPatcherTask creates a new patching job task based on PatcherTask structure and returns it.
func NewPatcherTask(runner *Runner, opts *TaskOpts, driver Driver, provider Provider, signer Signer) *PatcherTask {
	return &PatcherTask{
		driver:   driver,
		provider: provider,
		signer:   signer,
		opts:     opts,
		runner:   runner,
	}
}

// PatcherTask.Run creates a new patcher job based on the data passed to it
// and then triggers that job using the specified driver in the task.
func (p *PatcherTask) Run(ctx context.Context, req *PatcherRunRequest) error {
	remoteURL, err := p.provider.RemoteURL(req.AppID, req.Run.VCSMeta.RemoteURL, map[string]interface{}{"installation_id": req.InstallationID})
	if err != nil {
		return err
	}

	token, err := p.signer.GenerateToken(p.runner.ID, []string{ScopeAutofix}, nil, 30*time.Minute)
	if err != nil {
		return err
	}

	req.Run.VCSMeta.RemoteURL = remoteURL
	job, err := NewPatcherDriverJob(req.Run, &PatcherJobOpts{
		PublisherURL:         p.opts.RemoteHost + patcherPublishPath,
		PublisherToken:       token,
		SnippetStorageType:   p.opts.SnippetStorageType,
		SnippetStorageBucket: p.opts.SnippetStorageBucket,
		SentryDSN:            p.opts.SentryDSN,
		KubernetesOpts:       p.opts.KubernetesOpts,
	})
	if err != nil {
		return err
	}
	return p.driver.TriggerJob(ctx, job)
}
