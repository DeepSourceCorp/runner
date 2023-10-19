package contract

import (
	"context"
	"errors"
	"fmt"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
)

type RefreshRequest struct {
	Ctx          context.Context
	AppID        string `param:"app_id"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

func NewRefreshRequest(c echo.Context) (*RefreshRequest, error) {
	req := &RefreshRequest{}
	if err := c.Bind(req); err != nil {
		err = fmt.Errorf("refresh request bind error: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}
	if err := req.validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *RefreshRequest) validate() error {
	if r.AppID == "" || r.ClientID == "" || r.ClientSecret == "" || r.RefreshToken == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}
