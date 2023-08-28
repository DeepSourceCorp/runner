package orchestrator

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	artifact "github.com/DeepSourceCorp/artifacts/types"
)

const (
	analysisJobPrefix = "analysis-s"
)

var VolumeMounts = map[string]string{
	"codedir":      "/code",
	"artifactsdir": "/artifacts",
	"ssh":          "/home/runner/.ssh",
	"marvindir":    "/marvin",
}

// AnalysisDriverJob is a struct that implements the IDriverJob interface.
type AnalysisDriverJob struct {
	run                   *artifact.AnalysisRun
	check                 *artifact.Check
	analysisConfigBytes   []byte
	deepsourceConfigBytes []byte

	opts *AnalysisOpts
}

type AnalysisOpts struct {
	PublisherURL         string
	PublisherToken       string
	SnippetStorageType   string
	SnippetStorageBucket string

	SentryDSN string

	KubernetesOpts *KubernetesOpts
}

func NewAnalysisDriverJob(run *artifact.AnalysisRun, check artifact.Check, opts *AnalysisOpts) (JobCreator, error) {
	analyisConfig := NewMarvinAnalysisConfig(run, check)
	analysisConfigBytes, err := analyisConfig.Bytes()
	if err != nil {
		return nil, err
	}

	deepsourceConfigBytes, err := json.Marshal(run.Config)
	if err != nil {
		return nil, err
	}

	return &AnalysisDriverJob{
		run:                   run,
		check:                 &check,
		analysisConfigBytes:   analysisConfigBytes,
		deepsourceConfigBytes: deepsourceConfigBytes,
		opts:                  opts,
	}, nil
}

func (j *AnalysisDriverJob) Name() string {
	return analysisJobPrefix + j.run.RunSerial + "-" + j.run.RunID + "-" + j.check.CheckSeq
}

func (j *AnalysisDriverJob) Namespace() string {
	return j.opts.KubernetesOpts.Namespace
}

func (j *AnalysisDriverJob) JobLabels() map[string]string {
	return map[string]string{
		LabelNameApp:     j.Name(),
		LabelNameManager: "runner",
	}
}

func (j *AnalysisDriverJob) PodLabels() map[string]string {
	return map[string]string{
		LabelNameApp:      j.Name(),
		LabelNameRole:     "analysis",
		LabelNameAnalyzer: j.check.AnalyzerMeta.Shortcode,
	}
}

func (*AnalysisDriverJob) Volumes() []string {
	var volumes []string

	for k := range VolumeMounts {
		volumes = append(volumes, k)
	}
	return volumes
}

func (j *AnalysisDriverJob) NodeSelector() map[string]string {
	return j.opts.KubernetesOpts.NodeSelector
}

func (j *AnalysisDriverJob) Container() *Container {
	return &Container{
		Name:  "marvin",
		Image: j.getMarvinImageURL(),
		Limit: Resource{
			CPU:    j.check.AnalyzerMeta.CPULimit + "m",
			Memory: j.check.AnalyzerMeta.MemoryLimit + "Mi",
		},
		Requests: Resource{
			CPU:    j.check.AnalyzerMeta.CPULimit + "m",
			Memory: j.check.AnalyzerMeta.MemoryLimit + "Mi",
		},
		Cmd: []string{"/bin/sh"},
		Args: []string{
			"-c",
			strings.Join(
				[]string{
					MarvinCmdCpy,
					MarvinCmdBase,
					MarvinModeAnalyze,
					fmt.Sprintf("'%s'", string(j.analysisConfigBytes)),
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
			EnvNameMemoryLimit:              j.check.AnalyzerMeta.MemoryLimit + "Mi",
			EnvNameCPULimit:                 j.check.AnalyzerMeta.CPULimit + "m",
			EnvNameTimeLimit:                "1500",
			EnvNameOnPrem:                   "true",
			EnvNamePublisher:                "http",
			EnvNamePublisherURL:             j.opts.PublisherURL,
			EnvNamePublisherToken:           j.opts.PublisherToken,
			EnvNameResultTask:               AnalysisResultTask,
			EnvNameArtifactsCredentialsPath: "/credentials/credentials",
			EnvNameSentryDSN:                j.opts.SentryDSN,
		},
		VolumeMounts: VolumeMounts,
	}
}

func (j *AnalysisDriverJob) InitContainer() *Container {
	artifactsStr, err := json.Marshal(j.check.Artifacts)
	if err != nil {
		log.Println("cannot convert artifacts to json string,error=", err)
		artifactsStr = []byte("[]")
	}
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
			CoatArgNameCheckSeq, j.check.CheckSeq,
			CoatArgNameRemoteURL, j.run.VCSMeta.RemoteURL,
			CoatArgNameBaseBranch, j.run.VCSMeta.BaseBranch,
			CoatArgNameCheckoutOid, j.run.VCSMeta.CheckoutOID,
			CoatArgNameCloneSubmodules, strconv.FormatBool(j.run.VCSMeta.CloneSubmodules),
			CoatArgArtifacts, string(artifactsStr),
			CoatArgNameDecryptRemote, strconv.FormatBool(false),
		},
		Env: map[string]string{
			EnvNameSSHPrivateKey:            j.run.Keys.SSH.Private,
			EnvNameSSHPublicKey:             j.run.Keys.SSH.Public,
			EnvNamePublisher:                "http",
			EnvNamePublisherURL:             j.opts.PublisherURL,
			EnvNamePublisherToken:           j.opts.PublisherToken,
			EnvNameArtifactsCredentialsPath: "/credentials/credentials",
			EnvNameResultTask:               AnalysisResultTask,
			EnvNameSentryDSN:                j.opts.SentryDSN,
		},
		VolumeMounts: VolumeMounts,
	}
}

func (j *AnalysisDriverJob) ImagePullSecrets() []string {
	return j.opts.KubernetesOpts.ImagePullSecrets
}

func (j *AnalysisDriverJob) getMarvinImageURL() string {
	imagePrefix := ""
	if j.check.AnalyzerMeta.AnalyzerType == "core" {
		imagePrefix = "marvin-"
	}
	return j.opts.KubernetesOpts.ImageURL.JoinPath(fmt.Sprintf(
		"/%s%s:%s",
		imagePrefix,
		j.check.AnalyzerMeta.Shortcode,
		j.check.AnalyzerMeta.Version,
	)).String()
}

func (j *AnalysisDriverJob) getCoatImageURL() string {
	return j.opts.KubernetesOpts.ImageURL.JoinPath(fmt.Sprintf(
		"/coat:%s",
		CoatVersion,
	)).String()
}
