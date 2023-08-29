package orchestrator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	artifact "github.com/DeepSourceCorp/artifacts/types"
)

type TransformerJob struct {
	run                    *artifact.TransformerRun
	transformerConfigBytes []byte
	deepsourceConfigBytes  []byte

	opts *TransformerOpts
}

type TransformerOpts struct {
	PublisherURL   string
	PublisherToken string

	SentryDSN string

	KubernetesOpts *KubernetesOpts
}

func NewTransformerJob(run *artifact.TransformerRun, opts *TransformerOpts) (JobCreator, error) {
	transformerConfig := NewTransformerMarvinConfig(run)
	transformerConfigBytes, err := transformerConfig.Bytes()
	if err != nil {
		return nil, err
	}

	deepsourceConfigBytes, err := json.Marshal(run.Config)
	if err != nil {
		return nil, err
	}

	return &TransformerJob{
		run:                    run,
		transformerConfigBytes: transformerConfigBytes,
		deepsourceConfigBytes:  deepsourceConfigBytes,
		opts:                   opts,
	}, nil
}

func (j *TransformerJob) Name() string {
	return "transformer-s" + j.run.RunSerial + "-" + j.run.RunID
}

func (j *TransformerJob) Namespace() string {
	return j.opts.KubernetesOpts.Namespace
}

func (j *TransformerJob) JobLabels() map[string]string {
	return map[string]string{
		LabelNameManager: "runner",
		LabelNameApp:     j.Name(),
	}
}

func (j *TransformerJob) NodeSelector() map[string]string {
	return j.opts.KubernetesOpts.NodeSelector
}

func (j *TransformerJob) PodLabels() map[string]string {
	return map[string]string{
		LabelNameApp:  j.Name(),
		LabelNameRole: "transformer",
	}
}

func (*TransformerJob) Volumes() []string {
	var volumes []string

	for k := range VolumeMounts {
		volumes = append(volumes, k)
	}
	return volumes
}

func (j *TransformerJob) InitContainer() *Container {
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
		Cmd: []string{"/app/coat"},
		Args: []string{
			CoatArgNameRunID, j.run.RunID,
			CoatArgNameCheckSeq, "1",
			CoatArgNameRemoteURL, j.run.VCSMeta.RemoteURL,
			CoatArgNameBaseBranch, j.run.VCSMeta.BaseBranch,
			CoatArgNameCheckoutOid, j.run.VCSMeta.CheckoutOID,
			CoatArgNameCloneSubmodules, strconv.FormatBool(j.run.VCSMeta.CloneSubmodules),
			CoatArgNameDecryptRemote, strconv.FormatBool(false),
		},
		Env: map[string]string{
			EnvNamePublisher:                "http",
			EnvNamePublisherURL:             j.opts.PublisherURL,
			EnvNamePublisherToken:           j.opts.PublisherToken,
			EnvNameResultTask:               TransformerResultTask,
			EnvNameArtifactsCredentialsPath: "/credentials/credentials",
			EnvNameSentryDSN:                j.opts.SentryDSN,
		},
		VolumeMounts: VolumeMounts,
	}
}

func (j *TransformerJob) Container() *Container {
	return &Container{
		Name:  "marvin",
		Image: j.getMarvinImageURL(),
		Limit: Resource{
			CPU:    j.run.Transformer.Meta.CPULimit + "m",
			Memory: j.run.Transformer.Meta.MemoryLimit + "Mi",
		},
		Requests: Resource{
			CPU:    j.run.Transformer.Meta.CPULimit + "m",
			Memory: j.run.Transformer.Meta.MemoryLimit + "Mi",
		},
		Cmd: []string{"/bin/sh"},
		Args: []string{
			"-c",
			strings.Join(
				[]string{
					MarvinCmdCpy,
					MarvinCmdBase,
					MarvinModeTransform,
					fmt.Sprintf("'%s'", string(j.transformerConfigBytes)),
					MarvinCmdArgConfig,
					fmt.Sprintf("'%s'", string(j.deepsourceConfigBytes)),
				}, " "),
		},
		Env: map[string]string{
			EnvNameCodePath:                 "/code",
			EnvNameToolboxPath:              "/toolbox",
			EnvNameArtifactsCredentialsPath: "/credentials/credentials",
			EnvNameMemoryLimit:              j.run.Transformer.Meta.MemoryLimit + "Mi",
			EnvNameCPULimit:                 j.run.Transformer.Meta.CPULimit + "m",
			EnvNameTimeLimit:                "1500",
			EnvNameOnPrem:                   "true",
			EnvNamePublisher:                "http",
			EnvNamePublisherURL:             j.opts.PublisherURL,
			EnvNamePublisherToken:           j.opts.PublisherToken,
			EnvNameResultTask:               TransformerResultTask,
			EnvNameSentryDSN:                j.opts.SentryDSN,
		},
		VolumeMounts: VolumeMounts,
	}
}

func (j *TransformerJob) ImagePullSecrets() []string {
	return j.opts.KubernetesOpts.ImagePullSecrets
}

func (j *TransformerJob) getMarvinImageURL() string {
	return j.opts.KubernetesOpts.ImageURL.JoinPath(fmt.Sprintf(
		"/bumblebee:%s",
		j.run.Transformer.Meta.Version,
	)).String()
}

func (j *TransformerJob) getCoatImageURL() string {
	return j.opts.KubernetesOpts.ImageURL.JoinPath(fmt.Sprintf(
		"coat:%s",
		CoatVersion,
	)).String()
}
