package provider

import (
	"github.com/labstack/echo/v4"
)

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

type Facade struct {
	Adapter *Adapter
}

type Opts struct {
	Apps           map[string]*App
	GithubProvider Provider
}

func NewFacade(opts *Opts) *Facade {
	return &Facade{
		Adapter: NewAdapter(opts.Apps, opts.GithubProvider),
	}
}

func (f *Facade) AddRoutes(r Router) Router {
	r.AddRoute("*", "apps/:app_id/webhook", f.Adapter.HandleWebhook)
	r.AddRoute("*", "apps/:app_id/api/*", f.Adapter.HandleAPI)
	r.AddRoute("*", "apps/:app_id/installation/new", f.Adapter.HandleInstallation)
	return r
}
