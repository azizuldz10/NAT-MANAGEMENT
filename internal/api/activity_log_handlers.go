package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"nat-management-app/internal/middleware"
	"nat-management-app/internal/models"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ActivityLogHandler handles activity log API endpoints
type ActivityLogHandler struct {
	logService *services.ActivityLogService
	logger     *logrus.Logger
}

// NewActivityLogHandler creates a new activity log handler
func NewActivityLogHandler(logService *services.ActivityLogService, logger *logrus.Logger) *ActivityLogHandler {
	return &ActivityLogHandler{
		logService: logService,
		logger:     logger,
	}
}

// GetLogs handles GET /api/logs
func (h *ActivityLogHandler) GetLogs(c *gin.Context) {
	// Check if user is Administrator
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	if user.Role != models.RoleAdministrator {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Only administrators can view activity logs",
		})
		return
	}

	// Parse query parameters
	filter := &models.ActivityLogFilter{}

	// User filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}

	// Username filter
	filter.Username = c.Query("username")

	// Action type filter (single - backward compatibility)
	filter.ActionType = c.Query("action_type")

	// Action types filter (multi-select)
	if actionTypesStr := c.Query("action_types"); actionTypesStr != "" {
		filter.ActionTypes = strings.Split(actionTypesStr, ",")
	}

	// Resource type filter (single - backward compatibility)
	filter.ResourceType = c.Query("resource_type")

	// Resource types filter (multi-select)
	if resourceTypesStr := c.Query("resource_types"); resourceTypesStr != "" {
		filter.ResourceTypes = strings.Split(resourceTypesStr, ",")
	}

	// Status filter (single - backward compatibility)
	filter.Status = c.Query("status")

	// Statuses filter (multi-select)
	if statusesStr := c.Query("statuses"); statusesStr != "" {
		filter.Statuses = strings.Split(statusesStr, ",")
	}

	// Date range filter
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			filter.EndDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	// Pagination
	limit := 50 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	filter.Limit = limit

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	filter.Offset = offset

	// Get logs
	logs, total, err := h.logService.GetLogs(filter)
	if err != nil {
		h.logger.Errorf("Failed to get activity logs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve activity logs",
		})
		return
	}

	// Calculate pagination
	currentPage := (offset / limit) + 1
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, models.ActivityLogsResponse{
		Status: "success",
		Data:   logs,
		Total:  total,
		Pagination: &models.Pagination{
			CurrentPage: currentPage,
			PerPage:     limit,
			TotalPages:  totalPages,
			TotalItems:  total,
		},
	})
}

// GetLogByID handles GET /api/logs/:id
func (h *ActivityLogHandler) GetLogByID(c *gin.Context) {
	// Check if user is Administrator
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	if user.Role != models.RoleAdministrator {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Only administrators can view activity logs",
		})
		return
	}

	// Parse log ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid log ID",
		})
		return
	}

	// Get log
	log, err := h.logService.GetLogByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Status:  "error",
			Message: "Activity log not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   log,
	})
}

// GetLogStats handles GET /api/logs/stats
func (h *ActivityLogHandler) GetLogStats(c *gin.Context) {
	// Check if user is Administrator
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	if user.Role != models.RoleAdministrator {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Only administrators can view activity log statistics",
		})
		return
	}

	// Get stats
	stats, err := h.logService.GetLogStats()
	if err != nil {
		h.logger.Errorf("Failed to get activity log stats: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: "Failed to retrieve statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// DeleteOldLogs handles POST /api/logs/cleanup
func (h *ActivityLogHandler) DeleteOldLogs(c *gin.Context) {
	// Check if user is Administrator
	user, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Status:  "error",
			Message: "Authentication required",
		})
		return
	}

	if user.Role != models.RoleAdministrator {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Status:  "error",
			Message: "Only administrators can delete old logs",
		})
		return
	}

	// Parse request
	var req struct {
		DaysToKeep int `json:"days_to_keep" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Delete old logs
	deletedCount, err := h.logService.DeleteOldLogs(req.DaysToKeep)
	if err != nil {
		h.logger.Errorf("Failed to delete old logs: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Status:  "error",
			Message: "Failed to delete old logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Old logs deleted successfully",
		"data": gin.H{
			"deleted_count": deletedCount,
			"days_kept":     req.DaysToKeep,
		},
	})
}

// Helper function to create activity log (can be called from other handlers)
func CreateActivityLog(logService *services.ActivityLogService, log *models.ActivityLogCreate) {
	if err := logService.CreateLog(log); err != nil {
		// Log the error but don't fail the main operation
		// Activity logging is important but shouldn't break functionality
	}
}
