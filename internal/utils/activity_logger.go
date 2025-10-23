package utils

import (
	"encoding/json"
	"time"

	"nat-management-app/internal/models"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
)

// ActivityLogger provides enhanced activity logging with duration tracking and metadata
type ActivityLogger struct {
	logService *services.ActivityLogService
	c          *gin.Context
	startTime  time.Time
	actionType string
	resourceType string
	resourceID string
	description string
	metadata   map[string]interface{}
}

// NewActivityLogger creates a new activity logger for tracking an operation
func NewActivityLogger(logService *services.ActivityLogService, c *gin.Context) *ActivityLogger {
	return &ActivityLogger{
		logService: logService,
		c:          c,
		startTime:  time.Now(),
		metadata:   make(map[string]interface{}),
	}
}

// SetAction sets the action type and resource information
func (al *ActivityLogger) SetAction(actionType, resourceType, resourceID, description string) *ActivityLogger {
	al.actionType = actionType
	al.resourceType = resourceType
	al.resourceID = resourceID
	al.description = description
	return al
}

// AddMetadata adds metadata to the log
func (al *ActivityLogger) AddMetadata(key string, value interface{}) *ActivityLogger {
	al.metadata[key] = value
	return al
}

// AddBeforeState adds the "before" state to metadata (for updates)
func (al *ActivityLogger) AddBeforeState(before interface{}) *ActivityLogger {
	// Convert to map if needed
	beforeJSON, err := json.Marshal(before)
	if err == nil {
		var beforeMap map[string]interface{}
		if err := json.Unmarshal(beforeJSON, &beforeMap); err == nil {
			al.metadata["before"] = beforeMap
		}
	}
	return al
}

// AddAfterState adds the "after" state to metadata (for updates)
func (al *ActivityLogger) AddAfterState(after interface{}) *ActivityLogger {
	// Convert to map if needed
	afterJSON, err := json.Marshal(after)
	if err == nil {
		var afterMap map[string]interface{}
		if err := json.Unmarshal(afterJSON, &afterMap); err == nil {
			al.metadata["after"] = afterMap
		}
	}
	return al
}

// AddCircuitBreakerState adds circuit breaker state to metadata
func (al *ActivityLogger) AddCircuitBreakerState(routerName, state string) *ActivityLogger {
	if al.metadata["circuit_breaker"] == nil {
		al.metadata["circuit_breaker"] = make(map[string]string)
	}
	if cb, ok := al.metadata["circuit_breaker"].(map[string]string); ok {
		cb[routerName] = state
	}
	return al
}

// LogSuccess logs a successful operation
func (al *ActivityLogger) LogSuccess() {
	al.log(models.StatusSuccess, "")
}

// LogError logs a failed operation
func (al *ActivityLogger) LogError(errorMsg string) {
	al.log(models.StatusError, errorMsg)
}

// LogFailed logs a failed operation (validation, business logic failure)
func (al *ActivityLogger) LogFailed(failureMsg string) {
	al.log(models.StatusFailed, failureMsg)
}

// log creates the activity log entry
func (al *ActivityLogger) log(status, errorMessage string) {
	// Calculate duration
	duration := int(time.Since(al.startTime).Milliseconds())

	// Get user info from context
	user, exists := al.c.Get("user")
	if !exists {
		return // Can't log without user context
	}

	var userID *int
	var username, userRole string

	if u, ok := user.(*models.User); ok {
		userID = &u.ID
		username = u.Username
		userRole = string(u.Role)
	}

	// Parse device info from user agent
	userAgent := al.c.GetHeader("User-Agent")
	deviceInfo := ParseUserAgent(userAgent)

	// Create log entry
	logEntry := &models.ActivityLogCreate{
		UserID:       userID,
		Username:     username,
		UserRole:     userRole,
		ActionType:   al.actionType,
		ResourceType: al.resourceType,
		ResourceID:   al.resourceID,
		Description:  al.description,
		IPAddress:    al.c.ClientIP(),
		UserAgent:    userAgent,
		Status:       status,
		ErrorMessage: errorMessage,
		DurationMs:   &duration,
		DeviceInfo:   deviceInfo,
		Metadata:     al.metadata,
	}

	// Log it (don't fail main operation if logging fails)
	_ = al.logService.CreateLog(logEntry)
}

// QuickLog is a convenience function for simple logging without duration tracking
func QuickLog(logService *services.ActivityLogService, c *gin.Context, actionType, resourceType, resourceID, description string) {
	logger := NewActivityLogger(logService, c)
	logger.SetAction(actionType, resourceType, resourceID, description)
	logger.LogSuccess()
}

// QuickLogWithStatus is a convenience function for logging with custom status
func QuickLogWithStatus(logService *services.ActivityLogService, c *gin.Context, actionType, resourceType, resourceID, description, status, errorMsg string) {
	logger := NewActivityLogger(logService, c)
	logger.SetAction(actionType, resourceType, resourceID, description)
	if status == models.StatusSuccess {
		logger.LogSuccess()
	} else if status == models.StatusError {
		logger.LogError(errorMsg)
	} else {
		logger.LogFailed(errorMsg)
	}
}
