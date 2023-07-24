package artifact

import "github.com/labstack/echo/v4"

type CORSMiddleware struct {
	AllowedOrigin string
}

func (m *CORSMiddleware) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Access-Control-Allow-Origin", m.AllowedOrigin)
		c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Response().Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cache-Control, Pragma")
		c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
		return next(c)
	}
}
