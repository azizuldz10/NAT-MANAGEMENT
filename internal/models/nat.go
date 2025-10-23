package models

import "time"

// NATRouterConfig represents NAT-specific router configuration
type NATRouterConfig struct {
	Name           string `json:"name"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	TunnelEndpoint string `json:"tunnel_endpoint"`
	PublicONTURL   string `json:"public_ont_url"`
}

// ONTNATRule represents the specific ONT NAT rule data
type ONTNATRule struct {
	Router         string `json:"router"`
	ID             string `json:"id"`
	Chain          string `json:"chain"`
	Action         string `json:"action"`
	SrcAddress     string `json:"src_address"`
	DstAddress     string `json:"dst_address"`
	SrcPort        string `json:"src_port"`
	DstPort        string `json:"dst_port"`
	ToAddresses    string `json:"to_addresses"`
	ToPorts        string `json:"to_ports"`
	Protocol       string `json:"protocol"`
	Comment        string `json:"comment"`
	Disabled       bool   `json:"disabled"`
	Bytes          int64  `json:"bytes"`
	Packets        int64  `json:"packets"`
	TunnelEndpoint string `json:"tunnel_endpoint"`
	PublicONTURL   string `json:"public_ont_url"`
}

// ONTConfig represents current ONT configuration for a router
type ONTConfig struct {
	Found             bool   `json:"found"`
	CurrentIP         string `json:"current_ip,omitempty"`
	CurrentPort       string `json:"current_port,omitempty"`
	DstAddress        string `json:"dst_address,omitempty"`
	DstPort           string `json:"dst_port,omitempty"`
	Protocol          string `json:"protocol,omitempty"`
	Status            string `json:"status,omitempty"`
	Comment           string `json:"comment,omitempty"`
	TunnelEndpoint    string `json:"tunnel_endpoint,omitempty"`
	PublicONTURL      string `json:"public_ont_url,omitempty"`
	Bytes             int64  `json:"bytes,omitempty"`
	Packets           int64  `json:"packets,omitempty"`
	Error             string `json:"error,omitempty"`
	Message           string `json:"message,omitempty"`
}

// NATClient represents an online PPPoE client
type NATClient struct {
	Router    string `json:"router"`
	Username  string `json:"username"`
	IPAddress string `json:"ip_address"`
	CallerID  string `json:"caller_id"`
	Uptime    string `json:"uptime"`
	Encoding  string `json:"encoding"`
}

// NATUpdateRequest represents a request to update NAT rule
type NATUpdateRequest struct {
	Router string `json:"router" binding:"required"`
	IP     string `json:"ip" binding:"required"`
	Port   string `json:"port"`
}

// NATConfigsResponse represents the response for NAT configs API
type NATConfigsResponse struct {
	Status string               `json:"status"`
	Data   map[string]ONTConfig `json:"data"`
}

// NATClientsResponse represents the response for NAT clients API
type NATClientsResponse struct {
	Status string                     `json:"status"`
	Data   map[string][]NATClient     `json:"data"`
}

// NATUpdateResponse represents the response for NAT update API
type NATUpdateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// NATTestResponse represents the response for NAT test API
type NATTestResponse struct {
	Status string                           `json:"status"`
	Data   map[string]RouterConnectionTest `json:"data"`
}

// RouterConnectionTest represents router connection test result
type RouterConnectionTest struct {
	Status      string    `json:"status"`
	RouterName  string    `json:"router_name,omitempty"`
	Version     string    `json:"version,omitempty"`
	Board       string    `json:"board,omitempty"`
	Message     string    `json:"message,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// PPPoEStatusRequest represents a request to check PPPoE status
type PPPoEStatusRequest struct {
	Username         string `json:"username" binding:"required"`
	Router           string `json:"router,omitempty"`          // Optional: if specified, check only this router
	TestConnectivity bool   `json:"test_connectivity,omitempty"` // Optional: perform TCP connectivity test
}

// PPPoEStatusResult represents PPPoE status for a single router
type PPPoEStatusResult struct {
	Router             string        `json:"router"`
	IsOnline           bool          `json:"is_online"`
	IPAddress          string        `json:"ip_address,omitempty"`
	CallerID           string        `json:"caller_id,omitempty"`
	Uptime             string        `json:"uptime,omitempty"`
	Encoding           string        `json:"encoding,omitempty"`
	SessionTime        string        `json:"session_time,omitempty"`
	LastSeen           time.Time     `json:"last_seen,omitempty"`
	Message            string        `json:"message,omitempty"`

	// Connectivity Test Fields
	DeviceReachable    bool          `json:"device_reachable,omitempty"`
	ReachablePort      string        `json:"reachable_port,omitempty"`
	ConnectivityTime   time.Duration `json:"connectivity_time,omitempty"`
	ConnectivityStatus string        `json:"connectivity_status,omitempty"` // "reachable", "unreachable", "not_tested"
}

// PPPoEStatusResponse represents the response for PPPoE status check
type PPPoEStatusResponse struct {
	Status      string                         `json:"status"`
	Username    string                         `json:"username"`
	IsOnline    bool                          `json:"is_online"`
	OnlineCount int                           `json:"online_count"`
	Data        map[string]PPPoEStatusResult  `json:"data"`
	Message     string                        `json:"message,omitempty"`
	Timestamp   time.Time                     `json:"timestamp"`
}

// PPPoESearchHistory represents historical PPPoE search for future enhancement
type PPPoESearchHistory struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	UserID    int       `json:"user_id"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// PPPoEFuzzySearchRequest represents a request for fuzzy search
type PPPoEFuzzySearchRequest struct {
	Username string `json:"username" binding:"required"`
	Router   string `json:"router,omitempty"` // Optional: specific router to search
	Limit    int    `json:"limit,omitempty"`  // Optional: max results (default 5)
}

// PPPoEFuzzyMatch represents a fuzzy search match
type PPPoEFuzzyMatch struct {
	Username   string  `json:"username"`
	Router     string  `json:"router"`
	IPAddress  string  `json:"ip_address"`
	CallerID   string  `json:"caller_id"`
	Uptime     string  `json:"uptime"`
	Profile    string  `json:"profile"`
	Similarity float64 `json:"similarity"` // 0.0 to 1.0 similarity score (internal use)
	IsOnline   bool    `json:"is_online"`
}

// PPPoEFuzzySearchResponse represents fuzzy search response
type PPPoEFuzzySearchResponse struct {
	Status      string             `json:"status"`
	SearchTerm  string             `json:"search_term"`
	MatchCount  int                `json:"match_count"`
	Matches     []PPPoEFuzzyMatch  `json:"matches"`
	Message     string             `json:"message,omitempty"`
	Timestamp   time.Time          `json:"timestamp"`
}

// ============================================================================
// ROUTER MANAGEMENT MODELS
// ============================================================================

// Router represents a router configuration with full metadata
type Router struct {
	ID              string    `json:"id" binding:"required"`
	Name            string    `json:"name" binding:"required"`
	Host            string    `json:"host" binding:"required"`
	Port            int       `json:"port" binding:"required,min=1,max=65535"`
	Username        string    `json:"username" binding:"required"`
	Password        string    `json:"password" binding:"required"`
	TunnelEndpoint  string    `json:"tunnel_endpoint" binding:"required"`
	PublicONTURL    string    `json:"public_ont_url" binding:"required"`
	Enabled         bool      `json:"enabled"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RouterStorageConfig represents the complete router storage configuration
type RouterStorageConfig struct {
	Version       string                 `json:"version"`
	LastUpdated   time.Time              `json:"last_updated"`
	Description   string                 `json:"description"`
	Routers       []Router               `json:"routers"`
	AccessControl RouterAccessControl    `json:"access_control"`
	Metadata      RouterStorageMetadata  `json:"metadata"`
}

// RouterAccessControl represents role-based access control for routers
type RouterAccessControl struct {
	Roles map[string]RouterRole `json:"roles"`
}

// RouterRole represents a role with router access permissions
type RouterRole struct {
	Description string   `json:"description"`
	Routers     []string `json:"routers"`
	Permissions []string `json:"permissions"`
}

// RouterStorageMetadata represents metadata about the router storage
type RouterStorageMetadata struct {
	TotalRouters   int       `json:"total_routers"`
	ActiveRouters  int       `json:"active_routers"`
	LastBackup     time.Time `json:"last_backup"`
	BackupLocation string    `json:"backup_location"`
}

// RouterCreateRequest represents request to create a new router
type RouterCreateRequest struct {
	Name           string `json:"name" binding:"required"`
	Host           string `json:"host" binding:"required"`
	Port           int    `json:"port" binding:"required,min=1,max=65535"`
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	TunnelEndpoint string `json:"tunnel_endpoint" binding:"required"`
	PublicONTURL   string `json:"public_ont_url" binding:"required"`
	Description    string `json:"description"`
	Enabled        bool   `json:"enabled"`
}

// RouterUpdateRequest represents request to update an existing router
type RouterUpdateRequest struct {
	Name           string `json:"name" binding:"required"`
	Host           string `json:"host" binding:"required"`
	Port           int    `json:"port" binding:"required,min=1,max=65535"`
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	TunnelEndpoint string `json:"tunnel_endpoint" binding:"required"`
	PublicONTURL   string `json:"public_ont_url" binding:"required"`
	Description    string `json:"description"`
	Enabled        bool   `json:"enabled"`
}

// RouterTestRequest represents request to test router connection
type RouterTestRequest struct {
	RouterID string `json:"router_id" binding:"required"`
}

// RouterResponse represents a single router response (without sensitive data)
type RouterResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Host           string    `json:"host"`
	Port           int       `json:"port"`
	Username       string    `json:"username"`
	TunnelEndpoint string    `json:"tunnel_endpoint"`
	PublicONTURL   string    `json:"public_ont_url"`
	Enabled        bool      `json:"enabled"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	// Password is intentionally excluded for security
}

// RouterListResponse represents response for router list API
type RouterListResponse struct {
	Status  string           `json:"status"`
	Data    []RouterResponse `json:"data"`
	Message string           `json:"message,omitempty"`
	Total   int              `json:"total"`
}

// RouterDetailResponse represents response for single router API
type RouterDetailResponse struct {
	Status  string         `json:"status"`
	Data    RouterResponse `json:"data"`
	Message string         `json:"message,omitempty"`
}

// RouterCreateResponse represents response for router creation API
type RouterCreateResponse struct {
	Status  string         `json:"status"`
	Data    RouterResponse `json:"data"`
	Message string         `json:"message"`
}

// RouterUpdateResponse represents response for router update API
type RouterUpdateResponse struct {
	Status  string         `json:"status"`
	Data    RouterResponse `json:"data"`
	Message string         `json:"message"`
}

// RouterDeleteResponse represents response for router deletion API
type RouterDeleteResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// RouterTestResponse represents response for router connection test API
type RouterTestResponse struct {
	Status        string                  `json:"status"`
	RouterID      string                  `json:"router_id"`
	RouterName    string                  `json:"router_name"`
	TestResult    RouterConnectionTest    `json:"test_result"`
	Message       string                  `json:"message"`
}

// RouterValidationError represents validation errors for router operations
type RouterValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// RouterValidationResponse represents response with validation errors
type RouterValidationResponse struct {
	Status string                  `json:"status"`
	Errors []RouterValidationError `json:"errors"`
}

// RouterImportRequest represents request to import routers from JSON
type RouterImportRequest struct {
	Routers   []RouterCreateRequest `json:"routers" binding:"required"`
	Overwrite bool                  `json:"overwrite"` // Whether to overwrite existing routers
}

// RouterImportResponse represents response for router import API
type RouterImportResponse struct {
	Status      string   `json:"status"`
	Message     string   `json:"message"`
	Imported    int      `json:"imported"`
	Failed      int      `json:"failed"`
	Skipped     int      `json:"skipped"`
	FailedItems []string `json:"failed_items,omitempty"`
}

// RouterExportResponse represents response for router export API
type RouterExportResponse struct {
	Status    string           `json:"status"`
	Routers   []RouterResponse `json:"routers"`
	Total     int              `json:"total"`
	ExportedAt time.Time       `json:"exported_at"`
}

// RouterStatsResponse represents response for router statistics API
type RouterStatsResponse struct {
	Status          string                           `json:"status"`
	TotalRouters    int                              `json:"total_routers"`
	ActiveRouters   int                              `json:"active_routers"`
	DisabledRouters int                              `json:"disabled_routers"`
	ConnectionTests map[string]RouterConnectionTest  `json:"connection_tests"`
	LastUpdated     time.Time                        `json:"last_updated"`
}

// RouterBackupRequest represents request to backup router configurations
type RouterBackupRequest struct {
	IncludePasswords bool   `json:"include_passwords"`
	BackupLocation   string `json:"backup_location,omitempty"`
}

// RouterBackupResponse represents response for router backup API
type RouterBackupResponse struct {
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	BackupLocation string    `json:"backup_location"`
	BackupSize     int64     `json:"backup_size"`
	RouterCount    int       `json:"router_count"`
	BackupTime     time.Time `json:"backup_time"`
}

// ToRouter converts a RouterCreateRequest to a Router with generated metadata
func (req *RouterCreateRequest) ToRouter(id string) Router {
	now := time.Now()
	return Router{
		ID:             id,
		Name:           req.Name,
		Host:           req.Host,
		Port:           req.Port,
		Username:       req.Username,
		Password:       req.Password,
		TunnelEndpoint: req.TunnelEndpoint,
		PublicONTURL:   req.PublicONTURL,
		Enabled:        req.Enabled,
		Description:    req.Description,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// ToResponse converts a Router to RouterResponse (excluding sensitive data)
func (r *Router) ToResponse() RouterResponse {
	return RouterResponse{
		ID:             r.ID,
		Name:           r.Name,
		Host:           r.Host,
		Port:           r.Port,
		Username:       r.Username,
		TunnelEndpoint: r.TunnelEndpoint,
		PublicONTURL:   r.PublicONTURL,
		Enabled:        r.Enabled,
		Description:    r.Description,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// ToNATRouterConfig converts a Router to NATRouterConfig for backward compatibility
func (r *Router) ToNATRouterConfig() NATRouterConfig {
	return NATRouterConfig{
		Name:           r.Name,
		Host:           r.Host,
		Port:           r.Port,
		Username:       r.Username,
		Password:       r.Password,
		TunnelEndpoint: r.TunnelEndpoint,
		PublicONTURL:   r.PublicONTURL,
	}
}

// UpdateFromRequest updates router fields from RouterUpdateRequest
func (r *Router) UpdateFromRequest(req *RouterUpdateRequest) {
	r.Name = req.Name
	r.Host = req.Host
	r.Port = req.Port
	r.Username = req.Username
	r.Password = req.Password
	r.TunnelEndpoint = req.TunnelEndpoint
	r.PublicONTURL = req.PublicONTURL
	r.Description = req.Description
	r.Enabled = req.Enabled
	r.UpdatedAt = time.Now()
}

// ============================================================================
// ONT WIFI EXTRACTION MODELS
// ============================================================================

// ONTWiFiInfo represents WiFi information extracted from ONT device
type ONTWiFiInfo struct {
	ID             int       `json:"id" db:"id"`
	PPPoEUsername  string    `json:"pppoe_username" db:"pppoe_username"`
	Router         string    `json:"router" db:"router"`
	SSID           string    `json:"ssid" db:"ssid"`
	Password       string    `json:"password" db:"password"`
	Security       string    `json:"security" db:"security"`
	Encryption     string    `json:"encryption" db:"encryption"`
	Authentication string    `json:"authentication" db:"authentication"`
	ONTURL         string    `json:"ont_url" db:"ont_url"`
	ONTModel       string    `json:"ont_model" db:"ont_model"`
	ExtractedAt    time.Time `json:"extracted_at" db:"extracted_at"`
	ExtractedBy    string    `json:"extracted_by" db:"extracted_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ONTWiFiExtractRequest represents request to extract WiFi info from ONT
type ONTWiFiExtractRequest struct {
	ONTURL        string `json:"ont_url" binding:"required"`
	Username      string `json:"username"`                  // ONT login username (default: admin)
	Password      string `json:"password"`                  // ONT login password (default: admin)
	PPPoEUsername string `json:"pppoe_username"`            // Optional: PPPoE username for database record
	Router        string `json:"router"`                    // Optional: Router name for database record
	Debug         bool   `json:"debug"`                     // Optional: Enable debug mode in extractor
}

// ONTWiFiExtractFromNATRequest represents request to extract WiFi from existing NAT config
type ONTWiFiExtractFromNATRequest struct {
	Router        string `json:"router" binding:"required"`
	PPPoEUsername string `json:"pppoe_username"`            // Optional: for database record
	ONTUsername   string `json:"ont_username"`              // Optional: ONT login username
	ONTPassword   string `json:"ont_password"`              // Optional: ONT login password
	Debug         bool   `json:"debug"`                     // Optional: Enable debug mode
}

// ONTWiFiExtractResponse represents response for WiFi extraction
type ONTWiFiExtractResponse struct {
	Status       string        `json:"status"`
	Data         *ONTWiFiInfo  `json:"data,omitempty"`
	Message      string        `json:"message"`
	ExtractionTime time.Duration `json:"extraction_time,omitempty"` // Time taken for extraction
	Timestamp    time.Time     `json:"timestamp"`
}

// ONTWiFiHistoryRequest represents request to get WiFi extraction history
type ONTWiFiHistoryRequest struct {
	PPPoEUsername string `json:"pppoe_username,omitempty"` // Filter by PPPoE username
	Router        string `json:"router,omitempty"`         // Filter by router
	Limit         int    `json:"limit,omitempty"`          // Limit results (default: 50)
	Offset        int    `json:"offset,omitempty"`         // Offset for pagination
}

// ONTWiFiHistoryResponse represents response for WiFi history
type ONTWiFiHistoryResponse struct {
	Status  string          `json:"status"`
	Data    []ONTWiFiInfo   `json:"data"`
	Total   int             `json:"total"`
	Message string          `json:"message,omitempty"`
}

// ONTWiFiAvailabilityResponse represents response for checking webautomation availability
type ONTWiFiAvailabilityResponse struct {
	Status          string   `json:"status"`
	Available       bool     `json:"available"`
	Message         string   `json:"message"`
	SupportedModels []string `json:"supported_models,omitempty"`
	NodeVersion     string   `json:"node_version,omitempty"`
}

// ONTWiFiBulkExtractRequest represents request to extract WiFi from multiple ONTs
type ONTWiFiBulkExtractRequest struct {
	Targets []ONTWiFiExtractRequest `json:"targets" binding:"required"`
}

// ONTWiFiBulkExtractResult represents result for single extraction in bulk operation
type ONTWiFiBulkExtractResult struct {
	ONTURL       string        `json:"ont_url"`
	Success      bool          `json:"success"`
	Data         *ONTWiFiInfo  `json:"data,omitempty"`
	Error        string        `json:"error,omitempty"`
	ExtractTime  time.Duration `json:"extract_time"`
}

// ONTWiFiBulkExtractResponse represents response for bulk WiFi extraction
type ONTWiFiBulkExtractResponse struct {
	Status       string                     `json:"status"`
	TotalTargets int                        `json:"total_targets"`
	Successful   int                        `json:"successful"`
	Failed       int                        `json:"failed"`
	Results      []ONTWiFiBulkExtractResult `json:"results"`
	TotalTime    time.Duration              `json:"total_time"`
	Message      string                     `json:"message"`
	Timestamp    time.Time                  `json:"timestamp"`
}