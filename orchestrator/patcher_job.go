package orchestrator

import (
	"encoding/json"
	"fmt"
	"strconv"

	artifact "github.com/deepsourcelabs/artifacts/types"
)

const (
	patcherJobPrefix = "patcher-s"
)

// PatcherDriverJob represents the patcher job and the data required by it
// in the form of config or opts.
type PatcherDriverJob struct {
	run          *artifact.PatcherRun
	artifactData string

	opts *PatcherJobOpts
}

// PatcherJobOpts represents the data that needs to be passed to the patcher job like the results URL.
type PatcherJobOpts struct {
	PublisherURL         string
	PublisherToken       string
	SnippetStorageType   string
	SnippetStorageBucket string

	KubernetesOpts *KubernetesOpts
}

// NewPatcherDriverJob is responsible for creating the patcher run config and then returning an instance
// of the patcher job.
func NewPatcherDriverJob(run *artifact.PatcherRun, opts *PatcherJobOpts) (JobCreator, error) {
	artifactJSON, err := json.Marshal(run.Artifacts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal artifacts json,error=%v", err)
	}

	return &PatcherDriverJob{
		run:          run,
		artifactData: string(artifactJSON),
		opts:         opts,
	}, nil
}

func (j *PatcherDriverJob) taskID() string {
	return patcherJobPrefix + j.run.RunSerial + "-" + j.run.RunID
}

func (j *PatcherDriverJob) Name() string {
	return j.taskID()
}

func (j *PatcherDriverJob) Namespace() string {
	return j.opts.KubernetesOpts.Namespace
}

func (j *PatcherDriverJob) JobLabels() map[string]string {
	return map[string]string{
		LabelNameManager: "runner",
		LabelNameApp:     j.taskID(),
	}
}

func (j *PatcherDriverJob) NodeSelector() map[string]string {
	return j.opts.KubernetesOpts.NodeSelector
}

func (j *PatcherDriverJob) PodLabels() map[string]string {
	return map[string]string{
		LabelNameApp:  j.taskID(),
		LabelNameRole: "patcher",
	}
}

func (*PatcherDriverJob) Volumes() []string {
	var volumes []string

	for k := range VolumeMounts {
		volumes = append(volumes, k)
	}
	return volumes
}

func (j *PatcherDriverJob) Container() *Container {
	return &Container{
		Name:  "coat",
		Image: j.getCoatImageURL(),
		Limit: Resource{
			CPU:    CoatCPULimit,
			Memory: CoatMemoryLimit,
		},
		Requests: Resource{
			CPU:    CoatCPURequest,
			Memory: CoatMemoryRequest,
		},
		Cmd: []string{CoatCmdName},
		Args: []string{
			CoatArgNameRunID, j.run.RunID,
			CoatArgNameRemoteURL, j.run.VCSMeta.RemoteURL,
			CoatArgNameBaseBranch, j.run.VCSMeta.BaseBranch,
			CoatArgNameCheckoutOid, j.run.VCSMeta.CheckoutOID,
			CoatArgNameCloneSubmodules, strconv.FormatBool(j.run.VCSMeta.CloneSubmodules),
			CoatArgNameDecryptRemote, strconv.FormatBool(false),
			CoatArgPatchMeta, j.run.PatchMeta,
			CoatArgArtifacts, j.artifactData,
			CoatArgSnippetStorageType, j.opts.SnippetStorageType,
			CoatArgSnippetStorageBucket, j.opts.SnippetStorageBucket,
		},
		Env: map[string]string{
			EnvNameCodePath:                 "/code",
			EnvNameToolboxPath:              "/toolbox",
			EnvNameArtifactsCredentialsPath: "/credentials/credentials",
			EnvNameSSHPrivateKey:            j.run.Keys.SSH.Private,
			EnvNameSSHPublicKey:             j.run.Keys.SSH.Public,
			EnvNameTimeLimit:                "1500",
			EnvNameOnPrem:                   "true",
			EnvNamePublisher:                "http",
			EnvNamePublisherURL:             j.opts.PublisherURL,
			EnvNamePublisherToken:           j.opts.PublisherToken,
			EnvNameResultTask:               PatcherResultTask,
		},
		VolumeMounts: VolumeMounts,
	}
}

func (j *PatcherDriverJob) ImagePullSecrets() []string {
	return j.opts.KubernetesOpts.ImagePullSecrets
}

func (*PatcherDriverJob) InitContainer() *Container {
	return nil
}

func (j *PatcherDriverJob) getCoatImageURL() string {
	return j.opts.KubernetesOpts.ImageURL.JoinPath(fmt.Sprintf(
		"coat:%s",
		CoatVersion,
	)).String()
}
