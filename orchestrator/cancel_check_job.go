package orchestrator

import (
	artifact "github.com/DeepSourceCorp/artifacts/types"
)

// The prefix of the job to be cancelled.
const (
	jobPrefix = "analysis-s"
)

type CancelCheckDriverJob struct {
	run  *artifact.CancelCheckRun
	opts *CancelCheckOpts
}

type CancelCheckOpts struct {
	KubernetesOpts *KubernetesOpts
}

func NewCancelCheckDriverJob(run *artifact.CancelCheckRun, opts *CancelCheckOpts) (JobDeleter, error) {
	return &CancelCheckDriverJob{
		run:  run,
		opts: opts,
	}, nil
}

func (j *CancelCheckDriverJob) Name() string {
	return jobPrefix + j.run.AnalysisMeta.RunSerial + "-" + j.run.AnalysisMeta.RunID + "-" + j.run.AnalysisMeta.CheckSeq
}

func (j *CancelCheckDriverJob) Namespace() string {
	return j.opts.KubernetesOpts.Namespace
}
