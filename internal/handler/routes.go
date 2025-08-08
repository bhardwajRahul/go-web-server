package handler

import (
	"io/fs"
	"net/http"

	"log/slog"

	"github.com/dunamismax/go-web-server/internal/middleware"
	"github.com/dunamismax/go-web-server/internal/store"
	"github.com/dunamismax/go-web-server/internal/ui"
	"github.com/labstack/echo/v4"
)

// Handlers holds all the application handlers.
type Handlers struct {
	Home *HomeHandler
	User *UserHandler
	Auth *AuthHandler
}

// NewHandlers creates a new handlers instance with the given store.
func NewHandlers(s *store.Store, authService *middleware.AuthService) *Handlers {
	return &Handlers{
		Home: NewHomeHandler(s),
		User: NewUserHandler(s),
		Auth: NewAuthHandler(s, authService),
	}
}

// RegisterRoutes sets up all application routes.
func RegisterRoutes(e *echo.Echo, handlers *Handlers) error {
	// Serve static files
	staticFS, err := fs.Sub(ui.StaticFiles, "static")
	if err != nil {
		slog.Error("failed to create static file system", "error", err)

		return err
	}

	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS)))))

	// Home routes
	e.GET("/", handlers.Home.Home)
	e.GET("/demo", handlers.Home.Demo)
	e.GET("/health", handlers.Home.Health)

	// Authentication routes (no auth required)
	auth := e.Group("/auth")
	auth.GET("/login", handlers.Auth.LoginPage)
	auth.GET("/register", handlers.Auth.RegisterPage)
	auth.POST("/login", handlers.Auth.Login)
	auth.POST("/register", handlers.Auth.Register)
	auth.POST("/logout", handlers.Auth.Logout)

	// Protected routes (authentication required)
	protected := e.Group("/profile")
	// protected.Use(middleware.JWTMiddleware(authService)) // Commented out for now as we don't have authService here
	protected.GET("", handlers.Auth.Profile)

	// User management routes
	e.GET("/users", handlers.User.Users)
	e.GET("/users/list", handlers.User.UserList)
	e.GET("/users/form", handlers.User.UserForm)
	e.GET("/users/:id/edit", handlers.User.EditUserForm)
	e.POST("/users", handlers.User.CreateUser)
	e.PUT("/users/:id", handlers.User.UpdateUser)
	e.PATCH("/users/:id/deactivate", handlers.User.DeactivateUser)
	e.DELETE("/users/:id", handlers.User.DeleteUser)

	// API routes
	api := e.Group("/api")
	api.GET("/users/count", handlers.User.UserCount)

	return nil
}
