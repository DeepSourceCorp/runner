package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

type mockSessionStore struct {
	valid bool
}

func (m *mockSessionStore) IsValidSession(id string) bool {
	return m.valid
}

func TestSessionAuthenticationMiddleware_Middleware(t *testing.T) {
	t.Run("returns 401 if session cookie is not set", func(t *testing.T) {
		m := NewSessionMiddleware(&mockSessionStore{valid: true})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := echo.New().NewContext(req, rec)
		h := m.Middleware(func(c echo.Context) error {
			return c.JSON(http.StatusOK, "ok")
		})
		err := h(c)
		if err != nil {
			t.Error("expected no error")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("returns 401 if session cookie is empty", func(t *testing.T) {
		m := NewSessionMiddleware(&mockSessionStore{valid: true})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: "",
		})
		c := echo.New().NewContext(req, rec)
		h := m.Middleware(func(c echo.Context) error {
			return c.JSON(http.StatusOK, "ok")
		})
		err := h(c)
		if err != nil {
			t.Error("expected no error")
		}
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status code %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("returns 200 if oauthSession validates true", func(t *testing.T) {
		m := NewSessionMiddleware(&mockSessionStore{valid: true})
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		req.AddCookie(&http.Cookie{
			Name:  "session",
			Value: "test",
		})
		c := echo.New().NewContext(req, rec)
		h := m.Middleware(func(c echo.Context) error {
			return c.JSON(http.StatusOK, "ok")
		})
		err := h(c)
		if err != nil {
			t.Error("expected no error")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, rec.Code)
		}
	})

}
