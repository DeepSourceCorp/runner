package github

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

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
	apiProxyFactory     *APIProxyFactory
	webhookProxyFactory *WebhookProxyFactory
}

func NewHandler(apiF *APIProxyFactory, webhookF *WebhookProxyFactory) (*Handler, error) {
	return &Handler{
		apiProxyFactory:     apiF,
		webhookProxyFactory: webhookF,
	}, nil
}

type APIRequest struct {
	AppID          string
	InstallationID string
}

// HandleAPI handles the API request from DeepSource Cloud to Github.
func (h *Handler) HandleAPI(c echo.Context) error {
	req := &APIRequest{
		InstallationID: c.Request().Header.Get(HeaderInstallationID),
		AppID:          c.Param("app_id"),
	}

	client, err := h.apiProxyFactory.NewProxy(req.AppID, req.InstallationID)
	if err != nil {
		return c.JSON(ErrEchoResponse(HTTPErrInvalidRequest.WithInternal(err)))
	}

	proxyRes, err := client.Proxy(c.Request())
	if err != nil {
		slog.Error(fmt.Sprintf("failed to proxy request to github: %v", err))
		return c.JSON(ErrEchoResponse(HTTPErrUpstreamFailed.WithInternal(err)))
	}

	slog.Debug(fmt.Sprintf("got response code %d from github", proxyRes.StatusCode))

	responseBody, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		slog.Error("failed to read response body", err)
		return c.JSON(ErrEchoResponse(HTTPErrUpstreamBad.WithInternal(err)))
	}

	w := c.Response().Writer
	w.WriteHeader(proxyRes.StatusCode)
	if _, err := w.Write(responseBody); err != nil {
		slog.Error("failed to write response body", err)
		return c.JSON(ErrEchoResponse(HTTPErrUpstreamBad.WithInternal(err)))
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
		return ErrInvalidSignature
	}

	bodyReader := c.Request().Body
	defer bodyReader.Close()

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(bodyReader)
	if err != nil {
		slog.ErrorCtx(ctx, "failed to read request body", err)
		return c.JSON(ErrEchoResponse(HTTPErrInvalidRequest.WithInternal(err)))
	}
	c.Request().Body = io.NopCloser(strings.NewReader(buf.String()))
	client, err := h.webhookProxyFactory.New(req.AppID)
	if err != nil {
		slog.ErrorCtx(ctx, "failed to create webhook proxy client", err)
		return c.JSON(ErrEchoResponse(HTTPErrInvalidRequest.WithInternal(err)))
	}

	if err := client.VerifySignature(signature, buf.Bytes()); err != nil {
		slog.ErrorCtx(ctx, "failed to verify webhook signature", err)
		return c.JSON(ErrEchoResponse(HTTPErrSignatureMismatch.WithInternal(err)))
	}

	proxyRes, err := client.Proxy(c.Request())
	if err != nil {
		slog.ErrorCtx(ctx, "failed to proxy webhook request", err)
		return c.JSON(ErrEchoResponse(HTTPErrUpstreamFailed.WithInternal(err)))
	}

	responseBody, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		slog.ErrorCtx(ctx, "failed to read response body", err)
		return c.JSON(ErrEchoResponse(HTTPErrUpstreamBad.WithInternal(err)))
	}

	c.Response().Writer.WriteHeader(proxyRes.StatusCode)
	if _, err := c.Response().Writer.Write(responseBody); err != nil {
		slog.Error("failed to write response body", err)
		return c.JSON(ErrEchoResponse(HTTPErrUpstreamBad.WithInternal(err)))
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
		return c.JSON(ErrEchoResponse(HTTPErrInvalidRequest.WithInternal(err)))
	}
	client, err := h.apiProxyFactory.NewProxy(req.AppID, "")
	if err != nil {
		return c.JSON(ErrEchoResponse(HTTPErrInvalidRequest.WithInternal(err)))
	}

	installationURL := client.InstallationURL()
	return c.Redirect(http.StatusTemporaryRedirect, installationURL)
}
