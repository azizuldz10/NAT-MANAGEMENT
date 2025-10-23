/**
 * API Utilities with Toast Integration
 * Handles API requests with automatic error handling and toast notifications
 *
 * Usage:
 *   API.get('/api/users').then(data => {
 *     Toast.success('Users loaded!');
 *   });
 *
 *   API.post('/api/users', userData).catch(err => {
 *     // Errors are automatically shown as toasts
 *   });
 */

class APIClient {
    constructor(baseURL = '') {
        this.baseURL = baseURL;
        this.defaultHeaders = {
            'Content-Type': 'application/json'
        };
    }

    /**
     * Make an API request
     * @param {string} endpoint - API endpoint
     * @param {Object} options - Fetch options
     * @param {boolean} showErrorToast - Show error toast automatically (default: true)
     * @param {boolean} showSuccessToast - Show success toast automatically (default: false)
     * @param {string} successMessage - Custom success message
     */
    async request(endpoint, options = {}, config = {}) {
        const {
            showErrorToast = true,
            showSuccessToast = false,
            successMessage = 'Operation completed successfully'
        } = config;

        const url = this.baseURL + endpoint;
        const fetchOptions = {
            ...options,
            headers: {
                ...this.defaultHeaders,
                ...options.headers
            }
        };

        try {
            const response = await fetch(url, fetchOptions);
            const data = await response.json();

            if (!response.ok) {
                // Handle error response
                if (showErrorToast) {
                    this.handleErrorResponse(data, response.status);
                }
                throw new APIError(data, response.status);
            }

            // Handle success
            if (showSuccessToast && data.message) {
                Toast.success('Success', data.message);
            } else if (showSuccessToast) {
                Toast.success(successMessage);
            }

            return data;
        } catch (error) {
            if (error instanceof APIError) {
                throw error;
            }

            // Network or other errors
            if (showErrorToast) {
                Toast.error(
                    'Network Error',
                    'Unable to connect to the server. Please check your internet connection.',
                    0
                );
            }

            throw new APIError({
                status: 'error',
                code: 'NETWORK_ERROR',
                message: 'Network request failed'
            }, 0);
        }
    }

    handleErrorResponse(errorData, statusCode) {
        // Enhanced error response from backend
        if (errorData.code && errorData.message) {
            const title = this.getErrorTitle(errorData.code);
            const message = errorData.message;
            const suggestion = errorData.suggestion;

            // Handle field-level validation errors
            if (errorData.fields && errorData.fields.length > 0) {
                const fieldMessages = errorData.fields
                    .map(f => `• ${f.message}`)
                    .join('\n');

                Toast.show({
                    type: 'error',
                    title: title,
                    message: fieldMessages,
                    duration: 8000,
                    dismissible: true
                });
                return;
            }

            // Handle rate limiting
            if (errorData.code === 'RATE_LIMIT_EXCEEDED') {
                const retryAfter = errorData.retry_after || 60;
                Toast.show({
                    type: 'warning',
                    title: title,
                    message: `${message}. ${suggestion || `Please wait ${retryAfter} seconds.`}`,
                    duration: 0 // Don't auto-dismiss
                });
                return;
            }

            // Handle circuit breaker
            if (errorData.code === 'CIRCUIT_BREAKER_OPEN') {
                Toast.show({
                    type: 'warning',
                    title: 'Service Temporarily Unavailable',
                    message: suggestion || message,
                    duration: 0
                });
                return;
            }

            // Handle authentication errors
            if (errorData.code === 'UNAUTHORIZED' || errorData.code === 'SESSION_EXPIRED' || errorData.code === 'TOKEN_EXPIRED') {
                Toast.show({
                    type: 'error',
                    title: title,
                    message: message,
                    actions: [
                        {
                            text: 'Login',
                            primary: true,
                            onClick: () => {
                                window.location.href = '/login';
                            }
                        }
                    ],
                    duration: 0
                });
                return;
            }

            // General error with suggestion
            Toast.error(
                title,
                suggestion ? `${message}\n\n💡 ${suggestion}` : message,
                7000
            );
        } else {
            // Legacy error format
            Toast.error(
                'Error',
                errorData.message || 'An unexpected error occurred',
                5000
            );
        }
    }

    getErrorTitle(code) {
        const titles = {
            'UNAUTHORIZED': 'Authentication Required',
            'FORBIDDEN': 'Access Denied',
            'INVALID_CREDENTIALS': 'Invalid Credentials',
            'SESSION_EXPIRED': 'Session Expired',
            'TOKEN_EXPIRED': 'Token Expired',
            'VALIDATION_FAILED': 'Validation Error',
            'INVALID_INPUT': 'Invalid Input',
            'NOT_FOUND': 'Not Found',
            'ALREADY_EXISTS': 'Already Exists',
            'RATE_LIMIT_EXCEEDED': 'Too Many Requests',
            'ROUTER_OFFLINE': 'Router Offline',
            'CIRCUIT_BREAKER_OPEN': 'Service Unavailable',
            'DATABASE_ERROR': 'Database Error',
            'NETWORK_ERROR': 'Network Error',
            'TIMEOUT': 'Request Timeout',
            'INTERNAL_ERROR': 'Internal Server Error'
        };
        return titles[code] || 'Error';
    }

    // Convenience methods
    get(endpoint, config) {
        return this.request(endpoint, { method: 'GET' }, config);
    }

    post(endpoint, data, config) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        }, config);
    }

    put(endpoint, data, config) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        }, config);
    }

    patch(endpoint, data, config) {
        return this.request(endpoint, {
            method: 'PATCH',
            body: JSON.stringify(data)
        }, config);
    }

    delete(endpoint, config) {
        return this.request(endpoint, { method: 'DELETE' }, config);
    }
}

// Custom API Error class
class APIError extends Error {
    constructor(errorData, statusCode) {
        super(errorData.message || 'API Error');
        this.name = 'APIError';
        this.code = errorData.code;
        this.status = errorData.status;
        this.statusCode = statusCode;
        this.details = errorData.details;
        this.suggestion = errorData.suggestion;
        this.fields = errorData.fields;
        this.data = errorData;
    }
}

// Create global API instance
const API = new APIClient();

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { API, APIClient, APIError };
}
