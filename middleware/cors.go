package middleware

import "github.com/labstack/echo/v4"

const (
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
)

func CorsMiddleware(origin string) echo.MiddlewareFunc {
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
