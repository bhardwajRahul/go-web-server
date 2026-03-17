package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestRequestTimeoutSetsDeadlineOnContext(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RequestTimeout(50 * time.Millisecond)

	var hasDeadline bool
	err := mw(func(c echo.Context) error {
		_, hasDeadline = c.Request().Context().Deadline()
		return c.NoContent(http.StatusOK)
	})(c)
	if err != nil {
		t.Fatalf("RequestTimeout() error = %v", err)
	}

	if !hasDeadline {
		t.Fatal("expected request context to have a deadline")
	}
}

func TestRequestTimeoutReturnsAppTimeoutErrorAfterDeadline(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := RequestTimeout(10 * time.Millisecond)

	err := mw(func(c echo.Context) error {
		<-c.Request().Context().Done()
		return c.Request().Context().Err()
	})(c)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}

	if appErr.Code != http.StatusRequestTimeout {
		t.Fatalf("timeout status = %d, want %d", appErr.Code, http.StatusRequestTimeout)
	}

	if !errors.Is(appErr.Internal, context.DeadlineExceeded) {
		t.Fatalf("internal error = %v, want deadline exceeded", appErr.Internal)
	}
}
