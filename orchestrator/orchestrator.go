package orchestrator

import (
	"errors"
	"net/url"
	"time"
)

var ErrUnknownStorageProvider = errors.New("unknown storage provider")

type Storer interface {
	GenerateURL(bucket, object string) (string, error)
}

type Signer interface {
	GenerateToken(scope []string, expiry time.Duration) (string, error)
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

	KubernetesOpts *KubernetesOpts
}
