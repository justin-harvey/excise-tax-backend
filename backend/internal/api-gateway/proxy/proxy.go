// Package proxy provides HTTP reverse proxy functionality for routing requests to backend services.
package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ServiceConfig holds configuration for a backend service
type ServiceConfig struct {
	Name    string
	BaseURL string
	Timeout time.Duration
	// HealthCheckPath is the path to use for health checks
	HealthCheckPath string
}

// ProxyConfig holds the proxy configuration
type ProxyConfig struct {
	Services       map[string]ServiceConfig
	DefaultTimeout time.Duration
	MaxIdleConns   int
	IdleConnTimeout time.Duration
}

// ServiceProxy handles proxying requests to backend microservices
type ServiceProxy struct {
	config     ProxyConfig
	httpClient *http.Client
	logger     *zap.Logger
	mu         sync.RWMutex
	healthy    map[string]bool // Track health status of services
}

// NewServiceProxy creates a new service proxy instance
func NewServiceProxy(config ProxyConfig, logger *zap.Logger) *ServiceProxy {
	// Create HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConns / len(config.Services),
		IdleConnTimeout:     config.IdleConnTimeout,
		DisableKeepAlives:   false,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   config.DefaultTimeout,
	}

	proxy := &ServiceProxy{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		healthy:    make(map[string]bool),
	}

	// Initialize all services as healthy
	for name := range config.Services {
		proxy.healthy[name] = true
	}

	return proxy
}

// ProxyRequest proxies a request to the specified backend service
func (sp *ServiceProxy) ProxyRequest(serviceName string, stripPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get service configuration
		service, exists := sp.config.Services[serviceName]
		if !exists {
			sp.logger.Error("Service not found",
				zap.String("service", serviceName),
			)
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "Backend service not available",
				},
			})
			return
		}

		// Check service health
		if !sp.isServiceHealthy(serviceName) {
			sp.logger.Warn("Service is unhealthy",
				zap.String("service", serviceName),
			)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": gin.H{
					"code":    "SERVICE_UNHEALTHY",
					"message": "Backend service is temporarily unavailable",
				},
			})
			return
		}

		// Build target URL
		targetPath := c.Request.URL.Path
		if stripPrefix != "" {
			targetPath = strings.TrimPrefix(targetPath, stripPrefix)
		}

		targetURL, err := url.Parse(service.BaseURL + targetPath)
		if err != nil {
			sp.logger.Error("Failed to parse target URL",
				zap.String("service", serviceName),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to route request",
				},
			})
			return
		}

		// Preserve query parameters
		targetURL.RawQuery = c.Request.URL.RawQuery

		// Create proxied request
		proxyReq, err := http.NewRequestWithContext(
			c.Request.Context(),
			c.Request.Method,
			targetURL.String(),
			c.Request.Body,
		)
		if err != nil {
			sp.logger.Error("Failed to create proxy request",
				zap.String("service", serviceName),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to create proxy request",
				},
			})
			return
		}

		// Copy headers from original request
		sp.copyHeaders(c.Request.Header, proxyReq.Header)

		// Add custom headers
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			proxyReq.Header.Set("X-Request-ID", requestID)
		}

		// Add user context headers if available
		if userID, exists := c.Get("user_id"); exists {
			proxyReq.Header.Set("X-User-ID", fmt.Sprintf("%v", userID))
		}
		if userRole, exists := c.Get("user_role"); exists {
			proxyReq.Header.Set("X-User-Role", fmt.Sprintf("%v", userRole))
		}

		// Set timeout for this request
		timeout := service.Timeout
		if timeout == 0 {
			timeout = sp.config.DefaultTimeout
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		proxyReq = proxyReq.WithContext(ctx)

		// Execute request
		resp, err := sp.httpClient.Do(proxyReq)
		if err != nil {
			sp.logger.Error("Proxy request failed",
				zap.String("service", serviceName),
				zap.String("url", targetURL.String()),
				zap.Error(err),
			)
			c.JSON(http.StatusBadGateway, gin.H{
				"error": gin.H{
					"code":    "BACKEND_ERROR",
					"message": "Failed to communicate with backend service",
				},
			})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Copy response status and body
		c.Status(resp.StatusCode)
		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			sp.logger.Error("Failed to copy response body",
				zap.String("service", serviceName),
				zap.Error(err),
			)
		}
	}
}

// copyHeaders copies HTTP headers from source to destination
func (sp *ServiceProxy) copyHeaders(src, dst http.Header) {
	for key, values := range src {
		// Skip certain headers that shouldn't be forwarded
		if shouldSkipHeader(key) {
			continue
		}
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// shouldSkipHeader determines if a header should be skipped during forwarding
func shouldSkipHeader(header string) bool {
	skipHeaders := map[string]bool{
		"connection":          true,
		"keep-alive":          true,
		"proxy-authenticate":  true,
		"proxy-authorization": true,
		"te":                  true,
		"trailers":            true,
		"transfer-encoding":   true,
		"upgrade":             true,
	}
	return skipHeaders[strings.ToLower(header)]
}

// HealthCheck checks the health of a specific service
func (sp *ServiceProxy) HealthCheck(serviceName string) error {
	service, exists := sp.config.Services[serviceName]
	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	// Build health check URL
	healthURL := service.BaseURL
	if service.HealthCheckPath != "" {
		healthURL = service.BaseURL + service.HealthCheckPath
	} else {
		healthURL = service.BaseURL + "/health"
	}

	// Create request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Execute request
	resp, err := sp.httpClient.Do(req)
	if err != nil {
		sp.setServiceHealth(serviceName, false)
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		sp.setServiceHealth(serviceName, false)
		return fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
	}

	sp.setServiceHealth(serviceName, true)
	return nil
}

// GetServiceHealth returns the health status of all services
func (sp *ServiceProxy) GetServiceHealth() map[string]bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	health := make(map[string]bool)
	for name, status := range sp.healthy {
		health[name] = status
	}
	return health
}

// isServiceHealthy checks if a service is healthy
func (sp *ServiceProxy) isServiceHealthy(serviceName string) bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.healthy[serviceName]
}

// setServiceHealth sets the health status of a service
func (sp *ServiceProxy) setServiceHealth(serviceName string, healthy bool) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.healthy[serviceName] = healthy
}

// StartHealthChecks starts periodic health checks for all services
func (sp *ServiceProxy) StartHealthChecks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Initial health check
	for name := range sp.config.Services {
		go func(serviceName string) {
			if err := sp.HealthCheck(serviceName); err != nil {
				sp.logger.Warn("Initial health check failed",
					zap.String("service", serviceName),
					zap.Error(err),
				)
			}
		}(name)
	}

	// Periodic health checks
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for name := range sp.config.Services {
				go func(serviceName string) {
					if err := sp.HealthCheck(serviceName); err != nil {
						sp.logger.Error("Periodic health check failed",
							zap.String("service", serviceName),
							zap.Error(err),
						)
					}
				}(name)
			}
		}
	}
}
