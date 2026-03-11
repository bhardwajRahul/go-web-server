package handler

import (
	"log/slog"
	"net/http"

	"github.com/dunamismax/go-web-server/internal/middleware"
	"github.com/dunamismax/go-web-server/internal/store"
	"github.com/dunamismax/go-web-server/internal/view"
	"github.com/labstack/echo/v4"
)

// UserHandler handles all user-related HTTP requests including CRUD operations.
type UserHandler struct {
	store       *store.Store
	authService *middleware.SessionAuthService
}

// NewUserHandler creates a new UserHandler with the given store.
func NewUserHandler(s *store.Store, authService *middleware.SessionAuthService) *UserHandler {
	return &UserHandler{
		store:       s,
		authService: authService,
	}
}

// ManagedUserUpdateRequest represents the editable user fields from the CRUD form.
type ManagedUserUpdateRequest struct {
	Email           string `json:"email" form:"email" validate:"required,email"`
	Name            string `json:"name" form:"name" validate:"required,min=2,max=100"`
	Password        string `json:"password,omitempty" form:"password" validate:"omitempty,password"`
	ConfirmPassword string `json:"confirm_password,omitempty" form:"confirm_password"`
	Bio             string `json:"bio,omitempty" form:"bio" validate:"max=500"`
	AvatarURL       string `json:"avatar_url,omitempty" form:"avatar_url" validate:"omitempty,url"`
}

// Validate implements custom validation for ManagedUserUpdateRequest.
func (r ManagedUserUpdateRequest) Validate() error {
	if r.Password == "" && r.ConfirmPassword == "" {
		return nil
	}

	if r.Password == "" || r.ConfirmPassword == "" {
		return middleware.ValidationErrors{
			{Field: "confirm_password", Message: "password and confirmation are both required to change the password"},
		}
	}

	if r.Password != r.ConfirmPassword {
		return middleware.ValidationErrors{
			{Field: "confirm_password", Message: "passwords do not match"},
		}
	}

	return nil
}

// Users renders the main user management page.
func (h *UserHandler) Users(c echo.Context) error {
	token := setupCSRFHeaders(c)

	return renderWithCSRF(c,
		view.UsersContent(),       // HTMX component
		view.UsersWithCSRF(token), // Full page component with CSRF
		view.Users(),              // Basic component
	)
}

// UserList returns the list of users as HTML fragment.
func (h *UserHandler) UserList(c echo.Context) error {
	ctx := c.Request().Context()
	setupCSRFHeaders(c)

	users, err := h.store.ListUsers(ctx)
	if err != nil {
		return logAndReturnError(c, "fetch users", err, http.StatusInternalServerError, "Failed to fetch users")
	}

	return view.UserList(users).Render(ctx, c.Response().Writer)
}

// UserCount returns the count of active users.
func (h *UserHandler) UserCount(c echo.Context) error {
	ctx := c.Request().Context()
	setupCSRFHeaders(c)

	count, err := h.store.CountUsers(ctx)
	if err != nil {
		return logAndReturnError(c, "count users", err, http.StatusInternalServerError, "Failed to count users")
	}

	return view.UserCount(count).Render(ctx, c.Response().Writer)
}

// UserForm renders the user creation/edit form.
func (h *UserHandler) UserForm(c echo.Context) error {
	token := setupCSRFHeaders(c)
	return view.UserForm(nil, token).Render(c.Request().Context(), c.Response().Writer)
}

// EditUserForm renders the user edit form with existing data.
func (h *UserHandler) EditUserForm(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	user, err := h.store.GetUser(ctx, id)
	if err != nil {
		return logAndReturnError(c, "fetch user", err, http.StatusNotFound, "User not found")
	}

	token := setupCSRFHeaders(c)
	return view.UserForm(&user, token).Render(ctx, c.Response().Writer)
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return validationError(c, "Invalid request format", err)
	}

	if validationErrors := middleware.ValidateStruct(req); len(validationErrors) > 0 {
		return validationErrorWithDetails(c, "Validation failed", validationErrors)
	}

	if err := req.Validate(); err != nil {
		return validationErrorWithDetails(c, "Validation failed", err)
	}

	hashedPassword, err := h.authService.HashPasswordArgon2(req.Password)
	if err != nil {
		return internalError(c, "Failed to process password", err)
	}

	params := store.CreateUserParams{
		Email:        req.Email,
		Name:         req.Name,
		Bio:          stringPtr(req.Bio),
		AvatarUrl:    stringPtr(req.AvatarURL),
		PasswordHash: hashedPassword,
	}

	_, err = h.store.CreateUser(ctx, params)
	if err != nil {
		slog.Error("Failed to create user",
			"email", req.Email,
			"name", req.Name,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
		return databaseWriteError(c, err, "Failed to create user")
	}

	slog.Info("User created successfully",
		"name", req.Name,
		"email", req.Email,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userCreated")

	users, err := h.store.ListUsers(ctx)
	if err != nil {
		return logAndReturnError(c, "fetch updated users", err, http.StatusInternalServerError, "Failed to fetch updated users")
	}

	return view.UserList(users).Render(ctx, c.Response().Writer)
}

// UpdateUser updates an existing user.
func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	var req ManagedUserUpdateRequest
	if err := c.Bind(&req); err != nil {
		return validationError(c, "Invalid request format", err)
	}

	if validationErrors := middleware.ValidateStruct(req); len(validationErrors) > 0 {
		return validationErrorWithDetails(c, "Validation failed", validationErrors)
	}

	if err := req.Validate(); err != nil {
		return validationErrorWithDetails(c, "Validation failed", err)
	}

	params := store.UpdateUserParams{
		Email:     req.Email,
		Name:      req.Name,
		Bio:       stringPtr(req.Bio),
		AvatarUrl: stringPtr(req.AvatarURL),
		ID:        id,
	}

	if req.Password != "" {
		hashedPassword, err := h.authService.HashPasswordArgon2(req.Password)
		if err != nil {
			return internalError(c, "Failed to process password", err)
		}

		_, err = h.store.UpdateUserPassword(ctx, store.UpdateUserPasswordParams{
			Email:        req.Email,
			Name:         req.Name,
			Bio:          stringPtr(req.Bio),
			AvatarUrl:    stringPtr(req.AvatarURL),
			PasswordHash: hashedPassword,
			ID:           id,
		})
		if err != nil {
			slog.Error("Failed to update user with password",
				"id", id,
				"email", req.Email,
				"error", err,
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
			return databaseWriteError(c, err, "Failed to update user")
		}
	} else {
		_, err = h.store.UpdateUser(ctx, params)
		if err != nil {
			slog.Error("Failed to update user",
				"id", id,
				"email", req.Email,
				"error", err,
				"request_id", c.Response().Header().Get(echo.HeaderXRequestID))
			return databaseWriteError(c, err, "Failed to update user")
		}
	}

	slog.Info("User updated successfully",
		"id", id,
		"name", req.Name,
		"email", req.Email,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userUpdated")

	users, err := h.store.ListUsers(ctx)
	if err != nil {
		return logAndReturnError(c, "fetch updated users", err, http.StatusInternalServerError, "Failed to fetch updated users")
	}

	return view.UserList(users).Render(ctx, c.Response().Writer)
}

// DeactivateUser deactivates a user instead of deleting.
func (h *UserHandler) DeactivateUser(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	err = h.store.DeactivateUser(ctx, id)
	if err != nil {
		return logAndReturnError(c, "deactivate user", err, http.StatusInternalServerError, "Failed to deactivate user")
	}

	slog.Info("User deactivated successfully",
		"id", id,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Get the updated user and return the row
	user, err := h.store.GetUser(ctx, id)
	if err != nil {
		return logAndReturnError(c, "fetch updated user", err, http.StatusInternalServerError, "Failed to fetch updated user")
	}

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userDeactivated")

	return view.UserRow(user).Render(ctx, c.Response().Writer)
}

// DeleteUser permanently deletes a user.
func (h *UserHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := parseIDParam(c, "id")
	if err != nil {
		return err
	}

	err = h.store.DeleteUser(ctx, id)
	if err != nil {
		return logAndReturnError(c, "delete user", err, http.StatusInternalServerError, "Failed to delete user")
	}

	slog.Info("User deleted successfully",
		"id", id,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userDeleted")

	// Return empty response since the row should be removed
	return c.NoContent(http.StatusOK)
}
