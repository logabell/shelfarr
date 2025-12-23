package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/auth"
)

// AuthHandlers handles authentication-related API endpoints
type AuthHandlers struct {
	authService *auth.AuthService
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(authService *auth.AuthService) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
	}
}

// Login handles user login
func (h *AuthHandlers) Login(c echo.Context) error {
	var req auth.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and password are required"})
	}

	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
	}

	return c.JSON(http.StatusOK, resp)
}

// Refresh refreshes a JWT token
func (h *AuthHandlers) Refresh(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
	}

	resp, err := h.authService.RefreshToken(claims)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to refresh token"})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandlers) GetCurrentUser(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
	}

	user, err := h.authService.GetUserFromToken(claims)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
	}

	return c.JSON(http.StatusOK, auth.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		CanRead:   user.CanRead,
		CanDelete: user.CanDelete,
	})
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ChangePassword changes the user's password
func (h *AuthHandlers) ChangePassword(c echo.Context) error {
	claims, ok := c.Get("user").(*auth.Claims)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Authentication required"})
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Both old and new passwords are required"})
	}

	if len(req.NewPassword) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "New password must be at least 6 characters"})
	}

	if err := h.authService.ChangePassword(claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		if err == auth.ErrInvalidCredentials {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Current password is incorrect"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to change password"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Password changed successfully"})
}
