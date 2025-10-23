package api

import (
	"net/http"
	"strconv"

	"nat-management-app/internal/middleware"
	"nat-management-app/internal/models"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RouterHandler contains the router management API handlers
type RouterHandler struct {
	routerService      services.RouterServiceInterface
	natService         *services.NATService
	activityLogService *services.ActivityLogService
	logger             *logrus.Logger
}

// NewRouterHandler creates a new router API handler
func NewRouterHandler(routerService services.RouterServiceInterface, natService *services.NATService, activityLogService *services.ActivityLogService, logger *logrus.Logger) *RouterHandler {
	return &RouterHandler{
		routerService:      routerService,
		natService:         natService,
		activityLogService: activityLogService,
		logger:             logger,
	}
}

// GetRouters handles GET /api/routers - List all routers
func (h *RouterHandler) GetRouters(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Get routers with role-based filtering
	routers, err := h.routerService.GetAllRouters(string(userRole))
	if err != nil {
		h.logger.Errorf("Failed to get routers for role %s: %v", userRole, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve routers",
		})
		return
	}

	response := models.RouterListResponse{
		Status:  "success",
		Data:    routers,
		Total:   len(routers),
		Message: "Routers retrieved successfully",
	}

	h.logger.Infof("Retrieved %d routers for user role: %s", len(routers), userRole)
	c.JSON(http.StatusOK, response)
}

// GetRouter handles GET /api/routers/:id - Get specific router
func (h *RouterHandler) GetRouter(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	routerID := c.Param("id")
	if routerID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Router ID is required",
		})
		return
	}

	// Get router with access control
	router, err := h.routerService.GetRouter(routerID, string(userRole))
	if err != nil {
		h.logger.Errorf("Failed to get router %s for role %s: %v", routerID, userRole, err)

		if err.Error() == "router not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Status:  "error",
				Message: "Router not found",
			})
		} else if err.Error() == "access denied to router" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Status:  "error",
				Message: "Access denied to this router",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Status:  "error",
				Message: "Failed to retrieve router",
			})
		}
		return
	}

	response := models.RouterDetailResponse{
		Status:  "success",
		Data:    *router,
		Message: "Router retrieved successfully",
	}

	h.logger.Infof("Retrieved router %s for user role: %s", routerID, userRole)
	c.JSON(http.StatusOK, response)
}

// CreateRouter handles POST /api/routers - Create new router
func (h *RouterHandler) CreateRouter(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Only administrators can create routers
	if userRole != "Administrator" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Insufficient permissions to create router",
		})
		return
	}

	var req models.RouterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid router create request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Create router
	router, err := h.routerService.CreateRouter(&req, string(userRole))
	if err != nil {
		h.logger.Errorf("Failed to create router %s: %v", req.Name, err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	// Reload NAT service routers after creation
	if h.natService != nil {
		if reloadErr := h.natService.ReloadRouters(); reloadErr != nil {
			h.logger.Warnf("Failed to reload NAT service after router creation: %v", reloadErr)
			// Don't fail the request, just log the warning
		} else {
			h.logger.Infof("✅ NAT service reloaded successfully after router creation")
		}
	}

	response := models.RouterCreateResponse{
		Status:  "success",
		Data:    *router,
		Message: "Router created successfully. NAT service reloaded.",
	}

	h.logger.Infof("Created router %s (ID: %s) by user role: %s", router.Name, router.ID, userRole)

	// Log router creation
	if h.activityLogService != nil {
		currentUser, exists := middleware.GetUserFromContext(c)
		if exists {
			currentUserID := currentUser.ID
			h.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &currentUserID,
				Username:     currentUser.Username,
				UserRole:     string(currentUser.Role),
				ActionType:   models.ActionCreate,
				ResourceType: models.ResourceRouter,
				ResourceID:   router.Name,
				Description:  "Created router: " + router.Name,
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateRouter handles PUT /api/routers/:id - Update router
func (h *RouterHandler) UpdateRouter(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Only administrators can update routers
	if userRole != "Administrator" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Insufficient permissions to update router",
		})
		return
	}

	routerID := c.Param("id")
	if routerID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Router ID is required",
		})
		return
	}

	var req models.RouterUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid router update request for %s: %v", routerID, err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Update router
	router, err := h.routerService.UpdateRouter(routerID, &req, string(userRole))
	if err != nil {
		h.logger.Errorf("Failed to update router %s: %v", routerID, err)

		if err.Error() == "router not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Status:  "error",
				Message: "Router not found",
			})
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Status:  "error",
				Message: err.Error(),
			})
		}
		return
	}

	// Reload NAT service routers after update
	if h.natService != nil {
		if reloadErr := h.natService.ReloadRouters(); reloadErr != nil {
			h.logger.Warnf("Failed to reload NAT service after router update: %v", reloadErr)
			// Don't fail the request, just log the warning
		} else {
			h.logger.Infof("✅ NAT service reloaded successfully after router update")
		}
	}

	response := models.RouterUpdateResponse{
		Status:  "success",
		Data:    *router,
		Message: "Router updated successfully. NAT service reloaded.",
	}

	h.logger.Infof("Updated router %s by user role: %s", routerID, userRole)

	// Log router update
	if h.activityLogService != nil {
		currentUser, exists := middleware.GetUserFromContext(c)
		if exists {
			currentUserID := currentUser.ID
			h.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &currentUserID,
				Username:     currentUser.Username,
				UserRole:     string(currentUser.Role),
				ActionType:   models.ActionUpdate,
				ResourceType: models.ResourceRouter,
				ResourceID:   router.Name,
				Description:  "Updated router: " + router.Name,
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// DeleteRouter handles DELETE /api/routers/:id - Delete router
func (h *RouterHandler) DeleteRouter(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Only administrators can delete routers
	if userRole != "Administrator" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Insufficient permissions to delete router",
		})
		return
	}

	routerID := c.Param("id")
	if routerID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Router ID is required",
		})
		return
	}

	// Delete router
	err := h.routerService.DeleteRouter(routerID, string(userRole))
	if err != nil {
		h.logger.Errorf("Failed to delete router %s: %v", routerID, err)

		if err.Error() == "router not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Status:  "error",
				Message: "Router not found",
			})
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Status:  "error",
				Message: err.Error(),
			})
		}
		return
	}

	// Reload NAT service routers after deletion
	if h.natService != nil {
		if reloadErr := h.natService.ReloadRouters(); reloadErr != nil {
			h.logger.Warnf("Failed to reload NAT service after router deletion: %v", reloadErr)
			// Don't fail the request, just log the warning
		} else {
			h.logger.Infof("✅ NAT service reloaded successfully after router deletion")
		}
	}

	response := models.RouterDeleteResponse{
		Status:  "success",
		Message: "Router deleted successfully. NAT service reloaded.",
	}

	h.logger.Infof("Deleted router %s by user role: %s", routerID, userRole)

	// Log router deletion
	if h.activityLogService != nil {
		currentUser, exists := middleware.GetUserFromContext(c)
		if exists {
			currentUserID := currentUser.ID
			h.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &currentUserID,
				Username:     currentUser.Username,
				UserRole:     string(currentUser.Role),
				ActionType:   models.ActionDelete,
				ResourceType: models.ResourceRouter,
				ResourceID:   routerID,
				Description:  "Deleted router: " + routerID,
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	c.JSON(http.StatusOK, response)
}

// TestRouter handles POST /api/routers/:id/test - Test router connection
func (h *RouterHandler) TestRouter(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	routerID := c.Param("id")
	if routerID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Router ID is required",
		})
		return
	}

	// Test router connection
	testResult, err := h.routerService.TestRouter(routerID, string(userRole))
	if err != nil {
		h.logger.Errorf("Failed to test router %s for role %s: %v", routerID, userRole, err)

		if err.Error() == "router not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Status:  "error",
				Message: "Router not found",
			})
		} else if err.Error() == "access denied to router" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Status:  "error",
				Message: "Access denied to this router",
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Status:  "error",
				Message: "Failed to test router connection",
			})
		}
		return
	}

	h.logger.Infof("Tested router %s for user role %s: %s", routerID, userRole, testResult.TestResult.Status)
	c.JSON(http.StatusOK, testResult)
}

// GetRouterStats handles GET /api/routers/stats - Get router statistics
func (h *RouterHandler) GetRouterStats(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Get router statistics with role-based filtering
	stats, err := h.routerService.GetRouterStats(string(userRole))
	if err != nil {
		h.logger.Errorf("Failed to get router stats for role %s: %v", userRole, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve router statistics",
		})
		return
	}

	h.logger.Infof("Retrieved router stats for user role: %s (Total: %d, Active: %d)",
		userRole, stats.TotalRouters, stats.ActiveRouters)
	c.JSON(http.StatusOK, stats)
}

// ValidateRouter handles POST /api/routers/validate - Validate router configuration without saving
func (h *RouterHandler) ValidateRouter(c *gin.Context) {
	// Get user role from context
	_, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	var req models.RouterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid router validation request: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	// Basic validation (without creating the router)
	var errors []models.RouterValidationError

	if req.Name == "" {
		errors = append(errors, models.RouterValidationError{
			Field:   "name",
			Message: "Router name is required",
		})
	}

	if req.Host == "" {
		errors = append(errors, models.RouterValidationError{
			Field:   "host",
			Message: "Router host is required",
		})
	}

	if req.Port < 1 || req.Port > 65535 {
		errors = append(errors, models.RouterValidationError{
			Field:   "port",
			Message: "Port must be between 1 and 65535",
			Value:   strconv.Itoa(req.Port),
		})
	}

	if req.Username == "" {
		errors = append(errors, models.RouterValidationError{
			Field:   "username",
			Message: "Username is required",
		})
	}

	if req.Password == "" {
		errors = append(errors, models.RouterValidationError{
			Field:   "password",
			Message: "Password is required",
		})
	}

	if req.TunnelEndpoint == "" {
		errors = append(errors, models.RouterValidationError{
			Field:   "tunnel_endpoint",
			Message: "Tunnel endpoint is required",
		})
	}

	if req.PublicONTURL == "" {
		errors = append(errors, models.RouterValidationError{
			Field:   "public_ont_url",
			Message: "Public ONT URL is required",
		})
	}

	if len(errors) > 0 {
		response := models.RouterValidationResponse{
			Status: "error",
			Errors: errors,
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// If validation passes
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Router configuration is valid",
	})
}

// ReloadConfiguration handles POST /api/routers/reload - Reload router configuration from file
func (h *RouterHandler) ReloadConfiguration(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Only administrators can reload configuration
	if userRole != "Administrator" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Insufficient permissions to reload configuration",
		})
		return
	}

	// Reload configuration
	err := h.routerService.ReloadConfiguration()
	if err != nil {
		h.logger.Errorf("Failed to reload router configuration: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: "Failed to reload configuration: " + err.Error(),
		})
		return
	}

	h.logger.Infof("Router configuration reloaded by user role: %s", userRole)
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Router configuration reloaded successfully",
		"config_file": h.routerService.GetConfigurationPath(),
	})
}

// GetConfigurationInfo handles GET /api/routers/config - Get configuration file information
func (h *RouterHandler) GetConfigurationInfo(c *gin.Context) {
	// Get user role from context
	userRole, exists := middleware.GetUserRoleFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	// Only administrators can view configuration info
	if userRole != "Administrator" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Insufficient permissions to view configuration info",
		})
		return
	}

	configPath := h.routerService.GetConfigurationPath()

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"config_file": configPath,
		"message":     "Configuration information retrieved successfully",
	})
}