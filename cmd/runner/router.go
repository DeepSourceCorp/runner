package main

import (
	"github.com/labstack/echo/v4"
)

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

type Route struct {
	Method      string
	Path        string
	HandlerFunc echo.HandlerFunc
	Middleware  []echo.MiddlewareFunc
}
