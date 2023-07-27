package auth

import (
	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/deepsourcecorp/runner/auth/store"
	"github.com/deepsourcecorp/runner/auth/token"
	"github.com/labstack/echo/v4"
)

type Authentication struct {
	TokenHandlers     *token.Handler
	OAuthHandlers     *oauth.Handler
	TokenMiddleware   echo.MiddlewareFunc
	SessionMiddleware echo.MiddlewareFunc
}

func NewAuthentication(runner *model.Runner, deepsource *model.DeepSource, apps map[string]*oauth.App, store store.Store) *Authentication {
	oauthFactory := oauth.NewFactory(apps)
	tokenService := token.NewService(runner.ID, runner.PrivateKey)

	tokenHandlers := token.NewHandler(tokenService)
	oauthHandlers := oauth.NewHandler(runner, deepsource, store, oauthFactory)
	tokenMiddleware := token.BearerAuthMiddleware(tokenService)
	sessionMiddleware := token.SessionAuthMiddleware(tokenService)

	return &Authentication{
		TokenHandlers:     tokenHandlers,
		OAuthHandlers:     oauthHandlers,
		TokenMiddleware:   tokenMiddleware,
		SessionMiddleware: sessionMiddleware,
	}
}
