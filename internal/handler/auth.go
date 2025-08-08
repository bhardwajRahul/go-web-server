package handler

import (
	"log/slog"
	"net/http"

	"github.com/dunamismax/go-web-server/internal/middleware"
	"github.com/dunamismax/go-web-server/internal/store"
	"github.com/dunamismax/go-web-server/internal/view"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	store       *store.Store
	authService *middleware.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(s *store.Store, authService *middleware.AuthService) *AuthHandler {
	return &AuthHandler{
		store:       s,
		authService: authService,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Name            string `json:"name" validate:"required,min=2,max=100"`
	Password        string `json:"password" validate:"required,password"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	Bio             string `json:"bio,omitempty" validate:"max=500"`
	AvatarURL       string `json:"avatar_url,omitempty" validate:"omitempty,url"`
}

// Validate implements custom validation for RegisterRequest
func (r RegisterRequest) Validate() error {
	if r.Password != r.ConfirmPassword {
		return middleware.ValidationErrors{
			{Field: "confirm_password", Message: "passwords do not match"},
		}
	}
	return nil
}

// LoginPage renders the login page
func (h *AuthHandler) LoginPage(c echo.Context) error {
	// Check if user is already authenticated
	if user, exists := middleware.GetCurrentUser(c); exists && user.IsActive {
		// Redirect to dashboard or home page
		return c.Redirect(http.StatusFound, "/")
	}

	// Get CSRF token for the form
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == HtmxRequestHeader {
		component := view.LoginContent()
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Return full page with layout and CSRF token
	if token != "" {
		component := view.LoginWithCSRF(token)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Fallback to basic template
	component := view.Login()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// RegisterPage renders the registration page
func (h *AuthHandler) RegisterPage(c echo.Context) error {
	// Check if user is already authenticated
	if user, exists := middleware.GetCurrentUser(c); exists && user.IsActive {
		// Redirect to dashboard or home page
		return c.Redirect(http.StatusFound, "/")
	}

	// Get CSRF token for the form
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == HtmxRequestHeader {
		component := view.RegisterContent()
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Return full page with layout and CSRF token
	if token != "" {
		component := view.RegisterWithCSRF(token)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Fallback to basic template
	component := view.Register()
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse and validate request
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Invalid request format",
		).WithContext(c).WithInternal(err)
	}

	if validationErrors := middleware.ValidateStruct(req); len(validationErrors) > 0 {
		return middleware.NewAppErrorWithDetails(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Validation failed",
			validationErrors,
		).WithContext(c)
	}

	// Find user by email
	user, err := h.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		slog.Warn("Login attempt with invalid email",
			"email", req.Email,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeAuthentication,
			http.StatusUnauthorized,
			"Invalid email or password",
		).WithContext(c)
	}

	// For this demo, we'll assume users don't have passwords stored yet
	// In a real application, you'd validate the password here
	// if !middleware.CheckPassword(req.Password, user.PasswordHash) {
	//     return middleware.NewAppError(...)
	// }

	// Check if user is active
	if user.IsActive == nil || !*user.IsActive {
		return middleware.NewAppError(
			middleware.ErrorTypeAuthentication,
			http.StatusUnauthorized,
			"Account is inactive",
		).WithContext(c)
	}

	// Generate JWT token
	authUser := middleware.User{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: *user.IsActive,
	}

	token, err := h.authService.GenerateToken(authUser)
	if err != nil {
		slog.Error("Failed to generate JWT token",
			"user_id", user.ID,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to generate authentication token",
		).WithContext(c).WithInternal(err)
	}

	// Set authentication cookie
	h.authService.SetAuthCookie(c, token)

	slog.Info("User logged in successfully",
		"user_id", user.ID,
		"email", user.Email,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Return success response
	if c.Request().Header.Get("HX-Request") == HtmxRequestHeader {
		// For HTMX requests, trigger a redirect
		c.Response().Header().Set("HX-Redirect", "/")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Login successful",
		})
	}

	// For regular requests, redirect to home page
	return c.Redirect(http.StatusFound, "/")
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse and validate request
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Invalid request format",
		).WithContext(c).WithInternal(err)
	}

	if validationErrors := middleware.ValidateStruct(req); len(validationErrors) > 0 {
		return middleware.NewAppErrorWithDetails(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Validation failed",
			validationErrors,
		).WithContext(c)
	}

	// Custom validation
	if err := req.Validate(); err != nil {
		return middleware.NewAppErrorWithDetails(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Validation failed",
			err,
		).WithContext(c)
	}

	// Hash password (for demo purposes, we'll skip this since we don't have a password field)
	// hashedPassword, err := middleware.HashPassword(req.Password)
	// if err != nil {
	//     return middleware.NewAppError(...)
	// }

	// Create user
	var bioPtr *string
	if req.Bio != "" {
		bioPtr = &req.Bio
	}

	var avatarURLPtr *string
	if req.AvatarURL != "" {
		avatarURLPtr = &req.AvatarURL
	}

	params := store.CreateUserParams{
		Email:     req.Email,
		Name:      req.Name,
		Bio:       bioPtr,
		AvatarUrl: avatarURLPtr,
		// PasswordHash: hashedPassword, // Would add this field to the database
	}

	user, err := h.store.CreateUser(ctx, params)
	if err != nil {
		slog.Error("Failed to create user",
			"email", req.Email,
			"name", req.Name,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to create user account",
		).WithContext(c).WithInternal(err)
	}

	// Generate JWT token for automatic login
	authUser := middleware.User{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: *user.IsActive,
	}

	token, err := h.authService.GenerateToken(authUser)
	if err != nil {
		slog.Error("Failed to generate JWT token after registration",
			"user_id", user.ID,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to generate authentication token",
		).WithContext(c).WithInternal(err)
	}

	// Set authentication cookie
	h.authService.SetAuthCookie(c, token)

	slog.Info("User registered and logged in successfully",
		"user_id", user.ID,
		"email", user.Email,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Return success response
	if c.Request().Header.Get("HX-Request") == HtmxRequestHeader {
		// For HTMX requests, trigger a redirect
		c.Response().Header().Set("HX-Redirect", "/")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Registration successful",
		})
	}

	// For regular requests, redirect to home page
	return c.Redirect(http.StatusFound, "/")
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Clear authentication cookie
	h.authService.ClearAuthCookie(c)

	// Log the logout
	if user, exists := middleware.GetCurrentUser(c); exists {
		slog.Info("User logged out successfully",
			"user_id", user.ID,
			"email", user.Email,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
	}

	// Return success response
	if c.Request().Header.Get("HX-Request") == HtmxRequestHeader {
		// For HTMX requests, trigger a redirect
		c.Response().Header().Set("HX-Redirect", "/login")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Logout successful",
		})
	}

	// For regular requests, redirect to login page
	return c.Redirect(http.StatusFound, "/login")
}

// Profile handles user profile page
func (h *AuthHandler) Profile(c echo.Context) error {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		return c.Redirect(http.StatusFound, "/login")
	}

	// Get CSRF token
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == HtmxRequestHeader {
		component := view.ProfileContent(*user)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Return full page with layout and CSRF token
	if token != "" {
		component := view.ProfileWithCSRF(*user, token)
		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Fallback to basic template
	component := view.Profile(*user)
	return component.Render(c.Request().Context(), c.Response().Writer)
}