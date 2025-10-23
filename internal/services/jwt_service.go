package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"sync"
	"time"

	"nat-management-app/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// JWTService menangani JWT token generation dan validation dengan keamanan tinggi
type JWTService struct {
	logger           *logrus.Logger
	privateKey       *rsa.PrivateKey
	publicKey        *rsa.PublicKey
	refreshTokens    map[string]*models.RefreshToken
	mutex            sync.RWMutex
	rateLimiter      *rate.Limiter
	blacklistedTokens map[string]time.Time
	blacklistMutex   sync.RWMutex
}

// JWTClaims represents custom JWT claims dengan security enhancements
type JWTClaims struct {
	UserID    int         `json:"user_id"`
	Username  string      `json:"username"`
	Role      models.Role `json:"role"`
	TokenType string      `json:"token_type"` // "access" atau "refresh"
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent"`
	SessionID string      `json:"session_id"`
	jwt.RegisteredClaims
}

// RefreshToken represents refresh token data
type RefreshToken struct {
	Token     string    `json:"token"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	SessionID string    `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

// NewJWTService creates a new JWT service dengan keamanan tinggi
func NewJWTService(logger *logrus.Logger) (*JWTService, error) {
	// Generate RSA key pair untuk signing
	privateKey, err := generateRSAKeyPair()
	if err != nil {
		return nil, fmt.Errorf("gagal generate RSA key pair: %v", err)
	}

	service := &JWTService{
		logger:            logger,
		privateKey:        privateKey,
		publicKey:         &privateKey.PublicKey,
		refreshTokens:     make(map[string]*models.RefreshToken),
		rateLimiter:       rate.NewLimiter(rate.Every(time.Minute), 10), // 10 login per menit
		blacklistedTokens: make(map[string]time.Time),
	}

	// Start cleanup goroutines
	go service.cleanupExpiredTokens()
	go service.cleanupBlacklist()

	logger.Info("🔐 JWT Service initialized dengan RSA256 signing")
	return service, nil
}

// generateRSAKeyPair generates a secure RSA key pair
func generateRSAKeyPair() (*rsa.PrivateKey, error) {
	// Generate 2048-bit RSA key untuk keamanan yang baik
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// GenerateTokenPair creates access dan refresh token pair dengan keamanan tinggi
func (js *JWTService) GenerateTokenPair(user *models.User, ipAddress, userAgent string) (*models.TokenPair, error) {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	// Check rate limiting
	if !js.rateLimiter.Allow() {
		return nil, errors.New("terlalu banyak permintaan login, coba lagi nanti")
	}

	sessionID := js.generateSecureSessionID()
	now := time.Now()

	// Generate Access Token (short-lived: 15 menit)
	accessClaims := &JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: "access",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "nat-management-app",
			Subject:   fmt.Sprintf("user:%d", user.ID),
			ID:        js.generateSecureJTI(),
			Audience:  []string{"nat-management"},
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(js.privateKey)
	if err != nil {
		return nil, fmt.Errorf("gagal generate access token: %v", err)
	}

	// Generate Refresh Token (long-lived: 7 hari)
	refreshClaims := &JWTClaims{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		TokenType: "refresh",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "nat-management-app",
			Subject:   fmt.Sprintf("user:%d", user.ID),
			ID:        js.generateSecureJTI(),
			Audience:  []string{"nat-management"},
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(js.privateKey)
	if err != nil {
		return nil, fmt.Errorf("gagal generate refresh token: %v", err)
	}

	// Store refresh token
	refreshTokenData := &models.RefreshToken{
		Token:     refreshTokenString,
		UserID:    user.ID,
		SessionID: sessionID,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: now,
		LastUsed:  now,
	}
	js.refreshTokens[refreshTokenString] = refreshTokenData

	js.logger.Infof("🔐 JWT token pair generated untuk user: %s (session: %s)", user.Username, sessionID)

	return &models.TokenPair{
		AccessToken:           accessTokenString,
		RefreshToken:          refreshTokenString,
		AccessTokenExpiresAt:  accessClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.ExpiresAt.Time,
		TokenType:             "Bearer",
		SessionID:             sessionID,
	}, nil
}

// ValidateAccessToken validates access token dan return claims
func (js *JWTService) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	// Check if token is blacklisted
	if js.isTokenBlacklisted(tokenString) {
		return nil, errors.New("token sudah di-blacklist")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return js.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token tidak valid: %v", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("token claims tidak valid")
	}

	// Validate token type
	if claims.TokenType != "access" {
		return nil, errors.New("bukan access token")
	}

	// Additional security checks
	if claims.Issuer != "nat-management-app" {
		return nil, errors.New("token issuer tidak valid")
	}

	return claims, nil
}

// RefreshAccessToken generates new access token dari refresh token
func (js *JWTService) RefreshAccessToken(refreshTokenString, ipAddress, userAgent string) (*models.TokenPair, error) {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	// Check rate limiting
	if !js.rateLimiter.Allow() {
		return nil, errors.New("terlalu banyak permintaan refresh, coba lagi nanti")
	}

	// Validate refresh token
	refreshToken, err := jwt.ParseWithClaims(refreshTokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return js.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("refresh token tidak valid: %v", err)
	}

	refreshClaims, ok := refreshToken.Claims.(*JWTClaims)
	if !ok || !refreshToken.Valid {
		return nil, errors.New("refresh token claims tidak valid")
	}

	if refreshClaims.TokenType != "refresh" {
		return nil, errors.New("bukan refresh token")
	}

	// Check if refresh token exists in storage
	storedToken, exists := js.refreshTokens[refreshTokenString]
	if !exists {
		return nil, errors.New("refresh token tidak ditemukan")
	}

	// Security check: IP dan User Agent validation (optional, bisa direlaksasi untuk mobile)
	if storedToken.IPAddress != ipAddress {
		js.logger.Warnf("⚠️ IP address mismatch untuk refresh token dari user: %s", refreshClaims.Username)
		// Tidak langsung tolak, tapi log untuk monitoring
	}

	// Update last used
	storedToken.LastUsed = time.Now()

	// Create new access token dengan session yang sama
	now := time.Now()
	newAccessClaims := &JWTClaims{
		UserID:    refreshClaims.UserID,
		Username:  refreshClaims.Username,
		Role:      refreshClaims.Role,
		TokenType: "access",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		SessionID: refreshClaims.SessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "nat-management-app",
			Subject:   fmt.Sprintf("user:%d", refreshClaims.UserID),
			ID:        js.generateSecureJTI(),
			Audience:  []string{"nat-management"},
		},
	}

	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, newAccessClaims)
	newAccessTokenString, err := newAccessToken.SignedString(js.privateKey)
	if err != nil {
		return nil, fmt.Errorf("gagal generate new access token: %v", err)
	}

	js.logger.Infof("🔄 Access token refreshed untuk user: %s", refreshClaims.Username)

	return &models.TokenPair{
		AccessToken:           newAccessTokenString,
		RefreshToken:          refreshTokenString, // Keep same refresh token
		AccessTokenExpiresAt:  newAccessClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: storedToken.ExpiresAt,
		TokenType:             "Bearer",
		SessionID:             refreshClaims.SessionID,
	}, nil
}

// RevokeToken menambahkan token ke blacklist
func (js *JWTService) RevokeToken(tokenString string) error {
	js.blacklistMutex.Lock()
	defer js.blacklistMutex.Unlock()

	// Parse token untuk get expiration
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return js.publicKey, nil
	})

	if err != nil {
		// Even if parsing fails, add to blacklist
		js.blacklistedTokens[tokenString] = time.Now().Add(24 * time.Hour)
		return nil
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		js.blacklistedTokens[tokenString] = claims.ExpiresAt.Time
	} else {
		js.blacklistedTokens[tokenString] = time.Now().Add(24 * time.Hour)
	}

	js.logger.Infof("🚫 Token revoked dan ditambahkan ke blacklist")
	return nil
}

// RevokeAllTokensForUser revoke semua token untuk user tertentu
func (js *JWTService) RevokeAllTokensForUser(userID int) error {
	js.mutex.Lock()
	defer js.mutex.Unlock()

	count := 0
	for tokenString, refreshToken := range js.refreshTokens {
		if refreshToken.UserID == userID {
			delete(js.refreshTokens, tokenString)
			js.RevokeToken(tokenString)
			count++
		}
	}

	js.logger.Infof("🚫 Revoked %d refresh tokens untuk user ID: %d", count, userID)
	return nil
}

// GetUserFromToken extracts user info from valid token
func (js *JWTService) GetUserFromToken(tokenString string) (*models.User, error) {
	claims, err := js.ValidateAccessToken(tokenString)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Role:     claims.Role,
		IsActive: true, // Assuming active if token is valid
	}

	return user, nil
}

// Helper methods

func (js *JWTService) generateSecureSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func (js *JWTService) generateSecureJTI() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}

func (js *JWTService) isTokenBlacklisted(tokenString string) bool {
	js.blacklistMutex.RLock()
	defer js.blacklistMutex.RUnlock()

	_, exists := js.blacklistedTokens[tokenString]
	return exists
}

// Cleanup goroutines

func (js *JWTService) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		js.mutex.Lock()
		
		expiredCount := 0
		now := time.Now()
		
		for tokenString, refreshToken := range js.refreshTokens {
			if now.After(refreshToken.ExpiresAt) {
				delete(js.refreshTokens, tokenString)
				expiredCount++
			}
		}
		
		if expiredCount > 0 {
			js.logger.Infof("🧹 Cleaned up %d expired refresh tokens", expiredCount)
		}
		
		js.mutex.Unlock()
	}
}

func (js *JWTService) cleanupBlacklist() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		js.blacklistMutex.Lock()
		
		expiredCount := 0
		now := time.Now()
		
		for tokenString, expiresAt := range js.blacklistedTokens {
			if now.After(expiresAt) {
				delete(js.blacklistedTokens, tokenString)
				expiredCount++
			}
		}
		
		if expiredCount > 0 {
			js.logger.Infof("🧹 Cleaned up %d expired blacklisted tokens", expiredCount)
		}
		
		js.blacklistMutex.Unlock()
	}
}

// GetPublicKeyPEM returns public key dalam format PEM untuk eksternal validation
func (js *JWTService) GetPublicKeyPEM() (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(js.publicKey)
	if err != nil {
		return "", err
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM), nil
}
