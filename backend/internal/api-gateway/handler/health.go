// Package handler provides HTTP handlers for the API Gateway.
package handler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status       HealthStatus              `json:"status"`
	Services     map[string]HealthStatus   `json:"services"`
	Dependencies map[string]HealthStatus   `json:"dependencies"`
	Timestamp    time.Time                 `json:"timestamp"`
}

// ServiceHealthChecker is an interface for checking service health
type ServiceHealthChecker interface {
	HealthCheck(serviceName string) error
	GetServiceHealth() map[string]bool
}

// RedisHealthChecker is an interface for checking Redis health
type RedisHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// HealthHandler handles health check requests
type HealthHandler struct {
	serviceProxy ServiceHealthChecker
	redisClient  RedisHealthChecker
	logger       *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(serviceProxy ServiceHealthChecker, redisClient RedisHealthChecker, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		serviceProxy: serviceProxy,
		redisClient:  redisClient,
		logger:       logger,
	}
}

// Health handles GET /health endpoint
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:       HealthStatusHealthy,
		Services:     make(map[string]HealthStatus),
		Dependencies: make(map[string]HealthStatus),
		Timestamp:    time.Now(),
	}

	// Use WaitGroup to check all services in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Check backend services
	serviceNames := []string{"auth", "payment", "tax", "reporting"}
	for _, serviceName := range serviceNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			status := h.checkServiceHealth(ctx, name)

			mu.Lock()
			response.Services[name] = status
			if status != HealthStatusHealthy {
				response.Status = HealthStatusDegraded
			}
			mu.Unlock()
		}(serviceName)
	}

	// Check Redis
	wg.Add(1)
	go func() {
		defer wg.Done()

		status := h.checkRedisHealth(ctx)

		mu.Lock()
		response.Dependencies["redis"] = status
		if status != HealthStatusHealthy {
			response.Status = HealthStatusDegraded
		}
		mu.Unlock()
	}()

	// Wait for all checks to complete
	wg.Wait()

	// Determine HTTP status code
	httpStatus := http.StatusOK
	if response.Status == HealthStatusDegraded {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, response)
}

// checkServiceHealth checks the health of a backend service
func (h *HealthHandler) checkServiceHealth(ctx context.Context, serviceName string) HealthStatus {
	// Create a channel for the result
	resultChan := make(chan error, 1)

	go func() {
		err := h.serviceProxy.HealthCheck(serviceName)
		resultChan <- err
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		h.logger.Warn("Service health check timed out",
			zap.String("service", serviceName),
		)
		return HealthStatusUnhealthy
	case err := <-resultChan:
		if err != nil {
			h.logger.Debug("Service health check failed",
				zap.String("service", serviceName),
				zap.Error(err),
			)
			return HealthStatusUnhealthy
		}
		return HealthStatusHealthy
	}
}

// checkRedisHealth checks the health of Redis
func (h *HealthHandler) checkRedisHealth(ctx context.Context) HealthStatus {
	// Create a channel for the result
	resultChan := make(chan error, 1)

	go func() {
		err := h.redisClient.HealthCheck(ctx)
		resultChan <- err
	}()

	// Wait for result or timeout
	select {
	case <-ctx.Done():
		h.logger.Warn("Redis health check timed out")
		return HealthStatusUnhealthy
	case err := <-resultChan:
		if err != nil {
			h.logger.Debug("Redis health check failed",
				zap.Error(err),
			)
			return HealthStatusUnhealthy
		}
		return HealthStatusHealthy
	}
}

// Liveness handles GET /health/liveness endpoint (simple check for Kubernetes)
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Readiness handles GET /health/readiness endpoint (checks dependencies)
func (h *HealthHandler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check Redis connectivity
	if err := h.redisClient.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "redis unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
