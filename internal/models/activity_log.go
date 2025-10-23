package models

import (
	"encoding/json"
	"time"
)

// ActivityLog represents an audit log entry
type ActivityLog struct {
	ID           int                    `json:"id"`
	UserID       *int                   `json:"user_id,omitempty"`       // Nullable for deleted users
	Username     string                 `json:"username"`
	UserRole     string                 `json:"user_role,omitempty"`
	ActionType   string                 `json:"action_type"`             // LOGIN, LOGOUT, CREATE, UPDATE, DELETE, NAT_UPDATE, PPPOE_CHECK
	ResourceType string                 `json:"resource_type,omitempty"` // USER, ROUTER, NAT_RULE, PPPOE
	ResourceID   string                 `json:"resource_id,omitempty"`   // ID or name of resource
	Description  string                 `json:"description"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Status       string                 `json:"status"`                  // SUCCESS, FAILED, ERROR
	ErrorMessage string                 `json:"error_message,omitempty"`
	DurationMs   *int                   `json:"duration_ms,omitempty"`   // Operation duration in milliseconds
	DeviceInfo   *DeviceInfo            `json:"device_info,omitempty"`   // Device context (browser, OS, device type)
	Metadata     map[string]interface{} `json:"metadata,omitempty"`      // Additional data (before/after, circuit breaker state)
	CreatedAt    time.Time              `json:"created_at"`
}

// DeviceInfo represents device context extracted from user agent
type DeviceInfo struct {
	Browser     string `json:"browser,omitempty"`
	OS          string `json:"os,omitempty"`
	DeviceType  string `json:"device_type,omitempty"` // desktop, mobile, tablet
	IsMobile    bool   `json:"is_mobile"`
}

// ActivityLogCreate represents the data needed to create a new log entry
type ActivityLogCreate struct {
	UserID       *int                   `json:"user_id"`
	Username     string                 `json:"username"`
	UserRole     string                 `json:"user_role"`
	ActionType   string                 `json:"action_type"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	Description  string                 `json:"description"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Status       string                 `json:"status"`
	ErrorMessage string                 `json:"error_message"`
	DurationMs   *int                   `json:"duration_ms"`   // Operation duration in milliseconds
	DeviceInfo   *DeviceInfo            `json:"device_info"`   // Device context
	Metadata     map[string]interface{} `json:"metadata"`
}

// ActivityLogFilter represents filters for querying logs
type ActivityLogFilter struct {
	UserID        *int      `json:"user_id"`
	Username      string    `json:"username"`
	ActionType    string    `json:"action_type"`      // Single action type (deprecated, use ActionTypes)
	ActionTypes   []string  `json:"action_types"`     // Multi-select action types
	ResourceType  string    `json:"resource_type"`
	ResourceTypes []string  `json:"resource_types"`   // Multi-select resource types
	Status        string    `json:"status"`
	Statuses      []string  `json:"statuses"`         // Multi-select statuses
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Limit         int       `json:"limit"`
	Offset        int       `json:"offset"`
}

// ActivityLogsResponse represents the API response for logs listing
type ActivityLogsResponse struct {
	Status     string        `json:"status"`
	Data       []ActivityLog `json:"data"`
	Total      int           `json:"total"`
	Pagination *Pagination   `json:"pagination,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
}

// Action type constants
const (
	ActionLogin       = "LOGIN"
	ActionLogout      = "LOGOUT"
	ActionCreate      = "CREATE"
	ActionUpdate      = "UPDATE"
	ActionDelete      = "DELETE"
	ActionNATUpdate   = "NAT_UPDATE"
	ActionPPPoECheck  = "PPPOE_CHECK"
	ActionTest        = "TEST"
	ActionView        = "VIEW"
)

// Resource type constants
const (
	ResourceUser     = "USER"
	ResourceRouter   = "ROUTER"
	ResourceNATRule  = "NAT_RULE"
	ResourcePPPoE    = "PPPOE"
	ResourceAuth     = "AUTH"
)

// Status constants
const (
	StatusSuccess = "SUCCESS"
	StatusFailed  = "FAILED"
	StatusError   = "ERROR"
)

// MarshalMetadata marshals metadata map to JSON
func (a *ActivityLog) MarshalMetadata() ([]byte, error) {
	if a.Metadata == nil {
		return nil, nil
	}
	return json.Marshal(a.Metadata)
}

// UnmarshalMetadata unmarshals JSON to metadata map
func (a *ActivityLog) UnmarshalMetadata(data []byte) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, &a.Metadata)
}

// GetActionTypeLabel returns human-readable label for action type
func GetActionTypeLabel(actionType string) string {
	labels := map[string]string{
		ActionLogin:      "Login",
		ActionLogout:     "Logout",
		ActionCreate:     "Create",
		ActionUpdate:     "Update",
		ActionDelete:     "Delete",
		ActionNATUpdate:  "NAT Update",
		ActionPPPoECheck: "PPPoE Check",
		ActionTest:       "Test",
		ActionView:       "View",
	}
	if label, ok := labels[actionType]; ok {
		return label
	}
	return actionType
}

// GetResourceTypeLabel returns human-readable label for resource type
func GetResourceTypeLabel(resourceType string) string {
	labels := map[string]string{
		ResourceUser:    "User",
		ResourceRouter:  "Router",
		ResourceNATRule: "NAT Rule",
		ResourcePPPoE:   "PPPoE",
		ResourceAuth:    "Authentication",
	}
	if label, ok := labels[resourceType]; ok {
		return label
	}
	return resourceType
}
