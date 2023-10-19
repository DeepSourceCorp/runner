package session

import (
	"fmt"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
)

type RefreshRequest struct {
	Referrer     string
	RefreshToken string
}

func NewRefreshRequest(c echo.Context) (*RefreshRequest, error) {
	cookie, err := c.Cookie("refresh")
	if err != nil {
		err := fmt.Errorf("session/refresh: failed to parse cookie: %w", err)
		return nil, httperror.ErrBadRequest(err)
	}
	if cookie.Value == "" {
		err := fmt.Errorf("session/refresh: empty refresh token")
		return nil, httperror.ErrBadRequest(err)
	}

	return &RefreshRequest{
		Referrer:     c.QueryParam("redirect"),
		RefreshToken: cookie.Value,
	}, nil
}
