package models

// ErrorCode represents standardized error codes for programmatic handling
type ErrorCode string

const (
	// Authentication & Authorization Errors
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeSessionExpired     ErrorCode = "SESSION_EXPIRED"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"

	// Validation Errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// Resource Errors
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeConflict      ErrorCode = "CONFLICT"

	// Operation Errors
	ErrCodeOperationFailed ErrorCode = "OPERATION_FAILED"
	ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"
	ErrCodeNetworkError    ErrorCode = "NETWORK_ERROR"
	ErrCodeTimeout         ErrorCode = "TIMEOUT"

	// Rate Limiting
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Router/Connection Errors
	ErrCodeRouterOffline      ErrorCode = "ROUTER_OFFLINE"
	ErrCodeConnectionFailed   ErrorCode = "CONNECTION_FAILED"
	ErrCodeCircuitBreakerOpen ErrorCode = "CIRCUIT_BREAKER_OPEN"

	// Generic Error
	ErrCodeInternalError ErrorCode = "INTERNAL_ERROR"
)

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field       string `json:"field"`
	Message     string `json:"message"`
	Constraint  string `json:"constraint,omitempty"`  // e.g., "required", "min:6", "email"
	CurrentValue string `json:"current_value,omitempty"` // For debugging (sanitized)
}

// ErrorDetail provides detailed error information with actionable suggestions
type ErrorDetail struct {
	Status      string        `json:"status"`                // "error"
	Code        ErrorCode     `json:"code"`                  // Standardized error code
	Message     string        `json:"message"`               // User-friendly message
	Details     string        `json:"details,omitempty"`     // Technical details (optional)
	Suggestion  string        `json:"suggestion,omitempty"`  // Actionable suggestion
	Fields      []FieldError  `json:"fields,omitempty"`      // Field-level validation errors
	RetryAfter  int           `json:"retry_after,omitempty"` // Seconds to wait before retry
	HelpURL     string        `json:"help_url,omitempty"`    // Link to documentation
	RequestID   string        `json:"request_id,omitempty"`  // For support/debugging
	Timestamp   int64         `json:"timestamp,omitempty"`   // Unix timestamp
}

// ErrorResponse is the standard error response format (backward compatible)
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"` // Deprecated, use ErrorDetail
}

// NewErrorDetail creates a new detailed error response
func NewErrorDetail(code ErrorCode, message string) *ErrorDetail {
	return &ErrorDetail{
		Status:  "error",
		Code:    code,
		Message: message,
	}
}

// WithDetails adds technical details to the error
func (e *ErrorDetail) WithDetails(details string) *ErrorDetail {
	e.Details = details
	return e
}

// WithSuggestion adds an actionable suggestion
func (e *ErrorDetail) WithSuggestion(suggestion string) *ErrorDetail {
	e.Suggestion = suggestion
	return e
}

// WithFieldError adds a field validation error
func (e *ErrorDetail) WithFieldError(field, message, constraint string) *ErrorDetail {
	if e.Fields == nil {
		e.Fields = []FieldError{}
	}
	e.Fields = append(e.Fields, FieldError{
		Field:      field,
		Message:    message,
		Constraint: constraint,
	})
	return e
}

// WithFieldErrors adds multiple field validation errors
func (e *ErrorDetail) WithFieldErrors(fields []FieldError) *ErrorDetail {
	e.Fields = fields
	return e
}

// WithRetryAfter adds retry-after time for rate limiting
func (e *ErrorDetail) WithRetryAfter(seconds int) *ErrorDetail {
	e.RetryAfter = seconds
	return e
}

// WithHelpURL adds a documentation link
func (e *ErrorDetail) WithHelpURL(url string) *ErrorDetail {
	e.HelpURL = url
	return e
}

// WithRequestID adds a request ID for support
func (e *ErrorDetail) WithRequestID(requestID string) *ErrorDetail {
	e.RequestID = requestID
	return e
}

// WithTimestamp adds the current timestamp
func (e *ErrorDetail) WithTimestamp(timestamp int64) *ErrorDetail {
	e.Timestamp = timestamp
	return e
}

// Common error messages with suggestions
var (
	ErrUnauthorized = NewErrorDetail(
		ErrCodeUnauthorized,
		"Authentication required",
	).WithSuggestion("Please log in to access this resource")

	ErrForbidden = NewErrorDetail(
		ErrCodeForbidden,
		"You don't have permission to perform this action",
	).WithSuggestion("Contact your administrator if you need access")

	ErrInvalidCredentials = NewErrorDetail(
		ErrCodeInvalidCredentials,
		"Invalid username or password",
	).WithSuggestion("Please check your credentials and try again")

	ErrSessionExpired = NewErrorDetail(
		ErrCodeSessionExpired,
		"Your session has expired",
	).WithSuggestion("Please log in again to continue")

	ErrTokenExpired = NewErrorDetail(
		ErrCodeTokenExpired,
		"Your authentication token has expired",
	).WithSuggestion("Please refresh your token or log in again")

	ErrNotFound = NewErrorDetail(
		ErrCodeNotFound,
		"The requested resource was not found",
	).WithSuggestion("Please check the resource ID and try again")

	ErrAlreadyExists = NewErrorDetail(
		ErrCodeAlreadyExists,
		"A resource with this identifier already exists",
	).WithSuggestion("Please use a different identifier or update the existing resource")

	ErrRateLimitExceeded = NewErrorDetail(
		ErrCodeRateLimitExceeded,
		"Too many requests",
	).WithSuggestion("Please wait a moment before trying again")

	ErrRouterOffline = NewErrorDetail(
		ErrCodeRouterOffline,
		"Router is currently offline or unreachable",
	).WithSuggestion("Please check the router connection and try again later")

	ErrCircuitBreakerOpen = NewErrorDetail(
		ErrCodeCircuitBreakerOpen,
		"Service is temporarily unavailable due to repeated failures",
	).WithSuggestion("The system is protecting itself. Please try again in a few moments")

	ErrInternalError = NewErrorDetail(
		ErrCodeInternalError,
		"An unexpected error occurred",
	).WithSuggestion("Please try again later. If the problem persists, contact support")
)
