package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/deepsourcecorp/runner/auth/saml"
	"github.com/deepsourcecorp/runner/auth/store"
	"github.com/deepsourcecorp/runner/auth/token"
	"github.com/labstack/echo/v4"
)

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

type Facade struct {
	TokenHandlers     *token.Handler
	OAuthHandlers     *oauth.Handler
	SAMLHandlers      *saml.Handler
	TokenMiddleware   echo.MiddlewareFunc
	SessionMiddleware echo.MiddlewareFunc
}

type Opts struct {
	Runner     *model.Runner
	DeepSource *model.DeepSource
	Apps       map[string]*oauth.App
	SAML       *saml.Opts
	Store      store.Store
}

func New(ctx context.Context, opts *Opts, client *http.Client) (*Facade, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	if opts == nil {
		return nil, errors.New("opts cannot be nil")
	}

	// Initialize HTTP client
	if client == nil {
		client = http.DefaultClient
	}

	// Initialize token service and handlers
	tokenService := token.NewService(opts.Runner.ID, opts.Runner.PrivateKey)
	tokenHandlers := token.NewHandler(tokenService)

	// Initialize SAML middleware and handlers only if SAML is configured.
	var samlHandlers *saml.Handler
	if opts.SAML != nil {
		samlMiddleware, err := saml.NewSAMLMiddleware(ctx, opts.SAML, client)
		if err != nil {
			return nil, err
		}
		samlHandlers = saml.NewHandler(opts.Runner, opts.DeepSource, samlMiddleware, opts.Store)
	}

	// Initialize OAuth factory and handlers
	oauthFactory := oauth.NewFactory(opts.Apps)
	oauthHandlers := oauth.NewHandler(opts.Runner, opts.DeepSource, opts.Store, oauthFactory)

	// Initialize middlewares
	tokenMiddleware := token.BearerAuthMiddleware(tokenService)
	sessionMiddleware := token.SessionAuthMiddleware(tokenService)

	return &Facade{
		TokenHandlers:     tokenHandlers,
		OAuthHandlers:     oauthHandlers,
		TokenMiddleware:   tokenMiddleware,
		SessionMiddleware: sessionMiddleware,
		SAMLHandlers:      samlHandlers,
	}, nil
}

func (f *Facade) AddRoutes(r Router) Router {
	r.AddRoute(http.MethodPost, "/refresh", f.TokenHandlers.HandleRefresh)
	r.AddRoute(http.MethodPost, "/logout", f.TokenHandlers.HandleLogout)

	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/authorize", f.OAuthHandlers.HandleAuthorize)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/callback", f.OAuthHandlers.HandleCallback)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/session", f.OAuthHandlers.HandleSession)
	r.AddRoute(http.MethodPost, "/apps/:app_id/auth/token", f.OAuthHandlers.HandleToken)
	r.AddRoute(http.MethodPost, "/apps/:app_id/auth/refresh", f.OAuthHandlers.HandleRefresh)
	r.AddRoute(http.MethodGet, "/apps/:app_id/auth/user", f.OAuthHandlers.HandleUser)

	if f.SAMLHandlers != nil {
		r.AddRoute("*", "/saml/*", f.SAMLHandlers.SAMLHandler())
		r.AddRoute(http.MethodGet, "/apps/saml/auth/authorize", f.SAMLHandlers.AuthorizationHandler())
		r.AddRoute(http.MethodGet, "/apps/saml/auth/session", f.SAMLHandlers.HandleSession)
		r.AddRoute(http.MethodPost, "/apps/saml/auth/token", f.SAMLHandlers.HandleToken)
		r.AddRoute(http.MethodPost, "/apps/saml/auth/refresh", f.SAMLHandlers.HandleRefresh)
	}
	return r
}
