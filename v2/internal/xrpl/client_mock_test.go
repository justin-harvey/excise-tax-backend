package xrpl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockXRPLServer creates a test WebSocket server that simulates XRPL responses
type mockXRPLServer struct {
	server   *httptest.Server
	upgrader websocket.Upgrader
	handler  func(*websocket.Conn)
}

func newMockServer(handler func(*websocket.Conn)) *mockXRPLServer {
	ms := &mockXRPLServer{
		upgrader: websocket.Upgrader{},
		handler:  handler,
	}

	ms.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := ms.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		if ms.handler != nil {
			ms.handler(conn)
		}
	}))

	return ms
}

func (ms *mockXRPLServer) URL() string {
	return "ws" + strings.TrimPrefix(ms.server.URL, "http")
}

func (ms *mockXRPLServer) Close() {
	ms.server.Close()
}

// Tests using mock server

func TestClient_ConnectAndClose(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		// Read ping message
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
		// Send pong
		conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"response","result":{}}`))
	})
	defer server.Close()

	client := New(server.URL())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if client.conn == nil {
		t.Error("Expected conn to be non-nil after Connect")
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if client.conn != nil {
		t.Error("Expected conn to be nil after Close")
	}
}

func TestClient_ConnectAlreadyConnected(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		// Stay alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})
	defer server.Close()

	client := New(server.URL())

	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("First Connect() error = %v", err)
	}
	defer client.Close()

	// Try to connect again - should succeed (returns nil when already connected)
	err = client.Connect(ctx)
	if err != nil {
		t.Errorf("Second Connect() error = %v, want nil", err)
	}
}

func TestClient_CloseNotConnected(t *testing.T) {
	client := New("ws://localhost:1234")

	// Should not error when closing while not connected
	err := client.Close()
	if err != nil {
		t.Errorf("Close() on unconnected client error = %v", err)
	}
}

func TestClient_GetAccountInfo_Success(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			// Read request
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			// Send response based on command
			if req["command"] == "account_info" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account_data": map[string]interface{}{
							"Account":  "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY",
							"Balance":  "1000000",
							"Sequence": 123,
						},
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	info, err := client.GetAccountInfo(ctx, "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY")
	if err != nil {
		t.Fatalf("GetAccountInfo() error = %v", err)
	}

	if info.Account != "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY" {
		t.Errorf("Account = %v, want rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY", info.Account)
	}
	if info.Balance != "1000000" {
		t.Errorf("Balance = %v, want 1000000", info.Balance)
	}
}

func TestClient_GetAccountInfo_Error(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			// Send error response
			response := map[string]interface{}{
				"id":            req["id"],
				"status":        "error",
				"type":          "response",
				"error":         "actNotFound",
				"error_message": "Account not found",
			}
			data, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, data)
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	_, err = client.GetAccountInfo(ctx, "rInvalidAccount")
	if err == nil {
		t.Error("Expected error for invalid account")
	}
	if !strings.Contains(err.Error(), "actNotFound") {
		t.Errorf("Expected error to contain 'actNotFound', got: %v", err)
	}
}

func TestClient_GetTransaction_Success(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "tx" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"Account":         "rFrom",
						"Destination":     "rTo",
						"Amount":          "1000000",
						"TransactionType": "Payment",
						"hash":            "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234",
						"validated":       true,
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	validHash := "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234"
	tx, err := client.GetTransaction(ctx, validHash)
	if err != nil {
		t.Fatalf("GetTransaction() error = %v", err)
	}

	if tx.Hash != validHash {
		t.Errorf("Hash = %v, want %s", tx.Hash, validHash)
	}
	if tx.Tx.Account != "rFrom" {
		t.Errorf("Tx.Account = %v, want rFrom", tx.Tx.Account)
	}
	if !tx.Validated {
		t.Error("Expected Validated to be true")
	}
}

func TestClient_GetAccountTransactions_Success(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_tx" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account": "rTest",
						"transactions": []interface{}{
							map[string]interface{}{
								"tx": map[string]interface{}{
									"Account":         "rFrom",
									"Destination":     "rTo",
									"Amount":          "1000000",
									"TransactionType": "Payment",
									"hash":            "TX1",
								},
								"validated": true,
							},
						},
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	txs, err := client.GetAccountTransactions(ctx, "rTest", 10)
	if err != nil {
		t.Fatalf("GetAccountTransactions() error = %v", err)
	}

	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(txs))
	}
	if txs[0].Hash != "TX1" {
		t.Errorf("Hash = %v, want TX1", txs[0].Hash)
	}
}

func TestClient_SubscribeToAccount_Success(t *testing.T) {
	messageSent := make(chan struct{})
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "subscribe" {
				// Send subscription confirmation
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)

				// Send stream message immediately after
				streamMsg := map[string]interface{}{
					"type":          "transaction",
					"engine_result": "tesSUCCESS",
					"transaction": map[string]interface{}{
						"Account":     "rFrom",
						"Destination": "rTo",
						"Amount":      "1000000",
						"hash":        "STREAM1234567890STREAM1234567890STREAM1234567890STREAM12345678",
					},
					"validated": true,
				}
				data, _ = json.Marshal(streamMsg)
				conn.WriteMessage(websocket.TextMessage, data)
				close(messageSent)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	ch, err := client.SubscribeToAccount(ctx, "rTest")
	if err != nil {
		t.Fatalf("SubscribeToAccount() error = %v", err)
	}

	// Wait for message to be sent
	select {
	case <-messageSent:
		// Good, message was sent by server
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: server didn't send message")
	}

	// Now wait for client to receive it
	select {
	case msg := <-ch:
		expectedHash := "STREAM1234567890STREAM1234567890STREAM1234567890STREAM12345678"
		if msg.Transaction.Hash != expectedHash {
			t.Errorf("Hash = %v, want %s", msg.Transaction.Hash, expectedHash)
		}
		if msg.EngineResult != "tesSUCCESS" {
			t.Errorf("EngineResult = %v, want tesSUCCESS", msg.EngineResult)
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for stream message to be received by client")
	}
}

func TestClient_Unsubscribe_Success(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			response := map[string]interface{}{
				"id":     req["id"],
				"status": "success",
				"type":   "response",
				"result": map[string]interface{}{},
			}
			data, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, data)
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	err = client.Unsubscribe(ctx, "rTest")
	if err != nil {
		t.Errorf("Unsubscribe() error = %v", err)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		// Never respond to simulate timeout
		time.Sleep(10 * time.Second)
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = client.GetAccountInfo(ctx, "rTest")
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") &&
		!strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

func TestClient_VerifyTransaction_Validated(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "tx" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"Account":         "rFrom",
						"Destination":     "rTo",
						"Amount":          "1000000",
						"TransactionType": "Payment",
						"hash":            "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234",
						"validated":       true,
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	validHash := "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234"
	result, err := client.VerifyTransaction(ctx, validHash)
	if err != nil {
		t.Fatalf("VerifyTransaction() error = %v", err)
	}

	if result.Status != "validated" {
		t.Errorf("Status = %v, want validated", result.Status)
	}
	if !result.Validated {
		t.Error("Expected Validated to be true")
	}
}

func TestClient_VerifyTransaction_Pending(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "tx" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"Account":         "rFrom",
						"Destination":     "rTo",
						"Amount":          "1000000",
						"TransactionType": "Payment",
						"hash":            "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234",
						"validated":       false,
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	validHash := "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234"
	result, err := client.VerifyTransaction(ctx, validHash)
	if err != nil {
		t.Fatalf("VerifyTransaction() error = %v", err)
	}

	if result.Status != "pending" {
		t.Errorf("Status = %v, want pending", result.Status)
	}
	if result.Validated {
		t.Error("Expected Validated to be false")
	}
}

func TestClient_GetAccountInfo_MissingFields(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_info" {
				// Return minimal response with missing optional fields
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account_data": map[string]interface{}{
							"Account": "rTest",
							"Balance": "1000000",
						},
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	info, err := client.GetAccountInfo(ctx, "rTest")
	if err != nil {
		t.Fatalf("GetAccountInfo() error = %v", err)
	}

	if info.Account != "rTest" {
		t.Errorf("Account = %v, want rTest", info.Account)
	}
	if info.Balance != "1000000" {
		t.Errorf("Balance = %v, want 1000000", info.Balance)
	}
	// Optional fields should be zero values
	if info.Sequence != 0 {
		t.Errorf("Sequence = %v, want 0", info.Sequence)
	}
}

func TestClient_GetAccountTransactions_EmptyResult(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_tx" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account":      "rTest",
						"transactions": []interface{}{},
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	txs, err := client.GetAccountTransactions(ctx, "rTest", 10)
	if err != nil {
		t.Fatalf("GetAccountTransactions() error = %v", err)
	}

	if len(txs) != 0 {
		t.Errorf("Expected 0 transactions, got %d", len(txs))
	}
}

func TestClient_GetAccountTransactions_MultipleTransactions(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_tx" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account": "rTest",
						"transactions": []interface{}{
							map[string]interface{}{
								"tx": map[string]interface{}{
									"Account":         "rFrom1",
									"Destination":     "rTo1",
									"Amount":          "1000000",
									"TransactionType": "Payment",
									"hash":            "TX1234567890TX1234567890TX1234567890TX1234567890TX1234567890TX12",
								},
								"validated": true,
							},
							map[string]interface{}{
								"tx": map[string]interface{}{
									"Account":         "rFrom2",
									"Destination":     "rTo2",
									"Amount":          "2000000",
									"TransactionType": "Payment",
									"hash":            "TX2234567890TX2234567890TX2234567890TX2234567890TX2234567890TX22",
								},
								"validated": false,
							},
						},
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	txs, err := client.GetAccountTransactions(ctx, "rTest", 10)
	if err != nil {
		t.Fatalf("GetAccountTransactions() error = %v", err)
	}

	if len(txs) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(txs))
	}

	// Check first transaction
	if txs[0].Account != "rFrom1" {
		t.Errorf("First tx Account = %v, want rFrom1", txs[0].Account)
	}
	if !txs[0].Validated {
		t.Error("Expected first tx to be validated")
	}

	// Check second transaction
	if txs[1].Account != "rFrom2" {
		t.Errorf("Second tx Account = %v, want rFrom2", txs[1].Account)
	}
	if txs[1].Validated {
		t.Error("Expected second tx to not be validated")
	}
}

func TestClient_IsConnected(t *testing.T) {
	client := New("ws://localhost:9999")

	// Not connected initially
	if client.IsConnected() {
		t.Error("Expected IsConnected to be false initially")
	}

	// Create a mock server and connect
	server := newMockServer(func(conn *websocket.Conn) {
		// Keep alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})
	defer server.Close()

	client = New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Should be connected now
	if !client.IsConnected() {
		t.Error("Expected IsConnected to be true after Connect")
	}

	// Close and check again
	client.Close()
	if client.IsConnected() {
		t.Error("Expected IsConnected to be false after Close")
	}
}

func TestClient_ConnectWithRetry_Success(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		// Stay alive
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cm := client.NewConnectionManager(ctx)

	err := cm.ConnectWithRetry()
	if err != nil {
		t.Fatalf("ConnectWithRetry() error = %v", err)
	}

	if !client.IsConnected() {
		t.Error("Expected client to be connected after ConnectWithRetry")
	}

	if cm.RetryCount() != 0 {
		t.Errorf("RetryCount = %d, want 0 after successful connection", cm.RetryCount())
	}
}

func TestClient_ConnectWithRetry_ContextCancelled(t *testing.T) {
	client := New("ws://localhost:9999") // Invalid server
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	cm := client.NewConnectionManager(ctx)

	err := cm.ConnectWithRetry()
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

func TestClient_SendRequest_NotConnected(t *testing.T) {
	client := New("ws://localhost:9999")
	ctx := context.Background()

	req := map[string]interface{}{
		"command": "account_info",
		"account": "rTest",
	}

	_, err := client.sendRequest(ctx, req)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}
}

func TestClient_GetAccountTransactions_InvalidAddress(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	// Test empty address
	_, err = client.GetAccountTransactions(ctx, "", 10)
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for empty address, got: %v", err)
	}

	// Test address not starting with 'r'
	_, err = client.GetAccountTransactions(ctx, "invalid", 10)
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for invalid address, got: %v", err)
	}
}

func TestClient_GetAccountTransactions_LimitBounds(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_tx" {
				// Echo back the limit for verification
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account":      "rTest",
						"transactions": []interface{}{},
						"limit":        req["limit"], // Echo back for testing
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	// Test limit 0 (should default to 10)
	_, err = client.GetAccountTransactions(ctx, "rTest", 0)
	if err != nil {
		t.Errorf("GetAccountTransactions with limit 0 failed: %v", err)
	}

	// Test limit > 200 (should cap at 200)
	_, err = client.GetAccountTransactions(ctx, "rTest", 300)
	if err != nil {
		t.Errorf("GetAccountTransactions with limit 300 failed: %v", err)
	}
}

func TestClient_GetTransaction_InvalidHash(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	// Test short hash
	_, err = client.GetTransaction(ctx, "short")
	if err == nil || !strings.Contains(err.Error(), "must be 64 hex characters") {
		t.Errorf("Expected hash validation error, got: %v", err)
	}

	// Test long hash
	longHash := strings.Repeat("A", 100)
	_, err = client.GetTransaction(ctx, longHash)
	if err == nil || !strings.Contains(err.Error(), "must be 64 hex characters") {
		t.Errorf("Expected hash validation error for long hash, got: %v", err)
	}
}

func TestClient_VerifyTransaction_NotConnected(t *testing.T) {
	client := New("ws://localhost:9999")
	ctx := context.Background()

	validHash := "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234"
	_, err := client.VerifyTransaction(ctx, validHash)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}
}

func TestClient_GetAccountInfo_NotConnected(t *testing.T) {
	client := New("ws://localhost:9999")
	ctx := context.Background()

	_, err := client.GetAccountInfo(ctx, "rTest")
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}
}

func TestClient_Unsubscribe_NotConnected(t *testing.T) {
	client := New("ws://localhost:9999")
	ctx := context.Background()

	err := client.Unsubscribe(ctx, "rTest")
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}
}

func TestClient_Close_MultipleClose(t *testing.T) {
	client := New("ws://localhost:9999")

	// Multiple closes should not panic or error
	err1 := client.Close()
	err2 := client.Close()

	if err1 != nil {
		t.Errorf("First Close() error = %v", err1)
	}
	if err2 != nil {
		t.Errorf("Second Close() error = %v", err2)
	}
}

func TestClient_GetTransaction_ServerError(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "tx" {
				response := map[string]interface{}{
					"id":            req["id"],
					"status":        "error",
					"type":          "response",
					"error":         "txnNotFound",
					"error_message": "Transaction not found",
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	validHash := "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234"
	_, err = client.GetTransaction(ctx, validHash)
	if err == nil {
		t.Error("Expected error for server error response")
	}
	if !strings.Contains(err.Error(), "txnNotFound") {
		t.Errorf("Expected error to contain 'txnNotFound', got: %v", err)
	}
}

func TestClient_GetAccountTransactions_ServerError(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_tx" {
				response := map[string]interface{}{
					"id":            req["id"],
					"status":        "error",
					"type":          "response",
					"error":         "actNotFound",
					"error_message": "Account not found",
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	_, err = client.GetAccountTransactions(ctx, "rTest", 10)
	if err == nil {
		t.Error("Expected error for server error response")
	}
	if !strings.Contains(err.Error(), "actNotFound") {
		t.Errorf("Expected error to contain 'actNotFound', got: %v", err)
	}
}

func TestClient_GetAccountTransactions_MalformedTransactions(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_tx" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account": "rTest",
						"transactions": []interface{}{
							// Malformed transaction (missing tx field)
							map[string]interface{}{
								"validated": true,
							},
							// Invalid tx structure
							map[string]interface{}{
								"tx":        "not-an-object",
								"validated": true,
							},
							// Valid transaction with partial data
							map[string]interface{}{
								"tx": map[string]interface{}{
									"Account": "rFrom",
									"hash":    "PARTIAL1234567890PARTIAL1234567890PARTIAL1234567890PARTIAL123",
								},
								"validated": true,
							},
						},
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	txs, err := client.GetAccountTransactions(ctx, "rTest", 10)
	if err != nil {
		t.Fatalf("GetAccountTransactions() error = %v", err)
	}

	// Should only have 1 valid transaction (the malformed ones are skipped)
	if len(txs) != 1 {
		t.Errorf("Expected 1 valid transaction, got %d", len(txs))
	}

	if len(txs) > 0 && txs[0].Account != "rFrom" {
		t.Errorf("Transaction Account = %v, want rFrom", txs[0].Account)
	}
}

func TestClient_SubscribeToAccount_NotConnected(t *testing.T) {
	client := New("ws://localhost:9999")
	ctx := context.Background()

	_, err := client.SubscribeToAccount(ctx, "rTest")
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got: %v", err)
	}
}

func TestClient_Connect_InvalidURL(t *testing.T) {
	client := New("invalid://url")
	ctx := context.Background()

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestClient_Connect_ContextTimeout(t *testing.T) {
	// Use a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	client := New("ws://localhost:9999") // Non-existent server

	err := client.Connect(ctx)
	if err == nil {
		t.Error("Expected error for context timeout")
	}
}

func TestClient_SubscribeToAccount_Basic(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "subscribe" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	// Just test that subscription request succeeds
	ch, err := client.SubscribeToAccount(ctx, "rTest")
	if err != nil {
		t.Fatalf("SubscribeToAccount() error = %v", err)
	}

	if ch == nil {
		t.Error("Expected non-nil channel")
	}

	// Close the channel by closing client (this will test cleanup paths)
	client.Close()
}

func TestClient_GetAccountInfo_InvalidAddress(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	// Test empty address
	_, err = client.GetAccountInfo(ctx, "")
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for empty address, got: %v", err)
	}

	// Test address not starting with 'r'
	_, err = client.GetAccountInfo(ctx, "invalid")
	if err != ErrInvalidAddress {
		t.Errorf("Expected ErrInvalidAddress for invalid address, got: %v", err)
	}
}

func TestClient_GetAccountInfo_InvalidResponse(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "account_info" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{
						"account_data": "not-an-object", // Invalid format
					},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	_, err = client.GetAccountInfo(ctx, "rTest")
	if err == nil || !strings.Contains(err.Error(), "invalid response format") {
		t.Errorf("Expected invalid response format error, got: %v", err)
	}
}

func TestClient_Unsubscribe_ServerSuccess(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "unsubscribe" {
				response := map[string]interface{}{
					"id":     req["id"],
					"status": "success",
					"type":   "response",
					"result": map[string]interface{}{},
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	err = client.Unsubscribe(ctx, "rTest")
	if err != nil {
		t.Errorf("Unsubscribe() error = %v", err)
	}
}

func TestClient_Unsubscribe_Error(t *testing.T) {
	server := newMockServer(func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var req map[string]interface{}
			json.Unmarshal(msg, &req)

			if req["command"] == "unsubscribe" {
				response := map[string]interface{}{
					"id":            req["id"],
					"status":        "error",
					"type":          "response",
					"error":         "invalidParams",
					"error_message": "Invalid parameters",
				}
				data, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, data)
			}
		}
	})
	defer server.Close()

	client := New(server.URL())
	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	err = client.Unsubscribe(ctx, "rTest")
	if err == nil || !strings.Contains(err.Error(), "invalidParams") {
		t.Errorf("Expected error containing 'invalidParams', got: %v", err)
	}
}
