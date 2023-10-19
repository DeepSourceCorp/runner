package github

import (
	"net/http"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/common"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

type Handler struct {
	service *Service
	client  *http.Client
}

func NewHandler(service *Service, client *http.Client) *Handler {
	return &Handler{
		service: service,
		client:  client,
	}
}

func (h *Handler) HandleAPI(c echo.Context) error {
	req, err := NewAPIRequest(c)
	if err != nil {
		slog.Error("[github.Handler] failed to create contract for API", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	res, err := h.service.ForwardAPI(req)
	if err != nil {
		slog.Error("[github.Handler] failed to forward API request", slog.Any("err", err))
		return httperror.ErrUpstreamFailed(err)
	}

	return common.WriteResponse(c, res)
}

// HandleWebhook handles the webhook request from Github to DeepSource Cloud.
func (h *Handler) HandleWebhook(c echo.Context) error {
	req, err := NewWebhookRequest(c)
	if err != nil {
		return httperror.ErrMissingParams(err)
	}

	res, err := h.service.ForwardWebhook(req)
	if err != nil {
		return err
	}
	return common.WriteResponse(c, res)
}

// HandleInstallation redirects the user to the installation page on Github.
func (h *Handler) HandleInstallation(c echo.Context) error {
	req, err := NewInstallationRequest(c)
	if err != nil {
		return httperror.ErrMissingParams(err)
	}

	installationURL, err := h.service.InstallationURL(req)
	if err != nil {
		slog.Error("[github.Handler] failed to generate installation url", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}

	return c.Redirect(http.StatusTemporaryRedirect, installationURL)
}
