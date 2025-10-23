package services

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"nat-management-app/internal/database"
	"nat-management-app/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// RouterServiceDB handles router management operations using PostgreSQL database
type RouterServiceDB struct {
	logger              *logrus.Logger
	routerRepo          *database.RouterRepository
	accessControlRepo   *database.AccessControlRepository
	db                  *database.DB
	connectionPool      *RouterOSConnectionPool      // Connection pool for RouterOS
	circuitBreaker      *RouterCircuitBreaker        // Circuit breaker for fault tolerance
}

// NewRouterServiceDB creates a new database-backed router service instance
func NewRouterServiceDB(logger *logrus.Logger, db *database.DB) *RouterServiceDB {
	// Initialize connection pool with config from environment
	maxConnections := 5            // Max 5 connections per router
	idleTimeout := 5 * time.Minute // Close idle connections after 5 minutes
	maxLifetime := 30 * time.Minute // Recycle connections after 30 minutes

	pool := NewRouterOSConnectionPool(logger, maxConnections, idleTimeout, maxLifetime)

	// Initialize circuit breaker with config from environment
	failureThreshold := 3          // Open circuit after 3 failures
	circuitTimeout := 30 * time.Second // Wait 30 seconds before half-open

	circuitBreaker := NewRouterCircuitBreaker(logger, failureThreshold, circuitTimeout)

	return &RouterServiceDB{
		logger:            logger,
		routerRepo:        database.NewRouterRepository(db),
		accessControlRepo: database.NewAccessControlRepository(db),
		db:                db,
		connectionPool:    pool,
		circuitBreaker:    circuitBreaker,
	}
}

// GetAllRouters returns all routers with role-based filtering
func (rs *RouterServiceDB) GetAllRouters(userRole string) ([]models.RouterResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get all routers from database
	routers, err := rs.routerRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get routers: %w", err)
	}

	// Get allowed routers for user role
	allowedRouters, err := rs.getAllowedRouters(ctx, userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed routers: %w", err)
	}

	if len(allowedRouters) == 0 {
		return []models.RouterResponse{}, nil
	}

	var responses []models.RouterResponse
	for _, router := range routers {
		// Check if user has access to this router
		if rs.hasRouterAccess(router.Name, allowedRouters) {
			responses = append(responses, router.ToResponse())
		}
	}

	return responses, nil
}

// GetRouter returns a specific router by ID
func (rs *RouterServiceDB) GetRouter(routerID string, userRole string) (*models.RouterResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find router by ID
	foundRouter, err := rs.routerRepo.GetByID(ctx, routerID)
	if err != nil {
		return nil, fmt.Errorf("router not found: %w", err)
	}

	// Check access permissions
	allowedRouters, err := rs.getAllowedRouters(ctx, userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !rs.hasRouterAccess(foundRouter.Name, allowedRouters) {
		return nil, fmt.Errorf("access denied to router")
	}

	response := foundRouter.ToResponse()
	return &response, nil
}

// CreateRouter creates a new router
func (rs *RouterServiceDB) CreateRouter(req *models.RouterCreateRequest, userRole string) (*models.RouterResponse, error) {
	// Only administrators can create routers
	if userRole != "Administrator" {
		return nil, fmt.Errorf("insufficient permissions to create router")
	}

	// Validate request
	if err := rs.validateRouterRequest(req.Name, req.Host, req.Port, req.Username, req.Password); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if router name already exists
	exists, err := rs.routerRepo.Exists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check router existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("router name '%s' already exists", req.Name)
	}

	// Generate unique ID
	routerID := fmt.Sprintf("%s-%s", strings.ToLower(strings.ReplaceAll(req.Name, " ", "-")), uuid.New().String()[:8])

	// Create router
	newRouter := req.ToRouter(routerID)

	// Save to database
	if err := rs.routerRepo.Create(ctx, &newRouter); err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	response := newRouter.ToResponse()
	rs.logger.Infof("✅ Created new router: %s (ID: %s)", newRouter.Name, newRouter.ID)
	return &response, nil
}

// UpdateRouter updates an existing router
func (rs *RouterServiceDB) UpdateRouter(routerID string, req *models.RouterUpdateRequest, userRole string) (*models.RouterResponse, error) {
	// Only administrators can update routers
	if userRole != "Administrator" {
		return nil, fmt.Errorf("insufficient permissions to update router")
	}

	// Validate request
	if err := rs.validateRouterRequest(req.Name, req.Host, req.Port, req.Username, req.Password); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get existing router
	existingRouter, err := rs.routerRepo.GetByID(ctx, routerID)
	if err != nil {
		return nil, fmt.Errorf("router not found: %w", err)
	}

	// Check if new name conflicts with another router
	if existingRouter.Name != req.Name {
		exists, err := rs.routerRepo.Exists(ctx, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check router existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("router name '%s' already exists", req.Name)
		}
	}

	// Update router fields
	existingRouter.UpdateFromRequest(req)

	// Save to database
	if err := rs.routerRepo.Update(ctx, existingRouter); err != nil {
		return nil, fmt.Errorf("failed to update router: %w", err)
	}

	response := existingRouter.ToResponse()
	rs.logger.Infof("✅ Updated router: %s (ID: %s)", req.Name, routerID)
	return &response, nil
}

// DeleteRouter deletes a router
func (rs *RouterServiceDB) DeleteRouter(routerID string, userRole string) error {
	// Only administrators can delete routers
	if userRole != "Administrator" {
		return fmt.Errorf("insufficient permissions to delete router")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get router first to log the name
	router, err := rs.routerRepo.GetByID(ctx, routerID)
	if err != nil {
		return fmt.Errorf("router not found: %w", err)
	}

	// Delete from database
	if err := rs.routerRepo.Delete(ctx, routerID); err != nil {
		return fmt.Errorf("failed to delete router: %w", err)
	}

	rs.logger.Infof("🗑️ Deleted router: %s (ID: %s)", router.Name, routerID)
	return nil
}

// TestRouter tests connection to a specific router
func (rs *RouterServiceDB) TestRouter(routerID string, userRole string) (*models.RouterTestResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find router by ID
	foundRouter, err := rs.routerRepo.GetByID(ctx, routerID)
	if err != nil {
		return nil, fmt.Errorf("router not found: %w", err)
	}

	// Check access permissions
	allowedRouters, err := rs.getAllowedRouters(ctx, userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to check access: %w", err)
	}

	if !rs.hasRouterAccess(foundRouter.Name, allowedRouters) {
		return nil, fmt.Errorf("access denied to router")
	}

	// Test connection
	testResult := rs.testRouterConnection(*foundRouter)

	response := &models.RouterTestResponse{
		Status:     "success",
		RouterID:   foundRouter.ID,
		RouterName: foundRouter.Name,
		TestResult: testResult,
		Message:    testResult.Message,
	}

	if testResult.Status != "connected" {
		response.Status = "error"
	}

	return response, nil
}

// testRouterConnection performs actual connection test with circuit breaker and connection pooling
func (rs *RouterServiceDB) testRouterConnection(router models.Router) models.RouterConnectionTest {
	rs.logger.Debugf("🔄 Testing connection to %s:%d (circuit breaker + pooling)", router.Host, router.Port)

	// Check if circuit breaker allows the call
	if !rs.circuitBreaker.IsAvailable(router.Name) {
		circuitState := rs.circuitBreaker.GetState(router.Name)
		rs.logger.Warnf("⚠️ Circuit breaker is %s for %s, skipping connection test", circuitState, router.Name)
		return models.RouterConnectionTest{
			Status:    "disconnected",
			Message:   fmt.Sprintf("Circuit breaker is %s (router marked as unhealthy)", circuitState),
			Timestamp: time.Now(),
		}
	}

	var testResult models.RouterConnectionTest

	// Wrap connection test with circuit breaker
	err := rs.circuitBreaker.Call(router.Name, func() error {
		// First check if host is reachable via TCP
		timeout := 10 * time.Second
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", router.Host, router.Port), timeout)
		if err != nil {
			rs.logger.Warnf("⚠️ TCP connection failed for %s: %v", router.Name, err)
			testResult = models.RouterConnectionTest{
				Status:    "disconnected",
				Message:   fmt.Sprintf("TCP connection failed: %v", err),
				Timestamp: time.Now(),
			}
			return err
		}
		conn.Close()

		rs.logger.Debugf("✅ TCP connection successful for %s", router.Name)

		// Get connection from pool
		config := ConnectionConfig{
			Host:     router.Host,
			Port:     router.Port,
			Username: router.Username,
			Password: router.Password,
		}

		poolConn, err := rs.connectionPool.GetConnection(router.Name, config)
		if err != nil {
			rs.logger.Warnf("⚠️ Failed to get pooled connection for %s: %v", router.Name, err)
			testResult = models.RouterConnectionTest{
				Status:    "disconnected",
				Message:   fmt.Sprintf("Connection pool error: %v", err),
				Timestamp: time.Now(),
			}
			return err
		}

		// Ensure connection is released back to pool
		defer rs.connectionPool.ReleaseConnection(poolConn)

		// Get system info using pooled connection
		identityReply, err := poolConn.Client.Run("/system/identity/print")
		if err != nil {
			// Connection might be dead, close it
			rs.connectionPool.CloseConnection(poolConn)
			testResult = models.RouterConnectionTest{
				Status:    "disconnected",
				Message:   fmt.Sprintf("Failed to get system identity: %v", err),
				Timestamp: time.Now(),
			}
			return err
		}

		resourceReply, err := poolConn.Client.Run("/system/resource/print")
		if err != nil {
			// Connection might be dead, close it
			rs.connectionPool.CloseConnection(poolConn)
			testResult = models.RouterConnectionTest{
				Status:    "disconnected",
				Message:   fmt.Sprintf("Failed to get system resource: %v", err),
				Timestamp: time.Now(),
			}
			return err
		}

		var routerName, version, board string
		if len(identityReply.Re) > 0 {
			routerName = identityReply.Re[0].Map["name"]
		}
		if len(resourceReply.Re) > 0 {
			version = resourceReply.Re[0].Map["version"]
			board = resourceReply.Re[0].Map["board-name"]
		}

		rs.logger.Infof("✅ Successfully tested %s using pooled connection (circuit: CLOSED)", router.Name)

		testResult = models.RouterConnectionTest{
			Status:     "connected",
			RouterName: routerName,
			Version:    version,
			Board:      board,
			Message:    "Connection successful (pooled + circuit breaker)",
			Timestamp:  time.Now(),
		}

		return nil
	})

	if err != nil {
		// Circuit breaker recorded the failure
		rs.logger.Warnf("⚠️ Connection test failed for %s: %v (circuit breaker updated)", router.Name, err)
	}

	return testResult
}

// GetRouterStats returns statistics about all routers
func (rs *RouterServiceDB) GetRouterStats(userRole string) (*models.RouterStatsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	allowedRouters, err := rs.getAllowedRouters(ctx, userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowed routers: %w", err)
	}

	routers, err := rs.routerRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get routers: %w", err)
	}

	connectionTests := make(map[string]models.RouterConnectionTest)

	totalRouters := 0
	activeRouters := 0
	disabledRouters := 0

	for _, router := range routers {
		if rs.hasRouterAccess(router.Name, allowedRouters) {
			totalRouters++
			if router.Enabled {
				activeRouters++
				// Test connection for enabled routers only
				testResult := rs.testRouterConnection(router)
				connectionTests[router.ID] = testResult
			} else {
				disabledRouters++
			}
		}
	}

	return &models.RouterStatsResponse{
		Status:          "success",
		TotalRouters:    totalRouters,
		ActiveRouters:   activeRouters,
		DisabledRouters: disabledRouters,
		ConnectionTests: connectionTests,
		LastUpdated:     time.Now(),
	}, nil
}

// GetRoutersForNATService returns routers in the format expected by NATService
func (rs *RouterServiceDB) GetRoutersForNATService() (map[string]models.NATRouterConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	enabledRouters, err := rs.routerRepo.GetEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled routers: %w", err)
	}

	routers := make(map[string]models.NATRouterConfig)
	for _, router := range enabledRouters {
		routers[router.Name] = router.ToNATRouterConfig()
	}

	rs.logger.Infof("📡 Loaded %d enabled routers for NAT Service", len(routers))
	return routers, nil
}

// ReloadConfiguration reloads router configuration from database
func (rs *RouterServiceDB) ReloadConfiguration() error {
	// Database is always up-to-date, just do a health check
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rs.db.HealthCheck(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	rs.logger.Info("✅ Router configuration reloaded from database")
	return nil
}

// Helper methods

// validateRouterRequest validates router request parameters
func (rs *RouterServiceDB) validateRouterRequest(name, host string, port int, username, password string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("router name is required")
	}

	if strings.TrimSpace(host) == "" {
		return fmt.Errorf("router host is required")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("router port must be between 1 and 65535")
	}

	if strings.TrimSpace(username) == "" {
		return fmt.Errorf("router username is required")
	}

	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("router password is required")
	}

	// Validate IP address or hostname
	if net.ParseIP(host) == nil {
		// Not an IP, check if it's a valid hostname
		if !rs.isValidHostname(host) {
			return fmt.Errorf("invalid host: must be a valid IP address or hostname")
		}
	}

	return nil
}

// isValidHostname validates hostname format
func (rs *RouterServiceDB) isValidHostname(hostname string) bool {
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}
	// Simple hostname validation
	for _, char := range hostname {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '.') {
			return false
		}
	}
	return true
}

// getAllowedRouters returns list of allowed routers for a user role
func (rs *RouterServiceDB) getAllowedRouters(ctx context.Context, userRole string) ([]string, error) {
	routerNames, err := rs.accessControlRepo.GetRouterNamesByRole(ctx, userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to get router names for role: %w", err)
	}

	// Handle wildcard access - get all router names from database
	if len(routerNames) == 1 && routerNames[0] == "*" {
		allRouters, err := rs.routerRepo.GetAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get all routers: %w", err)
		}

		var allRouterNames []string
		for _, router := range allRouters {
			allRouterNames = append(allRouterNames, router.Name)
		}
		return allRouterNames, nil
	}

	return routerNames, nil
}

// hasRouterAccess checks if user has access to specific router
func (rs *RouterServiceDB) hasRouterAccess(routerName string, allowedRouters []string) bool {
	for _, allowed := range allowedRouters {
		if allowed == "*" || allowed == routerName {
			return true
		}
	}
	return false
}

// GetConfigurationPath returns the configuration source (PostgreSQL database)
func (rs *RouterServiceDB) GetConfigurationPath() string {
	return "PostgreSQL Database (Neon Serverless)"
}

// GetRouterConnection returns a pooled connection for a router (for health monitoring)
func (rs *RouterServiceDB) GetRouterConnection(routerID string) (*RouterOSConnection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get router details
	router, err := rs.routerRepo.GetByID(ctx, routerID)
	if err != nil {
		return nil, fmt.Errorf("router not found: %w", err)
	}

	// Get connection from pool
	config := ConnectionConfig{
		Host:     router.Host,
		Port:     router.Port,
		Username: router.Username,
		Password: router.Password,
	}

	poolConn, err := rs.connectionPool.GetConnection(router.Name, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return poolConn, nil
}

// Close closes all resources including connection pool
func (rs *RouterServiceDB) Close() {
	if rs.connectionPool != nil {
		rs.connectionPool.Close()
	}
	rs.logger.Info("✅ RouterServiceDB closed")
}
