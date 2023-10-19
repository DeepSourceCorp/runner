package gitlab

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type APIRequest struct {
	AppID       string
	HTTPRequest *http.Request
	Token       string
}

func NewAPIRequest(c echo.Context) (*APIRequest, error) {
	req := &APIRequest{
		AppID:       c.Param("app_id"),
		HTTPRequest: c.Request(),
	}

	authorization := c.Request().Header.Get("Authorization")
	parts := strings.Split(authorization, " ")
	if len(parts) != 2 {
		return nil, errors.New("invalid authorization header")
	}
	req.Token = parts[1]

	if err := req.validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *APIRequest) validate() error {
	if r.AppID == "" || r.Token == "" {
		return errors.New("missing app id or token")
	}
	return nil
}

type WebhookRequest struct {
	AppID       string
	Token       string
	HTTPRequest *http.Request
}

func NewWebhookRequest(c echo.Context) (*WebhookRequest, error) {
	req := &WebhookRequest{
		HTTPRequest: c.Request(),
		AppID:       c.Param("app_id"),
		Token:       c.Request().Header.Get("X-Gitlab-Token"),
	}

	if err := req.validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *WebhookRequest) validate() error {
	if r.AppID == "" || r.Token == "" {
		return errors.New("missing app id or token")
	}
	return nil
}

type RemoteURLRequest struct {
	SourceURL string
	Token     string
}

func NewRemoteURLRequest(sourceURL string, extra map[string]interface{}) *RemoteURLRequest {
	req := &RemoteURLRequest{
		SourceURL: sourceURL,
		Token:     extra["token"].(string),
	}

	return req
}
