package services

import (
	"nat-management-app/internal/models"
)

// RouterServiceInterface defines the interface for router management operations
type RouterServiceInterface interface {
	GetAllRouters(userRole string) ([]models.RouterResponse, error)
	GetRouter(routerID string, userRole string) (*models.RouterResponse, error)
	CreateRouter(req *models.RouterCreateRequest, userRole string) (*models.RouterResponse, error)
	UpdateRouter(routerID string, req *models.RouterUpdateRequest, userRole string) (*models.RouterResponse, error)
	DeleteRouter(routerID string, userRole string) error
	TestRouter(routerID string, userRole string) (*models.RouterTestResponse, error)
	GetRouterStats(userRole string) (*models.RouterStatsResponse, error)
	GetRoutersForNATService() (map[string]models.NATRouterConfig, error)
	ReloadConfiguration() error
	GetConfigurationPath() string
}

// AuthServiceInterface defines the interface for authentication operations
type AuthServiceInterface interface {
	Login(username, password, ipAddress, userAgent string) (*models.AuthResponse, error)
	LoginWithJWT(username, password, ipAddress, userAgent string) (*models.AuthResponse, error)
	Logout(sessionID string) error
	LogoutJWT(accessToken, refreshToken string) error
	ValidateSession(sessionID string) (*models.User, error)
	ValidateJWTToken(tokenString string) (*models.User, error)
	RefreshToken(refreshToken, ipAddress, userAgent string) (*models.AuthResponse, error)
	RevokeAllUserTokens(userID int) error
	GetJWTPublicKey() (string, error)
}
