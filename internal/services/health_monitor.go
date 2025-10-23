package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"nat-management-app/internal/models"

	"github.com/sirupsen/logrus"
)

// HealthStatus represents the health status of a router
type HealthStatus struct {
	RouterID          string     `json:"router_id"`
	RouterName        string     `json:"router_name"`
	Status            string     `json:"status"` // healthy, degraded, down
	LastChecked       time.Time  `json:"last_checked"`
	LastSeen          time.Time  `json:"last_seen"`
	ResponseTime      int64      `json:"response_time_ms"`
	ConsecutiveFails  int        `json:"consecutive_fails"`
	DownSince         *time.Time `json:"down_since,omitempty"`
	UptimePercent     float64    `json:"uptime_percent"`
	ErrorMessage      string     `json:"error_message,omitempty"`
	CheckCount        int64      `json:"check_count"`
	FailCount         int64      `json:"fail_count"`
	// Resource metrics
	ActiveConnections int     `json:"active_connections"`
	CPUUsage          float64 `json:"cpu_usage"`
	RAMUsage          float64 `json:"ram_usage"`
	RAMTotal          float64 `json:"ram_total"`
}

// HealthCache stores health status with TTL
type HealthCache struct {
	mu    sync.RWMutex
	cache map[string]*CachedHealth
}

// CachedHealth represents cached health data
type CachedHealth struct {
	Status    *HealthStatus
	ExpiresAt time.Time
}

// HealthMonitor monitors router health in background
type HealthMonitor struct {
	logger         *logrus.Logger
	routerService  *RouterServiceDB
	cache          *HealthCache
	states         map[string]*RouterState
	statesMu       sync.RWMutex
	checkInterval  time.Duration
	cacheTTL       time.Duration
	failThreshold  int
	ctx            context.Context
	cancel         context.CancelFunc
}

// RouterState tracks router state over time
type RouterState struct {
	CurrentStatus    string
	ConsecutiveFails int
	LastSeenAt       time.Time
	DownSince        *time.Time
	CheckCount       int64
	FailCount        int64
	TotalUptime      time.Duration
	LastCheckTime    time.Time
}

// NewHealthMonitor creates a new health monitor instance
func NewHealthMonitor(logger *logrus.Logger, routerService *RouterServiceDB) *HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &HealthMonitor{
		logger:        logger,
		routerService: routerService,
		cache: &HealthCache{
			cache: make(map[string]*CachedHealth),
		},
		states:        make(map[string]*RouterState),
		checkInterval: 30 * time.Second,  // Check every 30 seconds (match UI refresh)
		cacheTTL:      5 * time.Minute,   // Cache for 5 minutes
		failThreshold: 3,                 // Declare down after 3 consecutive fails
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start begins the health monitoring background worker
func (hm *HealthMonitor) Start() {
	hm.logger.Info("🏥 Health Monitor starting...")

	// Initial check
	go hm.checkAllRouters()

	// Start periodic checker
	go hm.healthWorker()

	// Start cache cleanup
	go hm.cacheCleanupWorker()

	hm.logger.Info("✅ Health Monitor started successfully")
}

// Stop stops the health monitoring
func (hm *HealthMonitor) Stop() {
	hm.logger.Info("⏹️ Stopping Health Monitor...")
	hm.cancel()
}

// healthWorker runs periodic health checks
func (hm *HealthMonitor) healthWorker() {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			hm.logger.Info("Health worker stopped")
			return
		case <-ticker.C:
			hm.checkAllRouters()
		}
	}
}

// checkAllRouters checks health of all routers
func (hm *HealthMonitor) checkAllRouters() {
	hm.logger.Debug("🔍 Starting health check for all routers...")

	// Get all routers (admin role to see all)
	routers, err := hm.routerService.GetAllRouters("Administrator")
	if err != nil {
		hm.logger.Errorf("Failed to get routers for health check: %v", err)
		return
	}

	hm.logger.Debugf("📊 Found %d routers to check", len(routers))

	// Check each router concurrently
	var wg sync.WaitGroup
	for _, router := range routers {
		wg.Add(1)
		go func(r models.RouterResponse) {
			defer wg.Done()

			if r.ID == "" {
				return
			}

			hm.checkRouter(r.ID, r.Name)
		}(router)
	}

	wg.Wait()
	hm.logger.Debug("✅ Health check completed for all routers")
}

// checkRouter performs health check on a single router
func (hm *HealthMonitor) checkRouter(routerID, routerName string) {
	startTime := time.Now()

	// Try to test connection (this reuses connection pool)
	_, err := hm.routerService.TestRouter(routerID, "Administrator")

	responseTime := time.Since(startTime).Milliseconds()

	// Get resource metrics if connection successful
	var activeConns int
	var cpuUsage, ramUsage, ramTotal float64

	if err == nil {
		// Get resource metrics from router
		metrics, metricsErr := hm.getRouterMetrics(routerID)
		if metricsErr == nil {
			activeConns = metrics.ActiveConnections
			cpuUsage = metrics.CPUUsage
			ramUsage = metrics.RAMUsage
			ramTotal = metrics.RAMTotal
			hm.logger.Debugf("📊 Metrics for %s: Conns=%d, CPU=%.1f%%, RAM=%.0f/%.0f MB",
				routerName, activeConns, cpuUsage, ramUsage, ramTotal)
		} else {
			hm.logger.Warnf("⚠️ Failed to get metrics for %s: %v", routerName, metricsErr)
		}
	}

	if err != nil {
		// Connection failed
		hm.logger.Warnf("Router %s health check failed: %v", routerName, err)
		hm.updateState(routerID, routerName, false, responseTime, err.Error(), 0, 0, 0, 0)
	} else {
		// Connection successful
		hm.logger.Debugf("Router %s is healthy (response: %dms)", routerName, responseTime)
		hm.updateState(routerID, routerName, true, responseTime, "", activeConns, cpuUsage, ramUsage, ramTotal)
	}
}

// updateState updates router state and cache
func (hm *HealthMonitor) updateState(routerID, routerName string, success bool, responseTime int64, errorMsg string, activeConns int, cpuUsage, ramUsage, ramTotal float64) {
	hm.statesMu.Lock()
	defer hm.statesMu.Unlock()

	state, exists := hm.states[routerID]
	if !exists {
		state = &RouterState{
			CurrentStatus: "unknown",
			LastSeenAt:    time.Now(),
			LastCheckTime: time.Now(),
		}
		hm.states[routerID] = state
	}

	now := time.Now()
	state.CheckCount++
	state.LastCheckTime = now

	if success {
		// Reset consecutive failures
		state.ConsecutiveFails = 0
		state.LastSeenAt = now

		// Calculate uptime if was down
		if state.CurrentStatus == "down" && state.DownSince != nil {
			downtime := now.Sub(*state.DownSince)
			state.TotalUptime += downtime
			state.DownSince = nil

			hm.logger.Infof("🟢 Router %s recovered (was down for %v)", routerName, downtime)
		}

		// Determine status based on response time
		if responseTime > 1000 {
			state.CurrentStatus = "degraded"
		} else {
			state.CurrentStatus = "healthy"
		}

	} else {
		// Increment failure count
		state.ConsecutiveFails++
		state.FailCount++

		// Declare down only after threshold
		if state.ConsecutiveFails >= hm.failThreshold && state.CurrentStatus != "down" {
			state.CurrentStatus = "down"
			downSince := now
			state.DownSince = &downSince

			hm.logger.Errorf("🔴 Router %s is DOWN (failed %d consecutive checks)",
				routerName, state.ConsecutiveFails)
		}
	}

	// Calculate uptime percentage
	uptimePercent := float64(100.0)
	if state.CheckCount > 0 {
		uptimePercent = (float64(state.CheckCount-state.FailCount) / float64(state.CheckCount)) * 100
	}

	// Create health status
	healthStatus := &HealthStatus{
		RouterID:          routerID,
		RouterName:        routerName,
		Status:            state.CurrentStatus,
		LastChecked:       now,
		LastSeen:          state.LastSeenAt,
		ResponseTime:      responseTime,
		ConsecutiveFails:  state.ConsecutiveFails,
		DownSince:         state.DownSince,
		UptimePercent:     uptimePercent,
		ErrorMessage:      errorMsg,
		CheckCount:        state.CheckCount,
		FailCount:         state.FailCount,
		ActiveConnections: activeConns,
		CPUUsage:          cpuUsage,
		RAMUsage:          ramUsage,
		RAMTotal:          ramTotal,
	}

	// Update cache
	hm.cache.Set(routerID, healthStatus, hm.cacheTTL)
	hm.logger.Debugf("💾 Cached health data for %s: %s (uptime: %.2f%%)", routerName, state.CurrentStatus, uptimePercent)
}

// GetHealth returns cached health status for a router
func (hm *HealthMonitor) GetHealth(routerID string) (*HealthStatus, bool) {
	return hm.cache.Get(routerID)
}

// GetAllHealth returns health status for all routers
func (hm *HealthMonitor) GetAllHealth() []*HealthStatus {
	hm.cache.mu.RLock()
	defer hm.cache.mu.RUnlock()

	statuses := make([]*HealthStatus, 0, len(hm.cache.cache))
	for _, cached := range hm.cache.cache {
		if time.Now().Before(cached.ExpiresAt) {
			statuses = append(statuses, cached.Status)
		}
	}

	return statuses
}

// Cache methods
func (hc *HealthCache) Get(routerID string) (*HealthStatus, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	cached, exists := hc.cache[routerID]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(cached.ExpiresAt) {
		return nil, false
	}

	return cached.Status, true
}

func (hc *HealthCache) Set(routerID string, status *HealthStatus, ttl time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.cache[routerID] = &CachedHealth{
		Status:    status,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// cacheCleanupWorker periodically cleans up expired cache entries
func (hm *HealthMonitor) cacheCleanupWorker() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			hm.cleanupExpiredCache()
		}
	}
}

func (hm *HealthMonitor) cleanupExpiredCache() {
	hm.cache.mu.Lock()
	defer hm.cache.mu.Unlock()

	now := time.Now()
	for routerID, cached := range hm.cache.cache {
		if now.After(cached.ExpiresAt) {
			delete(hm.cache.cache, routerID)
		}
	}
}

// RouterMetrics holds resource metrics from a router
type RouterMetrics struct {
	ActiveConnections int
	CPUUsage          float64
	RAMUsage          float64
	RAMTotal          float64
}

// getRouterMetrics fetches resource metrics from a router
func (hm *HealthMonitor) getRouterMetrics(routerID string) (*RouterMetrics, error) {
	metrics := &RouterMetrics{}

	// Get router connection from service
	conn, err := hm.routerService.GetRouterConnection(routerID)
	if err != nil {
		return nil, err
	}

	// Get active PPPoE connections count
	reply, err := conn.Client.Run("/ppp/active/print")
	if err == nil {
		metrics.ActiveConnections = len(reply.Re)
		hm.logger.Debugf("🔍 Router %s has %d active PPPoE connections", routerID, metrics.ActiveConnections)
	} else {
		hm.logger.Warnf("⚠️ Failed to get PPPoE connections for %s: %v", routerID, err)
	}

	// Get system resources (CPU and RAM)
	reply, err = conn.Client.Run("/system/resource/print")
	if err == nil && len(reply.Re) > 0 {
		resource := reply.Re[0].Map

		// Parse CPU usage
		if cpuLoad, ok := resource["cpu-load"]; ok {
			fmt.Sscanf(cpuLoad, "%f", &metrics.CPUUsage)
		}

		// Parse RAM usage (in bytes)
		if totalMem, ok := resource["total-memory"]; ok {
			var total int64
			fmt.Sscanf(totalMem, "%d", &total)
			metrics.RAMTotal = float64(total) / (1024 * 1024) // Convert to MB
		}

		if freeMem, ok := resource["free-memory"]; ok {
			var free int64
			fmt.Sscanf(freeMem, "%d", &free)
			usedMem := metrics.RAMTotal - (float64(free) / (1024 * 1024))
			metrics.RAMUsage = usedMem
		}
	}

	return metrics, nil
}
