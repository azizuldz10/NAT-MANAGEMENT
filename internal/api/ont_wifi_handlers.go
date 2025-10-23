package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"nat-management-app/internal/database"
	"nat-management-app/internal/models"
	"nat-management-app/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ONTWiFiHandler handles HTTP requests for ONT WiFi extraction operations
type ONTWiFiHandler struct {
	extractorService *services.ONTExtractorService
	wifiRepo         *database.ONTWiFiRepository
	natService       *services.NATService
	activityLogger   *services.ActivityLogService
	logger           *logrus.Logger
}

// NewONTWiFiHandler creates a new ONT WiFi handler
func NewONTWiFiHandler(
	extractorService *services.ONTExtractorService,
	wifiRepo *database.ONTWiFiRepository,
	natService *services.NATService,
	activityLogger *services.ActivityLogService,
	logger *logrus.Logger,
) *ONTWiFiHandler {
	return &ONTWiFiHandler{
		extractorService: extractorService,
		wifiRepo:         wifiRepo,
		natService:       natService,
		activityLogger:   activityLogger,
		logger:           logger,
	}
}

// ExtractWiFiInfo extracts WiFi information from an ONT device
// POST /api/ont/wifi/extract
func (h *ONTWiFiHandler) ExtractWiFiInfo(c *gin.Context) {
	var req models.ONTWiFiExtractRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ONTWiFiExtractResponse{
			Status:    "error",
			Message:   fmt.Sprintf("Invalid request: %v", err),
			Timestamp: time.Now(),
		})
		return
	}

	// Get user info from context
	username := getUsernameFromContext(c)

	// Start extraction
	startTime := time.Now()
	h.logger.Infof("Starting WiFi extraction for ONT: %s (requested by: %s)", req.ONTURL, username)

	// Extract WiFi info using webautomation
	wifiInfo, err := h.extractorService.ExtractWiFiInfo(req.ONTURL, req.Username, req.Password, req.Debug)
	if err != nil {
		h.logger.Errorf("WiFi extraction failed: %v", err)

		// Log activity
		if h.activityLogger != nil {
			_ = h.activityLogger.CreateLog(&models.ActivityLogCreate{
				Username:     username,
				ActionType:   "ONT_WIFI_EXTRACT",
				ResourceType: "ONT",
				ResourceID:   req.ONTURL,
				Description:  fmt.Sprintf("Failed to extract WiFi info: %v", err),
				Status:       models.StatusFailed,
				IPAddress:    c.ClientIP(),
			})
		}

		c.JSON(http.StatusInternalServerError, models.ONTWiFiExtractResponse{
			Status:    "error",
			Message:   fmt.Sprintf("WiFi extraction failed: %v", err),
			Timestamp: time.Now(),
		})
		return
	}

	// Add metadata
	wifiInfo.PPPoEUsername = req.PPPoEUsername
	wifiInfo.Router = req.Router
	wifiInfo.ExtractedBy = username

	// Save to database
	if err := h.wifiRepo.SaveWiFiInfo(c.Request.Context(), wifiInfo); err != nil {
		h.logger.Errorf("Failed to save WiFi info to database: %v", err)
		// Continue anyway - extraction was successful
	}

	// Log activity
	if h.activityLogger != nil {
		_ = h.activityLogger.CreateLog(&models.ActivityLogCreate{
			Username:     username,
			ActionType:   "ONT_WIFI_EXTRACT",
			ResourceType: "ONT",
			ResourceID:   req.ONTURL,
			Description:  fmt.Sprintf("Successfully extracted WiFi info (SSID: %s)", wifiInfo.SSID),
			Status:       models.StatusSuccess,
			IPAddress:    c.ClientIP(),
		})
	}

	extractionTime := time.Since(startTime)
	h.logger.Infof("WiFi extraction completed in %v", extractionTime)

	c.JSON(http.StatusOK, models.ONTWiFiExtractResponse{
		Status:         "success",
		Data:           wifiInfo,
		Message:        "WiFi information extracted successfully",
		ExtractionTime: extractionTime,
		Timestamp:      time.Now(),
	})
}

// ExtractWiFiFromNAT extracts WiFi info using existing NAT configuration
// POST /api/ont/wifi/extract-from-nat
func (h *ONTWiFiHandler) ExtractWiFiFromNAT(c *gin.Context) {
	var req models.ONTWiFiExtractFromNATRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ONTWiFiExtractResponse{
			Status:    "error",
			Message:   fmt.Sprintf("Invalid request: %v", err),
			Timestamp: time.Now(),
		})
		return
	}

	username := getUsernameFromContext(c)

	// Get NAT configuration for the router
	natConfig, err := h.natService.GetONTNATRule(req.Router)
	if err != nil {
		h.logger.Errorf("Failed to get NAT config: %v", err)
		c.JSON(http.StatusNotFound, models.ONTWiFiExtractResponse{
			Status:    "error",
			Message:   fmt.Sprintf("NAT configuration not found: %v", err),
			Timestamp: time.Now(),
		})
		return
	}

	// Check if PublicONTURL is available
	if natConfig.PublicONTURL == "" {
		c.JSON(http.StatusBadRequest, models.ONTWiFiExtractResponse{
			Status:    "error",
			Message:   "No public ONT URL configured for this router",
			Timestamp: time.Now(),
		})
		return
	}

	// Extract WiFi info
	startTime := time.Now()
	wifiInfo, err := h.extractorService.ExtractWiFiInfo(
		natConfig.PublicONTURL,
		req.ONTUsername,
		req.ONTPassword,
		req.Debug,
	)
	if err != nil {
		h.logger.Errorf("WiFi extraction failed: %v", err)
		c.JSON(http.StatusInternalServerError, models.ONTWiFiExtractResponse{
			Status:    "error",
			Message:   fmt.Sprintf("WiFi extraction failed: %v", err),
			Timestamp: time.Now(),
		})
		return
	}

	// Add metadata
	wifiInfo.PPPoEUsername = req.PPPoEUsername
	wifiInfo.Router = req.Router
	wifiInfo.ExtractedBy = username

	// Save to database
	if err := h.wifiRepo.SaveWiFiInfo(c.Request.Context(), wifiInfo); err != nil {
		h.logger.Errorf("Failed to save WiFi info: %v", err)
	}

	// Log activity
	if h.activityLogger != nil {
		_ = h.activityLogger.CreateLog(&models.ActivityLogCreate{
			Username:     username,
			ActionType:   "ONT_WIFI_EXTRACT_FROM_NAT",
			ResourceType: "ROUTER",
			ResourceID:   req.Router,
			Description:  fmt.Sprintf("Extracted WiFi info from NAT config (SSID: %s)", wifiInfo.SSID),
			Status:       models.StatusSuccess,
			IPAddress:    c.ClientIP(),
		})
	}

	extractionTime := time.Since(startTime)

	c.JSON(http.StatusOK, models.ONTWiFiExtractResponse{
		Status:         "success",
		Data:           wifiInfo,
		Message:        "WiFi information extracted successfully",
		ExtractionTime: extractionTime,
		Timestamp:      time.Now(),
	})
}

// GetWiFiHistory retrieves WiFi extraction history
// GET /api/ont/wifi/history
func (h *ONTWiFiHandler) GetWiFiHistory(c *gin.Context) {
	// Parse query parameters
	req := models.ONTWiFiHistoryRequest{
		PPPoEUsername: c.Query("pppoe_username"),
		Router:        c.Query("router"),
	}

	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			req.Limit = val
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			req.Offset = val
		}
	}

	// Query database
	history, total, err := h.wifiRepo.GetWiFiHistory(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to get WiFi history: %v", err)
		c.JSON(http.StatusInternalServerError, models.ONTWiFiHistoryResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to retrieve history: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.ONTWiFiHistoryResponse{
		Status:  "success",
		Data:    history,
		Total:   total,
		Message: fmt.Sprintf("Retrieved %d WiFi extraction records", len(history)),
	})
}

// GetLatestWiFiInfo retrieves the most recent WiFi info for a PPPoE user
// GET /api/ont/wifi/latest/:pppoe_username
func (h *ONTWiFiHandler) GetLatestWiFiInfo(c *gin.Context) {
	pppoeUsername := c.Param("pppoe_username")

	if pppoeUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "PPPoE username is required",
		})
		return
	}

	wifiInfo, err := h.wifiRepo.GetLatestWiFiInfoByPPPoE(c.Request.Context(), pppoeUsername)
	if err != nil {
		h.logger.Errorf("Failed to get latest WiFi info: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("No WiFi info found for user: %s", pppoeUsername),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   wifiInfo,
	})
}

// CheckAvailability checks if webautomation tools are available
// GET /api/ont/wifi/availability
func (h *ONTWiFiHandler) CheckAvailability(c *gin.Context) {
	err := h.extractorService.CheckWebautomationAvailability()

	if err != nil {
		c.JSON(http.StatusOK, models.ONTWiFiAvailabilityResponse{
			Status:    "error",
			Available: false,
			Message:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.ONTWiFiAvailabilityResponse{
		Status:          "success",
		Available:       true,
		Message:         "Webautomation tools are available and ready",
		SupportedModels: h.extractorService.GetSupportedModels(),
	})
}

// SearchBySSID searches WiFi information by SSID
// GET /api/ont/wifi/search
func (h *ONTWiFiHandler) SearchBySSID(c *gin.Context) {
	ssid := c.Query("ssid")
	if ssid == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "SSID parameter is required",
		})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 {
			limit = val
		}
	}

	results, err := h.wifiRepo.SearchWiFiBySSID(c.Request.Context(), ssid, limit)
	if err != nil {
		h.logger.Errorf("Failed to search WiFi by SSID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Search failed: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    results,
		"count":   len(results),
		"message": fmt.Sprintf("Found %d results for SSID: %s", len(results), ssid),
	})
}

// GetWiFiStats retrieves statistics about WiFi extraction records
// GET /api/ont/wifi/stats
func (h *ONTWiFiHandler) GetWiFiStats(c *gin.Context) {
	stats, err := h.wifiRepo.GetWiFiInfoStats(c.Request.Context())
	if err != nil {
		h.logger.Errorf("Failed to get WiFi stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": fmt.Sprintf("Failed to retrieve stats: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   stats,
	})
}

// Helper function to get username from Gin context
func getUsernameFromContext(c *gin.Context) string {
	if user, exists := c.Get("username"); exists {
		if username, ok := user.(string); ok {
			return username
		}
	}
	return "unknown"
}
