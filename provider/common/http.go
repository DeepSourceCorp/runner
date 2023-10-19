package common

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func WriteResponse(c echo.Context, res *http.Response) error {
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	for k, v := range res.Header {
		c.Response().Header().Set(k, v[0])
	}

	c.Response().WriteHeader(res.StatusCode)
	if _, err := c.Response().Write(body); err != nil {
		return err
	}
	c.Response().Flush()
	return nil
}
