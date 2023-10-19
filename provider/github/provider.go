package github

import (
	"net/http"

	"github.com/deepsourcecorp/runner/provider/common"
)

type ProviderOpts struct {
	Apps       map[string]*App
	Runner     *common.Runner
	DeepSource *common.DeepSource
	Client     *http.Client
}
