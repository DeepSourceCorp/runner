package orchestrator

import (
	"context"
	"time"
)

type Resource struct {
	CPU    string
	Memory string
}

type Container struct {
	Env          map[string]string
	VolumeMounts map[string]string
	Limit        Resource
	Requests     Resource
	Name         string
	Image        string
	Cmd          []string
	Args         []string
}

type Driver interface {
	TriggerJob(ctx context.Context, request JobCreator) error
	DeleteJob(ctx context.Context, request JobDeleter) error
	CleanExpiredJobs(ctx context.Context, namespace string, interval *time.Duration) error
}

// Type JobCreator interface defines methods to access data required for a job creation
// irrespective of the driver implementation.
type JobCreator interface {
	Name() string
	Namespace() string

	JobLabels() map[string]string
	PodLabels() map[string]string

	Volumes() []string

	Container() *Container
	InitContainer() *Container

	NodeSelector() map[string]string

	ImagePullSecrets() []string
}

// Type JobDeleter interface defines methods to access data required for a job deletion
// irrespective of the driver implementation.
type JobDeleter interface {
	Name() string
	Namespace() string
}
