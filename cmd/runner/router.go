package main

import (
	"github.com/labstack/echo/v4"
)

type Route struct {
	Method      string
	Path        string
	HandlerFunc echo.HandlerFunc
	Middleware  []echo.MiddlewareFunc
}

type Router struct {
	Routes []Route
	e      *echo.Echo
}

func NewRouter() *Router {
	return &Router{
		Routes: []Route{},
	}
}

func (r *Router) Setup() {
	for _, route := range r.Routes {
		if route.Method == "*" {
			r.e.Any(route.Path, route.HandlerFunc, route.Middleware...)
		}
		r.e.Add(route.Method, route.Path, route.HandlerFunc, route.Middleware...)
	}
}

func (r *Router) AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc) {
	r.Routes = append(r.Routes, Route{
		Method:      method,
		Path:        path,
		HandlerFunc: handlerFunc,
		Middleware:  middleware,
	})
}
