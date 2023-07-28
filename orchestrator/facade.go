package orchestrator

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	ErrMissingOpts = errors.New("missing opts")
)

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

type Opts struct {
	*TaskOpts
	*CleanerOpts
	Provider
	Signer
	Driver
	*Runner
}

type Facade struct {
	OrchestratorHandler *Handler
	Cleaner             *Cleaner
}

func New(opts *Opts) (*Facade, error) {
	if opts == nil || opts.TaskOpts == nil || opts.Provider == nil || opts.Signer == nil || opts.Driver == nil {
		return nil, ErrMissingOpts
	}
	cleaner := NewCleaner(opts.Driver, opts.CleanerOpts)
	handler := NewHandler(opts.TaskOpts, opts.Driver, opts.Provider, opts.Signer, opts.Runner)

	return &Facade{
		Cleaner:             cleaner,
		OrchestratorHandler: handler,
	}, nil
}

func (f *Facade) AddRoutes(router Router, middleware []echo.MiddlewareFunc) Router {
	router.AddRoute(http.MethodPost, "apps/:app_id/tasks/analysis", f.OrchestratorHandler.HandleAnalysis, middleware...)
	router.AddRoute(http.MethodPost, "apps/:app_id/tasks/autofix", f.OrchestratorHandler.HandleAutofix, middleware...)
	router.AddRoute(http.MethodPost, "apps/:app_id/tasks/transformer", f.OrchestratorHandler.HandleTransformer, middleware...)
	router.AddRoute(http.MethodPost, "apps/:app_id/tasks/cancelcheck", f.OrchestratorHandler.HandleCancelCheck, middleware...)
	router.AddRoute(http.MethodPost, "apps/:app_id/tasks/commit", f.OrchestratorHandler.HandlePatcher, middleware...)
	return router
}
