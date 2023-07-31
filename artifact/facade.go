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
	CORSMiddleware  echo.MiddlewareFunc
}

type Opts struct {
	Bucket        string
	AllowedOrigin string
	Storage       StorageClient
}

func New(ctx context.Context, opts *Opts) (*Facade, error) {
	if opts == nil || opts.Storage == nil {
		return nil, ErrMissingOpts
	}

	cors := corsMiddleware(opts.AllowedOrigin)

	return &Facade{
		ArtifactHandler: NewHandler(opts.Storage, opts.Bucket),
		CORSMiddleware:  cors,
	}, nil
}

func (f *Facade) AddRoutes(router Router, middleware []echo.MiddlewareFunc) Router {
	middleware = append([]echo.MiddlewareFunc{f.CORSMiddleware}, middleware...)
	router.AddRoute(http.MethodOptions, "apps/:app_id/artifacts/*", f.ArtifactHandler.HandleOptions, []echo.MiddlewareFunc{f.CORSMiddleware}...)
	router.AddRoute(http.MethodPost, "apps/:app_id/artifacts/analysis", f.ArtifactHandler.HandleAnalysis, middleware...)
	router.AddRoute(http.MethodPost, "apps/:app_id/artifacts/autofix", f.ArtifactHandler.HandleAutofix, middleware...)
	return router
}

func corsMiddleware(origin string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(HeaderAccessControlAllowOrigin, origin)
			c.Response().Header().Set(HeaderAccessControlAllowMethods, "GET, POST, OPTIONS")
			c.Response().Header().Set(HeaderAccessControlAllowHeaders, "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cache-Control, Pragma")
			c.Response().Header().Set(HeaderAccessControlAllowCredentials, "true")
			return next(c)
		}
	}
}
