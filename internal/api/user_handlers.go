package api

import (
	"net/http"
	"strconv"

	"nat-management-app/internal/middleware"
	"nat-management-app/internal/models"
	"nat-management-app/internal/services"
	"nat-management-app/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// UserHandler handles user management HTTP requests
type UserHandler struct {
	userService        *services.UserService
	activityLogService *services.ActivityLogService
	logger             *logrus.Logger
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService *services.UserService, activityLogService *services.ActivityLogService, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		userService:        userService,
		activityLogService: activityLogService,
		logger:             logger,
	}
}

// CreateUser handles POST /api/users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Invalid create user request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Create user
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		h.logger.Errorf("Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to create user: " + err.Error(),
		})
		return
	}

	h.logger.Infof("✅ User '%s' created successfully by admin", user.Username)

	// Log user creation
	if h.activityLogService != nil {
		currentUser, exists := middleware.GetUserFromContext(c)
		if exists {
			currentUserID := currentUser.ID
			h.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &currentUserID,
				Username:     currentUser.Username,
				UserRole:     string(currentUser.Role),
				ActionType:   models.ActionCreate,
				ResourceType: models.ResourceUser,
				ResourceID:   strconv.Itoa(user.ID),
				Description:  "Created user: " + user.Username,
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "User created successfully",
		"data":    user,
	})
}

// GetUser handles GET /api/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.Errorf("Error getting user %d: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
	})
}

// ListUsers handles GET /api/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Check if search query provided
	searchQuery := c.Query("search")
	if searchQuery != "" {
		// Search users
		users, err := h.userService.SearchUsers(searchQuery)
		if err != nil {
			h.logger.Errorf("Error searching users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to search users",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   users,
			"total":  len(users),
		})
		return
	}

	// List all users with pagination
	users, total, err := h.userService.ListUsers(limit, offset)
	if err != nil {
		h.logger.Errorf("Error listing users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateUser handles PUT /api/users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Start activity logger for duration tracking
	activityLog := utils.NewActivityLogger(h.activityLogService, c)

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	// Get existing user for before state
	existingUser, err := h.userService.GetUserByID(userID)
	if err != nil {
		activityLog.SetAction(models.ActionUpdate, models.ResourceUser, strconv.Itoa(userID), "Failed to update user (not found)")
		activityLog.LogFailed("User not found: " + err.Error())

		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "User not found",
		})
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warnf("Invalid update user request: %v", err)

		activityLog.SetAction(models.ActionUpdate, models.ResourceUser, strconv.Itoa(userID), "Failed to update user (invalid request)")
		activityLog.LogFailed("Invalid request data: " + err.Error())

		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request data: " + err.Error(),
		})
		return
	}

	// Capture before state
	activityLog.AddBeforeState(existingUser)

	// Update user
	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		h.logger.Errorf("Error updating user %d: %v", userID, err)

		activityLog.SetAction(models.ActionUpdate, models.ResourceUser, strconv.Itoa(userID), "Failed to update user")
		activityLog.LogError("Update failed: " + err.Error())

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to update user: " + err.Error(),
		})
		return
	}

	h.logger.Infof("✅ User %d updated successfully", userID)

	// Capture after state and log success with duration
	activityLog.SetAction(models.ActionUpdate, models.ResourceUser, strconv.Itoa(userID), "Updated user: "+user.Username)
	activityLog.AddAfterState(user)
	activityLog.AddMetadata("updated_fields", getUpdatedFields(existingUser, user))
	activityLog.LogSuccess()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User updated successfully",
		"data":    user,
	})
}

// getUpdatedFields compares before and after user states to identify changed fields
func getUpdatedFields(before, after *services.UserWithRouters) []string {
	var updated []string

	if before.FullName != after.FullName {
		updated = append(updated, "full_name")
	}
	if before.Email != after.Email {
		updated = append(updated, "email")
	}
	if before.IsActive != after.IsActive {
		updated = append(updated, "is_active")
	}
	if len(before.Routers) != len(after.Routers) {
		updated = append(updated, "routers")
	}

	return updated
}

// DeleteUser handles DELETE /api/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	// Check if trying to delete self
	currentUserID, exists := c.Get("user_id")
	if exists && currentUserID.(int) == userID {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  "error",
			"message": "Cannot delete your own account",
		})
		return
	}

	// Delete user (soft delete)
	err = h.userService.DeleteUser(userID)
	if err != nil {
		h.logger.Errorf("Error deleting user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to delete user: " + err.Error(),
		})
		return
	}

	h.logger.Infof("✅ User %d deleted successfully", userID)

	// Log user deletion
	if h.activityLogService != nil {
		currentUser, exists := middleware.GetUserFromContext(c)
		if exists {
			currentUserID := currentUser.ID
			h.activityLogService.CreateLog(&models.ActivityLogCreate{
				UserID:       &currentUserID,
				Username:     currentUser.Username,
				UserRole:     string(currentUser.Role),
				ActionType:   models.ActionDelete,
				ResourceType: models.ResourceUser,
				ResourceID:   strconv.Itoa(userID),
				Description:  "Deleted user ID: " + strconv.Itoa(userID),
				IPAddress:    c.ClientIP(),
				UserAgent:    c.GetHeader("User-Agent"),
				Status:       models.StatusSuccess,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User deleted successfully",
	})
}

// GetUserRouters handles GET /api/users/:id/routers
func (h *UserHandler) GetUserRouters(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	routers, err := h.userService.GetUserRouters(userID)
	if err != nil {
		h.logger.Errorf("Error getting user routers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get user routers",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   routers,
	})
}

// GetRouterUsers handles GET /api/routers/:name/users
func (h *UserHandler) GetRouterUsers(c *gin.Context) {
	routerName := c.Param("name")

	users, err := h.userService.GetRouterUsers(routerName)
	if err != nil {
		h.logger.Errorf("Error getting router users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get router users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   users,
		"total":  len(users),
	})
}

// GetUserStats handles GET /api/users/:id/stats
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	stats, err := h.userService.GetUserStats(userID)
	if err != nil {
		h.logger.Errorf("Error getting user stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to get user stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// ActivateUser handles PATCH /api/users/:id/activate
func (h *UserHandler) ActivateUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	// Get user and set is_active = true
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "User not found",
		})
		return
	}

	updateReq := services.UpdateUserRequest{
		FullName: user.FullName,
		Email:    user.Email,
		Routers:  user.Routers,
		IsActive: true,
	}

	updatedUser, err := h.userService.UpdateUser(userID, &updateReq)
	if err != nil {
		h.logger.Errorf("Error activating user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to activate user",
		})
		return
	}

	h.logger.Infof("✅ User %d activated successfully", userID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User activated successfully",
		"data":    updatedUser,
	})
}

// ChangeUserPassword handles PATCH /api/users/:id/password
func (h *UserHandler) ChangeUserPassword(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid user ID",
		})
		return
	}

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get user
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": "User not found",
		})
		return
	}

	// Update password
	updateReq := services.UpdateUserRequest{
		FullName: user.FullName,
		Email:    user.Email,
		Password: req.NewPassword,
		Routers:  user.Routers,
		IsActive: user.IsActive,
	}

	_, err = h.userService.UpdateUser(userID, &updateReq)
	if err != nil {
		h.logger.Errorf("Error changing password for user %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to change password",
		})
		return
	}

	h.logger.Infof("✅ Password changed for user %d", userID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Password changed successfully",
	})
}
