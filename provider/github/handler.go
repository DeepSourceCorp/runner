package github

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

const (
	HeaderInstallationID = "X-Installation-Id"
)

var (
	ErrConfigNotFound = errors.New("config not found")
)

type Handler struct {
	appFactory          *AppFactory
	webhookProxyFactory *WebhookProxyFactory
	httpClient          *http.Client
}

func NewHandler(webhookF *WebhookProxyFactory, appFactory *AppFactory, httpClient *http.Client) (*Handler, error) {
	return &Handler{
		appFactory:          appFactory,
		webhookProxyFactory: webhookF,
	}, nil
}

func (h *Handler) HandleAPI(c echo.Context) error {
	req, err := NewAPIRequest(c)
	if err != nil {
		slog.Error("invalid request", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	app := h.appFactory.GetApp(req.AppID)
	if app == nil {
		slog.Error("app not found", slog.Any("app_id", req.AppID))
		return httperror.ErrAppInvalid(nil)
	}

	installationClient := NewInstallationClient(app, req.InstallationID, h.httpClient)

	accessToken, err := installationClient.AccessToken()
	if err != nil {
		slog.Error("failed to generate access token", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	targetResponse, err := installationClient.Proxy(c.Request(), accessToken)
	if err != nil {
		slog.Error("failed to proxy request", slog.Any("err", err))
		return httperror.ErrUpstreamFailed(err)
	}

	defer targetResponse.Body.Close()

	slog.Info("got response code from github", slog.Any("status_code", targetResponse.StatusCode))

	responseBody, err := io.ReadAll(targetResponse.Body)
	if err != nil {
		slog.Error("failed to read response body", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	w := c.Response().Writer
	w.WriteHeader(targetResponse.StatusCode)
	if _, err := w.Write(responseBody); err != nil {
		slog.Error("failed to write response body", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	c.Response().Flush()
	return nil
}

type WebhookRequest struct {
	AppID string `param:"app_id"`
}

// HandleWebhook handles the webhook request from Github to DeepSource Cloud.
func (h *Handler) HandleWebhook(c echo.Context) error {
	ctx := c.Request().Context()
	req := &WebhookRequest{
		AppID: c.Param("app_id"),
	}

	signature := c.Request().Header.Get(HeaderGithubSignature)
	if signature == "" {
		slog.ErrorCtx(ctx, "missing signature header")
		return httperror.ErrBadRequest(ErrInvalidSignature)
	}

	bodyReader := c.Request().Body
	defer bodyReader.Close()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(bodyReader)
	if err != nil {
		slog.Error("failed to read request body", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	c.Request().Body = io.NopCloser(strings.NewReader(buf.String()))
	client, err := h.webhookProxyFactory.New(req.AppID)
	if err != nil {
		slog.Error("failed to create webhook proxy", slog.Any("err", err))
		return httperror.ErrAppInvalid(err)
	}

	if err := client.VerifySignature(signature, buf.Bytes()); err != nil {
		slog.Error("failed to verify signature", slog.Any("err", err))
		return httperror.ErrUnauthorized(err)
	}

	proxyRes, err := client.Proxy(c.Request())
	if err != nil {
		slog.Error("failed to proxy request", slog.Any("err", err))
		return httperror.ErrUpstreamFailed(err)
	}

	responseBody, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		slog.Error("failed to read response body", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	c.Response().Writer.WriteHeader(proxyRes.StatusCode)
	if _, err := c.Response().Writer.Write(responseBody); err != nil {
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
		return "", ErrConfigNotFound
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
