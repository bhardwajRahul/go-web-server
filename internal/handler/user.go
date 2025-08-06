package handler

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/dunamismax/go-web-server/internal/middleware"
	"github.com/dunamismax/go-web-server/internal/store"
	"github.com/dunamismax/go-web-server/internal/view"
	"github.com/labstack/echo/v4"
)

const (
	htmxRequestHeader = "true"
)

// UserHandler handles all user-related HTTP requests including CRUD operations.
type UserHandler struct {
	store *store.Store
}

// NewUserHandler creates a new UserHandler with the given store.
func NewUserHandler(s *store.Store) *UserHandler {
	return &UserHandler{
		store: s,
	}
}

// Users renders the main user management page.
func (h *UserHandler) Users(c echo.Context) error {
	// Get CSRF token for initial requests
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	// Check if this is an HTMX request for partial content
	if c.Request().Header.Get("HX-Request") == HtmxRequestHeader {
		component := view.UsersContent()

		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Return full page with layout and CSRF token
	if token != "" {
		component := view.UsersWithCSRF(token)

		return component.Render(c.Request().Context(), c.Response().Writer)
	}

	// Fallback to basic template
	component := view.Users()

	return component.Render(c.Request().Context(), c.Response().Writer)
}

// UserList returns the list of users as HTML fragment.
func (h *UserHandler) UserList(c echo.Context) error {
	ctx := c.Request().Context()

	// Set CSRF token in response header for HTMX to pick up
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	users, err := h.store.ListUsers(ctx)
	if err != nil {
		slog.Error("Failed to fetch users",
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to fetch users",
		).WithContext(c).WithInternal(err)
	}

	component := view.UserList(users)

	return component.Render(ctx, c.Response().Writer)
}

// UserCount returns the count of active users.
func (h *UserHandler) UserCount(c echo.Context) error {
	ctx := c.Request().Context()

	// Set CSRF token in response header for HTMX to pick up
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	count, err := h.store.CountUsers(ctx)
	if err != nil {
		slog.Error("Failed to count users",
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to count users",
		).WithContext(c).WithInternal(err)
	}

	// Update metrics
	middleware.UpdateActiveUsers(count)

	component := view.UserCount(count)

	return component.Render(ctx, c.Response().Writer)
}

// UserForm renders the user creation/edit form.
func (h *UserHandler) UserForm(c echo.Context) error {
	// Set CSRF token in response header for HTMX to pick up
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	component := view.UserForm(nil, token)

	return component.Render(c.Request().Context(), c.Response().Writer)
}

// EditUserForm renders the user edit form with existing data.
func (h *UserHandler) EditUserForm(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Invalid user ID format",
		).WithContext(c).WithInternal(err)
	}

	user, err := h.store.GetUser(ctx, id)
	if err != nil {
		slog.Error("Failed to fetch user",
			"id", id,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeNotFound,
			http.StatusNotFound,
			"User not found",
		).WithContext(c).WithInternal(err)
	}

	// Set CSRF token in response header for HTMX to pick up
	token := middleware.GetCSRFToken(c)
	if token != "" {
		c.Response().Header().Set("X-CSRF-Token", token)
	}

	component := view.UserForm(&user, token)

	return component.Render(ctx, c.Response().Writer)
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.FormValue("name")
	email := c.FormValue("email")
	bio := c.FormValue("bio")
	avatarURL := c.FormValue("avatar_url")

	// Validate required fields
	if name == "" || email == "" {
		return middleware.NewAppErrorWithDetails(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Validation failed",
			map[string]string{
				"name":  "Name is required",
				"email": "Email is required",
			},
		).WithContext(c)
	}

	var bioSQL sql.NullString
	if bio != "" {
		bioSQL = sql.NullString{String: bio, Valid: true}
	}

	var avatarURLSQL sql.NullString
	if avatarURL != "" {
		avatarURLSQL = sql.NullString{String: avatarURL, Valid: true}
	}

	params := store.CreateUserParams{
		Email:     email,
		Name:      name,
		Bio:       bioSQL,
		AvatarUrl: avatarURLSQL,
	}

	_, err := h.store.CreateUser(ctx, params)
	if err != nil {
		slog.Error("Failed to create user",
			"name", name,
			"email", email,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to create user",
		).WithContext(c).WithInternal(err)
	}

	slog.Info("User created successfully",
		"name", name,
		"email", email,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Record metrics
	middleware.RecordUserCreated()

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userCreated")

	users, err := h.store.ListUsers(ctx)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to fetch updated users",
		).WithContext(c).WithInternal(err)
	}

	component := view.UserList(users)

	return component.Render(ctx, c.Response().Writer)
}

// UpdateUser updates an existing user.
func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Invalid user ID format",
		).WithContext(c).WithInternal(err)
	}

	name := c.FormValue("name")
	bio := c.FormValue("bio")
	avatarURL := c.FormValue("avatar_url")

	if name == "" {
		return middleware.NewAppErrorWithDetails(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Validation failed",
			map[string]string{"name": "Name is required"},
		).WithContext(c)
	}

	var bioSQL sql.NullString
	if bio != "" {
		bioSQL = sql.NullString{String: bio, Valid: true}
	}

	var avatarURLSQL sql.NullString
	if avatarURL != "" {
		avatarURLSQL = sql.NullString{String: avatarURL, Valid: true}
	}

	params := store.UpdateUserParams{
		Name:      name,
		Bio:       bioSQL,
		AvatarUrl: avatarURLSQL,
		ID:        id,
	}

	_, err = h.store.UpdateUser(ctx, params)
	if err != nil {
		slog.Error("Failed to update user",
			"id", id,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to update user",
		).WithContext(c).WithInternal(err)
	}

	slog.Info("User updated successfully",
		"id", id,
		"name", name,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userUpdated")

	users, err := h.store.ListUsers(ctx)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to fetch updated users",
		).WithContext(c).WithInternal(err)
	}

	component := view.UserList(users)

	return component.Render(ctx, c.Response().Writer)
}

// DeactivateUser deactivates a user instead of deleting.
func (h *UserHandler) DeactivateUser(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Invalid user ID format",
		).WithContext(c).WithInternal(err)
	}

	err = h.store.DeactivateUser(ctx, id)
	if err != nil {
		slog.Error("Failed to deactivate user",
			"id", id,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to deactivate user",
		).WithContext(c).WithInternal(err)
	}

	slog.Info("User deactivated successfully",
		"id", id,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Get the updated user and return the row
	user, err := h.store.GetUser(ctx, id)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to fetch updated user",
		).WithContext(c).WithInternal(err)
	}

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userDeactivated")

	component := view.UserRow(user)

	return component.Render(ctx, c.Response().Writer)
}

// DeleteUser permanently deletes a user.
func (h *UserHandler) DeleteUser(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrorTypeValidation,
			http.StatusBadRequest,
			"Invalid user ID format",
		).WithContext(c).WithInternal(err)
	}

	err = h.store.DeleteUser(ctx, id)
	if err != nil {
		slog.Error("Failed to delete user",
			"id", id,
			"error", err,
			"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

		return middleware.NewAppError(
			middleware.ErrorTypeInternal,
			http.StatusInternalServerError,
			"Failed to delete user",
		).WithContext(c).WithInternal(err)
	}

	slog.Info("User deleted successfully",
		"id", id,
		"request_id", c.Response().Header().Get(echo.HeaderXRequestID))

	// Trigger custom event for HTMX
	c.Response().Header().Set("HX-Trigger", "userDeleted")

	// Return empty response since the row should be removed
	return c.NoContent(http.StatusOK)
}
