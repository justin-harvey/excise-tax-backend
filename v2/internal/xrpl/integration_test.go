//go:build integration
// +build integration

package xrpl

import (
	"context"
	"os"
	"testing"
	"time"
)

// These tests require a connection to XRPL testnet and should be run with:
// go test -tags=integration ./internal/xrpl

const (
	// Use testnet for integration tests
	testnetURL = "wss://s.altnet.rippletest.net:51233"

	// Known test account with activity (XRPL test faucet funded account)
	testAccount = "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY"

	// Known transaction hash on testnet (replace with a valid one)
	// This should be updated to a recent transaction for reliable testing
	testTxHash = "C71F385124008A436842B56DEF8196B0621762FBD9464F2510EE9C3D1A3322DA"

	// Test timeout
	testTimeout = 30 * time.Second
)

// skipIfShort skips the test if -short flag is provided
func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
}

// setupClient creates and connects a client for testing
func setupClient(t *testing.T, ctx context.Context) *Client {
	t.Helper()

	// Allow override of testnet URL via environment
	url := os.Getenv("XRPL_TEST_URL")
	if url == "" {
		url = testnetURL
	}

	client := New(url)

	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Failed to connect to XRPL: %v", err)
	}

	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close client: %v", err)
		}
	})

	return client
}

func TestIntegration_Connect(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := New(testnetURL)

	// Test connection
	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}

	// Verify connected
	if !client.IsConnected() {
		t.Error("IsConnected() = false, want true after successful connect")
	}

	// Test close
	err = client.Close()
	if err != nil {
		t.Errorf("Close() failed: %v", err)
	}

	// Verify disconnected
	if client.IsConnected() {
		t.Error("IsConnected() = true, want false after close")
	}
}

func TestIntegration_ConnectWithTimeout(t *testing.T) {
	skipIfShort(t)

	// Very short timeout to test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	client := New(testnetURL)

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Connect() with expired context should fail")
		client.Close()
	}
}

func TestIntegration_GetAccountInfo(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// Test valid account
	info, err := client.GetAccountInfo(ctx, testAccount)
	if err != nil {
		t.Fatalf("GetAccountInfo() failed: %v", err)
	}

	// Verify response structure
	if info.Account == "" {
		t.Error("GetAccountInfo() returned empty Account field")
	}
	if info.Balance == "" {
		t.Error("GetAccountInfo() returned empty Balance field")
	}
	if info.Sequence == 0 {
		t.Error("GetAccountInfo() returned zero Sequence")
	}

	t.Logf("Account: %s", info.Account)
	t.Logf("Balance: %s drops", info.Balance)
	t.Logf("Sequence: %d", info.Sequence)
}

func TestIntegration_GetAccountInfo_InvalidAddress(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// Test invalid address
	_, err := client.GetAccountInfo(ctx, "invalid_address")
	if err == nil {
		t.Error("GetAccountInfo() with invalid address should fail")
	}
}

func TestIntegration_GetAccountInfo_NonexistentAccount(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// This is a valid address format but likely doesn't exist
	// XRPL should return an error for accounts that don't exist
	nonExistentAccount := "rN7n7otQDd6FczFgLdlqtyMVrn3HMfXbZA"

	_, err := client.GetAccountInfo(ctx, nonExistentAccount)
	// We expect this to fail (account not found)
	if err == nil {
		t.Log("Note: Test account exists, test may need updating")
	}
}

func TestIntegration_GetTransaction(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// First get recent transactions to find a valid hash
	txs, err := client.GetAccountTransactions(ctx, testAccount, 1)
	if err != nil {
		t.Fatalf("GetAccountTransactions() failed: %v", err)
	}

	if len(txs) == 0 {
		t.Skip("No transactions available for test account")
	}

	txHash := txs[0].Hash

	// Now test getting that transaction
	tx, err := client.GetTransaction(ctx, txHash)
	if err != nil {
		t.Fatalf("GetTransaction() failed: %v", err)
	}

	if tx.Hash != txHash {
		t.Errorf("GetTransaction() hash = %s, want %s", tx.Hash, txHash)
	}

	t.Logf("Transaction: %s", tx.Hash)
	t.Logf("Type: %s", tx.Tx.TransactionType)
	t.Logf("Validated: %v", tx.Validated)
}

func TestIntegration_GetTransaction_InvalidHash(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// Test with completely invalid hash format
	_, err := client.GetTransaction(ctx, "invalid")
	if err == nil {
		t.Error("GetTransaction() with invalid hash should fail")
	}
}

func TestIntegration_VerifyTransaction(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// First get a recent transaction to verify
	txs, err := client.GetAccountTransactions(ctx, testAccount, 1)
	if err != nil {
		t.Fatalf("GetAccountTransactions() failed: %v", err)
	}

	if len(txs) == 0 {
		t.Skip("No transactions available for test account")
	}

	txHash := txs[0].Hash

	// Verify the transaction
	result, err := client.VerifyTransaction(ctx, txHash)
	if err != nil {
		t.Fatalf("VerifyTransaction() failed: %v", err)
	}

	if result.Hash != txHash {
		t.Errorf("VerifyTransaction() hash = %s, want %s", result.Hash, txHash)
	}

	if result.Status == "" {
		t.Error("VerifyTransaction() returned empty status")
	}

	// Validated transactions should have status "validated"
	if result.Validated && result.Status != "validated" {
		t.Errorf("VerifyTransaction() status = %s, want 'validated' for validated tx", result.Status)
	}

	t.Logf("Transaction %s status: %s (validated: %v)", txHash, result.Status, result.Validated)
}

func TestIntegration_VerifyTransaction_InvalidHash(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// Test with invalid hash format
	_, err := client.VerifyTransaction(ctx, "invalid")
	if err == nil {
		t.Error("VerifyTransaction() with invalid hash should fail")
	}
}

func TestIntegration_GetAccountTransactions(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	tests := []struct {
		name  string
		limit int
	}{
		{"default limit", 10},
		{"small limit", 5},
		{"single transaction", 1},
		{"large limit", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txs, err := client.GetAccountTransactions(ctx, testAccount, tt.limit)
			if err != nil {
				t.Fatalf("GetAccountTransactions() failed: %v", err)
			}

			// Note: We might get fewer than limit if account has fewer transactions
			t.Logf("Retrieved %d transactions (requested %d)", len(txs), tt.limit)

			// Verify each transaction has required fields
			for i, tx := range txs {
				if tx.Hash == "" {
					t.Errorf("Transaction %d has empty Hash", i)
				}
				if tx.TransactionType == "" {
					t.Errorf("Transaction %d has empty TransactionType", i)
				}
			}
		})
	}
}

func TestIntegration_GetAccountTransactions_InvalidAddress(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	_, err := client.GetAccountTransactions(ctx, "invalid", 10)
	if err == nil {
		t.Error("GetAccountTransactions() with invalid address should fail")
	}
}

func TestIntegration_SubscribeAndUnsubscribe(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// Subscribe to account
	stream, err := client.SubscribeToAccount(ctx, testAccount)
	if err != nil {
		t.Fatalf("SubscribeToAccount() failed: %v", err)
	}

	// Give it a moment to establish subscription
	time.Sleep(1 * time.Second)

	// Test unsubscribe
	err = client.Unsubscribe(ctx, testAccount)
	if err != nil {
		t.Errorf("Unsubscribe() failed: %v", err)
	}

	// Stream should be closed after unsubscribe
	select {
	case _, ok := <-stream:
		if ok {
			t.Error("Stream should be closed after unsubscribe")
		}
	case <-time.After(2 * time.Second):
		t.Error("Stream was not closed after unsubscribe")
	}
}

func TestIntegration_SubscribeToAccount_InvalidAddress(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	_, err := client.SubscribeToAccount(ctx, "invalid")
	if err == nil {
		t.Error("SubscribeToAccount() with invalid address should fail")
	}
}

func TestIntegration_MultipleRequests(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// Make multiple concurrent requests
	const numRequests = 10

	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			_, err := client.GetAccountInfo(ctx, testAccount)
			errors <- err
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		if err := <-errors; err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}
}

func TestIntegration_ReconnectAfterClose(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := New(testnetURL)

	// Connect
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Initial Connect() failed: %v", err)
	}

	// Close
	if err := client.Close(); err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// Connect again
	if err := client.Connect(ctx); err != nil {
		t.Fatalf("Reconnect() failed: %v", err)
	}

	// Verify it works
	_, err := client.GetAccountInfo(ctx, testAccount)
	if err != nil {
		t.Errorf("GetAccountInfo() after reconnect failed: %v", err)
	}

	client.Close()
}

func TestIntegration_ContextCancellation(t *testing.T) {
	skipIfShort(t)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	client := setupClient(t, ctx)

	// Create a context that we'll cancel
	reqCtx, reqCancel := context.WithCancel(ctx)

	// Start a request in a goroutine
	done := make(chan error, 1)
	go func() {
		_, err := client.GetAccountInfo(reqCtx, testAccount)
		done <- err
	}()

	// Cancel immediately
	reqCancel()

	// Wait for result
	select {
	case err := <-done:
		// We expect either success (if request completed before cancel)
		// or an error related to context cancellation
		if err != nil {
			t.Logf("Request cancelled as expected: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("Request did not complete or get cancelled in time")
	}
}

func TestIntegration_DropsToXRP(t *testing.T) {
	skipIfShort(t)

	tests := []struct {
		drops string
		want  string
	}{
		{"1000000", "1.000000"},
		{"777495588", "777.495588"},
		{"1", "0.000001"},
		{"0", "0.000000"},
	}

	for _, tt := range tests {
		t.Run(tt.drops, func(t *testing.T) {
			got, err := DropsToXRP(tt.drops)
			if err != nil {
				t.Fatalf("DropsToXRP(%s) error = %v", tt.drops, err)
			}
			if got != tt.want {
				t.Errorf("DropsToXRP(%s) = %s, want %s", tt.drops, got, tt.want)
			}
		})
	}
}
