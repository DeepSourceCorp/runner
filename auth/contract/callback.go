package contract

import (
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
)

type CallbackRequest struct {
	AppID string `param:"app_id"`
	Code  string `query:"code"`
	State string `query:"state"`
}

func NewCallbackRequest(c echo.Context) (*CallbackRequest, error) {
	req := &CallbackRequest{}
	if err := c.Bind(req); err != nil {
		return nil, fmt.Errorf("callback request bind error: %w", err)
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (r *CallbackRequest) Validate() error {
	if r.AppID == "" || r.Code == "" || r.State == "" {
		return errors.New("callback request validation failed")
	}
	return nil
}
