package xrpl

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Connection retry settings
	initialRetryDelay = 1 * time.Second
	maxRetryDelay     = 30 * time.Second
	retryMultiplier   = 2.0

	// Health check settings
	pingInterval = 30 * time.Second
	pongTimeout  = 10 * time.Second
	writeTimeout = 10 * time.Second
)

// ConnectionManager handles connection lifecycle with reconnection and health checks
type ConnectionManager struct {
	client        *Client
	ctx           context.Context
	cancel        context.CancelFunc
	retryCount    int
	lastConnected time.Time
	mu            sync.RWMutex
	healthTicker  *time.Ticker
}

// NewConnectionManager creates a new connection manager for the client
func (c *Client) NewConnectionManager(ctx context.Context) *ConnectionManager {
	ctx, cancel := context.WithCancel(ctx)
	return &ConnectionManager{
		client: c,
		ctx:    ctx,
		cancel: cancel,
	}
}

// ConnectWithRetry establishes connection with automatic retry on failure
func (cm *ConnectionManager) ConnectWithRetry() error {
	for {
		err := cm.client.Connect(cm.ctx)
		if err == nil {
			cm.mu.Lock()
			cm.retryCount = 0
			cm.lastConnected = time.Now()
			cm.mu.Unlock()

			// Start health checks
			go cm.healthCheck()
			return nil
		}

		// Check if context is cancelled
		select {
		case <-cm.ctx.Done():
			return cm.ctx.Err()
		default:
		}

		// Calculate backoff delay
		cm.mu.Lock()
		cm.retryCount++
		retryCount := cm.retryCount
		cm.mu.Unlock()

		delay := cm.calculateBackoff(retryCount)

		select {
		case <-cm.ctx.Done():
			return cm.ctx.Err()
		case <-time.After(delay):
			// Continue to next retry
		}
	}
}

// calculateBackoff returns exponential backoff delay
func (cm *ConnectionManager) calculateBackoff(attempt int) time.Duration {
	delay := float64(initialRetryDelay) * math.Pow(retryMultiplier, float64(attempt-1))

	if delay > float64(maxRetryDelay) {
		delay = float64(maxRetryDelay)
	}

	return time.Duration(delay)
}

// healthCheck performs periodic ping/pong health checks
func (cm *ConnectionManager) healthCheck() {
	cm.healthTicker = time.NewTicker(pingInterval)
	defer cm.healthTicker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-cm.client.done:
			return
		case <-cm.healthTicker.C:
			if err := cm.ping(); err != nil {
				// Connection unhealthy, attempt reconnect
				go cm.handleDisconnect()
				return
			}
		}
	}
}

// ping sends a ping message and waits for pong
func (cm *ConnectionManager) ping() error {
	cm.client.mu.Lock()
	conn := cm.client.conn
	cm.client.mu.Unlock()

	if conn == nil {
		return ErrNotConnected
	}

	// Set pong handler
	pongReceived := make(chan struct{}, 1)
	conn.SetPongHandler(func(appData string) error {
		select {
		case pongReceived <- struct{}{}:
		default:
		}
		return nil
	})

	// Set write deadline
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Send ping
	if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}

	// Wait for pong
	select {
	case <-pongReceived:
		return nil
	case <-time.After(pongTimeout):
		return fmt.Errorf("pong timeout")
	case <-cm.ctx.Done():
		return cm.ctx.Err()
	}
}

// handleDisconnect handles connection loss and attempts reconnect
func (cm *ConnectionManager) handleDisconnect() {
	cm.client.mu.Lock()
	if cm.client.conn != nil {
		cm.client.conn.Close()
		cm.client.conn = nil
	}
	cm.client.mu.Unlock()

	// Attempt to reconnect
	_ = cm.ConnectWithRetry()
}

// Shutdown gracefully shuts down the connection with timeout
func (cm *ConnectionManager) Shutdown(timeout time.Duration) error {
	// Cancel context to stop health checks and retries
	cm.cancel()

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- cm.client.Close()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// RetryCount returns the current retry count
func (cm *ConnectionManager) RetryCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.retryCount
}

// LastConnected returns the time of last successful connection
func (cm *ConnectionManager) LastConnected() time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.lastConnected
}
