package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-routeros/routeros"
	"github.com/sirupsen/logrus"
)

// RouterOSConnection wraps a RouterOS client with metadata
type RouterOSConnection struct {
	Client     *routeros.Client
	RouterName string
	LastUsed   time.Time
	InUse      bool
	Created    time.Time
}

// RouterOSConnectionPool manages a pool of RouterOS connections
type RouterOSConnectionPool struct {
	logger          *logrus.Logger
	connections     map[string][]*RouterOSConnection // RouterName -> Connections
	mu              sync.Mutex
	maxConnections  int           // Max connections per router
	idleTimeout     time.Duration // Idle connection timeout
	maxLifetime     time.Duration // Max connection lifetime
	cleanupInterval time.Duration // Cleanup interval
	stopCleanup     chan struct{}
}

// ConnectionConfig holds configuration for connection pool
type ConnectionConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// NewRouterOSConnectionPool creates a new connection pool
func NewRouterOSConnectionPool(logger *logrus.Logger, maxConnections int, idleTimeout, maxLifetime time.Duration) *RouterOSConnectionPool {
	pool := &RouterOSConnectionPool{
		logger:          logger,
		connections:     make(map[string][]*RouterOSConnection),
		maxConnections:  maxConnections,
		idleTimeout:     idleTimeout,
		maxLifetime:     maxLifetime,
		cleanupInterval: 30 * time.Second,
		stopCleanup:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go pool.cleanupRoutine()

	logger.Infof("🏊 RouterOS Connection Pool initialized (max: %d per router, idle timeout: %v, max lifetime: %v)",
		maxConnections, idleTimeout, maxLifetime)

	return pool
}

// GetConnection retrieves or creates a connection for a router
func (pool *RouterOSConnectionPool) GetConnection(routerName string, config ConnectionConfig) (*RouterOSConnection, error) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	// Try to find an idle connection
	if conns, exists := pool.connections[routerName]; exists {
		for _, conn := range conns {
			// Check if connection is idle and still healthy
			if !conn.InUse {
				// Check if connection is still alive
				if pool.isHealthy(conn) {
					conn.InUse = true
					conn.LastUsed = time.Now()
					pool.logger.Debugf("♻️ Reusing existing connection for router: %s", routerName)
					return conn, nil
				} else {
					// Connection is dead, remove it
					pool.logger.Warnf("⚠️ Found dead connection for %s, removing...", routerName)
					pool.removeConnection(routerName, conn)
				}
			}
		}
	}

	// No idle connection available, create new one if under limit
	if len(pool.connections[routerName]) >= pool.maxConnections {
		return nil, fmt.Errorf("connection pool limit reached for router %s (max: %d)", routerName, pool.maxConnections)
	}

	// Create new connection
	client, err := routeros.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port), config.Username, config.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RouterOS: %w", err)
	}

	conn := &RouterOSConnection{
		Client:     client,
		RouterName: routerName,
		LastUsed:   time.Now(),
		InUse:      true,
		Created:    time.Now(),
	}

	// Add to pool
	pool.connections[routerName] = append(pool.connections[routerName], conn)
	pool.logger.Infof("✅ Created new connection for router: %s (total: %d)", routerName, len(pool.connections[routerName]))

	return conn, nil
}

// ReleaseConnection marks a connection as idle and returns it to pool
func (pool *RouterOSConnectionPool) ReleaseConnection(conn *RouterOSConnection) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	conn.InUse = false
	conn.LastUsed = time.Now()
	pool.logger.Debugf("↩️ Released connection for router: %s", conn.RouterName)
}

// CloseConnection closes and removes a specific connection from pool
func (pool *RouterOSConnectionPool) CloseConnection(conn *RouterOSConnection) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if conn.Client != nil {
		conn.Client.Close()
	}
	pool.removeConnection(conn.RouterName, conn)
	pool.logger.Debugf("🔒 Closed connection for router: %s", conn.RouterName)
}

// removeConnection removes a connection from pool (must be called with lock held)
func (pool *RouterOSConnectionPool) removeConnection(routerName string, connToRemove *RouterOSConnection) {
	conns := pool.connections[routerName]
	for i, conn := range conns {
		if conn == connToRemove {
			// Close client if still open
			if conn.Client != nil {
				conn.Client.Close()
			}
			// Remove from slice
			pool.connections[routerName] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	// Clean up empty router entry
	if len(pool.connections[routerName]) == 0 {
		delete(pool.connections, routerName)
	}
}

// isHealthy checks if a connection is still healthy
func (pool *RouterOSConnectionPool) isHealthy(conn *RouterOSConnection) bool {
	// Check if connection exceeded max lifetime
	if time.Since(conn.Created) > pool.maxLifetime {
		pool.logger.Debugf("Connection for %s exceeded max lifetime", conn.RouterName)
		return false
	}

	// Try a simple command to verify connection is alive
	_, err := conn.Client.Run("/system/identity/print")
	if err != nil {
		pool.logger.Debugf("Health check failed for %s: %v", conn.RouterName, err)
		return false
	}

	return true
}

// cleanupRoutine periodically cleans up idle and stale connections
func (pool *RouterOSConnectionPool) cleanupRoutine() {
	ticker := time.NewTicker(pool.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pool.cleanup()
		case <-pool.stopCleanup:
			pool.logger.Info("🛑 Stopping connection pool cleanup routine")
			return
		}
	}
}

// cleanup removes idle and stale connections
func (pool *RouterOSConnectionPool) cleanup() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	now := time.Now()
	totalCleaned := 0

	for routerName, conns := range pool.connections {
		var keepConns []*RouterOSConnection

		for _, conn := range conns {
			shouldKeep := true

			// Remove connections that are idle for too long
			if !conn.InUse && now.Sub(conn.LastUsed) > pool.idleTimeout {
				pool.logger.Debugf("🧹 Cleaning up idle connection for %s (idle for %v)", routerName, now.Sub(conn.LastUsed))
				if conn.Client != nil {
					conn.Client.Close()
				}
				shouldKeep = false
				totalCleaned++
			}

			// Remove connections that exceeded max lifetime
			if now.Sub(conn.Created) > pool.maxLifetime {
				pool.logger.Debugf("🧹 Cleaning up old connection for %s (age: %v)", routerName, now.Sub(conn.Created))
				if conn.Client != nil {
					conn.Client.Close()
				}
				shouldKeep = false
				totalCleaned++
			}

			// Remove unhealthy connections that are not in use
			if !conn.InUse && !pool.isHealthy(conn) {
				pool.logger.Debugf("🧹 Cleaning up unhealthy connection for %s", routerName)
				shouldKeep = false
				totalCleaned++
			}

			if shouldKeep {
				keepConns = append(keepConns, conn)
			}
		}

		if len(keepConns) > 0 {
			pool.connections[routerName] = keepConns
		} else {
			delete(pool.connections, routerName)
		}
	}

	if totalCleaned > 0 {
		pool.logger.Infof("🧹 Connection pool cleanup: removed %d stale connections", totalCleaned)
	}
}

// Close shuts down the connection pool and closes all connections
func (pool *RouterOSConnectionPool) Close() {
	pool.logger.Info("🔒 Closing RouterOS connection pool...")

	// Stop cleanup routine
	close(pool.stopCleanup)

	pool.mu.Lock()
	defer pool.mu.Unlock()

	totalClosed := 0
	for _, conns := range pool.connections {
		for _, conn := range conns {
			if conn.Client != nil {
				conn.Client.Close()
				totalClosed++
			}
		}
	}

	// Clear all connections
	pool.connections = make(map[string][]*RouterOSConnection)

	pool.logger.Infof("✅ Connection pool closed (%d connections closed)", totalClosed)
}

// GetStats returns connection pool statistics
func (pool *RouterOSConnectionPool) GetStats() map[string]interface{} {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	totalConnections := 0
	activeConnections := 0
	idleConnections := 0
	routerStats := make(map[string]map[string]int)

	for routerName, conns := range pool.connections {
		routerActive := 0
		routerIdle := 0

		for _, conn := range conns {
			totalConnections++
			if conn.InUse {
				activeConnections++
				routerActive++
			} else {
				idleConnections++
				routerIdle++
			}
		}

		routerStats[routerName] = map[string]int{
			"total":  len(conns),
			"active": routerActive,
			"idle":   routerIdle,
		}
	}

	return map[string]interface{}{
		"total_connections":  totalConnections,
		"active_connections": activeConnections,
		"idle_connections":   idleConnections,
		"routers":            len(pool.connections),
		"router_stats":       routerStats,
	}
}
