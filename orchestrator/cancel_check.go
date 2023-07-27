package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	artifact "github.com/deepsourcelabs/artifacts/types"
	"github.com/google/uuid"
)

const (
	cancelCheckPublishPath = "/api/runner/cancel-check/results"
)

type CancelCheckTask struct {
	opts   *TaskOpts
	driver Driver
	signer Signer
	client *http.Client
}

// NewCancelCheckTask registers a new cancel check task with the supplied properties of
// driver, provider facade and license store.
func NewCancelCheckTask(opts *TaskOpts, driver Driver, signer Signer, client *http.Client) *CancelCheckTask {
	t := &CancelCheckTask{
		opts:   opts,
		driver: driver,
		signer: signer,
		client: client,
	}
	if t.client == nil {
		t.client = http.DefaultClient
	}
	return t
}

// Run creates the template for the cancel check job and triggers the cancel check job.
func (t *CancelCheckTask) Run(ctx context.Context, run *artifact.CancelCheckRun) error {
	job, err := NewCancelCheckDriverJob(run, &CancelCheckOpts{
		KubernetesOpts: t.opts.KubernetesOpts,
	})
	if err != nil {
		return err
	}

	result := artifact.CancelCheckResult{}
	if err := t.driver.DeleteJob(ctx, job); err != nil {
		result.RunID = run.RunID
		result.Status = artifact.Status{
			Code:     5001,
			HMessage: "Error cancelling the check",
			Err:      err.Error(),
		}
	} else {
		result.RunID = run.RunID
		result.Status = artifact.Status{
			Code:     5000,
			HMessage: "Check cancelled successfully",
			Err:      "",
		}
	}

	payload := artifact.CancelCheckResultCeleryTask{
		ID:      uuid.NewString(),
		Task:    "cancel-check",
		KWArgs:  result,
		Retries: 0,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	publisherURL := t.opts.RemoteHost + cancelCheckPublishPath
	publisherToken, err := t.signer.GenerateToken([]string{ScopeAnalysis}, nil, 30*time.Minute)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", publisherURL, bytes.NewReader(payloadJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+publisherToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
