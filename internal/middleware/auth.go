package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthConfig holds JWT authentication configuration
type AuthConfig struct {
	SigningKey     []byte
	TokenDuration  time.Duration
	RefreshDuration time.Duration
	Issuer         string
	CookieName     string
	CookieSecure   bool
	CookieHTTPOnly bool
}

// DefaultAuthConfig provides sensible defaults for JWT authentication
var DefaultAuthConfig = AuthConfig{
	SigningKey:      []byte("your-secret-key"), // Should be set from environment
	TokenDuration:   time.Hour * 24,           // 24 hours
	RefreshDuration: time.Hour * 24 * 7,       // 7 days
	Issuer:          "go-web-server",
	CookieName:      "auth_token",
	CookieSecure:    true,
	CookieHTTPOnly:  true,
}

// Claims represents JWT claims with user information
type Claims struct {
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	jwt.RegisteredClaims
}

// User represents authenticated user information
type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// AuthService provides JWT token operations
type AuthService struct {
	config AuthConfig
}

// NewAuthService creates a new authentication service
func NewAuthService(config AuthConfig) *AuthService {
	return &AuthService{config: config}
}

// NewAuthServiceWithDefaults creates a new authentication service with default config
func NewAuthServiceWithDefaults() *AuthService {
	return &AuthService{config: DefaultAuthConfig}
}

// GenerateToken generates a JWT token for a user
func (a *AuthService) GenerateToken(user User) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: user.IsActive,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.config.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.config.Issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.config.SigningKey)
}

// ValidateToken validates a JWT token and returns the claims
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.config.SigningKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// SetAuthCookie sets an authentication cookie
func (a *AuthService) SetAuthCookie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     a.config.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(a.config.TokenDuration.Seconds()),
		Secure:   a.config.CookieSecure,
		HttpOnly: a.config.CookieHTTPOnly,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

// ClearAuthCookie clears the authentication cookie
func (a *AuthService) ClearAuthCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     a.config.CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   a.config.CookieSecure,
		HttpOnly: a.config.CookieHTTPOnly,
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if a password matches the hashed password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// JWTMiddleware creates a JWT authentication middleware
func JWTMiddleware(authService *AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to get token from Authorization header first
			authHeader := c.Request().Header.Get("Authorization")
			var tokenString string

			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// Try to get token from cookie
				cookie, err := c.Cookie(authService.config.CookieName)
				if err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				return NewAppError(
					ErrorTypeAuthentication,
					http.StatusUnauthorized,
					"Authentication required",
				).WithContext(c)
			}

			// Validate token
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				return NewAppError(
					ErrorTypeAuthentication,
					http.StatusUnauthorized,
					"Invalid or expired token",
				).WithContext(c).WithInternal(err)
			}

			// Check if user is active
			if !claims.IsActive {
				return NewAppError(
					ErrorTypeAuthentication,
					http.StatusUnauthorized,
					"User account is inactive",
				).WithContext(c)
			}

			// Store user information in context
			user := User{
				ID:       claims.UserID,
				Email:    claims.Email,
				Name:     claims.Name,
				IsActive: claims.IsActive,
			}
			c.Set("user", user)
			c.Set("user_id", claims.UserID)

			return next(c)
		}
	}
}

// OptionalJWTMiddleware creates a JWT middleware that doesn't require authentication
func OptionalJWTMiddleware(authService *AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to get token from Authorization header first
			authHeader := c.Request().Header.Get("Authorization")
			var tokenString string

			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// Try to get token from cookie
				cookie, err := c.Cookie(authService.config.CookieName)
				if err == nil {
					tokenString = cookie.Value
				}
			}

			// If no token, continue without authentication
			if tokenString == "" {
				return next(c)
			}

			// Validate token if present
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				// Invalid token, continue without authentication
				return next(c)
			}

			// Check if user is active
			if !claims.IsActive {
				// Inactive user, continue without authentication
				return next(c)
			}

			// Store user information in context
			user := User{
				ID:       claims.UserID,
				Email:    claims.Email,
				Name:     claims.Name,
				IsActive: claims.IsActive,
			}
			c.Set("user", user)
			c.Set("user_id", claims.UserID)

			return next(c)
		}
	}
}

// GetCurrentUser retrieves the current authenticated user from context
func GetCurrentUser(c echo.Context) (*User, bool) {
	user := c.Get("user")
	if user == nil {
		return nil, false
	}

	if u, ok := user.(User); ok {
		return &u, true
	}

	return nil, false
}

// GetCurrentUserID retrieves the current authenticated user ID from context
func GetCurrentUserID(c echo.Context) (int64, bool) {
	userID := c.Get("user_id")
	if userID == nil {
		return 0, false
	}

	if id, ok := userID.(int64); ok {
		return id, true
	}

	return 0, false
}

// RequireAuth is a helper middleware that requires authentication
func RequireAuth(authService *AuthService) echo.MiddlewareFunc {
	return JWTMiddleware(authService)
}

// AdminOnly middleware that requires admin privileges (example)
func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user, exists := GetCurrentUser(c)
			if !exists {
				return NewAppError(
					ErrorTypeAuthentication,
					http.StatusUnauthorized,
					"Authentication required",
				).WithContext(c)
			}

			// Example: Check if user is admin (you'd need to add this field to your user model)
			// For now, we'll just check if user exists and is active
			if !user.IsActive {
				return NewAppError(
					ErrorTypeAuthorization,
					http.StatusForbidden,
					"Admin privileges required",
				).WithContext(c)
			}

			return next(c)
		}
	}
}