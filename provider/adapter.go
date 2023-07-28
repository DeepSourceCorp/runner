package provider

import (
	"fmt"
	"net/http"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/github"
	"github.com/labstack/echo/v4"
)

var (
	ErrNoVCS = httperror.Error{Message: "no VCS provider found for the given app ID"}
	ErrNoApp = httperror.Error{Message: "no app found for the given app ID"}
)

// Adapter is a unified interface for all VCS providers.  We maintain a map
// of app IDs to VCS providers, and the appropriate provider is chosen based
// on the app ID in the request.  The adapter then delegates the request to
// the chosen provider.
type Adapter struct {
	providers map[string]Provider
	apps      map[string]*App
}

// New creates a new provider facade.
func NewAdapter(apps map[string]*App, githubProvider *github.Handler) *Adapter {
	return &Adapter{
		providers: map[string]Provider{
			"github": githubProvider,
		},
		apps: apps,
	}
}

// HandleAPI handles API requests for a specific app.
func (a *Adapter) HandleAPI(c echo.Context) error {
	provider, err := a.getProvider(c.Param("app_id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrNoVCS)
	}
	return provider.HandleAPI(c)
}

// HandleWebhook handles webhook requests for a specific app.
func (a *Adapter) HandleWebhook(c echo.Context) error {
	provider, err := a.getProvider(c.Param("app_id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrNoVCS)
	}
	return provider.HandleWebhook(c)
}

// HandleInstallation handles installation requests for a specific app.
// This is only implemented by some providers.
func (a *Adapter) HandleInstallation(c echo.Context) error {
	provider, err := a.getProvider(c.Param("app_id"))
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrNoVCS)
	}
	return provider.HandleInstallation(c)
}

// AuthenticatedRemoteURL returns an authenticated remote URL for a specific app,
// installation, and source URL.
func (a *Adapter) AuthenticatedRemoteURL(appID, installationID, srcURL string) (string, error) {
	provider, err := a.getProvider(appID)
	if err != nil {
		return "", fmt.Errorf("failed to generate remote url: %w", err)
	}
	return provider.AuthenticatedRemoteURL(appID, installationID, srcURL)
}

// getProvider retrieves the VCS provider based on the given appID.
func (a *Adapter) getProvider(appID string) (Provider, error) {
	app := a.apps[appID]
	if app == nil {
		return nil, ErrNoApp
	}
	return a.providers[app.Provider], nil
}

// Provider interface defines the methods that each VCS provider should implement.
type Provider interface {
	HandleAPI(c echo.Context) error
	HandleWebhook(c echo.Context) error
	HandleInstallation(c echo.Context) error
	AuthenticatedRemoteURL(appID, installationID, srcURL string) (string, error)
}

// App represents an application with a specific VCS provider.
type App struct {
	Provider string
}
