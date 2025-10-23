package models

import (
	"time"
)

// Role represents user roles in the NAT management system
type Role string

const (
	RoleAdministrator Role = "Administrator"
	RoleHeadBranch1   Role = "Head Branch 1"
	RoleHeadBranch2   Role = "Head Branch 2"
	RoleHeadBranch3   Role = "Head Branch 3"
)

// User represents a user in the system
type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"password,omitempty"` // Omit in JSON responses
	FullName    string    `json:"full_name"`
	Email       string    `json:"email"`
	Role        Role      `json:"role"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// UserSession represents an active user session
type UserSession struct {
	SessionID string    `json:"session_id"`
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents authentication responses
type AuthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PasswordChangeRequest represents a password change request
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// TokenPair represents JWT access and refresh token pair
type TokenPair struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"` // "Bearer"
	SessionID             string    `json:"session_id"`
}

// RefreshToken represents stored refresh token data
type RefreshToken struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

// RefreshTokenRequest represents a request to refresh access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// GetNATRouterAccess returns the routers for NAT management access
// DEPRECATED: Use RouterService.GetAvailableRoutersWithFilter() for dynamic access control
// This function is kept for backward compatibility but uses static mappings
func GetNATRouterAccess(role Role) []string {
	// Static fallback access map - should match router configuration access control
	accessMap := map[Role][]string{
		RoleAdministrator: {
			"DARUSSALAM",
			"SAMSAT",
			"LANE1",
			"LANE2",
			"BT JAYA/PK JAYA",
			"LANE4",
			"SUKAWANGI",
		},
		RoleHeadBranch1: {
			"DARUSSALAM",
			"SAMSAT",
			"LANE1",
		},
		RoleHeadBranch2: {
			"LANE2",
			"LANE4",
		},
		RoleHeadBranch3: {
			"BT JAYA/PK JAYA",
		},
	}

	if routers, exists := accessMap[role]; exists {
		return routers
	}
	return []string{} // No access by default
}

// GetRoleForRouterAccess returns the role string for router access control
// This bridges the gap between auth roles and router access control roles
func GetRoleForRouterAccess(role Role) string {
	return string(role)
}

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleAdministrator, RoleHeadBranch1, RoleHeadBranch2, RoleHeadBranch3:
		return true
	default:
		return false
	}
}

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}

// GetRoleDisplayName returns a human-readable name for the role
func (r Role) GetRoleDisplayName() string {
	displayNames := map[Role]string{
		RoleAdministrator: "Administrator",
		RoleHeadBranch1:   "Head Branch 1 (DARUSSALAM, SAMSAT, LANE1)",
		RoleHeadBranch2:   "Head Branch 2 (LANE2, LANE4)",
		RoleHeadBranch3:   "Head Branch 3 (BT JAYA/PK JAYA)",
	}
	
	if name, exists := displayNames[r]; exists {
		return name
	}
	return string(r)
}

// HasFullAccess checks if the role has full system access
func (r Role) HasFullAccess() bool {
	return r == RoleAdministrator
}

// CanAccessNATManagement checks if the role can access NAT management
func (r Role) CanAccessNATManagement() bool {
	// All roles can access NAT management, but filtered by their router access
	return true
}
