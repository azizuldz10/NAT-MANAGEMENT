package api

import (
	"fmt"
	"net/http"

	"nat-management-app/internal/middleware"
	"nat-management-app/internal/models"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// NATHandler contains the NAT API handlers
type NATHandler struct {
	natService         *services.NATService
	userService        *services.UserService // Added for user-specific router access
	activityLogService *services.ActivityLogService
	logger             *logrus.Logger
}

// NewNATHandler creates a new NAT API handler
func NewNATHandler(natService *services.NATService, userService *services.UserService, activityLogService *services.ActivityLogService, logger *logrus.Logger) *NATHandler {
	return &NATHandler{
		natService:         natService,
		userService:        userService,
		activityLogService: activityLogService,
		logger:             logger,
	}
}

// getAllowedRoutersForUser gets allowed routers for a user
// Priority: 1. User-specific routers (user_routers table), 2. Role-based (router_access_control)
func (h *NATHandler) getAllowedRoutersForUser(c *gin.Context) []string {
	// Get user from context
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		// Fallback to role-based if user not found
		userRole, roleExists := middleware.GetUserRoleFromContext(c)
		if roleExists {
			return h.natService.GetAvailableRoutersWithFilter(models.GetRoleForRouterAccess(userRole))
		}
		return []string{}
	}

	// Try to get user-specific routers from user_routers table
	userRouterNames, err := h.userService.GetUserRouters(user.ID)
	if err != nil {
		h.logger.Warnf("Failed to get user-specific routers for user ID %d: %v", user.ID, err)
		// Fallback to role-based access
		return h.natService.GetAvailableRoutersWithFilter(models.GetRoleForRouterAccess(user.Role))
	}

	// If user has specific router assignments, use those
	if len(userRouterNames) > 0 {
		h.logger.Debugf("User ID %d has access to %d user-specific routers for NAT operations", user.ID, len(userRouterNames))
		return userRouterNames
	}

	// Fallback to role-based access if no user-specific routers
	h.logger.Debugf("User ID %d has no specific routers, falling back to role-based access for NAT", user.ID)
	return h.natService.GetAvailableRoutersWithFilter(models.GetRoleForRouterAccess(user.Role))
}

// GetNATConfigs handles GET /api/nat/configs
func (h *NATHandler) GetNATConfigs(c *gin.Context) {
	// Get user role from context (for authentication check)
	_, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Get all configs
	allConfigs := h.natService.GetAllONTConfigs()

	// Filter configs based on user-specific or role-based router access
	allowedRouters := h.getAllowedRoutersForUser(c)
	filteredConfigs := make(map[string]models.ONTConfig)
	
	for routerName, config := range allConfigs {
		// Check if user has access to this router
		hasAccess := false
		for _, allowed := range allowedRouters {
			if routerName == allowed {
				hasAccess = true
				break
			}
		}
		
		if hasAccess {
			filteredConfigs[routerName] = config
		}
	}

	response := models.NATConfigsResponse{
		Status: "success",
		Data:   filteredConfigs,
	}

	c.JSON(http.StatusOK, response)
}

// GetNATClients handles GET /api/nat/clients
func (h *NATHandler) GetNATClients(c *gin.Context) {
	// Get user role from context (for authentication check)
	_, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Get all clients
	allClients := h.natService.GetAllClients()

	// Filter clients based on user-specific or role-based router access
	allowedRouters := h.getAllowedRoutersForUser(c)
	filteredClients := make(map[string][]models.NATClient)
	
	for routerName, clients := range allClients {
		// Check if user has access to this router
		hasAccess := false
		for _, allowed := range allowedRouters {
			if routerName == allowed {
				hasAccess = true
				break
			}
		}
		
		if hasAccess {
			filteredClients[routerName] = clients
		}
	}

	response := models.NATClientsResponse{
		Status: "success",
		Data:   filteredClients,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateNATRule handles POST /api/nat/update
func (h *NATHandler) UpdateNATRule(c *gin.Context) {
	// Get user role from context (for authentication check)
	_, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	var req models.NATUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Format request tidak valid",
		})
		return
	}

	// Validate required fields
	if req.Router == "" || req.IP == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Router name dan IP address wajib diisi",
		})
		return
	}

	// Check if user has access to this router
	allowedRouters := h.getAllowedRoutersForUser(c)
	hasAccess := false
	for _, allowed := range allowedRouters {
		if req.Router == allowed {
			hasAccess = true
			break
		}
	}
	
	if !hasAccess {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Tidak memiliki akses ke router ini",
		})
		return
	}

	// Default port if not provided
	if req.Port == "" {
		req.Port = "80"
	}

	// Update NAT rule
	err := h.natService.UpdateONTNATRule(req.Router, req.IP, req.Port)
	if err != nil {
		h.logger.Errorf("Failed to update NAT rule for %s: %v", req.Router, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Log NAT update
	if h.activityLogService != nil {
		user, exists := middleware.GetUserFromContext(c)
		if exists {
			userID := user.ID
			h.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &userID,
				Username:     user.Username,
				UserRole:     string(user.Role),
				ActionType:   models.ActionNATUpdate,
				ResourceType: models.ResourceNATRule,
				ResourceID:   req.Router,
				Description:  fmt.Sprintf("Updated NAT rule for %s to %s:%s", req.Router, req.IP, req.Port),
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	response := models.NATUpdateResponse{
		Status:  "success",
		Message: fmt.Sprintf("NAT rule untuk %s berhasil diupdate ke %s:%s", req.Router, req.IP, req.Port),
	}

	c.JSON(http.StatusOK, response)
}

// TestNATConnections handles GET /api/nat/test
func (h *NATHandler) TestNATConnections(c *gin.Context) {
	// Get user role from context (for authentication check)
	_, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Get all test results
	allResults := h.natService.TestAllConnections()

	// Filter results based on user-specific or role-based router access
	allowedRouters := h.getAllowedRoutersForUser(c)
	filteredResults := make(map[string]models.RouterConnectionTest)
	
	for routerName, result := range allResults {
		// Check if user has access to this router
		hasAccess := false
		for _, allowed := range allowedRouters {
			if routerName == allowed {
				hasAccess = true
				break
			}
		}
		
		if hasAccess {
			filteredResults[routerName] = result
		}
	}

	response := models.NATTestResponse{
		Status: "success",
		Data:   filteredResults,
	}

	c.JSON(http.StatusOK, response)
}

// GetNATStatus handles GET /api/nat/status - for health check
func (h *NATHandler) GetNATStatus(c *gin.Context) {
	// Get user role from context (for authentication check)
	_, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Test NAT service health
	allConfigs := h.natService.GetAllONTConfigs()

	// Filter configs based on user-specific or role-based router access
	allowedRouters := h.getAllowedRoutersForUser(c)
	filteredConfigs := make(map[string]models.ONTConfig)
	
	for routerName, config := range allConfigs {
		// Check if user has access to this router
		hasAccess := false
		for _, allowed := range allowedRouters {
			if routerName == allowed {
				hasAccess = true
				break
			}
		}
		
		if hasAccess {
			filteredConfigs[routerName] = config
		}
	}
	
	foundCount := 0
	totalRouters := len(filteredConfigs)
	
	for _, config := range filteredConfigs {
		if config.Found {
			foundCount++
		}
	}

	healthStatus := "healthy"
	if foundCount == 0 {
		healthStatus = "unhealthy"
	} else if foundCount < totalRouters {
		healthStatus = "partial"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "success",
		"health":        healthStatus,
		"total_routers": totalRouters,
		"configured":    foundCount,
		"message":       fmt.Sprintf("%d/%d router memiliki ONT NAT rules yang terkonfigurasi", foundCount, totalRouters),
	})
}

// CheckPPPoEStatus handles POST /api/pppoe/check
func (h *NATHandler) CheckPPPoEStatus(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	var req models.PPPoEStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Format request tidak valid: " + err.Error(),
		})
		return
	}

	// Validate username
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Username PPPoE harus diisi",
		})
		return
	}

	// Get routers based on user-specific or role-based access
	allowedRouters := h.getAllowedRoutersForUser(c)

	// If specific router requested, check if user has access
	if req.Router != "" {
		hasAccess := false
		for _, allowed := range allowedRouters {
			if req.Router == allowed {
				hasAccess = true
				break
			}
		}
		
		if !hasAccess {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Status:  "error",
				Message: "Tidak memiliki akses ke router ini",
			})
			return
		}
	}

	// Check PPPoE status across accessible routers or specific router
	var result *models.PPPoEStatusResponse
	if req.Router != "" {
		// Check specific router
		result = h.natService.CheckPPPoEStatus(req.Username, req.TestConnectivity, req.Router)
	} else {
		// Check only accessible routers for this user
		result = h.natService.CheckPPPoEStatusWithRouterFilter(req.Username, allowedRouters, req.TestConnectivity)
	}

	if result.Status == "error" {
		c.JSON(http.StatusBadRequest, result)
		return
	}

	h.logger.Infof("PPPoE status checked for user: %s (router: %s, connectivity test: %t) by role: %s - Online: %t", req.Username, req.Router, req.TestConnectivity, userRole, result.IsOnline)

	// Log PPPoE check
	if h.activityLogService != nil {
		user, exists := middleware.GetUserFromContext(c)
		if exists {
			userID := user.ID
			h.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &userID,
				Username:     user.Username,
				UserRole:     string(user.Role),
				ActionType:   models.ActionPPPoECheck,
				ResourceType: models.ResourcePPPoE,
				ResourceID:   req.Username,
				Description:  fmt.Sprintf("Checked PPPoE status for: %s (Online: %t)", req.Username, result.IsOnline),
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	c.JSON(http.StatusOK, result)
}

// CheckPPPoEStatusByGET handles GET /api/pppoe/check/:username (alternative endpoint)
func (h *NATHandler) CheckPPPoEStatusByGET(c *gin.Context) {
	// Get user role from context
	_, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Username PPPoE harus diisi dalam URL",
		})
		return
	}

	// Check PPPoE status (no connectivity test for GET endpoint)
	result := h.natService.CheckPPPoEStatus(username, false)
	
	if result.Status == "error" {
		c.JSON(http.StatusBadRequest, result)
		return
	}

	h.logger.Infof("PPPoE status checked via GET for user: %s - Online: %t", username, result.IsOnline)
	c.JSON(http.StatusOK, result)
}

// GetPPPoERouters handles GET /api/pppoe/routers
func (h *NATHandler) GetPPPoERouters(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Get routers based on user-specific or role-based access from RouterService
	// This ensures PPPoE Checker is synchronized with the router database
	allowedRouters := h.getAllowedRoutersForUser(c)

	h.logger.Infof("📋 PPPoE Routers loaded for role %s: %d routers", userRole, len(allowedRouters))

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   allowedRouters,
	})
}

// FuzzySearchPPPoE handles POST /api/pppoe/fuzzy-search
func (h *NATHandler) FuzzySearchPPPoE(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	var req models.PPPoEFuzzySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Format request tidak valid: " + err.Error(),
		})
		return
	}

	// Validate username
	if req.Username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Search term harus diisi",
		})
		return
	}

	// Set default limit if not provided
	if req.Limit <= 0 {
		req.Limit = 5
	}

	// Get routers based on user-specific or role-based access from RouterService
	// This ensures fuzzy search is synchronized with the router database
	allowedRouters := h.getAllowedRoutersForUser(c)

	// If specific router requested, check if user has access
	if req.Router != "" {
		hasAccess := false
		for _, allowed := range allowedRouters {
			if req.Router == allowed {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Status:  "error",
				Message: "Tidak memiliki akses ke router ini",
			})
			return
		}
	}

	// Perform fuzzy search with dynamic router filtering
	result := h.natService.FuzzySearchPPPoEWithRouterFilter(req.Username, req.Router, req.Limit, allowedRouters)

	if result.Status == "error" {
		c.JSON(http.StatusBadRequest, result)
		return
	}

	h.logger.Infof("PPPoE fuzzy search for: %s (router: %s) by role: %s - Found: %d matches", req.Username, req.Router, userRole, result.MatchCount)
	c.JSON(http.StatusOK, result)
}