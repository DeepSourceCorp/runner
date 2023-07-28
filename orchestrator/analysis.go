package orchestrator

import (
	"context"
	"log"
	"sync"
	"time"

	artifact "github.com/deepsourcelabs/artifacts/types"
	"golang.org/x/exp/slog"
)

const (
	analysisPublishPath = "/api/runner/analysis/results"
)

type AnalysisTask struct {
	driver   Driver
	provider Provider
	signer   Signer
	opts     *TaskOpts
}

func NewAnalysisTask(opts *TaskOpts, driver Driver, provider Provider, signer Signer) *AnalysisTask {
	return &AnalysisTask{
		opts:     opts,
		driver:   driver,
		signer:   signer, // used for generating the auth token.
		provider: provider,
	}
}

type AnalysisRunRequest struct {
	Run            *artifact.AnalysisRun
	AppID          string
	InstallationID string
}

// Run executes the analysis task for the given analysis run.
// For each check in the run, it creates a new analysis driver job
// and triggers the job in a separate goroutine. The function waits
// for all jobs to complete before returning.
//
// The context is used to control the overall execution of the task.
// The run parameter contains the information about the analysis run
// to be executed.
//
// If any of the driver jobs fail, Run logs the error and returns it.
// If all jobs complete successfully, Run returns nil.
//
// Example usage:
//
//	err := task.Run(ctx, run)
//	if err != nil {
//	  log.Fatal(err)
//	}
//
// Run is safe for concurrent use.
func (t *AnalysisTask) Run(ctx context.Context, req *AnalysisRunRequest) error {
	remoteURL, err := t.provider.AuthenticatedRemoteURL(req.AppID, req.InstallationID, req.Run.VCSMeta.RemoteURL)
	if err != nil {
		return err
	}
	req.Run.VCSMeta.RemoteURL = remoteURL
	var wg sync.WaitGroup // initialize waitgroup
	for _, check := range req.Run.Checks {
		slog.Info("creating analysis job for check", check.CheckSeq)

		token, err := t.signer.GenerateToken("", []string{ScopeAnalysis}, nil, 30*time.Minute)
		if err != nil {
			slog.Error("failed to generate token for analysis job, err= %v", err)
			return err
		}

		log.Printf("creating analysis job for check %s", check.CheckSeq)
		job, err := NewAnalysisDriverJob(
			req.Run,
			check,
			&AnalysisOpts{
				PublisherURL:         t.opts.RemoteHost + analysisPublishPath,
				PublisherToken:       token,
				SnippetStorageType:   t.opts.SnippetStorageType,
				SnippetStorageBucket: t.opts.SnippetStorageBucket,
				KubernetesOpts:       t.opts.KubernetesOpts,
			},
		)
		if err != nil {
			slog.Error("failed to create analysis job for check sequence= %d, err= %v", check.CheckSeq, err)
			return err
		}

		wg.Add(1) // add to waitgroup
		go func(job JobCreator) {
			defer wg.Done() // mark job as done when function completes
			if err = t.driver.TriggerJob(ctx, job); err != nil {
				slog.Error("failed to trigger analysis job, name= %d, err= %v", job.Name(), err)
			}
		}(job)
	}
	wg.Wait() // wait for all jobs to complete
	return nil
}
