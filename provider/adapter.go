package provider

import (
	"errors"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/router"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

// Provider interface defines the methods that each VCS provider should implement.
type Handler interface {
	HandleAPI(c echo.Context) error
	HandleWebhook(c echo.Context) error
	HandleInstallation(c echo.Context) error
}

type Authenticator interface {
	RemoteURL(appID string, sourceURL string, extra map[string]interface{}) (string, error)
}

var (
	ErrNoProvider      = errors.New("no provider found for app")
	ErrNoAuthenticator = errors.New("no authenticator found for app")
)

// Adapter is a unified interface for all VCS providers.  We maintain a map
// of app IDs to VCS providers, and the appropriate provider is chosen based
// on the app ID in the request.  The adapter then delegates the request to
// the chosen provider.
type Adapter struct {
	handlers       map[string]Handler
	authenticators map[string]Authenticator
}

// New creates a new provider facade.
func NewAdapter(handlers map[string]Handler, authenticators map[string]Authenticator) *Adapter {
	return &Adapter{
		handlers:       handlers,
		authenticators: authenticators,
	}
}

func (a *Adapter) AddRoutes(r router.Router) router.Router {
	r.AddRoute("*", "apps/:app_id/webhook", a.HandleWebhook)
	r.AddRoute("*", "apps/:app_id/api/*", a.HandleAPI)
	r.AddRoute("*", "apps/:app_id/installation/new", a.HandleInstallation)
	return r
}

// HandleAPI handles API requests for a specific app.
func (a *Adapter) HandleAPI(c echo.Context) error {
	provider := a.handlers[c.Param("app_id")]
	if provider == nil {
		slog.Error("no provider found for app", slog.Any("app_id", c.Param("app_id")))
		return httperror.ErrBadRequest(ErrNoProvider)
	}
	return provider.HandleAPI(c)
}

// HandleWebhook handles webhook requests for a specific app.
func (a *Adapter) HandleWebhook(c echo.Context) error {
	provider := a.handlers[c.Param("app_id")]
	if provider == nil {
		slog.Error("no provider found for app", slog.Any("app_id", c.Param("app_id")))
		return httperror.ErrBadRequest(ErrNoProvider)
	}
	return provider.HandleWebhook(c)
}

// HandleInstallation handles installation requests for a specific app.
// This is only implemented by some providers.
func (a *Adapter) HandleInstallation(c echo.Context) error {
	provider := a.handlers[c.Param("app_id")]
	if provider == nil {
		slog.Error("no provider found for app", slog.Any("app_id", c.Param("app_id")))
		return httperror.ErrBadRequest(ErrNoProvider)
	}
	return provider.HandleInstallation(c)
}

// AuthenticatedRemoteURL returns an authenticated remote URL for a specific app,
// installation, and source URL.
func (a *Adapter) RemoteURL(appID string, sourceURL string, extra map[string]any) (string, error) {
	authenticator := a.authenticators[appID]
	if authenticator == nil {
		slog.Error("no authenticator found for app", slog.Any("app_id", appID))
		return "", httperror.ErrBadRequest(ErrNoAuthenticator)
	}
	return authenticator.RemoteURL(appID, sourceURL, extra)
}

// App represents an application with a specific VCS provider.
type App struct {
	Provider string
}
