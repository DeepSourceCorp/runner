package contract

import (
	"errors"
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
)

type AuthorizationRequest struct {
	AppID    string `param:"app_id"`
	ClientID string `query:"client_id"`
	Scopes   []string
	State    string `query:"state"`
}

func NewAuthorizationRequest(c echo.Context) (*AuthorizationRequest, error) {
	req := &AuthorizationRequest{}
	if err := c.Bind(req); err != nil {
		return nil, fmt.Errorf("authorize request bind error: %w", err)
	}
	req.Scopes = strings.Split(c.QueryParam("scopes"), ",")
	if err := req.validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *AuthorizationRequest) validate() error {
	if r.AppID == "" || r.ClientID == "" || r.State == "" {
		return errors.New("authorization request validation failed")
	}
	return nil
}
