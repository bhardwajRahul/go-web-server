package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSanitizeString(t *testing.T) {
	t.Parallel()

	got := SanitizeString("  O'Reilly\x00  ", DefaultSanitizeConfig)
	want := "O'Reilly"

	if got != want {
		t.Fatalf("SanitizeString() = %q, want %q", got, want)
	}
}

func TestSanitizeMiddlewareNormalizesFormAndQueryValues(t *testing.T) {
	t.Parallel()

	e := echo.New()
	form := url.Values{
		"name": {"  O'Reilly\x00  "},
	}

	req := httptest.NewRequest(http.MethodPost, "/users?tab=%20active%20", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Sanitize()(func(c echo.Context) error {
		if got, want := c.FormValue("name"), "O'Reilly"; got != want {
			t.Fatalf("c.FormValue(name) = %q, want %q", got, want)
		}

		if got, want := c.QueryParam("tab"), "active"; got != want {
			t.Fatalf("c.QueryParam(tab) = %q, want %q", got, want)
		}

		return c.NoContent(http.StatusNoContent)
	})

	if err := handler(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
}
