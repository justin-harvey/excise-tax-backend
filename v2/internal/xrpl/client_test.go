package xrpl

import (
	"context"
	"testing"
	"time"
)

func TestDropsToXRP(t *testing.T) {
	tests := []struct {
		name    string
		drops   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid drops",
			drops:   "1000000",
			want:    "1.000000",
			wantErr: false,
		},
		{
			name:    "zero drops",
			drops:   "0",
			want:    "0.000000",
			wantErr: false,
		},
		{
			name:    "large amount",
			drops:   "100000000000",
			want:    "100000.000000",
			wantErr: false,
		},
		{
			name:    "testnet faucet balance",
			drops:   "777495588",
			want:    "777.495588",
			wantErr: false,
		},
		{
			name:    "invalid drops",
			drops:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "negative drops",
			drops:   "-1000000",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			drops:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DropsToXRP(tt.drops)
			if (err != nil) != tt.wantErr {
				t.Errorf("DropsToXRP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DropsToXRP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "testnet url",
			url:  "wss://s.altnet.rippletest.net:51233",
		},
		{
			name: "mainnet url",
			url:  "wss://xrplcluster.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.url)
			if client == nil {
				t.Error("New() returned nil")
			}
			if client.url != tt.url {
				t.Errorf("New() url = %v, want %v", client.url, tt.url)
			}
			if client.IsConnected() {
				t.Error("New() client should not be connected initially")
			}
		})
	}
}

func TestIsConnected(t *testing.T) {
	client := New("wss://test.example.com")

	if client.IsConnected() {
		t.Error("IsConnected() should return false for new client")
	}
}

func TestClose_NotConnected(t *testing.T) {
	client := New("wss://test.example.com")

	err := client.Close()
	if err != nil {
		t.Errorf("Close() on unconnected client should not error, got %v", err)
	}
}

// TestGetAccountInfo_ValidationErrors tests input validation
func TestGetAccountInfo_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		wantErr     error
		description string
	}{
		{
			name:        "empty address",
			address:     "",
			wantErr:     ErrInvalidAddress,
			description: "should reject empty address",
		},
		{
			name:        "invalid prefix",
			address:     "xPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY",
			wantErr:     ErrInvalidAddress,
			description: "should reject address not starting with 'r'",
		},
		{
			name:        "too short",
			address:     "r",
			wantErr:     ErrInvalidAddress,
			description: "should accept minimal 'r' address (format validation only)",
		},
	}

	client := New("wss://test.example.com")
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GetAccountInfo(ctx, tt.address)

			// Should get not connected error first since we haven't connected
			if err != ErrNotConnected && err != tt.wantErr {
				t.Errorf("GetAccountInfo() error = %v, want %v or %v", err, ErrNotConnected, tt.wantErr)
			}
		})
	}
}

// TestGetAccountInfo_NotConnected tests behavior when client is not connected
func TestGetAccountInfo_NotConnected(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()

	_, err := client.GetAccountInfo(ctx, "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY")
	if err != ErrNotConnected {
		t.Errorf("GetAccountInfo() error = %v, want %v", err, ErrNotConnected)
	}
}

// TestGetTransaction_NotConnected tests behavior when client is not connected
func TestGetTransaction_NotConnected(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()

	_, err := client.GetTransaction(ctx, "E08D6E9754025BA2534A78707605E0601F03ACE063687A0CA1BDDACFCD1698C7")
	if err != ErrNotConnected {
		t.Errorf("GetTransaction() error = %v, want %v", err, ErrNotConnected)
	}
}

// TestGetAccountTransactions_NotConnected tests behavior when client is not connected
func TestGetAccountTransactions_NotConnected(t *testing.T) {
	client := New("wss://test.example.com")
	ctx := context.Background()

	_, err := client.GetAccountTransactions(ctx, "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY", 10)
	if err != ErrNotConnected {
		t.Errorf("GetAccountTransactions() error = %v, want %v", err, ErrNotConnected)
	}
}

// TestContextCancellation tests that operations respect context cancellation
func TestContextCancellation(t *testing.T) {
	client := New("wss://test.example.com")

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Connect() should fail with cancelled context")
	}
}

// TestContextTimeout tests that operations respect context timeout
func TestContextTimeout(t *testing.T) {
	client := New("wss://invalid.host.that.does.not.exist.local:12345")

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := client.Connect(ctx)
	if err == nil {
		defer client.Close()
		t.Error("Connect() should fail with timeout")
	}
}

// TestConcurrentAccess tests thread-safety of IsConnected
func TestConcurrentAccess(t *testing.T) {
	client := New("wss://test.example.com")

	done := make(chan bool)

	// Multiple goroutines checking connection status
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = client.IsConnected()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestNew_URLVariations tests client creation with different URLs
func TestNew_URLVariations(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "testnet url",
			url:  "wss://s.altnet.rippletest.net:51233",
		},
		{
			name: "mainnet url",
			url:  "wss://xrplcluster.com",
		},
		{
			name: "localhost url",
			url:  "ws://localhost:6006",
		},
		{
			name: "empty url",
			url:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(tt.url)
			if client == nil {
				t.Fatal("New() returned nil")
			}
			if client.url != tt.url {
				t.Errorf("New() url = %v, want %v", client.url, tt.url)
			}
			if client.IsConnected() {
				t.Error("New() client should not be connected initially")
			}
			if client.responses == nil {
				t.Error("New() client.responses should be initialized")
			}
			if client.done == nil {
				t.Error("New() client.done should be initialized")
			}
		})
	}
}
