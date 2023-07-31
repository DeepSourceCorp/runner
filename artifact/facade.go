package artifact

import (
	"context"
	"errors"
	"net/http"

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
}

type Opts struct {
	Bucket  string
	Storage StorageClient
}

func New(ctx context.Context, opts *Opts) (*Facade, error) {
	if opts == nil || opts.Storage == nil {
		return nil, ErrMissingOpts
	}

	return &Facade{
		ArtifactHandler: NewHandler(opts.Storage, opts.Bucket),
	}, nil
}

func (f *Facade) AddRoutes(router Router, middleware []echo.MiddlewareFunc) Router {
	router.AddRoute(http.MethodPost, "apps/:app_id/artifacts/analysis", f.ArtifactHandler.HandleAnalysis, middleware...)
	router.AddRoute(http.MethodPost, "apps/:app_id/artifacts/autofix", f.ArtifactHandler.HandleAutofix, middleware...)
	return router
}
