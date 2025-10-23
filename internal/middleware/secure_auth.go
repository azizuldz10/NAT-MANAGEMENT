package middleware

import (
	"errors"
	"fmt"
	"nat-management-app/config"
	"nat-management-app/internal/utils"
	"net/http"
	"os"
	"strings"
	"sync"

	"nat-management-app/internal/models"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// SecureAuthMiddleware handles secure JWT authentication dengan enhanced security
type SecureAuthMiddleware struct {
	authService     services.AuthServiceInterface
	logger          *logrus.Logger
	rateLimiter     *rate.Limiter
	rateLimitConfig *config.RateLimitConfig
}

// NewSecureAuthMiddleware creates enhanced JWT-based auth middleware
func NewSecureAuthMiddleware(authService services.AuthServiceInterface, logger *logrus.Logger) *SecureAuthMiddleware {
	rateLimitConfig := config.LoadRateLimitConfig()

	return &SecureAuthMiddleware{
		authService: authService,
		logger:      logger,
		// Use per-second rate derived from requests-per-minute; burst capped to the same minute allowance
		rateLimiter:     rate.NewLimiter(rate.Limit(float64(rateLimitConfig.RequestsPerMinute)/60.0), rateLimitConfig.RequestsPerMinute),
		rateLimitConfig: rateLimitConfig,
	}
}

// RequireJWTAuth middleware untuk JWT authentication dengan security enhancements
func (sam *SecureAuthMiddleware) RequireJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rate limiting handled via RateLimitByIP middleware per route group

		user, err := sam.getCurrentUserFromJWT(c)
		if err != nil {
			sam.handleUnauthorized(c, "JWT Authentication required: "+err.Error())
			return
		}

		// Set user context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)
		c.Set("username", user.Username)

		// Security headers
		sam.setSecurityHeaders(c)

		c.Next()
	}
}

// getCurrentUserFromJWT extracts and validates JWT token
func (sam *SecureAuthMiddleware) getCurrentUserFromJWT(c *gin.Context) (*models.User, error) {
	// Get token from Authorization header (preferred method)
	var tokenString string
	authHeader := c.GetHeader("Authorization")

	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Fallback: try to get from cookie (untuk backward compatibility)
		if cookie, err := c.Cookie("access_token"); err == nil {
			tokenString = cookie
		}
	}

	if tokenString == "" {
		return nil, errors.New("missing authorization token")
	}

	// Validate JWT token
	user, err := sam.authService.ValidateJWTToken(tokenString)
	if err != nil {
		sam.logger.Warnf("🔒 JWT validation failed from IP %s: %v", c.ClientIP(), err)
		return nil, err
	}

	return user, nil
}

// setSecurityHeaders sets important security headers
func (sam *SecureAuthMiddleware) setSecurityHeaders(c *gin.Context) {
	// Security headers untuk melindungi aplikasi
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' https://cdnjs.cloudflare.com; img-src 'self' data:; connect-src 'self'")

	// Remove server information
	c.Header("Server", "")
}

// handleUnauthorized handles unauthorized access dengan improved security
func (sam *SecureAuthMiddleware) handleUnauthorized(c *gin.Context, message string) {
	sam.logger.Warnf("🔒 Unauthorized access attempt from IP: %s, User-Agent: %s",
		c.ClientIP(), c.GetHeader("User-Agent"))

	// Clear any invalid tokens
	c.SetCookie("access_token", "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	if sam.isAPIRequest(c) {
		c.JSON(http.StatusUnauthorized, models.AuthResponse{
			Status:  "error",
			Message: "Authentication required",
		})
	} else {
		// Redirect to login dengan redirect parameter
		c.Redirect(http.StatusFound, "/login?redirect="+c.Request.URL.Path)
	}
	c.Abort()
}

// isAPIRequest checks if request is API call
func (sam *SecureAuthMiddleware) isAPIRequest(c *gin.Context) bool {
	return strings.HasPrefix(c.Request.URL.Path, "/api/") ||
		strings.Contains(c.GetHeader("Accept"), "application/json") ||
		strings.Contains(c.GetHeader("Content-Type"), "application/json")
}

// SecureCORSWithAuth implements secure CORS policy
func (sam *SecureAuthMiddleware) SecureCORSWithAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Build allowed origins from ALLOWED_ORIGINS env (comma-separated), plus defaults and current host
		allowedOrigins := map[string]bool{}
		envOrigins := os.Getenv("ALLOWED_ORIGINS")
		if envOrigins != "" {
			for _, o := range strings.Split(envOrigins, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					allowedOrigins[o] = true
				}
			}
		} else {
			// Sensible defaults
			allowedOrigins["http://localhost:8080"] = true
			allowedOrigins["http://127.0.0.1:8080"] = true
		}
		// Allow current host dynamically (http/https)
		if h := c.Request.Host; h != "" {
			allowedOrigins["http://"+h] = true
			allowedOrigins["https://"+h] = true
		}

		// Deteksi origin lokal (private network) dan host yang cocok dengan server saat ini
		isLocalOrMatchingHost := func(o string) bool {
			if o == "" {
				return false
			}

			// Normalisasi scheme
			var hostPort string
			if strings.HasPrefix(o, "http://") {
				hostPort = strings.TrimPrefix(o, "http://")
			} else if strings.HasPrefix(o, "https://") {
				hostPort = strings.TrimPrefix(o, "https://")
			} else {
				hostPort = o
			}

			// Ambil host tanpa path
			if idx := strings.Index(hostPort, "/"); idx >= 0 {
				hostPort = hostPort[:idx]
			}
			// Pisahkan host dan port
			host := hostPort
			if hIdx := strings.Index(hostPort, ":"); hIdx >= 0 {
				host = hostPort[:hIdx]
			}

			// Cek kecocokan dengan host server (Host header berisi host:port)
			serverHost := c.Request.Host // contoh: "192.168.1.33:3000"
			if serverHost != "" {
				if strings.HasPrefix(o, "http://"+serverHost) || strings.HasPrefix(o, "https://"+serverHost) {
					return true
				}
				// Juga cek hanya host tanpa port
				srvHost := serverHost
				if pIdx := strings.Index(serverHost, ":"); pIdx >= 0 {
					srvHost = serverHost[:pIdx]
				}
				if strings.HasPrefix(o, "http://"+srvHost) || strings.HasPrefix(o, "https://"+srvHost) {
					return true
				}
			}

			// Cek private network: 127.0.0.1, localhost, 10.x.x.x, 192.168.x.x, 172.16-31.x.x
			if host == "localhost" || host == "127.0.0.1" {
				return true
			}
			if strings.HasPrefix(host, "10.") {
				return true
			}
			if strings.HasPrefix(host, "192.168.") {
				return true
			}
			if strings.HasPrefix(host, "172.") {
				parts := strings.Split(host, ".")
				if len(parts) >= 2 {
					// Validasi second octet berada di range 16-31
					switch parts[1] {
					case "16", "17", "18", "19", "20", "21", "22", "23", "24", "25", "26", "27", "28", "29", "30", "31":
						return true
					}
				}
			}
			return false
		}

		// Keputusan CORS
		if origin == "" {
			// Same-origin requests (browser tidak mengirim Origin)
			// Tidak perlu set Allow-Origin; lanjutkan tanpa abort
		} else if allowedOrigins[origin] || isLocalOrMatchingHost(origin) {
			// Izinkan origin yang cocok
			c.Header("Access-Control-Allow-Origin", origin) // Specific origin
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			// Origin tidak diizinkan
			sam.logger.Warnf("🚫 CORS blocked: unauthorized origin: %s from IP: %s", origin, c.ClientIP())
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Set CORS headers (security-first)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Accept, Content-Type, Authorization, X-Requested-With, X-CSRF-Token")
		c.Header("Access-Control-Expose-Headers", "X-Total-Count, X-Rate-Limit-Remaining")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours cache
		c.Header("Vary", "Origin")                  // ensure caches respect per-origin responses

		// Preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Additional security headers
		sam.setSecurityHeaders(c)

		c.Next()
	}
}

// RateLimitByIP implements IP-based rate limiting
func (sam *SecureAuthMiddleware) RateLimitByIP() gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)
	var mu sync.Mutex

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Check if IP is whitelisted (skip rate limiting for admin IPs)
		if sam.rateLimitConfig.IsWhitelisted(ip) {
			sam.logger.Debugf("✅ IP %s is whitelisted, bypassing rate limit", ip)
			c.Next()
			return
		}

		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			// Derive per-second rate from configured requests-per-minute, burst = minute allowance
			rpm := sam.rateLimitConfig.RequestsPerMinute
			perSecond := rate.Limit(float64(rpm) / 60.0)
			if perSecond <= 0 {
				perSecond = 1 // minimum 1 req/sec
			}
			if rpm <= 0 {
				rpm = 1
			}
			limiter = rate.NewLimiter(perSecond, rpm)
			limiters[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			sam.logger.Warnf("⚠️ Rate limit exceeded for IP: %s", ip)
			utils.RespondRateLimitExceeded(c, 60)
			c.Abort()
			return
		}

		c.Next()
	}
}

// LoginRateLimit implements strict rate limiting for login attempts
func (sam *SecureAuthMiddleware) LoginRateLimit() gin.HandlerFunc {
	limiters := make(map[string]*rate.Limiter)
	var mu sync.Mutex

	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Check if IP is whitelisted (skip rate limiting for admin IPs)
		if sam.rateLimitConfig.IsWhitelisted(ip) {
			sam.logger.Debugf("✅ IP %s is whitelisted, bypassing login rate limit", ip)
			c.Next()
			return
		}

		mu.Lock()
		limiter, exists := limiters[ip]
		if !exists {
			// Use configured login rate limit (attempts per minute → per-second rate with minute burst)
			limiter = rate.NewLimiter(rate.Limit(float64(sam.rateLimitConfig.LoginAttemptsPerMinute)/60.0), sam.rateLimitConfig.LoginAttemptsPerMinute)
			limiters[ip] = limiter
		}
		mu.Unlock()

		if !limiter.Allow() {
			sam.logger.Warnf("🚨 Login rate limit exceeded for IP: %s (Environment: %s, Limit: %d/min)",
				ip, sam.rateLimitConfig.Environment, sam.rateLimitConfig.LoginAttemptsPerMinute)
			utils.RespondRateLimitExceeded(c, 60)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityLogger logs security-related events
func (sam *SecureAuthMiddleware) SecurityLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Log format dengan security information
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s\" %s \"%s\" \"%s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.Request.Referer(),
			param.ErrorMessage,
		)
	})
}

// GetUserFromContext safely extracts user from context
func GetUserFromSecureContext(c *gin.Context) (*models.User, bool) {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u, true
		}
	}
	return nil, false
}

// GetUserRoleFromSecureContext safely extracts user role from context
func GetUserRoleFromSecureContext(c *gin.Context) (models.Role, bool) {
	if role, exists := c.Get("user_role"); exists {
		if r, ok := role.(models.Role); ok {
			return r, true
		}
	}
	return "", false
}
