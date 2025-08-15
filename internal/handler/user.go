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

	params := store.CreateUserParams{
		Email:     email,
		Name:      name,
		Bio:       stringPtr(bio),
		AvatarUrl: stringPtr(avatarURL),
	}

	_, err := h.store.CreateUser(ctx, params)
	if err != nil {
		return logAndReturnError(c, "create user", err, http.StatusInternalServerError, "Failed to create user")
	}

	slog.Info("User created successfully",
		"name", name,
		"email", email,
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

	params := store.UpdateUserParams{
		Name:      name,
		Bio:       stringPtr(bio),
		AvatarUrl: stringPtr(avatarURL),
		ID:        id,
	}

	_, err = h.store.UpdateUser(ctx, params)
	if err != nil {
		return logAndReturnError(c, "update user", err, http.StatusInternalServerError, "Failed to update user")
	}

	slog.Info("User updated successfully",
		"id", id,
		"name", name,
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
