package contract

import (
	"context"
	"errors"
	"strings"

	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
)

type UserRequest struct {
	Ctx         context.Context
	AppID       string `param:"app_id"`
	AccessToken string
}

func NewUserRequest(c echo.Context) (*UserRequest, error) {
	req := &UserRequest{}
	req.AppID = c.Param("app_id")

	authorization := c.Request().Header.Get("Authorization")
	parts := strings.Split(authorization, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, httperror.ErrBadRequest(errors.New("invalid authorization header"))
	}

	req.AccessToken = parts[1]

	if err := req.validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func (r *UserRequest) validate() error {
	if r.AppID == "" || r.AccessToken == "" {
		return httperror.ErrBadRequest(errors.New("missing params"))
	}
	return nil
}

type UserResponse struct {
	UserName string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	ID       string `json:"id"`
}
