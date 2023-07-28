package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/deepsourcecorp/runner/auth/jwtutil"
	"github.com/deepsourcecorp/runner/auth/model"
	"github.com/deepsourcecorp/runner/auth/oauth"
	"github.com/deepsourcecorp/runner/auth/saml"
	"github.com/deepsourcecorp/runner/auth/store"
	"github.com/deepsourcecorp/runner/auth/token"
	"github.com/deepsourcecorp/runner/httperror"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
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

	deepsourceVerifier := jwtutil.NewVerifier(opts.DeepSource.PublicKey)
	tokenMiddleware := DeepSourceTokenMiddleware(opts.Runner.ID, deepsourceVerifier)

	// Initialize token service and handlers
	runnerSigner := jwtutil.NewSigner(opts.Runner.PrivateKey)
	runnerVerifier := jwtutil.NewVerifier(&opts.Runner.PrivateKey.PublicKey)
	tokenService := token.NewService(runnerSigner, runnerVerifier)
	tokenHandlers := token.NewHandler(opts.Runner, tokenService)
	sessionMiddleware := token.SessionAuthMiddleware(opts.Runner.ID, tokenService)

	// Initialize SAML middleware and handlers only if SAML is configured.
	var samlHandlers *saml.Handler
	if opts.SAML != nil {
		samlMiddleware, err := saml.NewSAMLMiddleware(ctx, opts.SAML, client)
		if err != nil {
			return nil, err
		}
		samlHandlers = saml.NewHandler(opts.Runner, opts.DeepSource, samlMiddleware, tokenService, opts.Store)
	}

	// Initialize OAuth factory and handlers
	oauthFactory := oauth.NewFactory(opts.Apps)
	oauthHandlers := oauth.NewHandler(opts.Runner, opts.DeepSource, opts.Store, tokenService, oauthFactory)

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

var ErrInvalidToken = httperror.Error{Message: "invalid token"}

func DeepSourceTokenMiddleware(runnerID string, verifier *jwtutil.Verifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authrorization := c.Request().Header.Get("Authorization")
			parts := strings.SplitN(authrorization, " ", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				slog.Error("not a valid bearer token", slog.Any("header", authrorization))
				return c.JSON(http.StatusUnauthorized, ErrInvalidToken)
			}

			claims, err := verifier.Verify(parts[1])
			if err != nil {
				slog.Error("failed to verify token", slog.Any("err", err))
				return c.JSON(http.StatusUnauthorized, ErrInvalidToken)
			}

			if claims["runner_id"] != runnerID {
				slog.Error("runner id mismatch", slog.Any("runner-id", claims["runner-id"]))
				return c.JSON(http.StatusUnauthorized, ErrInvalidToken)
			}

			// Token is valid
			return next(c)
		}
	}
}
