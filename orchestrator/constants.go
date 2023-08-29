package orchestrator

const (
	CoatVersion = "latest"

	LabelNameManager  = "manager"
	LabelNameRole     = "role"
	LabelNameApp      = "application"
	LabelNameAnalyzer = "analyzer"

	EnvNameCodePath                 = "CODE_PATH"
	EnvNameToolboxPath              = "TOOLBOX_PATH"
	EnvNameMemoryLimit              = "MEMORY_LIMIT"
	EnvNameCPULimit                 = "CPU_LIMIT"
	EnvNameTimeLimit                = "TIME_LIMIT"
	EnvNameOnPrem                   = "ON_PREM"
	EnvNameSSHPrivateKey            = "SSH_PRIVATE_KEY"
	EnvNameSSHPublicKey             = "SSH_PUBLIC_KEY"
	EnvNamePublisher                = "PUBLISHER"
	EnvNamePublisherURL             = "RESULT_HTTP_URL"
	EnvNamePublisherToken           = "RESULT_HTTP_TOKEN"
	EnvNameResultTask               = "RESULT_RMQ_TASK"
	EnvNameArtifactsCredentialsPath = "ARTIFACTS_CREDENTIALS_PATH"
	EnvNameArtifactsSecretName      = "TASK_ARTIFACT_SECRET_NAME"
	EnvNameSentryDSN                = "SENTRY_DSN"

	MarvinCmdCpy       = "cp /marvin/marvin /toolbox &&"
	MarvinCmdBase      = "/toolbox/marvin"
	MarvinCmdArgConfig = "--config"

	MarvinModeAnalyze   = "--analyze"
	MarvinModeAutofix   = "--autofix"
	MarvinModeTransform = "--transform"

	MarvinSnippetStorageType   = "--snippet-storage-type"
	MarvinSnippetStorageBucket = "--snippet-storage-bucket"

	CoatCmdName = "/app/coat"

	CoatArgNameRunID            = "--run-id"
	CoatArgNameCheckSeq         = "--check-seq"
	CoatArgNameRemoteURL        = "--remote-url"
	CoatArgNameCheckoutOid      = "--checkout-oid"
	CoatArgNameCloneSubmodules  = "--clone-submodules"
	CoatArgTestCoverageArtifact = "--artifacts"
	CoatArgNameDecryptRemote    = "--decrypt-remote-url"
	CoatArgSnippetStorageBucket = "--snippet-storage-bucket"
	CoatArgSnippetStorageType   = "--snippet-storage-type"
	CoatArgNameBaseBranch       = "--base-branch"
	CoatArgPatchMeta            = "--patch-meta"
	CoatArgArtifacts            = "--artifacts"

	CoatCPULimit      = "1400m"
	CoatMemoryLimit   = "4000Mi"
	CoatCPURequest    = "300m"
	CoatMemoryRequest = "500Mi"

	ScopeAnalysis  = "analysis.*"
	ScopeAutofix   = "autofix.*"
	ScopeTransform = "transform.*"

	AnalysisResultTask    = "contrib.atlas.tasks.store_analysis_run_result"
	AutofixResultTask     = "contrib.atlas.tasks.store_autofix_run_result"
	TransformerResultTask = "contrib.atlas.tasks.store_transformer_run_result"
	CancelCheckResultTask = "contrib.atlas.tasks.confirm_check_cancellation"
	PatcherResultTask     = "contrib.runner.tasks.store_autofix_committer_result"
)
