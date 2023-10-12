package authentication

import (
	"net/http"

	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/deepsourcecorp/runner/auth/session"
	"github.com/labstack/echo/v4"
)

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

type Facade struct {
	oauthHandler      *oauth.Handler
	sessionHandler    *oauth.Handler
	sessionMiddleware *session.Middleware
}

func InitializeFacade(oauthHandler *oauth.Handler, sessionHandler *oauth.Handler, sessionMiddleware *session.Middleware) *Facade {
	return &Facade{
		oauthHandler:      oauthHandler,
		sessionHandler:    sessionHandler,
		sessionMiddleware: sessionMiddleware,
	}
}

func (f *Facade) AddRoutes(r Router) Router {
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/authorize", f.oauthHandler.HandleAuthorize)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/callback", f.oauthHandler.HandleCallback)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/session", f.oauthHandler.HandleSession)
	r.AddRoute(http.MethodPost, "/apps/:app_id/auth/token", f.oauthHandler.HandleToken)
	r.AddRoute(http.MethodPost, "/apps/:app_id/auth/refresh", f.oauthHandler.HandleRefresh)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/user", f.oauthHandler.HandleUser)
	return r
}
