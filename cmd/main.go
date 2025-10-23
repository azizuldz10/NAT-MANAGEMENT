package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"nat-management-app/config"
	"nat-management-app/internal/api"
	"nat-management-app/internal/database"
	"nat-management-app/internal/middleware"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("⚠️ Warning: .env file not found, using environment variables\n")
	}

	// Load configuration
	cfg := config.Load()

	// Setup logger
	logger := setupLogger(cfg.Debug)
	logger.Info("🚀 Starting NAT Management Application with PostgreSQL...")

	// Initialize PostgreSQL database connection
	db, err := database.NewDB(logger)
	if err != nil {
		logger.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("✅ Database connection established")

	// Create services with database backend
	routerService := services.NewRouterServiceDB(logger, db)
	natService := services.NewNATService(logger, routerService)
	authService := services.NewAuthServiceDB(logger, db)
	userService := services.NewUserService(db, logger)
	activityLogService := services.NewActivityLogService(db, logger)

	// Create ONT WiFi extractor service
	ontExtractorService := services.NewONTExtractorService(logger)

	// Create ONT WiFi repository
	ontWiFiRepo := database.NewONTWiFiRepository(db)

	// Note: Health Monitor feature disabled (not needed yet)

	// Setup Gin
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Create middleware dengan security enhancements
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	secureAuthMiddleware := middleware.NewSecureAuthMiddleware(authService, logger)

	// Setup secure CORS dan security middleware
	router.Use(secureAuthMiddleware.SecureCORSWithAuth())
	router.Use(secureAuthMiddleware.SecurityLogger())
	// Global rate limiting not needed - specific rate limits applied per route group

	// Load HTML templates with error handling
	templateRoot := "web/templates"
	if _, err := os.Stat(templateRoot); os.IsNotExist(err) {
		logger.Warn("📁 web/templates directory not found, trying alternative paths...")
		if _, err := os.Stat("./templates"); err == nil {
			templateRoot = "./templates"
			logger.Info("📁 Using ./templates for templates (recursive)")
		} else {
			logger.Error("❌ No templates directory found! Application may not work correctly")
		}
	} else {
		logger.Info("📁 Using web/templates for templates (recursive)")
	}

	if err := loadTemplates(router, templateRoot); err != nil {
		logger.Fatalf("❌ Failed to load templates from %s: %v", templateRoot, err)
	}

	// Serve static files
	staticPath := "web/static"
	if _, err := os.Stat("web/static"); os.IsNotExist(err) {
		logger.Warn("📁 web/static directory not found, trying alternative paths...")
		if _, err := os.Stat("./static"); err == nil {
			staticPath = "./static"
			logger.Info("📁 Using ./static for static files")
		} else {
			logger.Error("❌ No static directory found! CSS/JS may not load correctly")
		}
	} else {
		logger.Info("📁 Using web/static for static files")
	}
	router.Static("/static", staticPath)

	// Serve SS assets (logos/screenshots), so /SS/logo.jpeg works on login page
	assetsPath := "SS"
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		logger.Warn("📁 SS directory not found; logo may not load from /SS/logo.jpeg")
	} else {
		router.Static("/SS", assetsPath)
		logger.Info("📁 Serving SS at /SS")
	}
	// Serve image assets (branding), so /image/logo.png works on login page
	imagesPath := "image"
	if _, err := os.Stat(imagesPath); os.IsNotExist(err) {
		logger.Warn("📁 image directory not found; logo may not load from /image/logo.png")
	} else {
		router.Static("/image", imagesPath)
		logger.Info("📁 Serving image at /image")
	}
	
	// Create API handlers
	natHandler := api.NewNATHandler(natService, userService, activityLogService, logger)
	authHandler := api.NewAuthHandler(authService, routerService, userService, activityLogService, logger)
	routerHandler := api.NewRouterHandler(routerService, natService, activityLogService, logger)
	userHandler := api.NewUserHandler(userService, activityLogService, logger)
	activityLogHandler := api.NewActivityLogHandler(activityLogService, logger)
	ontWiFiHandler := api.NewONTWiFiHandler(ontExtractorService, ontWiFiRepo, natService, activityLogService, logger)
	// monitoringHandler removed - feature disabled

	// Public routes (no authentication required)
	router.GET("/login", loginHandler)

	// Health check endpoints (for monitoring and load balancers)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "NAT Management Application",
			"version": "v4.2",
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		// Check database connection
		ctx := c.Request.Context()
		if err := db.Pool.Ping(ctx); err != nil {
			logger.Warnf("❌ Database ping failed in readiness check: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"reason": "database connection failed",
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"database": "connected",
			"service": "NAT Management Application",
		})
	})

	// Authentication API routes with environment-aware rate limiting
	authGroup := router.Group("/api/auth")
	authGroup.Use(secureAuthMiddleware.LoginRateLimit()) // Environment-based rate limiting (lenient in dev, strict in production)
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/refresh", authHandler.RefreshToken)
		authGroup.GET("/check", authHandler.CheckAuth)
		authGroup.GET("/jwt-public-key", authHandler.GetJWTPublicKey)
	}

	// Protected routes (authentication required) - Support both session dan JWT
	protected := router.Group("/")
	protected.Use(authMiddleware.RequireAuth()) // Keep session-based untuk backward compatibility
	{
		protected.GET("/", natManagementHandler) // Main NAT Management page
		protected.GET("/nat", natManagementHandler) // Alternative route
		protected.GET("/pppoe", pppoeCheckerHandler) // PPPoE Status Checker page
		// Monitoring page removed - feature disabled
		// ONT WiFi Extractor removed - functionality integrated into router cards
		protected.GET("/routers", routerManagementHandler) // Router Management page (Admin only)
		protected.GET("/users", userManagementHandler) // User Management page (Admin only)
		protected.GET("/logs", activityLogsHandler) // Activity Logs page (Admin only)
	}
	
	// Protected API routes (JWT authentication required)
	apiGroup := router.Group("/api")
	apiGroup.Use(secureAuthMiddleware.RequireJWTAuth()) // Use JWT for API endpoints
	{
		// User info
		apiGroup.GET("/auth/me", authHandler.Me)

		// Router Management API routes (Administrator only)
		routerGroup := apiGroup.Group("/routers")
		{
			routerGroup.GET("", routerHandler.GetRouters)
			routerGroup.POST("", routerHandler.CreateRouter)
			routerGroup.GET("/:id", routerHandler.GetRouter)
			routerGroup.PUT("/:id", routerHandler.UpdateRouter)
			routerGroup.DELETE("/:id", routerHandler.DeleteRouter)
			routerGroup.POST("/:id/test", routerHandler.TestRouter)
			routerGroup.GET("/stats", routerHandler.GetRouterStats)
			routerGroup.POST("/validate", routerHandler.ValidateRouter)
			routerGroup.POST("/reload", routerHandler.ReloadConfiguration)
			routerGroup.GET("/config", routerHandler.GetConfigurationInfo)
		}

		// User Management API routes (Administrator only)
		userGroup := apiGroup.Group("/users")
		{
			userGroup.GET("", userHandler.ListUsers)
			userGroup.POST("", userHandler.CreateUser)
			userGroup.GET("/:id", userHandler.GetUser)
			userGroup.PUT("/:id", userHandler.UpdateUser)
			userGroup.DELETE("/:id", userHandler.DeleteUser)
			userGroup.GET("/:id/routers", userHandler.GetUserRouters)
			userGroup.GET("/:id/stats", userHandler.GetUserStats)
			userGroup.PATCH("/:id/activate", userHandler.ActivateUser)
			userGroup.PATCH("/:id/password", userHandler.ChangeUserPassword)
		}

		// NAT Management API routes
		natGroup := apiGroup.Group("/nat")
		{
			natGroup.GET("/configs", natHandler.GetNATConfigs)
			natGroup.GET("/clients", natHandler.GetNATClients)
			natGroup.POST("/update", natHandler.UpdateNATRule)
			natGroup.GET("/test", natHandler.TestNATConnections)
			natGroup.GET("/status", natHandler.GetNATStatus)
		}

		// PPPoE Status Checking API routes
		pppoeGroup := apiGroup.Group("/pppoe")
		{
			pppoeGroup.POST("/check", natHandler.CheckPPPoEStatus)
			pppoeGroup.GET("/check/:username", natHandler.CheckPPPoEStatusByGET)
			pppoeGroup.GET("/routers", natHandler.GetPPPoERouters)
			pppoeGroup.POST("/fuzzy-search", natHandler.FuzzySearchPPPoE)
		}

		// Activity Logs API routes (Administrator only)
		logsGroup := apiGroup.Group("/logs")
		{
			logsGroup.GET("", activityLogHandler.GetLogs)
			logsGroup.GET("/:id", activityLogHandler.GetLogByID)
			logsGroup.GET("/stats", activityLogHandler.GetLogStats)
			logsGroup.POST("/cleanup", activityLogHandler.DeleteOldLogs)
		}

		// ONT WiFi Extraction API routes
		ontWiFiGroup := apiGroup.Group("/ont/wifi")
		{
			ontWiFiGroup.POST("/extract", ontWiFiHandler.ExtractWiFiInfo)
			ontWiFiGroup.POST("/extract-from-nat", ontWiFiHandler.ExtractWiFiFromNAT)
			ontWiFiGroup.GET("/history", ontWiFiHandler.GetWiFiHistory)
			ontWiFiGroup.GET("/latest/:pppoe_username", ontWiFiHandler.GetLatestWiFiInfo)
			ontWiFiGroup.GET("/search", ontWiFiHandler.SearchBySSID)
			ontWiFiGroup.GET("/stats", ontWiFiHandler.GetWiFiStats)
			ontWiFiGroup.GET("/availability", ontWiFiHandler.CheckAvailability)
		}

		// Monitoring API routes removed - feature disabled
	}

	// Print startup information
	printStartupInfo(cfg)

	// Create HTTP server
	address := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	srv := &http.Server{
		Addr:    address,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("🌐 NAT Management Server starting on http://%s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("❌ Failed to start server: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-quit
	logger.Info("📴 Shutting down NAT Management server...")

	// Create shutdown context with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("⚠️ Server forced to shutdown: %v", err)
	} else {
		logger.Info("✅ Server gracefully stopped")
	}

	logger.Info("🔒 Closing RouterOS connection pool...")
	routerService.Close()

	logger.Info("🔒 Closing database connections...")
	db.Close()

	logger.Info("👋 NAT Management Application shutdown complete")
}

// natManagementHandler serves the main NAT management page
func natManagementHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "nat_management.html", gin.H{
		"Title": "NAT Management Dashboard",
	})
}

// loginHandler serves the login page
func loginHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title": "Login - NAT Management",
	})
}

// pppoeCheckerHandler serves the PPPoE status checker page
func pppoeCheckerHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "pppoe_checker.html", gin.H{
		"Title": "PPPoE Status Checker",
	})
}

// routerManagementHandler serves the router management page (Administrator only)
func routerManagementHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "router_management.html", gin.H{
		"Title": "Router Management",
	})
}

// userManagementHandler serves the user management page (Administrator only)
func userManagementHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "user_management.html", gin.H{
		"Title": "User Management",
	})
}

// activityLogsHandler serves the activity logs page (Administrator only)
func activityLogsHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "activity_logs.html", gin.H{
		"Title": "Activity Logs",
	})
}

// setupLogger configures the logger
func setupLogger(debug bool) *logrus.Logger {
	logger := logrus.New()
	
	if debug {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	// Use custom formatter
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	return logger
}

// loadTemplates safely loads HTML templates (recursive) including partials/layouts
func loadTemplates(router *gin.Engine, templateRoot string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Panic while loading templates: %v\n", r)
		}
	}()

	// Collect all .html files under templateRoot (recursive)
	var files []string
	err := filepath.Walk(templateRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".html") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no template files found under %s", templateRoot)
	}

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return err
	}
	router.SetHTMLTemplate(tmpl)
	return nil
}

// printStartupInfo prints application startup information
func printStartupInfo(cfg *config.Config) {
	separator := strings.Repeat("=", 80)
	fmt.Println(separator)
	fmt.Println("🚀 NAT Management Dashboard - PostgreSQL Edition")
	fmt.Println(separator)
	fmt.Println("Features:")
	fmt.Println("- 🔐 Role-based Authentication (Admin, Head Branch 1-3)")
	fmt.Println("- 🗄️ PostgreSQL Database (Neon Serverless)")
	fmt.Println("- 🌐 NAT Configuration Management")
	fmt.Println("- 👥 Online Clients Monitoring")
	fmt.Println("- 🔧 Router Connection Testing")
	fmt.Println("- 🔍 PPPoE Status Checker")
	fmt.Println("- 📡 ONT WiFi Extraction (Auto-detect 4+ models)")
	fmt.Println("- 📱 Full Responsive UI/UX")
	fmt.Println("- 🎯 Comprehensive Network Management")
	fmt.Println("- 🔄 Real-time Database Sync")
	fmt.Println("- ❤️ Health Check Endpoints (/health, /ready)")
	fmt.Println("- 🏊 Connection Pooling (5 per router, auto cleanup)")
	fmt.Println("- 🔌 Circuit Breaker (fault tolerance, auto recovery)")
	fmt.Println(separator)
	fmt.Println("Default Users (from Database):")
	fmt.Println("- admin/admin123 (Administrator - All Routers)")
	fmt.Println("- head1/head123 (Branch 1 - SAMSAT, LANE1)")
	fmt.Println("- head2/head123 (Branch 2 - LANE2, LANE4)")
	fmt.Println("- head3/head123 (Branch 3 - BT JAYA/PK JAYA, SUKAWANGI)")
	fmt.Println(separator)
	fmt.Printf("🌐 Web Interface: http://localhost:%s\n", cfg.ServerPort)
	fmt.Println("⚡ Press Ctrl+C to stop")
	fmt.Println(separator)
}
