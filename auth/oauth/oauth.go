package oauth

import (
	"net/http"

	"github.com/deepsourcecorp/runner/auth/common"
	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/labstack/echo/v4"
)

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

type Opts struct {
	Runner       *common.Runner
	Deepsource   *common.DeepSource
	SessionStore session.Store
	Apps         Apps
}

type OAuth struct {
	handler *Handler
}

func New(opts *Opts) *OAuth {
	sessionService := session.NewService(opts.Runner, opts.Deepsource, opts.SessionStore)
	service := NewService(opts.Apps, sessionService)
	handler := NewHandler(service)
	return &OAuth{
		handler: handler,
	}
}

func (o *OAuth) AddRoutes(r Router) Router {
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/authorize", o.handler.HandleAuthorize)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/callback", o.handler.HandleCallback)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/session", o.handler.HandleSession)
	r.AddRoute(http.MethodPost, "/apps/:app_id/auth/token", o.handler.HandleToken)
	r.AddRoute(http.MethodPost, "/apps/:app_id/auth/refresh", o.handler.HandleRefresh)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/user", o.handler.HandleUser)
	return r
}
