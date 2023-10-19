package gitlab

import (
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/deepsourcecorp/runner/provider/common"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleAPI(c echo.Context) error {
	req, err := NewAPIRequest(c)
	if err != nil {
		slog.Error("[gitlab.handler] failed to create contract for API")
		return httperror.ErrBadRequest(err)
	}

	res, err := h.service.ForwardAPI(req)
	if err != nil {
		slog.Error("[gitlab.handler] failed to forward API request")
		return httperror.ErrUpstreamFailed(err)
	}

	return common.WriteResponse(c, res)
}

// HandleWebhook handles the webhook request from Gitlab to DeepSource Cloud.
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

func (h *Handler) HandleInstallation(c echo.Context) error {
	return httperror.ErrUnsupported(nil)
}
