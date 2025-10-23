package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"nat-management-app/internal/database"
	"nat-management-app/internal/models"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// AuthServiceDB handles authentication for NAT Management using PostgreSQL
type AuthServiceDB struct {
	logger     *logrus.Logger
	mutex      sync.RWMutex
	userRepo   *database.UserRepository
	sessions   map[string]*models.UserSession // JWT-based, sessions kept in memory temporarily
	jwtService *JWTService
	db         *database.DB
}

// NewAuthServiceDB creates a new database-backed AuthService instance
func NewAuthServiceDB(logger *logrus.Logger, db *database.DB) *AuthServiceDB {
	// Initialize JWT service
	jwtService, err := NewJWTService(logger)
	if err != nil {
		logger.Fatalf("Failed to initialize JWT service: %v", err)
	}

	service := &AuthServiceDB{
		logger:     logger,
		userRepo:   database.NewUserRepository(db),
		sessions:   make(map[string]*models.UserSession),
		jwtService: jwtService,
		db:         db,
	}

	// Start session cleanup for backward compatibility
	go service.sessionCleanup()

	logger.Info("✅ AuthServiceDB initialized with PostgreSQL backend")
	return service
}

// Login authenticates a user and creates a session
func (as *AuthServiceDB) Login(username, password, ipAddress, userAgent string) (*models.AuthResponse, error) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find user by username from database
	user, err := as.userRepo.GetByUsername(ctx, username)
	if err != nil {
		as.logger.Warnf("🔒 Login failed: user not found: %s", username)
		return &models.AuthResponse{
			Status:  "error",
			Message: "Username atau password salah",
		}, errors.New("invalid credentials")
	}

	if !user.IsActive {
		as.logger.Warnf("🔒 Login failed: user inactive: %s", username)
		return &models.AuthResponse{
			Status:  "error",
			Message: "User tidak aktif",
		}, errors.New("user inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		as.logger.Warnf("🔒 Login failed: invalid password for: %s", username)
		return &models.AuthResponse{
			Status:  "error",
			Message: "Username atau password salah",
		}, errors.New("invalid credentials")
	}

	// Create session
	sessionID := as.generateSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	session := &models.UserSession{
		SessionID: sessionID,
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	as.sessions[sessionID] = session

	// Update last login time in database
	if err := as.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		as.logger.Warnf("Failed to update last login time: %v", err)
	}

	// Create response user (without password)
	responseUser := &models.User{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}

	as.logger.Infof("✅ Login successful: %s (%s) from %s", username, user.Role, ipAddress)

	// Get router access from user_routers table
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	// Query user_routers table directly
	rows, err := as.db.Pool.Query(ctx2, "SELECT router_name FROM user_routers WHERE user_id = $1 ORDER BY router_name ASC", user.ID)
	routerAccess := []string{}

	if err != nil {
		as.logger.Warnf("Failed to query user router access: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var routerName string
			if err := rows.Scan(&routerName); err == nil {
				routerAccess = append(routerAccess, routerName)
			}
		}
	}

	return &models.AuthResponse{
		Status:  "success",
		Message: "Login berhasil",
		Data: map[string]interface{}{
			"user":              responseUser,
			"session_id":        sessionID,
			"expires_at":        expiresAt,
			"nat_router_access": routerAccess,
		},
	}, nil
}

// LoginWithJWT authenticates user dan generate JWT token pair
func (as *AuthServiceDB) LoginWithJWT(username, password, ipAddress, userAgent string) (*models.AuthResponse, error) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find user by username from database
	user, err := as.userRepo.GetByUsername(ctx, username)
	if err != nil {
		as.logger.Warnf("🔒 JWT Login failed: user not found: %s", username)
		return &models.AuthResponse{
			Status:  "error",
			Message: "Username atau password salah",
		}, errors.New("invalid credentials")
	}

	if !user.IsActive {
		as.logger.Warnf("🔒 JWT Login failed: user inactive: %s", username)
		return &models.AuthResponse{
			Status:  "error",
			Message: "User tidak aktif",
		}, errors.New("user inactive")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		as.logger.Warnf("🔒 JWT Login failed: invalid password for: %s", username)
		return &models.AuthResponse{
			Status:  "error",
			Message: "Username atau password salah",
		}, errors.New("invalid credentials")
	}

	// Generate JWT token pair
	tokenPair, err := as.jwtService.GenerateTokenPair(user, ipAddress, userAgent)
	if err != nil {
		as.logger.Errorf("Failed to generate JWT tokens for %s: %v", username, err)
		return &models.AuthResponse{
			Status:  "error",
			Message: "Gagal generate authentication token",
		}, err
	}

	// Update last login time in database
	if err := as.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		as.logger.Warnf("Failed to update last login time: %v", err)
	}

	// Create response user (without password)
	responseUser := &models.User{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}

	as.logger.Infof("✅ JWT Login successful: %s (%s) from %s", username, user.Role, ipAddress)

	// Get router access from user_routers table
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	// Query user_routers table directly
	rows, err := as.db.Pool.Query(ctx2, "SELECT router_name FROM user_routers WHERE user_id = $1 ORDER BY router_name ASC", user.ID)
	routerAccess := []string{}

	if err != nil {
		as.logger.Warnf("Failed to query user router access: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var routerName string
			if err := rows.Scan(&routerName); err == nil {
				routerAccess = append(routerAccess, routerName)
			}
		}
	}

	return &models.AuthResponse{
		Status:  "success",
		Message: "Login berhasil",
		Data: map[string]interface{}{
			"user":              responseUser,
			"tokens":            tokenPair,
			"nat_router_access": routerAccess,
		},
	}, nil
}

// Logout removes a user session
func (as *AuthServiceDB) Logout(sessionID string) error {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	session, exists := as.sessions[sessionID]
	if !exists {
		return errors.New("session not found")
	}

	delete(as.sessions, sessionID)
	as.logger.Infof("👋 Logout: %s", session.Username)

	return nil
}

// ValidateSession checks if a session is valid and returns the user
func (as *AuthServiceDB) ValidateSession(sessionID string) (*models.User, error) {
	as.mutex.RLock()
	defer as.mutex.RUnlock()

	session, exists := as.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		go func() {
			as.mutex.Lock()
			delete(as.sessions, sessionID)
			as.mutex.Unlock()
		}()
		return nil, errors.New("session expired")
	}

	// Get user data from database
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	user, err := as.userRepo.GetByID(ctx, session.UserID)
	if err != nil || !user.IsActive {
		return nil, errors.New("user not found or inactive")
	}

	// Return user without password
	responseUser := &models.User{
		ID:          user.ID,
		Username:    user.Username,
		FullName:    user.FullName,
		Email:       user.Email,
		Role:        user.Role,
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}

	return responseUser, nil
}

// RefreshToken generates new access token dari refresh token
func (as *AuthServiceDB) RefreshToken(refreshToken, ipAddress, userAgent string) (*models.AuthResponse, error) {
	tokenPair, err := as.jwtService.RefreshAccessToken(refreshToken, ipAddress, userAgent)
	if err != nil {
		as.logger.Warnf("🔄 Token refresh failed: %v", err)
		return &models.AuthResponse{
			Status:  "error",
			Message: "Gagal refresh token",
		}, err
	}

	as.logger.Infof("🔄 Token refreshed successfully")

	return &models.AuthResponse{
		Status:  "success",
		Message: "Token berhasil direfresh",
		Data: map[string]interface{}{
			"tokens": tokenPair,
		},
	}, nil
}

// ValidateJWTToken validates JWT access token
func (as *AuthServiceDB) ValidateJWTToken(tokenString string) (*models.User, error) {
	return as.jwtService.GetUserFromToken(tokenString)
}

// LogoutJWT revokes JWT tokens
func (as *AuthServiceDB) LogoutJWT(accessToken, refreshToken string) error {
	if accessToken != "" {
		err := as.jwtService.RevokeToken(accessToken)
		if err != nil {
			as.logger.Warnf("Failed to revoke access token: %v", err)
		}
	}

	if refreshToken != "" {
		err := as.jwtService.RevokeToken(refreshToken)
		if err != nil {
			as.logger.Warnf("Failed to revoke refresh token: %v", err)
		}
	}

	as.logger.Infof("👋 JWT Logout successful")
	return nil
}

// RevokeAllUserTokens revokes semua token untuk user
func (as *AuthServiceDB) RevokeAllUserTokens(userID int) error {
	return as.jwtService.RevokeAllTokensForUser(userID)
}

// GetJWTPublicKey returns JWT public key untuk external validation
func (as *AuthServiceDB) GetJWTPublicKey() (string, error) {
	return as.jwtService.GetPublicKeyPEM()
}

// generateSessionID generates a random session ID
func (as *AuthServiceDB) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// sessionCleanup removes expired sessions periodically
func (as *AuthServiceDB) sessionCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		as.mutex.Lock()

		expiredCount := 0
		for sessionID, session := range as.sessions {
			if time.Now().After(session.ExpiresAt) {
				delete(as.sessions, sessionID)
				expiredCount++
			}
		}

		if expiredCount > 0 {
			as.logger.Infof("🧹 Cleaned up %d expired sessions", expiredCount)
		}

		as.mutex.Unlock()
	}
}
