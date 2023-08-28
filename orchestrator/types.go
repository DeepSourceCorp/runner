package orchestrator

import (
	"errors"
	"net/url"
	"time"
)

const (
	DriverTypePrinter = "printer"
	DriverTypeK8s     = "k8s"
)

var ErrUnknownStorageProvider = errors.New("unknown storage provider")

type Storer interface {
	GenerateURL(bucket, object string) (string, error)
}

type Signer interface {
	GenerateToken(issuer string, scope []string, claims map[string]interface{}, expiry time.Duration) (string, error)
}

type KubernetesOpts struct {
	Namespace        string
	NodeSelector     map[string]string
	ImageURL         url.URL
	ImagePullSecrets []string
}

type TaskOpts struct {
	RemoteHost           string
	SnippetStorageType   string
	SnippetStorageBucket string
	
	SentryDSN            string

	KubernetesOpts *KubernetesOpts
}
