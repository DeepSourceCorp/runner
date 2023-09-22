package github

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type APIRequest struct {
	AppID          string
	InstallationID string
	HTTPRequest    *http.Request
}

func NewAPIRequest(c echo.Context) (*APIRequest, error) {
	req := &APIRequest{
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get(HeaderInstallationID),
		HTTPRequest:    c.Request(),
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func (r *APIRequest) Validate() error {
	if r.AppID == "" || r.InstallationID == "" {
		return errors.New("missing app id or installation id")
	}
	return nil
}

type WebhookRequest struct {
	AppID       string
	Signature   string
	HTTPRequest *http.Request
}

func NewWebhookRequest(c echo.Context) (*WebhookRequest, error) {
	req := &WebhookRequest{
		HTTPRequest: c.Request(),
		Signature:   c.Request().Header.Get("X-Hub-Signature-256"),
		AppID:       c.Param("app_id"),
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *WebhookRequest) Validate() error {
	if r.AppID == "" || r.Signature == "" {
		return errors.New("missing app id or signature")
	}
	return nil
}
