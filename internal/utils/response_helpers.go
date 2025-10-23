package utils

import (
	"fmt"
	"net/http"
	"time"

	"nat-management-app/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// RespondWithError sends a user-friendly error response
func RespondWithError(c *gin.Context, statusCode int, err *models.ErrorDetail) {
	// Add timestamp if not set
	if err.Timestamp == 0 {
		err.Timestamp = time.Now().Unix()
	}

	c.JSON(statusCode, err)
}

// RespondUnauthorized sends an unauthorized error
func RespondUnauthorized(c *gin.Context) {
	RespondWithError(c, http.StatusUnauthorized, models.ErrUnauthorized)
}

// RespondForbidden sends a forbidden error
func RespondForbidden(c *gin.Context) {
	RespondWithError(c, http.StatusForbidden, models.ErrForbidden)
}

// RespondNotFound sends a not found error with custom resource type
func RespondNotFound(c *gin.Context, resourceType string) {
	err := models.NewErrorDetail(
		models.ErrCodeNotFound,
		fmt.Sprintf("%s not found", resourceType),
	).WithSuggestion(fmt.Sprintf("Please check the %s ID and try again", resourceType))

	RespondWithError(c, http.StatusNotFound, err)
}

// RespondAlreadyExists sends an already exists error
func RespondAlreadyExists(c *gin.Context, resourceType, identifier string) {
	err := models.NewErrorDetail(
		models.ErrCodeAlreadyExists,
		fmt.Sprintf("%s '%s' already exists", resourceType, identifier),
	).WithSuggestion(fmt.Sprintf("Please use a different %s identifier", resourceType))

	RespondWithError(c, http.StatusConflict, err)
}

// RespondValidationError sends a validation error with field details
func RespondValidationError(c *gin.Context, validationErr error) {
	err := models.NewErrorDetail(
		models.ErrCodeValidationFailed,
		"Validation failed",
	).WithSuggestion("Please check the highlighted fields and correct the errors")

	// Parse validator errors
	if ve, ok := validationErr.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			fieldErr := models.FieldError{
				Field:      getJSONFieldName(fe),
				Message:    getValidationMessage(fe),
				Constraint: fe.Tag(),
			}
			err.Fields = append(err.Fields, fieldErr)
		}
	} else {
		// Generic validation error
		err.Details = validationErr.Error()
	}

	RespondWithError(c, http.StatusBadRequest, err)
}

// RespondInvalidInput sends an invalid input error
func RespondInvalidInput(c *gin.Context, field, message string) {
	err := models.NewErrorDetail(
		models.ErrCodeInvalidInput,
		"Invalid input provided",
	).WithFieldError(field, message, "validation").
		WithSuggestion("Please check the input and try again")

	RespondWithError(c, http.StatusBadRequest, err)
}

// RespondRateLimitExceeded sends a rate limit error
func RespondRateLimitExceeded(c *gin.Context, retryAfterSeconds int) {
	err := models.ErrRateLimitExceeded.
		WithRetryAfter(retryAfterSeconds).
		WithSuggestion(fmt.Sprintf("Please wait %d seconds before trying again", retryAfterSeconds))

	RespondWithError(c, http.StatusTooManyRequests, err)
}

// RespondRouterOffline sends a router offline error
func RespondRouterOffline(c *gin.Context, routerName string) {
	err := models.NewErrorDetail(
		models.ErrCodeRouterOffline,
		fmt.Sprintf("Router '%s' is currently offline", routerName),
	).WithSuggestion("Please check the router connection or try a different router").
		WithDetails(fmt.Sprintf("Router '%s' failed health check", routerName))

	RespondWithError(c, http.StatusServiceUnavailable, err)
}

// RespondCircuitBreakerOpen sends a circuit breaker error
func RespondCircuitBreakerOpen(c *gin.Context, routerName string, retryAfterSeconds int) {
	err := models.ErrCircuitBreakerOpen.
		WithDetails(fmt.Sprintf("Circuit breaker is OPEN for router '%s'", routerName)).
		WithRetryAfter(retryAfterSeconds).
		WithSuggestion(fmt.Sprintf("The router is experiencing issues. Please try again in %d seconds", retryAfterSeconds))

	RespondWithError(c, http.StatusServiceUnavailable, err)
}

// RespondDatabaseError sends a database error
func RespondDatabaseError(c *gin.Context, operation string) {
	err := models.NewErrorDetail(
		models.ErrCodeDatabaseError,
		"Database operation failed",
	).WithDetails(fmt.Sprintf("Failed to %s", operation)).
		WithSuggestion("Please try again. If the problem persists, contact support")

	RespondWithError(c, http.StatusInternalServerError, err)
}

// RespondInternalError sends an internal server error
func RespondInternalError(c *gin.Context, details string) {
	err := models.ErrInternalError.WithDetails(details)
	RespondWithError(c, http.StatusInternalServerError, err)
}

// RespondOperationFailed sends an operation failed error with custom message
func RespondOperationFailed(c *gin.Context, operation, reason, suggestion string) {
	err := models.NewErrorDetail(
		models.ErrCodeOperationFailed,
		fmt.Sprintf("Failed to %s", operation),
	).WithDetails(reason).
		WithSuggestion(suggestion)

	RespondWithError(c, http.StatusBadRequest, err)
}

// getJSONFieldName extracts the JSON field name from validator field error
func getJSONFieldName(fe validator.FieldError) string {
	// Try to get JSON tag, fallback to field name
	// This is a simplified version - you might want to enhance this
	// to properly parse struct tags
	field := fe.Field()

	// Convert first letter to lowercase for JSON convention
	if len(field) > 0 {
		return string(field[0]+32) + field[1:]
	}
	return field
}

// getValidationMessage returns a user-friendly validation error message
func getValidationMessage(fe validator.FieldError) string {
	field := getJSONFieldName(fe)

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid (constraint: %s)", field, fe.Tag())
	}
}

// RespondSuccess sends a success response (for consistency)
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   data,
	})
}

// RespondSuccessWithMessage sends a success response with a message
func RespondSuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

// RespondCreated sends a created response
func RespondCreated(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}
