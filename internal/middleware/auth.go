package middleware

import (
	"errors"
	"net/http"
	"strings"

	"nat-management-app/internal/models"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware handles authentication for NAT Management
type AuthMiddleware struct {
	authService services.AuthServiceInterface
	logger      *logrus.Logger
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(authService services.AuthServiceInterface, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth middleware that requires authentication for NAT access
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := am.getCurrentUser(c)
		if err != nil {
			am.handleUnauthorized(c, "NAT Authentication required")
			return
		}

		// Set user in context for use in handlers
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("username", user.Username)

		c.Next()
	}
}

// getCurrentUser extracts user from session or JWT (hybrid approach)
func (am *AuthMiddleware) getCurrentUser(c *gin.Context) (*models.User, error) {
	// Try JWT first (access_token cookie) - New authentication method
	if cookie, err := c.Cookie("access_token"); err == nil {
		if user, err := am.authService.ValidateJWTToken(cookie); err == nil {
			return user, nil
		}
	}
	
	// Try JWT from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if user, err := am.authService.ValidateJWTToken(tokenString); err == nil {
			return user, nil
		}
	}

	// Fallback to session-based authentication for backward compatibility
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return nil, errors.New("no valid authentication found")
	}

	if sessionID == "" {
		return nil, errors.New("empty session ID")
	}

	// Validate session
	user, err := am.authService.ValidateSession(sessionID)
	if err != nil {
		// Clear invalid cookie
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		return nil, err
	}

	return user, nil
}

// handleUnauthorized handles unauthorized access
func (am *AuthMiddleware) handleUnauthorized(c *gin.Context, message string) {
	// Check if request expects JSON
	if am.isAPIRequest(c) {
		c.JSON(http.StatusUnauthorized, models.AuthResponse{
			Status:  "error",
			Message: message,
		})
	} else {
		// Redirect to login page
		c.Redirect(http.StatusFound, "/login?redirect="+c.Request.URL.Path)
	}
	c.Abort()
}

// isAPIRequest checks if the request is an API request
func (am *AuthMiddleware) isAPIRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/") ||
		strings.Contains(c.GetHeader("Accept"), "application/json") ||
		strings.Contains(c.GetHeader("Content-Type"), "application/json")
}

// GetUserFromContext gets user from gin context
func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u, true
		}
	}
	return nil, false
}

// GetUserRoleFromContext gets user role from gin context
func GetUserRoleFromContext(c *gin.Context) (models.Role, bool) {
	if role, exists := c.Get("user_role"); exists {
		if r, ok := role.(models.Role); ok {
			return r, true
		}
	}
	return "", false
}

// CORS middleware for NAT Management
func (am *AuthMiddleware) CORSWithAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
