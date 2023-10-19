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

type InstallationRequest struct {
	AppID string `param:"app_id"`
}

func NewInstallationRequest(c echo.Context) (*InstallationRequest, error) {
	req := &InstallationRequest{}
	if err := c.Bind(req); err != nil {
		return nil, err
	}
	return req, nil
}

type RemoteURLRequest struct {
	AppID          string
	InstallationID string
	SourceURL      string
}

func NewRemoteURLRequest(appID, sourceURL string, extra map[string]interface{}) *RemoteURLRequest {
	return &RemoteURLRequest{
		AppID:          appID,
		SourceURL:      sourceURL,
		InstallationID: extra["installation_id"].(string),
	}
}
