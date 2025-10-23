package services

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"nat-management-app/internal/models"

	"github.com/go-routeros/routeros"
	"github.com/sirupsen/logrus"
)

// CachedData represents cached response with timestamp
type CachedData struct {
	Data      interface{}
	Timestamp time.Time
}

// NATService handles NAT management operations
type NATService struct {
	logger        *logrus.Logger
	mutex         sync.RWMutex
	routers       map[string]models.NATRouterConfig
	routerService RouterServiceInterface
	// 🔥 CACHE OPTIMIZATION: Response caching untuk faster subsequent requests
	configsCache  *CachedData
	clientsCache  *CachedData
	testCache     *CachedData
	cacheMutex    sync.RWMutex
	cacheTTL      time.Duration
}

// NewNATService creates a new NAT service instance with dynamic router loading
func NewNATService(logger *logrus.Logger, routerService RouterServiceInterface) *NATService {
	service := &NATService{
		logger:        logger,
		routerService: routerService,
		routers:       make(map[string]models.NATRouterConfig),
		cacheTTL:      30 * time.Second, // 🔥 Cache for 30 seconds
	}

	// Load router configurations from dynamic storage
	if err := service.loadRoutersFromDynamicStorage(); err != nil {
		logger.Errorf("Failed to load routers from dynamic storage: %v", err)
		// Fall back to empty routers map - admin can add routers via UI
		service.routers = make(map[string]models.NATRouterConfig)
	}

	return service
}

// loadRoutersFromDynamicStorage loads router configurations from RouterService
func (ns *NATService) loadRoutersFromDynamicStorage() error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// Get routers from RouterService in NAT format
	dynamicRouters, err := ns.routerService.GetRoutersForNATService()
	if err != nil {
		return fmt.Errorf("failed to get routers from RouterService: %v", err)
	}

	// Update the routers map
	ns.routers = dynamicRouters
	ns.logger.Infof("✅ Loaded %d routers from dynamic storage", len(ns.routers))

	return nil
}

// ReloadRouters reloads router configurations from storage
func (ns *NATService) ReloadRouters() error {
	ns.logger.Infof("🔄 Reloading router configurations...")

	// Reload RouterService configuration first
	if err := ns.routerService.ReloadConfiguration(); err != nil {
		return fmt.Errorf("failed to reload RouterService configuration: %v", err)
	}

	// Load updated router configurations
	return ns.loadRoutersFromDynamicStorage()
}

// ConnectRouter establishes connection to a specific router with retry logic
func (ns *NATService) ConnectRouter(routerName string) (*routeros.Client, error) {
	config, exists := ns.routers[routerName]
	if !exists {
		return nil, fmt.Errorf("router %s not configured", routerName)
	}

	// ⚡ OPTIMIZED: Reduced retries and timeout for faster response
	maxRetries := 2 // Reduced from 3 to 2
	baseTimeout := 8 * time.Second // Reduced from 15s to 8s
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		timeout := baseTimeout * time.Duration(attempt)

		ns.logger.Debugf("🔄 Attempt %d/%d: Connecting to %s at %s:%d (timeout: %v)",
			attempt, maxRetries, routerName, config.Host, config.Port, timeout)

		// Test TCP connection first
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port), timeout)
		if err != nil {
			lastErr = err
			ns.logger.Warnf("⚠️  Attempt %d: TCP connection to %s failed: %v", attempt, routerName, err)

			if attempt < maxRetries {
				backoff := time.Duration(attempt) * 1 * time.Second // Reduced from 2s to 1s
				ns.logger.Debugf("⏳ Waiting %v before retry...", backoff)
				time.Sleep(backoff)
				continue
			}

			return nil, fmt.Errorf("failed to connect to %s after %d attempts: %v", routerName, maxRetries, lastErr)
		}
		conn.Close()

		// Try RouterOS API connection
		client, err := routeros.Dial(fmt.Sprintf("%s:%d", config.Host, config.Port), config.Username, config.Password)
		if err != nil {
			lastErr = err
			ns.logger.Warnf("⚠️  Attempt %d: RouterOS API auth to %s failed: %v", attempt, routerName, err)

			if attempt < maxRetries {
				backoff := time.Duration(attempt) * 1 * time.Second // Reduced from 2s to 1s
				ns.logger.Debugf("⏳ Waiting %v before retry...", backoff)
				time.Sleep(backoff)
				continue
			}

			return nil, fmt.Errorf("RouterOS API auth to %s failed after %d attempts: %v", routerName, maxRetries, lastErr)
		}

		ns.logger.Infof("✅ Successfully connected to %s on attempt %d", routerName, attempt)
		return client, nil
	}

	return nil, fmt.Errorf("unexpected error: failed to connect to %s", routerName)
}

// GetONTNATRule retrieves the ONT NAT rule with comment 'REMOTE ONT PELANGGAN'
func (ns *NATService) GetONTNATRule(routerName string) (*models.ONTNATRule, error) {
	client, err := ns.ConnectRouter(routerName)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Get firewall NAT rules
	reply, err := client.Run("/ip/firewall/nat/print", "=.proplist=.id,chain,action,src-address,dst-address,src-port,dst-port,to-addresses,to-ports,protocol,comment,disabled,bytes,packets")
	if err != nil {
		return nil, fmt.Errorf("failed to get NAT rules: %v", err)
	}

	// Find rule with comment "REMOTE ONT PELANGGAN"
	for _, re := range reply.Re {
		comment := re.Map["comment"]
		if strings.Contains(strings.ToUpper(comment), "REMOTE ONT PELANGGAN") {
			rule := &models.ONTNATRule{
				Router:         routerName,
				ID:             re.Map[".id"],
				Chain:          re.Map["chain"],
				Action:         re.Map["action"],
				SrcAddress:     re.Map["src-address"],
				DstAddress:     re.Map["dst-address"],
				SrcPort:        re.Map["src-port"],
				DstPort:        re.Map["dst-port"],
				ToAddresses:    re.Map["to-addresses"],
				ToPorts:        re.Map["to-ports"],
				Protocol:       re.Map["protocol"],
				Comment:        re.Map["comment"],
				Disabled:       re.Map["disabled"] == "true",
				TunnelEndpoint: ns.routers[routerName].TunnelEndpoint,
				PublicONTURL:   ns.routers[routerName].PublicONTURL,
			}

			// Parse bytes and packets
			if bytes, err := strconv.ParseInt(re.Map["bytes"], 10, 64); err == nil {
				rule.Bytes = bytes
			}
			if packets, err := strconv.ParseInt(re.Map["packets"], 10, 64); err == nil {
				rule.Packets = packets
			}

			return rule, nil
		}
	}

	return nil, fmt.Errorf("ONT NAT rule not found")
}

// UpdateONTNATRule updates the ONT NAT rule with new IP and port
// IMPORTANT: Only updates to-addresses, does NOT create new NAT rule
func (ns *NATService) UpdateONTNATRule(routerName, newIP, newPort string) error {
	if !ns.validateIP(newIP) {
		return fmt.Errorf("invalid IP address: %s", newIP)
	}

	if !ns.validatePort(newPort) {
		return fmt.Errorf("invalid port: %s", newPort)
	}

	// Get current rule
	currentRule, err := ns.GetONTNATRule(routerName)
	if err != nil {
		return fmt.Errorf("ONT NAT rule not found in %s: %v", routerName, err)
	}

	// Connect and update
	client, err := ns.ConnectRouter(routerName)
	if err != nil {
		return err
	}
	defer client.Close()

	// Update existing NAT rule - only change to-addresses and to-ports
	_, err = client.Run("/ip/firewall/nat/set", "=.id="+currentRule.ID, "=to-addresses="+newIP, "=to-ports="+newPort)
	if err != nil {
		return fmt.Errorf("failed to update NAT rule: %v", err)
	}

	// 🔥 Invalidate cache after update
	ns.invalidateCache()

	ns.logger.Infof("✓ ONT NAT rule updated in %s: %s:%s", routerName, newIP, newPort)
	return nil
}

// GetAllONTConfigs retrieves ONT NAT configurations from all routers
// ⚡ OPTIMIZED: Parallel execution with goroutines for faster response
// 🔥 CACHE OPTIMIZATION: Return cached data if still fresh (30s TTL)
func (ns *NATService) GetAllONTConfigs() map[string]models.ONTConfig {
	// Check cache first
	ns.cacheMutex.RLock()
	if ns.configsCache != nil && time.Since(ns.configsCache.Timestamp) < ns.cacheTTL {
		cached := ns.configsCache.Data.(map[string]models.ONTConfig)
		ns.cacheMutex.RUnlock()
		ns.logger.Debugf("⚡ Returning cached ONT configs (age: %v)", time.Since(ns.configsCache.Timestamp))
		return cached
	}
	ns.cacheMutex.RUnlock()

	// Cache miss or expired - fetch fresh data
	configs := make(map[string]models.ONTConfig)
	var mu sync.Mutex
	var wg sync.WaitGroup

	ns.logger.Debugf("🚀 Starting parallel ONT config fetch for %d routers", len(ns.routers))
	startTime := time.Now()

	for routerName := range ns.routers {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			rule, err := ns.GetONTNATRule(name)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				configs[name] = models.ONTConfig{
					Found:   false,
					Error:   err.Error(),
					Message: "ONT NAT rule not found",
				}
				return
			}

			status := "enabled"
			if rule.Disabled {
				status = "disabled"
			}

			configs[name] = models.ONTConfig{
				Found:          true,
				CurrentIP:      rule.ToAddresses,
				CurrentPort:    rule.ToPorts,
				DstAddress:     rule.DstAddress,
				DstPort:        rule.DstPort,
				Protocol:       rule.Protocol,
				Status:         status,
				Comment:        rule.Comment,
				TunnelEndpoint: rule.TunnelEndpoint,
				PublicONTURL:   rule.PublicONTURL,
				Bytes:          rule.Bytes,
				Packets:        rule.Packets,
			}
		}(routerName)
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	ns.logger.Infof("✅ Parallel ONT config fetch completed in %v for %d routers", elapsed, len(configs))

	// Update cache
	ns.cacheMutex.Lock()
	ns.configsCache = &CachedData{
		Data:      configs,
		Timestamp: time.Now(),
	}
	ns.cacheMutex.Unlock()

	return configs
}

// GetRouterClients retrieves online clients from a specific router
func (ns *NATService) GetRouterClients(routerName string) ([]models.NATClient, error) {
	client, err := ns.ConnectRouter(routerName)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// Get PPPoE active connections
	reply, err := client.Run("/ppp/active/print", "=.proplist=name,address,caller-id,uptime,encoding")
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %v", err)
	}

	var clients []models.NATClient
	for _, re := range reply.Re {
		client := models.NATClient{
			Router:    routerName,
			Username:  re.Map["name"],
			IPAddress: re.Map["address"],
			CallerID:  re.Map["caller-id"],
			Uptime:    re.Map["uptime"],
			Encoding:  re.Map["encoding"],
		}
		clients = append(clients, client)
	}

	ns.logger.Infof("Retrieved %d clients from %s", len(clients), routerName)
	return clients, nil
}

// GetAllClients retrieves online clients from all routers
// ⚡ OPTIMIZED: Parallel execution with goroutines for faster response
// 🔥 CACHE OPTIMIZATION: Return cached data if still fresh (30s TTL)
func (ns *NATService) GetAllClients() map[string][]models.NATClient {
	// Check cache first
	ns.cacheMutex.RLock()
	if ns.clientsCache != nil && time.Since(ns.clientsCache.Timestamp) < ns.cacheTTL {
		cached := ns.clientsCache.Data.(map[string][]models.NATClient)
		ns.cacheMutex.RUnlock()
		ns.logger.Debugf("⚡ Returning cached clients (age: %v)", time.Since(ns.clientsCache.Timestamp))
		return cached
	}
	ns.cacheMutex.RUnlock()

	// Cache miss or expired - fetch fresh data
	allClients := make(map[string][]models.NATClient)
	var mu sync.Mutex
	var wg sync.WaitGroup

	ns.logger.Debugf("🚀 Starting parallel client fetch for %d routers", len(ns.routers))
	startTime := time.Now()

	for routerName := range ns.routers {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			clients, err := ns.GetRouterClients(name)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				ns.logger.Errorf("Failed to get clients from %s: %v", name, err)
				allClients[name] = []models.NATClient{}
				return
			}
			allClients[name] = clients
		}(routerName)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	totalClients := 0
	for _, clients := range allClients {
		totalClients += len(clients)
	}

	ns.logger.Infof("✅ Parallel client fetch completed in %v: %d clients from %d routers", elapsed, totalClients, len(allClients))

	// Update cache
	ns.cacheMutex.Lock()
	ns.clientsCache = &CachedData{
		Data:      allClients,
		Timestamp: time.Now(),
	}
	ns.cacheMutex.Unlock()

	return allClients
}

// TestRouterConnection tests connection to a specific router
func (ns *NATService) TestRouterConnection(routerName string) models.RouterConnectionTest {
	client, err := ns.ConnectRouter(routerName)
	if err != nil {
		return models.RouterConnectionTest{
			Status:    "disconnected",
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
	}
	defer client.Close()

	// Get system info
	identityReply, err := client.Run("/system/identity/print")
	if err != nil {
		return models.RouterConnectionTest{
			Status:    "disconnected",
			Message:   fmt.Sprintf("Failed to get system identity: %v", err),
			Timestamp: time.Now(),
		}
	}

	resourceReply, err := client.Run("/system/resource/print")
	if err != nil {
		return models.RouterConnectionTest{
			Status:    "disconnected",
			Message:   fmt.Sprintf("Failed to get system resource: %v", err),
			Timestamp: time.Now(),
		}
	}

	var routerName_actual, version, board string
	if len(identityReply.Re) > 0 {
		routerName_actual = identityReply.Re[0].Map["name"]
	}
	if len(resourceReply.Re) > 0 {
		version = resourceReply.Re[0].Map["version"]
		board = resourceReply.Re[0].Map["board-name"]
	}

	return models.RouterConnectionTest{
		Status:     "connected",
		RouterName: routerName_actual,
		Version:    version,
		Board:      board,
		Message:    "Connection successful",
		Timestamp:  time.Now(),
	}
}

// TestAllConnections tests connections to all routers
// ⚡ OPTIMIZED: Parallel execution with goroutines for faster response
// 🔥 CACHE OPTIMIZATION: Return cached data if still fresh (30s TTL)
func (ns *NATService) TestAllConnections() map[string]models.RouterConnectionTest {
	// Check cache first
	ns.cacheMutex.RLock()
	if ns.testCache != nil && time.Since(ns.testCache.Timestamp) < ns.cacheTTL {
		cached := ns.testCache.Data.(map[string]models.RouterConnectionTest)
		ns.cacheMutex.RUnlock()
		ns.logger.Debugf("⚡ Returning cached connection test (age: %v)", time.Since(ns.testCache.Timestamp))
		return cached
	}
	ns.cacheMutex.RUnlock()

	// Cache miss or expired - test connections
	results := make(map[string]models.RouterConnectionTest)
	var mu sync.Mutex
	var wg sync.WaitGroup

	ns.logger.Debugf("🚀 Starting parallel connection test for %d routers", len(ns.routers))
	startTime := time.Now()

	for routerName := range ns.routers {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			result := ns.TestRouterConnection(name)

			mu.Lock()
			defer mu.Unlock()
			results[name] = result
		}(routerName)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	connectedCount := 0
	for _, result := range results {
		if result.Status == "connected" {
			connectedCount++
		}
	}

	ns.logger.Infof("✅ Parallel connection test completed in %v: %d/%d routers connected", elapsed, connectedCount, len(results))

	// Update cache
	ns.cacheMutex.Lock()
	ns.testCache = &CachedData{
		Data:      results,
		Timestamp: time.Now(),
	}
	ns.cacheMutex.Unlock()

	return results
}

// validateIP validates IP address format
func (ns *NATService) validateIP(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}
	for _, part := range parts {
		if num, err := strconv.Atoi(part); err != nil || num < 0 || num > 255 {
			return false
		}
	}
	return true
}

// validatePort validates port number
func (ns *NATService) validatePort(port string) bool {
	if portNum, err := strconv.Atoi(port); err != nil || portNum < 1 || portNum > 65535 {
		return false
	}
	return true
}

// CheckPPPoEStatus checks if a specific PPPoE username is online across all routers or specific router
// CheckPPPoEStatusWithRouterFilter checks PPPoE status only on allowed routers
func (ns *NATService) CheckPPPoEStatusWithRouterFilter(username string, allowedRouters []string, testConnectivity bool) *models.PPPoEStatusResponse {
	return ns.checkPPPoEStatusInternal(username, "", allowedRouters, testConnectivity)
}

func (ns *NATService) CheckPPPoEStatus(username string, testConnectivity bool, specificRouter ...string) *models.PPPoEStatusResponse {
	if len(specificRouter) > 0 && specificRouter[0] != "" {
		return ns.checkPPPoEStatusInternal(username, specificRouter[0], nil, testConnectivity)
	}
	// Check all routers (no filtering)
	return ns.checkPPPoEStatusInternal(username, "", nil, testConnectivity)
}

// checkPPPoEStatusInternal is the internal implementation that supports router filtering
func (ns *NATService) checkPPPoEStatusInternal(username, specificRouter string, allowedRouters []string, testConnectivity bool) *models.PPPoEStatusResponse {
	response := &models.PPPoEStatusResponse{
		Status:      "success",
		Username:    username,
		IsOnline:    false,
		OnlineCount: 0,
		Data:        make(map[string]models.PPPoEStatusResult),
		Timestamp:   time.Now(),
	}

	if username == "" {
		response.Status = "error"
		response.Message = "Username tidak boleh kosong"
		return response
	}

	// Determine which routers to check
	var routersToCheck []string
	if specificRouter != "" {
		// Check specific router only
		if _, exists := ns.routers[specificRouter]; !exists {
			response.Status = "error"
			response.Message = fmt.Sprintf("Router %s tidak ditemukan", specificRouter)
			return response
		}
		routersToCheck = []string{specificRouter}
	} else if allowedRouters != nil {
		// Check only allowed routers (role-based filtering)
		for _, routerName := range allowedRouters {
			if _, exists := ns.routers[routerName]; exists {
				routersToCheck = append(routersToCheck, routerName)
			}
		}
	} else {
		// Check all routers
		for routerName := range ns.routers {
			routersToCheck = append(routersToCheck, routerName)
		}
	}

	// Check specified routers
	for _, routerName := range routersToCheck {
		result := ns.checkPPPoEOnRouterWithConnectivity(routerName, username, testConnectivity)
		response.Data[routerName] = result

		if result.IsOnline {
			response.IsOnline = true
			response.OnlineCount++
		}
	}

	if response.IsOnline {
		if len(routersToCheck) == 1 {
			response.Message = fmt.Sprintf("User %s ditemukan online di router %s", username, routersToCheck[0])
		} else {
			response.Message = fmt.Sprintf("User %s ditemukan online di %d router", username, response.OnlineCount)
		}
	} else {
		if len(routersToCheck) == 1 {
			response.Message = fmt.Sprintf("User %s tidak ditemukan online di router %s", username, routersToCheck[0])
		} else {
			response.Message = fmt.Sprintf("User %s tidak ditemukan online di router yang dapat diakses", username)
		}
	}

	ns.logger.Infof("PPPoE Status Check: %s - Online: %t (%d routers)", username, response.IsOnline, response.OnlineCount)
	return response
}

// checkPPPoEOnRouter checks PPPoE status on a specific router
func (ns *NATService) checkPPPoEOnRouter(routerName, username string) models.PPPoEStatusResult {
	return ns.checkPPPoEOnRouterWithConnectivity(routerName, username, false)
}

// testDeviceConnectivity tests if the device at given IP is actually reachable via TCP
func (ns *NATService) testDeviceConnectivity(ipAddress string) (bool, string, time.Duration) {
	// Common ports for ONT/modem/customer devices
	testPorts := []string{"80", "8080", "443", "22", "23", "8081"}

	startTime := time.Now()

	for _, port := range testPorts {
		address := fmt.Sprintf("%s:%s", ipAddress, port)

		// Test TCP connection with 2-second timeout per port
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			conn.Close()
			duration := time.Since(startTime)
			ns.logger.Debugf("✅ Device %s reachable on port %s (took %v)", ipAddress, port, duration)
			return true, port, duration
		}

		ns.logger.Debugf("⚠️  Port %s on %s unreachable: %v", port, ipAddress, err)
	}

	duration := time.Since(startTime)
	ns.logger.Warnf("❌ Device %s unreachable on all tested ports (took %v)", ipAddress, duration)
	return false, "", duration
}

// checkPPPoEOnRouterWithConnectivity checks PPPoE status with optional connectivity test
func (ns *NATService) checkPPPoEOnRouterWithConnectivity(routerName, username string, testConnectivity bool) models.PPPoEStatusResult {
	result := models.PPPoEStatusResult{
		Router:             routerName,
		IsOnline:           false,
		ConnectivityStatus: "not_tested",
		LastSeen:           time.Now(),
	}

	client, err := ns.ConnectRouter(routerName)
	if err != nil {
		result.Message = fmt.Sprintf("Gagal koneksi ke router: %v", err)
		return result
	}
	defer client.Close()

	// Get PPPoE active connections
	reply, err := client.Run("/ppp/active/print", "=.proplist=name,address,caller-id,uptime,encoding", fmt.Sprintf("?name=%s", username))
	if err != nil {
		result.Message = fmt.Sprintf("Gagal mengambil data PPPoE: %v", err)
		return result
	}

	// Check if user is found in active connections
	for _, re := range reply.Re {
		if re.Map["name"] == username {
			result.IsOnline = true
			result.IPAddress = re.Map["address"]
			result.CallerID = re.Map["caller-id"]
			result.Uptime = re.Map["uptime"]
			result.Encoding = re.Map["encoding"]
			result.SessionTime = re.Map["uptime"] // Same as uptime for active sessions
			result.Message = "User aktif (PPPoE session active)"

			// Perform connectivity test if requested
			if testConnectivity && result.IPAddress != "" {
				ns.logger.Infof("🔍 Testing device connectivity for %s at %s", username, result.IPAddress)
				reachable, port, duration := ns.testDeviceConnectivity(result.IPAddress)

				result.DeviceReachable = reachable
				result.ReachablePort = port
				result.ConnectivityTime = duration

				if reachable {
					result.ConnectivityStatus = "reachable"
					result.Message = fmt.Sprintf("User aktif dan device reachable (port %s, %dms)", port, duration.Milliseconds())
				} else {
					result.ConnectivityStatus = "unreachable"
					result.Message = fmt.Sprintf("User aktif tapi device tidak merespon (tested %dms)", duration.Milliseconds())
				}
			}

			break
		}
	}

	if !result.IsOnline {
		result.Message = "User tidak ditemukan di router ini"
	}

	return result
}

// GetPPPoEHistory gets recent PPPoE searches (placeholder for future implementation)
func (ns *NATService) GetPPPoEHistory(userID int, limit int) []models.PPPoESearchHistory {
	// Placeholder - could be implemented with database storage later
	return []models.PPPoESearchHistory{}
}

// CheckMultiplePPPoEStatus checks status for multiple usernames
func (ns *NATService) CheckMultiplePPPoEStatus(usernames []string) map[string]*models.PPPoEStatusResponse {
	results := make(map[string]*models.PPPoEStatusResponse)

	for _, username := range usernames {
		results[username] = ns.CheckPPPoEStatus(username, false) // No connectivity test for bulk checks
	}

	return results
}

// GetAvailableRouters returns list of available router names
func (ns *NATService) GetAvailableRouters() []string {
	ns.mutex.RLock()
	defer ns.mutex.RUnlock()

	var routers []string
	for routerName := range ns.routers {
		routers = append(routers, routerName)
	}
	return routers
}

// GetAvailableRoutersWithFilter returns list of router names accessible to a user role
func (ns *NATService) GetAvailableRoutersWithFilter(userRole string) []string {
	if ns.routerService == nil {
		// Fallback to all routers if RouterService not available
		return ns.GetAvailableRouters()
	}

	// Get allowed routers from RouterService access control
	allRouters, err := ns.routerService.GetAllRouters(userRole)
	if err != nil {
		ns.logger.Errorf("Failed to get filtered routers for role %s: %v", userRole, err)
		return []string{} // Return empty list on error
	}

	var routerNames []string
	for _, router := range allRouters {
		// Only include routers that are enabled and exist in our routers map
		if router.Enabled {
			if _, exists := ns.routers[router.Name]; exists {
				routerNames = append(routerNames, router.Name)
			}
		}
	}

	return routerNames
}

// RefreshRouterIfNeeded checks if router configuration needs refreshing and does so if necessary
func (ns *NATService) RefreshRouterIfNeeded() error {
	// This can be called before operations to ensure we have the latest router configurations
	// For now, we'll implement a simple approach - in production this could be optimized with caching
	if len(ns.routers) == 0 {
		ns.logger.Infof("🔄 No routers loaded, refreshing from storage...")
		return ns.loadRoutersFromDynamicStorage()
	}
	return nil
}

// FuzzySearchPPPoEWithRouterFilter performs fuzzy search with router access filtering
// ⚡ OPTIMIZED: Parallel execution for faster multi-router search
func (ns *NATService) FuzzySearchPPPoEWithRouterFilter(searchTerm string, specificRouter string, limit int, allowedRouters []string) *models.PPPoEFuzzySearchResponse {
	response := &models.PPPoEFuzzySearchResponse{
		Status:     "success",
		SearchTerm: searchTerm,
		Matches:    []models.PPPoEFuzzyMatch{},
		Timestamp:  time.Now(),
	}

	if searchTerm == "" {
		response.Status = "error"
		response.Message = "Search term tidak boleh kosong"
		return response
	}

	if limit <= 0 {
		limit = 5
	}

	// Determine which routers to search
	var routersToSearch []string
	if specificRouter != "" {
		// Check if specific router is allowed
		hasAccess := false
		for _, allowed := range allowedRouters {
			if specificRouter == allowed {
				hasAccess = true
				break
			}
		}
		if hasAccess {
			if _, exists := ns.routers[specificRouter]; exists {
				routersToSearch = []string{specificRouter}
			}
		}
	} else {
		// Search only allowed routers
		for _, routerName := range allowedRouters {
			if _, exists := ns.routers[routerName]; exists {
				routersToSearch = append(routersToSearch, routerName)
			}
		}
	}

	// ⚡ PARALLEL SEARCH: Search in each allowed router concurrently
	var allMatches []models.PPPoEFuzzyMatch
	var mu sync.Mutex
	var wg sync.WaitGroup

	ns.logger.Debugf("🚀 Starting parallel fuzzy search in %d routers", len(routersToSearch))
	startTime := time.Now()

	for _, routerName := range routersToSearch {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			matches := ns.searchPPPoEInRouter(name, searchTerm)

			mu.Lock()
			allMatches = append(allMatches, matches...)
			mu.Unlock()
		}(routerName)
	}

	wg.Wait()
	elapsed := time.Since(startTime)

	// Sort matches by similarity score (highest first)
	ns.sortMatchesBySimilarity(allMatches)

	// Limit results
	if len(allMatches) > limit {
		allMatches = allMatches[:limit]
	}

	response.Matches = allMatches
	response.MatchCount = len(allMatches)

	if len(allMatches) > 0 {
		response.Message = fmt.Sprintf("Ditemukan %d username serupa dengan '%s'", len(allMatches), searchTerm)
	} else {
		response.Message = fmt.Sprintf("Tidak ditemukan username serupa dengan '%s' di router yang dapat diakses", searchTerm)
	}

	ns.logger.Infof("✅ Parallel fuzzy search completed in %v: '%s' - Found %d matches in %d routers", elapsed, searchTerm, len(allMatches), len(routersToSearch))
	return response
}

// FuzzySearchPPPoE performs fuzzy search for similar PPPoE usernames
func (ns *NATService) FuzzySearchPPPoE(searchTerm string, specificRouter string, limit int) *models.PPPoEFuzzySearchResponse {
	response := &models.PPPoEFuzzySearchResponse{
		Status:     "success",
		SearchTerm: searchTerm,
		Matches:    []models.PPPoEFuzzyMatch{},
		Timestamp:  time.Now(),
	}

	if searchTerm == "" {
		response.Status = "error"
		response.Message = "Search term tidak boleh kosong"
		return response
	}

	if limit <= 0 || limit > 10 {
		limit = 5 // Default limit
	}

	// Determine which routers to search
	var routersToSearch []string
	if specificRouter != "" {
		if _, exists := ns.routers[specificRouter]; !exists {
			response.Status = "error"
			response.Message = fmt.Sprintf("Router %s tidak ditemukan", specificRouter)
			return response
		}
		routersToSearch = []string{specificRouter}
	} else {
		for routerName := range ns.routers {
			routersToSearch = append(routersToSearch, routerName)
		}
	}

	// Collect all matches from all routers
	var allMatches []models.PPPoEFuzzyMatch

	for _, routerName := range routersToSearch {
		matches := ns.searchPPPoEInRouter(routerName, searchTerm)
		allMatches = append(allMatches, matches...)
	}

	// Sort by similarity score (highest first)
	ns.sortMatchesBySimilarity(allMatches)

	// Limit results
	if len(allMatches) > limit {
		allMatches = allMatches[:limit]
	}

	response.Matches = allMatches
	response.MatchCount = len(allMatches)

	if response.MatchCount > 0 {
		response.Message = fmt.Sprintf("Ditemukan %d username serupa dengan '%s'", response.MatchCount, searchTerm)
	} else {
		response.Message = fmt.Sprintf("Tidak ditemukan username yang serupa dengan '%s'", searchTerm)
	}

	ns.logger.Infof("PPPoE Fuzzy Search: %s - Found: %d matches", searchTerm, response.MatchCount)
	return response
}

// searchPPPoEInRouter searches for similar usernames in a specific router
func (ns *NATService) searchPPPoEInRouter(routerName, searchTerm string) []models.PPPoEFuzzyMatch {
	var matches []models.PPPoEFuzzyMatch

	client, err := ns.ConnectRouter(routerName)
	if err != nil {
		ns.logger.Errorf("Failed to connect to router %s for fuzzy search: %v", routerName, err)
		return matches
	}
	defer client.Close()

	// Get all active PPPoE connections
	activeReply, err := client.Run("/ppp/active/print", "=.proplist=name,address,caller-id,uptime,encoding,service")
	if err != nil {
		ns.logger.Errorf("Failed to get PPPoE active connections from %s: %v", routerName, err)
		return matches
	}

	// Get PPPoE secrets to get profile names
	secretsReply, err := client.Run("/ppp/secret/print", "=.proplist=name,profile")
	if err != nil {
		ns.logger.Errorf("Failed to get PPPoE secrets from %s: %v", routerName, err)
		// Continue with active connections only if secrets fail
		secretsReply = &routeros.Reply{}
	}

	// Build a map of username -> profile from secrets
	profileMap := make(map[string]string)
	for _, secretRe := range secretsReply.Re {
		username := secretRe.Map["name"]
		profile := secretRe.Map["profile"]
		if username != "" && profile != "" {
			profileMap[username] = profile
		}
	}

	// Process each active connection and calculate similarity
	for _, re := range activeReply.Re {
		username := re.Map["name"]
		if username == "" {
			continue
		}

		similarity := ns.calculateSimilarity(searchTerm, username)
		
		// Only include if similarity is above threshold (0.3 = 30%)
		if similarity >= 0.3 {
			// Get profile from secrets map, fallback to service field or default
			profile := profileMap[username]
			if profile == "" {
				profile = re.Map["service"]
				if profile == "" {
					profile = "default"
				}
			}
			
			match := models.PPPoEFuzzyMatch{
				Username:   username,
				Router:     routerName,
				IPAddress:  re.Map["address"],
				CallerID:   re.Map["caller-id"],
				Uptime:     re.Map["uptime"],
				Profile:    profile,
				Similarity: similarity,
				IsOnline:   true,
			}
			matches = append(matches, match)
		}
	}

	return matches
}

// calculateSimilarity calculates similarity between two strings using multiple algorithms
func (ns *NATService) calculateSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if s1 == s2 {
		return 1.0
	}

	// Combine multiple similarity algorithms for better matching
	
	// 1. Exact substring match (highest weight for prefix/suffix matching)
	substringScore := ns.substringMatchScore(s1, s2)
	
	// 2. Levenshtein distance based similarity
	levenshteinScore := ns.levenshteinSimilarity(s1, s2)
	
	// 3. Common subsequence similarity
	lcsScore := ns.longestCommonSubsequenceScore(s1, s2)
	
	// 4. Wildcard pattern matching (for area-based names like ahmadkukun, budikukun)
	patternScore := ns.patternMatchScore(s1, s2)

	// Weighted combination (prioritize pattern matching for area-based names)
	finalScore := (substringScore * 0.3) + (levenshteinScore * 0.2) + (lcsScore * 0.2) + (patternScore * 0.3)
	
	return finalScore
}

// substringMatchScore calculates score based on substring matching
func (ns *NATService) substringMatchScore(s1, s2 string) float64 {
	if strings.Contains(s1, s2) || strings.Contains(s2, s1) {
		shorter := len(s1)
		longer := len(s2)
		if shorter > longer {
			shorter, longer = longer, shorter
		}
		return float64(shorter) / float64(longer)
	}
	return 0.0
}

// levenshteinSimilarity calculates similarity using Levenshtein distance
func (ns *NATService) levenshteinSimilarity(s1, s2 string) float64 {
	distance := ns.levenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - (float64(distance) / float64(maxLen))
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (ns *NATService) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			deletion := matrix[i-1][j] + 1
			insertion := matrix[i][j-1] + 1
			substitution := matrix[i-1][j-1] + cost
			matrix[i][j] = min(deletion, min(insertion, substitution))
		}
	}

	return matrix[len(s1)][len(s2)]
}

// longestCommonSubsequenceScore calculates similarity based on LCS
func (ns *NATService) longestCommonSubsequenceScore(s1, s2 string) float64 {
	lcsLen := ns.longestCommonSubsequence(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}
	if maxLen == 0 {
		return 1.0
	}
	return float64(lcsLen) / float64(maxLen)
}

// longestCommonSubsequence calculates the length of the LCS
func (ns *NATService) longestCommonSubsequence(s1, s2 string) int {
	m, n := len(s1), len(s2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

// patternMatchScore calculates score for area-based name patterns (e.g., ahmadkukun, budikukun)
func (ns *NATService) patternMatchScore(s1, s2 string) float64 {
	// Extract potential area names (common suffixes/prefixes)
	commonAreas := []string{
		"kukun", "cipanas", "sukatani", "darussalam", "samsat", "cikarang", 
		"sukawangi", "jaya", "lane4", "lane", "bt", "pk", "kp",
	}

	score := 0.0
	
	// Check if both strings contain the same area pattern
	for _, area := range commonAreas {
		s1HasArea := strings.Contains(s1, area)
		s2HasArea := strings.Contains(s2, area)
		
		if s1HasArea && s2HasArea {
			// Both contain the same area, give high score
			score += 0.8
			
			// Additional score for similar name part (before/after area)
			s1WithoutArea := strings.ReplaceAll(s1, area, "")
			s2WithoutArea := strings.ReplaceAll(s2, area, "")
			
			if len(s1WithoutArea) > 0 && len(s2WithoutArea) > 0 {
				namePartSimilarity := ns.levenshteinSimilarity(s1WithoutArea, s2WithoutArea)
				score += namePartSimilarity * 0.2
			}
			
			break // Found matching area, no need to check others
		}
	}

	// Cap the score at 1.0
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// sortMatchesBySimilarity sorts matches by similarity score (descending)
func (ns *NATService) sortMatchesBySimilarity(matches []models.PPPoEFuzzyMatch) {
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].Similarity < matches[j].Similarity {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// invalidateCache clears all caches when data is modified
// 🔥 CACHE OPTIMIZATION: Force refresh after updates
func (ns *NATService) invalidateCache() {
	ns.cacheMutex.Lock()
	defer ns.cacheMutex.Unlock()

	ns.configsCache = nil
	ns.clientsCache = nil
	ns.testCache = nil

	ns.logger.Debug("🔥 Cache invalidated - fresh data will be fetched on next request")
}
