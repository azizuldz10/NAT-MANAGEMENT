package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal operation
	StateOpen                          // Circuit is open (failing)
	StateHalfOpen                      // Testing if service recovered
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern for router connections
type CircuitBreaker struct {
	logger           *logrus.Logger
	failureThreshold int           // Number of failures before opening circuit
	timeout          time.Duration // Time to wait before attempting half-open
	maxHalfOpenReqs  int           // Max requests allowed in half-open state
}

// CircuitBreakerState tracks state for a specific router
type CircuitBreakerState struct {
	state            CircuitState
	failures         int
	lastFailureTime  time.Time
	lastSuccessTime  time.Time
	halfOpenRequests int
	mu               sync.RWMutex
}

// RouterCircuitBreaker manages circuit breakers for all routers
type RouterCircuitBreaker struct {
	breaker *CircuitBreaker
	states  map[string]*CircuitBreakerState // routerName -> state
	mu      sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(logger *logrus.Logger, failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		logger:           logger,
		failureThreshold: failureThreshold,
		timeout:          timeout,
		maxHalfOpenReqs:  1, // Only allow 1 request in half-open state
	}
}

// NewRouterCircuitBreaker creates a new router circuit breaker manager
func NewRouterCircuitBreaker(logger *logrus.Logger, failureThreshold int, timeout time.Duration) *RouterCircuitBreaker {
	return &RouterCircuitBreaker{
		breaker: NewCircuitBreaker(logger, failureThreshold, timeout),
		states:  make(map[string]*CircuitBreakerState),
	}
}

// getOrCreateState gets or creates circuit breaker state for a router
func (rcb *RouterCircuitBreaker) getOrCreateState(routerName string) *CircuitBreakerState {
	rcb.mu.RLock()
	state, exists := rcb.states[routerName]
	rcb.mu.RUnlock()

	if exists {
		return state
	}

	// Create new state
	rcb.mu.Lock()
	defer rcb.mu.Unlock()

	// Double-check after acquiring write lock
	if state, exists := rcb.states[routerName]; exists {
		return state
	}

	state = &CircuitBreakerState{
		state:           StateClosed,
		failures:        0,
		lastFailureTime: time.Time{},
		lastSuccessTime: time.Now(),
	}
	rcb.states[routerName] = state

	return state
}

// Call executes a function with circuit breaker protection
func (rcb *RouterCircuitBreaker) Call(routerName string, fn func() error) error {
	state := rcb.getOrCreateState(routerName)

	state.mu.Lock()
	currentState := state.state
	state.mu.Unlock()

	// Check if circuit is open
	if currentState == StateOpen {
		// Check if timeout has elapsed
		state.mu.RLock()
		elapsed := time.Since(state.lastFailureTime)
		state.mu.RUnlock()

		if elapsed < rcb.breaker.timeout {
			return fmt.Errorf("circuit breaker is OPEN for %s (timeout: %v remaining)",
				routerName, rcb.breaker.timeout-elapsed)
		}

		// Transition to half-open
		state.mu.Lock()
		state.state = StateHalfOpen
		state.halfOpenRequests = 0
		state.mu.Unlock()

		rcb.breaker.logger.Infof("🔄 Circuit breaker for %s transitioned to HALF-OPEN (testing recovery)", routerName)
	}

	// Check if in half-open state
	if currentState == StateHalfOpen {
		state.mu.Lock()
		if state.halfOpenRequests >= rcb.breaker.maxHalfOpenReqs {
			state.mu.Unlock()
			return fmt.Errorf("circuit breaker is HALF-OPEN for %s (max test requests reached)", routerName)
		}
		state.halfOpenRequests++
		state.mu.Unlock()
	}

	// Execute function
	err := fn()

	if err != nil {
		// Record failure
		rcb.recordFailure(routerName, state)
		return err
	}

	// Record success
	rcb.recordSuccess(routerName, state)
	return nil
}

// recordFailure records a failure and potentially opens the circuit
func (rcb *RouterCircuitBreaker) recordFailure(routerName string, state *CircuitBreakerState) {
	state.mu.Lock()
	defer state.mu.Unlock()

	state.failures++
	state.lastFailureTime = time.Now()

	// If in half-open state, immediately open on failure
	if state.state == StateHalfOpen {
		state.state = StateOpen
		state.failures = rcb.breaker.failureThreshold
		rcb.breaker.logger.Warnf("🚨 Circuit breaker for %s reopened (failed during half-open test)", routerName)
		return
	}

	// Check if threshold reached
	if state.failures >= rcb.breaker.failureThreshold {
		state.state = StateOpen
		rcb.breaker.logger.Warnf("🚨 Circuit breaker OPENED for %s (failures: %d/%d, timeout: %v)",
			routerName, state.failures, rcb.breaker.failureThreshold, rcb.breaker.timeout)
	} else {
		rcb.breaker.logger.Debugf("⚠️ Circuit breaker failure for %s (%d/%d)",
			routerName, state.failures, rcb.breaker.failureThreshold)
	}
}

// recordSuccess records a success and potentially closes the circuit
func (rcb *RouterCircuitBreaker) recordSuccess(routerName string, state *CircuitBreakerState) {
	state.mu.Lock()
	defer state.mu.Unlock()

	previousState := state.state
	state.failures = 0
	state.lastSuccessTime = time.Now()
	state.state = StateClosed

	if previousState == StateHalfOpen {
		rcb.breaker.logger.Infof("✅ Circuit breaker for %s CLOSED (recovery successful)", routerName)
	} else if previousState == StateOpen {
		rcb.breaker.logger.Infof("✅ Circuit breaker for %s recovered", routerName)
	}
}

// GetState returns the current state of a router's circuit breaker
func (rcb *RouterCircuitBreaker) GetState(routerName string) CircuitState {
	state := rcb.getOrCreateState(routerName)
	state.mu.RLock()
	defer state.mu.RUnlock()
	return state.state
}

// GetStats returns statistics for all circuit breakers
func (rcb *RouterCircuitBreaker) GetStats() map[string]map[string]interface{} {
	rcb.mu.RLock()
	defer rcb.mu.RUnlock()

	stats := make(map[string]map[string]interface{})

	for routerName, state := range rcb.states {
		state.mu.RLock()
		stats[routerName] = map[string]interface{}{
			"state":             state.state.String(),
			"failures":          state.failures,
			"last_failure_time": state.lastFailureTime,
			"last_success_time": state.lastSuccessTime,
		}
		state.mu.RUnlock()
	}

	return stats
}

// Reset resets the circuit breaker for a specific router
func (rcb *RouterCircuitBreaker) Reset(routerName string) {
	state := rcb.getOrCreateState(routerName)
	state.mu.Lock()
	defer state.mu.Unlock()

	state.state = StateClosed
	state.failures = 0
	state.halfOpenRequests = 0
	state.lastSuccessTime = time.Now()

	rcb.breaker.logger.Infof("🔄 Circuit breaker for %s manually reset", routerName)
}

// IsAvailable checks if a router is available (circuit not open)
func (rcb *RouterCircuitBreaker) IsAvailable(routerName string) bool {
	state := rcb.getOrCreateState(routerName)
	state.mu.RLock()
	defer state.mu.RUnlock()

	if state.state == StateOpen {
		// Check if timeout has elapsed
		if time.Since(state.lastFailureTime) >= rcb.breaker.timeout {
			return true // Will transition to half-open on next call
		}
		return false
	}

	return true
}
