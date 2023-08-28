package orchestrator

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	artifact "github.com/DeepSourceCorp/artifacts/types"
)

const (
	autofixJobPrefix = "autofix-s"
)

type AutofixDriverJob struct {
	run                   *artifact.AutofixRun
	autofixConfigBytes    []byte
	deepsourceConfigBytes []byte

	opts *AutofixOpts
}

type AutofixOpts struct {
	PublisherURL         string
	PublisherToken       string
	SnippetStorageType   string
	SnippetStorageBucket string

	KubernetesOpts *KubernetesOpts
}

func NewAutofixDriverJob(run *artifact.AutofixRun, opts *AutofixOpts) (JobCreator, error) {
	autofixConfig, err := NewAutofixConfig(run)
	if err != nil {
		return nil, err
	}
	autofixConfigBytes, err := autofixConfig.Bytes()
	if err != nil {
		return nil, err
	}

	deepsourceConfigBytes, err := json.Marshal(run.Config)
	if err != nil {
		return nil, err
	}
	return &AutofixDriverJob{
		run:                   run,
		autofixConfigBytes:    autofixConfigBytes,
		deepsourceConfigBytes: deepsourceConfigBytes,

		opts: opts,
	}, nil
}

func (j *AutofixDriverJob) Name() string {
	return autofixJobPrefix + j.run.RunSerial + "-" + j.run.RunID + "-1"
}

func (j *AutofixDriverJob) Namespace() string {
	return j.opts.KubernetesOpts.Namespace
}

func (j *AutofixDriverJob) NodeSelector() map[string]string {
	return j.opts.KubernetesOpts.NodeSelector
}

func (j *AutofixDriverJob) JobLabels() map[string]string {
	return map[string]string{
		LabelNameManager: "runner",
		LabelNameApp:     j.Name(),
	}
}

func (j *AutofixDriverJob) PodLabels() map[string]string {
	return map[string]string{
		LabelNameApp:      j.Name(),
		LabelNameRole:     "autofix",
		LabelNameAnalyzer: j.run.Autofixer.AutofixMeta.Shortcode,
	}
}

func (*AutofixDriverJob) Volumes() []string {
	var volumes []string

	for k := range VolumeMounts {
		volumes = append(volumes, k)
	}
	return volumes
}

func (j *AutofixDriverJob) InitContainer() *Container {
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
			CoatArgNameCheckSeq, "1",
			CoatArgNameRemoteURL, j.run.VCSMeta.RemoteURL,
			CoatArgNameBaseBranch, j.run.VCSMeta.BaseBranch,
			CoatArgNameCheckoutOid, j.run.VCSMeta.CheckoutOID,
			CoatArgNameCloneSubmodules, strconv.FormatBool(j.run.VCSMeta.CloneSubmodules),
			CoatArgNameDecryptRemote, strconv.FormatBool(false),
		},
		Env: map[string]string{
			EnvNameSSHPrivateKey:            j.run.Keys.SSH.Private,
			EnvNameSSHPublicKey:             j.run.Keys.SSH.Public,
			EnvNamePublisher:                "http",
			EnvNamePublisherURL:             j.opts.PublisherURL,
			EnvNamePublisherToken:           j.opts.PublisherToken,
			EnvNameResultTask:               AutofixResultTask,
			EnvNameArtifactsCredentialsPath: "/credentials/credentials",
		},
		VolumeMounts: VolumeMounts,
	}
}

func (j *AutofixDriverJob) Container() *Container {
	return &Container{
		Name:  "marvin",
		Image: j.getMarvinImageURL(),
		Limit: Resource{
			CPU:    j.run.Autofixer.AutofixMeta.CPULimit + "m",
			Memory: j.run.Autofixer.AutofixMeta.MemoryLimit + "Mi",
		},
		Requests: Resource{
			CPU:    j.run.Autofixer.AutofixMeta.CPULimit + "m",
			Memory: j.run.Autofixer.AutofixMeta.MemoryLimit + "Mi",
		},
		Cmd: []string{"/bin/sh"},
		Args: []string{
			"-c",
			strings.Join(
				[]string{
					MarvinCmdCpy,
					MarvinCmdBase,
					MarvinModeAutofix,
					fmt.Sprintf("'%s'", string(j.autofixConfigBytes)),
					MarvinCmdArgConfig,
					fmt.Sprintf("'%s'", string(j.deepsourceConfigBytes)),
					MarvinSnippetStorageType,
					j.opts.SnippetStorageType,
					MarvinSnippetStorageBucket,
					j.opts.SnippetStorageBucket,
				}, " "),
		},
		Env: map[string]string{
			EnvNameCodePath:                 "/code",
			EnvNameToolboxPath:              "/toolbox",
			EnvNameMemoryLimit:              j.run.Autofixer.AutofixMeta.MemoryLimit + "Mi",
			EnvNameCPULimit:                 j.run.Autofixer.AutofixMeta.CPULimit + "m",
			EnvNameTimeLimit:                "1500",
			EnvNameOnPrem:                   "true",
			EnvNamePublisher:                "http",
			EnvNamePublisherURL:             j.opts.PublisherURL,
			EnvNamePublisherToken:           j.opts.PublisherToken,
			EnvNameResultTask:               AutofixResultTask,
			EnvNameArtifactsCredentialsPath: "/credentials/credentials",
		},
		VolumeMounts: VolumeMounts,
	}
}

func (j *AutofixDriverJob) ImagePullSecrets() []string {
	return j.opts.KubernetesOpts.ImagePullSecrets
}

func (j *AutofixDriverJob) getMarvinImageURL() string {
	return j.opts.KubernetesOpts.ImageURL.JoinPath(fmt.Sprintf(
		"/%s%s:%s",
		"marvin-",
		j.run.Autofixer.AutofixMeta.Shortcode,
		j.run.Autofixer.AutofixMeta.Version,
	)).String()
}

func (j *AutofixDriverJob) getCoatImageURL() string {
	return j.opts.KubernetesOpts.ImageURL.JoinPath(fmt.Sprintf(
		"/coat:%s",
		CoatVersion,
	)).String()
}
