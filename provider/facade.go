package provider

import (
	"github.com/deepsourcecorp/runner/provider/github"
	"github.com/labstack/echo/v4"
)

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

type Facade struct {
	Adapter *Adapter
}

func NewFacade(apps map[string]*App, githubProvider *github.Handler) *Facade {
	return &Facade{
		Adapter: NewAdapter(apps, githubProvider),
	}
}

func (f *Facade) AddRoutes(r Router) Router {
	r.AddRoute("*", "apps/:app_id/webhook", f.Adapter.HandleWebhook)
	r.AddRoute("*", "apps/:app_id/api/*", f.Adapter.HandleAPI)
	r.AddRoute("*", "apps/:app_id/installation/new", f.Adapter.HandleInstallation)
	return r
}
