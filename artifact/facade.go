package artifact

import (
	"context"
	"errors"
	"net/http"

	"github.com/deepsourcecorp/runner/middleware"

	"github.com/labstack/echo/v4"
)

const (
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
)

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

var (
	ErrMissingOpts = errors.New("missing required options")
)

type Facade struct {
	ArtifactHandler *Handler
	allowedOrigin   string
}

type Opts struct {
	AllowedOrigin string // For CORS
	Bucket        string
	Storage       StorageClient
}

func New(ctx context.Context, opts *Opts) (*Facade, error) {
	if opts == nil || opts.Storage == nil {
		return nil, ErrMissingOpts
	}

	return &Facade{
		allowedOrigin:   opts.AllowedOrigin,
		ArtifactHandler: NewHandler(opts.Storage, opts.Bucket),
	}, nil
}

func (f *Facade) AddRoutes(router Router, m []echo.MiddlewareFunc) Router {
	cors := middleware.CorsMiddleware(f.allowedOrigin)
	router.AddRoute(http.MethodOptions, "apps/:app_id/artifacts", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, cors)

	m = append([]echo.MiddlewareFunc{cors}, m...)
	router.AddRoute(http.MethodPost, "apps/:app_id/artifacts/analysis", f.ArtifactHandler.HandleAnalysis, m...)
	router.AddRoute(http.MethodPost, "apps/:app_id/artifacts/autofix", f.ArtifactHandler.HandleAutofix, m...)

	return router
}
