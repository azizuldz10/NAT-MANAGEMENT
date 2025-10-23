package api

import (
	"net/http"
	"strings"

	"nat-management-app/internal/middleware"
	"nat-management-app/internal/models"
	"nat-management-app/internal/services"
	"nat-management-app/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthHandler handles authentication for NAT Management
type AuthHandler struct {
	authService        services.AuthServiceInterface
	routerService      services.RouterServiceInterface
	userService        *services.UserService // Added to get user-specific routers
	activityLogService *services.ActivityLogService
	logger             *logrus.Logger
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService services.AuthServiceInterface, routerService services.RouterServiceInterface, userService *services.UserService, activityLogService *services.ActivityLogService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		routerService:      routerService,
		userService:        userService,
		activityLogService: activityLogService,
		logger:             logger,
	}
}

// Login handles POST /api/auth/login dengan JWT implementation
func (ah *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Use enhanced validation error response
		utils.RespondValidationError(c, err)
		return
	}

	// Validate required fields manually (additional validation)
	if strings.TrimSpace(req.Username) == "" {
		utils.RespondInvalidInput(c, "username", "Username is required")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		utils.RespondInvalidInput(c, "password", "Password is required")
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Attempt JWT login
	response, err := ah.authService.LoginWithJWT(req.Username, req.Password, ipAddress, userAgent)
	if err != nil {
		// Log failed login attempt with enhanced logging
		if ah.activityLogService != nil {
			deviceInfo := utils.ParseUserAgent(userAgent)
			ah.activityLogService.CreateLog(&models.ActivityLogCreate{
				Username:     req.Username,
				ActionType:   models.ActionLogin,
				ResourceType: models.ResourceAuth,
				Description:  "Failed login attempt",
				IPAddress:    ipAddress,
				UserAgent:    userAgent,
				DeviceInfo:   deviceInfo,
				Status:       models.StatusFailed,
				ErrorMessage: err.Error(),
			})
		}

		// Send user-friendly error response
		errDetail := models.ErrInvalidCredentials.
			WithDetails("Login failed for user: " + req.Username).
			WithSuggestion("Please verify your username and password. If you've forgotten your credentials, contact your administrator.")

		utils.RespondWithError(c, http.StatusUnauthorized, errDetail)
		return
	}

	// Extract tokens from response untuk set cookies
	if data, ok := response.Data.(map[string]interface{}); ok {
		if tokens, exists := data["tokens"]; exists {
			if tokenPair, ok := tokens.(*models.TokenPair); ok {
				// Set secure JWT cookies
				domain := ""
				secure := false  // Allow HTTP for development, set true for production
				httpOnly := true

				// Set access token cookie (short-lived)
				c.SetSameSite(http.SameSiteLaxMode) // Use Lax for development compatibility
				c.SetCookie("access_token", tokenPair.AccessToken, 900, "/", domain, secure, httpOnly) // 15 minutes

				// Set refresh token cookie (long-lived) - More secure
				c.SetCookie("refresh_token", tokenPair.RefreshToken, 604800, "/", domain, secure, httpOnly) // 7 days

				ah.logger.Infof("🍪 JWT cookies set untuk user: %s", req.Username)
			}
		}

		// Log successful login with enhanced logging
		if ah.activityLogService != nil {
			// Extract user info from response
			if user, ok := data["user"].(*models.User); ok {
				userID := user.ID
				deviceInfo := utils.ParseUserAgent(userAgent)
				ah.activityLogService.CreateLog(&models.ActivityLogCreate{
					UserID:       &userID,
					Username:     req.Username,
					UserRole:     string(user.Role),
					ActionType:   models.ActionLogin,
					ResourceType: models.ResourceAuth,
					Description:  "Successful login",
					IPAddress:    ipAddress,
					UserAgent:    userAgent,
					DeviceInfo:   deviceInfo,
					Status:       models.StatusSuccess,
				})
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles POST /api/auth/logout dengan JWT revocation
func (ah *AuthHandler) Logout(c *gin.Context) {
	var accessToken, refreshToken string

	// Get access token dari header atau cookie
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		accessToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Fallback to cookie
		if cookie, err := c.Cookie("access_token"); err == nil {
			accessToken = cookie
		}
	}

	// Get refresh token dari cookie
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		refreshToken = cookie
	}

	// Revoke JWT tokens
	if accessToken != "" || refreshToken != "" {
		err := ah.authService.LogoutJWT(accessToken, refreshToken)
		if err != nil {
			ah.logger.Warnf("Failed to revoke JWT tokens: %v", err)
		}
	}

	// Clear all auth cookies dengan secure settings
	domain := ""
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", "", -1, "/", domain, false, true)    // Allow HTTP for development
	c.SetCookie("refresh_token", "", -1, "/", domain, false, true)   // Allow HTTP for development
	c.SetCookie("session_id", "", -1, "/", domain, false, true)      // Legacy session cleanup

	ah.logger.Infof("👋 JWT Logout completed, all tokens revoked")

	// Log successful logout
	if ah.activityLogService != nil {
		// Get user from context if available
		user, exists := middleware.GetUserFromContext(c)
		if exists {
			userID := user.ID
			ah.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &userID,
				Username:     user.Username,
				UserRole:     string(user.Role),
				ActionType:   models.ActionLogout,
				ResourceType: models.ResourceAuth,
				Description:  "User logged out",
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Status:  "success",
		Message: "Logout berhasil, semua token telah direvoke",
	})
}

// Me handles GET /api/auth/me - returns current user info
func (ah *AuthHandler) Me(c *gin.Context) {
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		utils.RespondUnauthorized(c)
		return
	}

	// Get NAT router access from database
	// Priority: 1. User-specific routers (user_routers table), 2. Role-based routers (router_access_control)
	routerAccess, err := ah.getRouterAccessForUserID(user.ID, string(user.Role))
	if err != nil {
		ah.logger.Warnf("Failed to get router access for user %s: %v", user.Username, err)
		// Fallback to empty array if error
		routerAccess = []string{}
	}

	// Add NAT router access info
	data := gin.H{
		"user":              user,
		"nat_router_access": routerAccess,
		"permissions": gin.H{
			"can_access_nat":  user.Role.CanAccessNATManagement(),
			"has_full_access": user.Role.HasFullAccess(),
		},
	}

	utils.RespondSuccess(c, data)
}

// getRouterAccessForUser gets the list of routers accessible to a user role from database
func (ah *AuthHandler) getRouterAccessForUser(userRole string) ([]string, error) {
	// Get routers from database via RouterService
	routers, err := ah.routerService.GetAllRouters(userRole)
	if err != nil {
		return nil, err
	}

	// Extract router names
	var routerNames []string
	for _, router := range routers {
		routerNames = append(routerNames, router.Name)
	}

	ah.logger.Debugf("User role '%s' has access to %d routers from database", userRole, len(routerNames))
	return routerNames, nil
}

// getRouterAccessForUserID gets routers for a specific user
// Priority: 1. User-specific routers (user_routers table), 2. Role-based (router_access_control)
func (ah *AuthHandler) getRouterAccessForUserID(userID int, userRole string) ([]string, error) {
	// Try to get user-specific router access from user_routers table
	userRouterNames, err := ah.userService.GetUserRouters(userID)
	if err != nil {
		ah.logger.Warnf("Failed to get user-specific routers for user ID %d: %v", userID, err)
		// Fallback to role-based access
		return ah.getRouterAccessForUser(userRole)
	}

	// If user has specific router assignments, use those
	if len(userRouterNames) > 0 {
		ah.logger.Debugf("User ID %d has access to %d user-specific routers", userID, len(userRouterNames))
		return userRouterNames, nil
	}

	// Fallback to role-based access if no user-specific routers
	ah.logger.Debugf("User ID %d has no specific routers, falling back to role-based access", userID)
	return ah.getRouterAccessForUser(userRole)
}

// CheckAuth handles GET /api/auth/check - check if user is authenticated (support both session and JWT)
func (ah *AuthHandler) CheckAuth(c *gin.Context) {
	// Try JWT first (from Authorization header or access_token cookie)
	var user *models.User
	var authMethod string
	
	// Check JWT token dari Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if jwtUser, err := ah.authService.ValidateJWTToken(tokenString); err == nil {
			user = jwtUser
			authMethod = "jwt_header"
		}
	}
	
	// Fallback: Check JWT dari access_token cookie
	if user == nil {
		if cookie, err := c.Cookie("access_token"); err == nil {
			if jwtUser, err := ah.authService.ValidateJWTToken(cookie); err == nil {
				user = jwtUser
				authMethod = "jwt_cookie"
			}
		}
	}
	
	// Fallback: Check session-based auth untuk backward compatibility
	if user == nil {
		if sessionUser, exists := middleware.GetUserFromContext(c); exists {
			user = sessionUser
			authMethod = "session"
		}
	}
	
	// Final fallback: Check session cookie directly
	if user == nil {
		if sessionID, err := c.Cookie("session_id"); err == nil {
			if sessionUser, err := ah.authService.ValidateSession(sessionID); err == nil {
				user = sessionUser
				authMethod = "session_cookie"
			}
		}
	}

	if user == nil {
		errDetail := models.ErrUnauthorized.
			WithSuggestion("Please log in to access this resource. Your session may have expired.")
		utils.RespondWithError(c, http.StatusUnauthorized, errDetail)
		return
	}

	ah.logger.Debugf("👤 User authenticated via %s: %s", authMethod, user.Username)

	// Get NAT router access from database
	routerAccess, err := ah.getRouterAccessForUser(string(user.Role))
	if err != nil {
		ah.logger.Warnf("Failed to get router access for user %s: %v", user.Username, err)
		// Fallback to empty array if error
		routerAccess = []string{}
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Status: "success",
		Data: gin.H{
			"authenticated":     true,
			"user":              user,
			"auth_method":       authMethod,
			"nat_router_access": routerAccess,
		},
	})
}

// RefreshToken handles POST /api/auth/refresh untuk refresh JWT tokens
func (ah *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	
	// Try to get refresh token dari request body atau cookie
	if err := c.ShouldBindJSON(&req); err != nil {
		// Fallback: get dari cookie
		if cookie, cookieErr := c.Cookie("refresh_token"); cookieErr == nil {
			req.RefreshToken = cookie
		} else {
			c.JSON(http.StatusBadRequest, models.AuthResponse{
				Status:  "error",
				Message: "Refresh token tidak ditemukan: " + err.Error(),
			})
			return
		}
	}

	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, models.AuthResponse{
			Status:  "error",
			Message: "Refresh token wajib diisi",
		})
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Attempt token refresh
	response, err := ah.authService.RefreshToken(req.RefreshToken, ipAddress, userAgent)
	if err != nil {
		ah.logger.Warnf("🔄 Token refresh failed dari IP %s: %v", ipAddress, err)
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Update cookies dengan new access token
	if data, ok := response.Data.(map[string]interface{}); ok {
		if tokens, exists := data["tokens"]; exists {
			if tokenPair, ok := tokens.(*models.TokenPair); ok {
				domain := ""
				secure := false  // Allow HTTP for development
				httpOnly := true
				
				// Update access token cookie
				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie("access_token", tokenPair.AccessToken, 900, "/", domain, secure, httpOnly) // 15 minutes
				
				ah.logger.Infof("🔄 Access token refreshed dan cookie updated")
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetJWTPublicKey handles GET /api/auth/jwt-public-key untuk external services
func (ah *AuthHandler) GetJWTPublicKey(c *gin.Context) {
	publicKey, err := ah.authService.GetJWTPublicKey()
	if err != nil {
		ah.logger.Errorf("Failed to get JWT public key: %v", err)
		c.JSON(http.StatusInternalServerError, models.AuthResponse{
			Status:  "error",
			Message: "Gagal mendapatkan public key",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"public_key": publicKey,
		"algorithm":  "RS256",
	})
}

