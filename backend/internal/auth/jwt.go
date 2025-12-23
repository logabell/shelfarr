package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("invalid token")
	ErrUserNotFound       = errors.New("user not found")
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	jwt.RegisteredClaims
}

// AuthService handles authentication operations
type AuthService struct {
	db        *gorm.DB
	secretKey []byte
	tokenTTL  time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(db *gorm.DB, secretKey string, tokenTTL time.Duration) *AuthService {
	// If no secret key provided, generate a random one
	if secretKey == "" {
		key := make([]byte, 32)
		rand.Read(key)
		secretKey = base64.StdEncoding.EncodeToString(key)
	}

	if tokenTTL == 0 {
		tokenTTL = 7 * 24 * time.Hour // 7 days default
	}

	return &AuthService{
		db:        db,
		secretKey: []byte(secretKey),
		tokenTTL:  tokenTTL,
	}
}

// User model for authentication
type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	Email        string
	IsAdmin      bool
	CanRead      bool
	CanDelete    bool
	RemoteUser   string `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	User      UserInfo `json:"user"`
}

// UserInfo represents user information returned in responses
type UserInfo struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	IsAdmin   bool   `json:"isAdmin"`
	CanRead   bool   `json:"canRead"`
	CanDelete bool   `json:"canDelete"`
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	var user User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate token
	token, expiresAt, err := s.generateToken(&user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			IsAdmin:   user.IsAdmin,
			CanRead:   user.CanRead,
			CanDelete: user.CanDelete,
		},
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// RefreshToken generates a new token for a valid user
func (s *AuthService) RefreshToken(claims *Claims) (*LoginResponse, error) {
	var user User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, ErrUserNotFound
	}

	token, expiresAt, err := s.generateToken(&user)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		User: UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			IsAdmin:   user.IsAdmin,
			CanRead:   user.CanRead,
			CanDelete: user.CanDelete,
		},
	}, nil
}

// GetUserFromToken gets the full user from a token
func (s *AuthService) GetUserFromToken(claims *Claims) (*User, error) {
	var user User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

// CreateUser creates a new user with hashed password
func (s *AuthService) CreateUser(username, password, email string, isAdmin bool) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		Username:     username,
		PasswordHash: string(hash),
		Email:        email,
		IsAdmin:      isAdmin,
		CanRead:      true,
		CanDelete:    isAdmin,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user User
	if err := s.db.First(&user, userID).Error; err != nil {
		return ErrUserNotFound
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.db.Save(&user).Error
}

// EnsureAdminExists creates a default admin user if none exists
func (s *AuthService) EnsureAdminExists() error {
	var count int64
	s.db.Model(&User{}).Where("is_admin = ?", true).Count(&count)
	
	if count == 0 {
		// Create default admin
		_, err := s.CreateUser("admin", "admin", "", true)
		return err
	}
	return nil
}

func (s *AuthService) generateToken(user *User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.tokenTTL)
	
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// JWTMiddleware creates Echo middleware for JWT authentication
func JWTMiddleware(authService *AuthService) echo.MiddlewareFunc {
	// Check if authentication is disabled (for development)
	authDisabled := os.Getenv("AUTH_DISABLED") == "true"
	
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication for certain paths
			path := c.Path()
			if path == "/health" || path == "/api/v1/auth/login" || strings.HasPrefix(path, "/static") {
				return next(c)
			}

			// If auth is disabled (dev mode), allow all requests
			if authDisabled {
				// Set default dev user
				c.Set("userId", uint(1))
				c.Set("isAdmin", true)
				return next(c)
			}

			// Get token from Authorization header
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Parse "Bearer <token>"
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization format")
			}

			// Validate token
			claims, err := authService.ValidateToken(parts[1])
			if err != nil {
				if errors.Is(err, ErrTokenExpired) {
					return echo.NewHTTPError(http.StatusUnauthorized, "token expired")
				}
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			// Store claims in context
			c.Set("user", claims)
			c.Set("userId", claims.UserID)
			c.Set("isAdmin", claims.IsAdmin)

			return next(c)
		}
	}
}

// SSOMiddleware creates Echo middleware for header-based SSO authentication
func SSOMiddleware(authService *AuthService, headerName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			remoteUser := c.Request().Header.Get(headerName)
			if remoteUser == "" {
				return next(c) // No SSO header, continue to normal auth
			}

			// Find or create user by remote user header
			var user User
			err := authService.db.Where("remote_user = ?", remoteUser).First(&user).Error
			if err != nil {
				// Create new user from SSO
				user = User{
					Username:   remoteUser,
					RemoteUser: remoteUser,
					CanRead:    true,
				}
				authService.db.Create(&user)
			}

			// Store user info in context
			c.Set("user", &Claims{
				UserID:   user.ID,
				Username: user.Username,
				IsAdmin:  user.IsAdmin,
			})
			c.Set("userId", user.ID)
			c.Set("isAdmin", user.IsAdmin)

			return next(c)
		}
	}
}

// RequireAdmin middleware ensures the user is an admin
func RequireAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isAdmin, ok := c.Get("isAdmin").(bool)
			if !ok || !isAdmin {
				return echo.NewHTTPError(http.StatusForbidden, "admin access required")
			}
			return next(c)
		}
	}
}

// RequirePermission middleware ensures the user has a specific permission
func RequirePermission(db *gorm.DB, permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get("user").(*Claims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			// Get full user to check permissions
			var user User
			if err := db.First(&user, claims.UserID).Error; err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
			}

			switch permission {
			case "read":
				if !user.CanRead && !user.IsAdmin {
					return echo.NewHTTPError(http.StatusForbidden, "read permission required")
				}
			case "delete":
				if !user.CanDelete && !user.IsAdmin {
					return echo.NewHTTPError(http.StatusForbidden, "delete permission required")
				}
			case "admin":
				if !user.IsAdmin {
					return echo.NewHTTPError(http.StatusForbidden, "admin permission required")
				}
			}

			return next(c)
		}
	}
}

