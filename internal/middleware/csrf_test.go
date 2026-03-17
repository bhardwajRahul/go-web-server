package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCSRFSafeRequestReusesExistingCookieToken(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: DefaultCSRFConfig.CookieName, Value: "existing-token"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CSRF()(func(c echo.Context) error {
		if got := GetCSRFToken(c); got != "existing-token" {
			t.Fatalf("GetCSRFToken() = %q, want %q", got, "existing-token")
		}
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler() error = %v", err)
	}

	if setCookie := rec.Header().Get(echo.HeaderSetCookie); setCookie != "" {
		t.Fatalf("expected no new Set-Cookie header, got %q", setCookie)
	}

	if got := rec.Header().Get(echo.HeaderXCSRFToken); got != "existing-token" {
		t.Fatalf("X-CSRF-Token = %q, want %q", got, "existing-token")
	}
}

func TestCSRFSafeRequestCreatesCookieWhenMissing(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CSRF()(func(c echo.Context) error {
		if got := GetCSRFToken(c); got == "" {
			t.Fatal("expected CSRF token in context")
		}
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler() error = %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected CSRF cookie to be set")
	}

	if cookies[0].Name != DefaultCSRFConfig.CookieName {
		t.Fatalf("cookie name = %q, want %q", cookies[0].Name, DefaultCSRFConfig.CookieName)
	}

	if got := rec.Header().Get(echo.HeaderXCSRFToken); got == "" {
		t.Fatal("expected X-CSRF-Token response header")
	}
}

func TestCSRFUnsafeRequestRotatesTokenAndReturnsHeader(t *testing.T) {
	t.Parallel()

	e := echo.New()
	form := url.Values{}
	form.Set("csrf_token", "existing-token")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.AddCookie(&http.Cookie{Name: DefaultCSRFConfig.CookieName, Value: "existing-token"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := CSRF()(func(c echo.Context) error {
		token := GetCSRFToken(c)
		if token == "" || token == "existing-token" {
			t.Fatalf("expected rotated CSRF token, got %q", token)
		}
		return c.NoContent(http.StatusOK)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler() error = %v", err)
	}

	got := rec.Header().Get(echo.HeaderXCSRFToken)
	if got == "" || got == "existing-token" {
		t.Fatalf("X-CSRF-Token = %q, want rotated token", got)
	}
}
