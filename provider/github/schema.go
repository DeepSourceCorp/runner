package github

import (
	"errors"

	"github.com/labstack/echo/v4"
)

type APIRequest struct {
	AppID          string
	InstallationID string
}

func NewAPIRequest(c echo.Context) (*APIRequest, error) {
	req := &APIRequest{
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get(HeaderInstallationID),
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
