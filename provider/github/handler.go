package github

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/model"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

const ()

type Handler struct {
	apiService     *APIService
	webhookService *WebhookService

	appFactory *AppFactory

	httpClient *http.Client
}

func NewHandler(
	webhookService *WebhookService,
	apiService *APIService,
	appFactory *AppFactory,
	_ *model.Runner,
	_ *model.DeepSource,
	httpClient *http.Client) (*Handler, error) {
	return &Handler{
		apiService:     apiService,
		webhookService: webhookService,

		appFactory: appFactory,

		httpClient: httpClient,
	}, nil
}

func (h *Handler) HandleAPI(c echo.Context) error {
	req, err := NewAPIRequest(c)
	if err != nil {
		return httperror.ErrMissingParams(err)
	}
	res, err := h.apiService.Process(req)
	if err != nil {
		return httperror.ErrUpstreamFailed(err)
	}

	return h.writeResponse(c, res)
}

// HandleWebhook handles the webhook request from Github to DeepSource Cloud.
func (h *Handler) HandleWebhook(c echo.Context) error {
	req, err := NewWebhookRequest(c)
	if err != nil {
		return httperror.ErrMissingParams(err)
	}

	res, err := h.webhookService.Process(req)
	if err != nil {
		return err
	}
	return h.writeResponse(c, res)
}

func (*Handler) writeResponse(c echo.Context, res *http.Response) error {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("failed to read response body", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	c.Response().Writer.WriteHeader(res.StatusCode)
	if _, err := c.Response().Writer.Write(body); err != nil {
		slog.Error("failed to write response body", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	c.Response().Flush()
	return nil
}

type InstallationRequest struct {
	AppID string `param:"app_id"`
}

// HandleInstallation redirects the user to the installation page on Github.
func (h *Handler) HandleInstallation(c echo.Context) error {
	req := &InstallationRequest{}
	if err := c.Bind(req); err != nil {
		slog.Error("failed to bind request", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	app := h.appFactory.GetApp(req.AppID)
	if app == nil {
		slog.Error("app not found", slog.Any("app_id", req.AppID))
		return httperror.ErrAppInvalid(nil)
	}
	return c.Redirect(http.StatusTemporaryRedirect, app.InstallationURL())
}

func (h *Handler) AuthenticatedRemoteURL(appID, installationID string, srcURL string) (string, error) {
	app := h.appFactory.GetApp(appID)
	if app == nil {
		return "", ErrAppNotFound
	}

	installationClient := NewInstallationClient(app, installationID, h.httpClient)

	token, err := installationClient.AccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate authenticated remote url: %w", err)
	}

	u, err := url.Parse(srcURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse url: %w", err)
	}
	u.User = url.UserPassword("x-access-token", token)
	return u.String(), nil
}
