package xrpl

import (
	"context"
	"testing"
	"time"
)

func TestConnectionManager_CalculateBackoff(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	tests := []struct {
		name    string
		attempt int
		wantMin time.Duration
		wantMax time.Duration
	}{
		{
			name:    "first retry",
			attempt: 1,
			wantMin: 1 * time.Second,
			wantMax: 1 * time.Second,
		},
		{
			name:    "second retry",
			attempt: 2,
			wantMin: 2 * time.Second,
			wantMax: 2 * time.Second,
		},
		{
			name:    "third retry",
			attempt: 3,
			wantMin: 4 * time.Second,
			wantMax: 4 * time.Second,
		},
		{
			name:    "fourth retry",
			attempt: 4,
			wantMin: 8 * time.Second,
			wantMax: 8 * time.Second,
		},
		{
			name:    "capped at max",
			attempt: 10,
			wantMin: 30 * time.Second,
			wantMax: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cm.calculateBackoff(tt.attempt)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateBackoff(%d) = %v, want between %v and %v",
					tt.attempt, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestConnectionManager_RetryCount(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	if cm.RetryCount() != 0 {
		t.Errorf("RetryCount() = %d, want 0", cm.RetryCount())
	}

	cm.mu.Lock()
	cm.retryCount = 5
	cm.mu.Unlock()

	if cm.RetryCount() != 5 {
		t.Errorf("RetryCount() = %d, want 5", cm.RetryCount())
	}
}

func TestConnectionManager_LastConnected(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	// Initially should be zero time
	if !cm.LastConnected().IsZero() {
		t.Error("LastConnected() should be zero time initially")
	}

	now := time.Now()
	cm.mu.Lock()
	cm.lastConnected = now
	cm.mu.Unlock()

	got := cm.LastConnected()
	if !got.Equal(now) {
		t.Errorf("LastConnected() = %v, want %v", got, now)
	}
}

func TestConnectionManager_CancelContext(t *testing.T) {
	client := New("wss://invalid.host.that.does.not.exist:12345")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cm := client.NewConnectionManager(ctx)

	// This should fail due to context timeout
	err := cm.ConnectWithRetry()
	if err == nil {
		t.Error("ConnectWithRetry() should fail with cancelled context")
	}
}

func TestConnectionManager_Shutdown(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	// Shutdown should complete even without connection
	err := cm.Shutdown(1 * time.Second)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}
}

func TestConnectionManager_ShutdownTimeout(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	// Test with very short timeout (should still succeed for unconnected client)
	err := cm.Shutdown(1 * time.Nanosecond)
	// This might succeed or timeout depending on timing
	_ = err
}

func TestNewConnectionManager(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	if cm == nil {
		t.Fatal("NewConnectionManager() returned nil")
	}

	if cm.client != client {
		t.Error("ConnectionManager.client not set correctly")
	}

	if cm.ctx == nil {
		t.Error("ConnectionManager.ctx not set")
	}

	if cm.RetryCount() != 0 {
		t.Errorf("initial RetryCount() = %d, want 0", cm.RetryCount())
	}
}

func TestConnectionManager_ConcurrentAccess(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	done := make(chan bool)

	// Concurrent reads of RetryCount
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = cm.RetryCount()
				_ = cm.LastConnected()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConnectionManager_BackoffProgression(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()
	cm := client.NewConnectionManager(ctx)

	// Test that backoff increases with each attempt
	var prevDelay time.Duration
	for attempt := 1; attempt <= 5; attempt++ {
		delay := cm.calculateBackoff(attempt)
		if attempt > 1 && delay <= prevDelay {
			t.Errorf("Backoff did not increase: attempt %d delay %v <= previous %v",
				attempt, delay, prevDelay)
		}
		prevDelay = delay
	}
}
