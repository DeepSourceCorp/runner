package orchestrator

type Provider interface {
	RemoteURL(appID string, sourceURL string, extra map[string]interface{}) (string, error)
}
