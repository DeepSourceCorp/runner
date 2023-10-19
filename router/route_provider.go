package router

import "github.com/labstack/echo/v4"

type Router interface {
	AddRoute(method string, path string, handlerFunc echo.HandlerFunc, middleware ...echo.MiddlewareFunc)
}

type RouteProvider interface {
	AddRoutes(r Router) Router
}
