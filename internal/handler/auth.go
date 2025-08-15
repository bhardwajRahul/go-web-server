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
	authService *middleware.SessionAuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(s *store.Store, authService *middleware.SessionAuthService) *AuthHandler {
	return &AuthHandler{
		store:       s,
		authService: authService,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=1"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email           string `json:"email" form:"email" validate:"required,email"`
	Name            string `json:"name" form:"name" validate:"required,min=2,max=100"`
	Password        string `json:"password" form:"password" validate:"required,password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" validate:"required"`
	Bio             string `json:"bio,omitempty" form:"bio" validate:"max=500"`
	AvatarURL       string `json:"avatar_url,omitempty" form:"avatar_url" validate:"omitempty,url"`
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
	if user, exists := h.authService.GetCurrentUser(c); exists && user.IsActive {
		return c.Redirect(http.StatusFound, RouteHome)
	}

	token := middleware.GetCSRFToken(c)

	return renderWithCSRF(c,
		view.LoginContent(),       // HTMX component
		view.LoginWithCSRF(token), // Full page component with CSRF
		view.Login(),              // Basic component
	)
}

// RegisterPage renders the registration page
func (h *AuthHandler) RegisterPage(c echo.Context) error {
	// Check if user is already authenticated
	if user, exists := h.authService.GetCurrentUser(c); exists && user.IsActive {
		return c.Redirect(http.StatusFound, RouteHome)
	}

	token := middleware.GetCSRFToken(c)

	return renderWithCSRF(c,
		view.RegisterContent(),       // HTMX component
		view.RegisterWithCSRF(token), // Full page component with CSRF
		view.Register(),              // Basic component
	)
}

// Login handles user login
func (h *AuthHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse and validate request
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return validationError(c, "Invalid request format", err)
	}

	if validationErrors := middleware.ValidateStruct(req); len(validationErrors) > 0 {
		return validationErrorWithDetails(c, "Validation failed", validationErrors)
	}

	// Find user by email
	user, err := h.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		slog.Warn("Login attempt with invalid email",
			"email", req.Email,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return authenticationError(c, "Invalid email or password")
	}

	// Verify password if user has a password hash
	if user.PasswordHash != nil {
		valid, err := h.authService.VerifyPasswordArgon2(req.Password, *user.PasswordHash)
		if err != nil {
			slog.Error("Password verification failed",
				"error", err,
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
			return internalError(c, "Authentication error", err)
		}
		if !valid {
			return authenticationError(c, "Invalid email or password")
		}
	} else {
		// For demo users without passwords, allow any password
		slog.Warn("User logging in without password set", "email", req.Email)
	}

	// Check if user is active
	if user.IsActive == nil || !*user.IsActive {
		return authenticationError(c, "Account is inactive")
	}

	// Create user session
	authUser := middleware.User{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: *user.IsActive,
	}

	err = h.authService.LoginUser(c, authUser)
	if err != nil {
		slog.Error("Failed to create user session",
			"user_id", user.ID,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to create user session",
		).WithContext(c).WithInternal(err)
	}

	slog.Info("User logged in successfully",
		"user_id", user.ID,
		"email", user.Email,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Return success response
	return redirectOrHtmx(c, RouteHome, MsgLoginSuccess)
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse and validate request
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return validationError(c, "Invalid request format", err)
	}

	if validationErrors := middleware.ValidateStruct(req); len(validationErrors) > 0 {
		return validationErrorWithDetails(c, "Validation failed", validationErrors)
	}

	// Custom validation
	if err := req.Validate(); err != nil {
		return validationErrorWithDetails(c, "Validation failed", err)
	}

	// Hash password using Argon2id
	hashedPassword, err := h.authService.HashPasswordArgon2(req.Password)
	if err != nil {
		slog.Error("Failed to hash password",
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to process password",
		).WithContext(c).WithInternal(err)
	}

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
		Email:        req.Email,
		Name:         req.Name,
		Bio:          bioPtr,
		AvatarUrl:    avatarURLPtr,
		PasswordHash: &hashedPassword,
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

	// Create user session for automatic login
	authUser := middleware.User{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: *user.IsActive,
	}

	err = h.authService.LoginUser(c, authUser)
	if err != nil {
		slog.Error("Failed to create user session after registration",
			"user_id", user.ID,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to create user session",
		).WithContext(c).WithInternal(err)
	}

	slog.Info("User registered and logged in successfully",
		"user_id", user.ID,
		"email", user.Email,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Return success response
	return redirectOrHtmx(c, RouteHome, MsgRegisterSuccess)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Log the logout
	if user, exists := h.authService.GetCurrentUser(c); exists {
		slog.Info("User logged out successfully",
			"user_id", user.ID,
			"email", user.Email,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
	}

	// Destroy user session
	err := h.authService.LogoutUser(c)
	if err != nil {
		slog.Error("Failed to destroy session",
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
	}

	// Return success response
	return redirectOrHtmx(c, RouteLogin, MsgLogoutSuccess)
}

// Profile handles user profile page
func (h *AuthHandler) Profile(c echo.Context) error {
	user, exists := h.authService.GetCurrentUser(c)
	if !exists {
		return c.Redirect(http.StatusFound, RouteLogin)
	}

	token := middleware.GetCSRFToken(c)

	return renderWithCSRF(c,
		view.ProfileContent(*user),         // HTMX component
		view.ProfileWithCSRF(*user, token), // Full page component with CSRF
		view.Profile(*user),                // Basic component
	)
}
