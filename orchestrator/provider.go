package orchestrator

type Provider interface {
	AuthenticatedRemoteURL(appID, installationID string, srcURL string) (string, error)
}
