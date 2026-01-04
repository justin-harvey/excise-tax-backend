package xrpl

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// ErrNotConnected indicates the client is not connected to XRPL
	ErrNotConnected = errors.New("not connected to XRPL")
	// ErrInvalidAddress indicates an invalid XRPL address
	ErrInvalidAddress = errors.New("invalid XRPL address")
	// ErrTimeout indicates a request timeout
	ErrTimeout = errors.New("request timeout")
)

// Client represents an XRPL WebSocket client
type Client struct {
	url       string
	conn      *websocket.Conn
	mu        sync.RWMutex
	writeMu   sync.Mutex // Separate mutex for WebSocket writes
	nextID    int
	responses map[int]chan Response
	streams   map[string]chan StreamMessage
	done      chan struct{}
	streamsMu sync.RWMutex
}

// New creates a new XRPL client with the given WebSocket URL
func New(url string) *Client {
	return &Client{
		url:       url,
		responses: make(map[int]chan Response),
		streams:   make(map[string]chan StreamMessage),
		done:      make(chan struct{}),
	}
}

// Connect establishes a WebSocket connection to XRPL
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil // Already connected
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, c.url, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", c.url, err)
	}

	c.conn = conn
	c.done = make(chan struct{})

	// Start reading responses in background
	go c.readLoop()

	return nil
}

// Close closes the WebSocket connection
func (c *Client) Close() error {
	c.mu.Lock()

	if c.conn == nil {
		c.mu.Unlock()
		return nil
	}

	// Close the done channel first to signal readLoop
	close(c.done)

	// Close the connection
	err := c.conn.Close()
	c.conn = nil

	c.mu.Unlock()

	// Give readLoop a moment to exit
	time.Sleep(10 * time.Millisecond)

	return err
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}

// GetAccountInfo retrieves account information for the given address
func (c *Client) GetAccountInfo(ctx context.Context, address string) (*AccountInfo, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	// Basic address validation (XRPL addresses start with 'r')
	if len(address) == 0 || address[0] != 'r' {
		return nil, ErrInvalidAddress
	}

	req := map[string]interface{}{
		"command":      "account_info",
		"account":      address,
		"ledger_index": "validated",
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("request failed: %s", resp.Error)
	}

	// Extract account_data from result
	accountData, ok := resp.Result["account_data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid response format")
	}

	info := &AccountInfo{}

	if account, ok := accountData["Account"].(string); ok {
		info.Account = account
	}
	if balance, ok := accountData["Balance"].(string); ok {
		info.Balance = balance
	}
	if seq, ok := accountData["Sequence"].(float64); ok {
		info.Sequence = int64(seq)
	}
	if count, ok := accountData["OwnerCount"].(float64); ok {
		info.OwnerCount = int(count)
	}
	if txn, ok := accountData["PreviousTxnID"].(string); ok {
		info.PreviousTxn = txn
	}

	return info, nil
}

// GetTransaction retrieves a transaction by its hash
func (c *Client) GetTransaction(ctx context.Context, hash string) (*TxResult, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	// Basic hash validation (64 hex characters)
	if len(hash) != 64 {
		return nil, fmt.Errorf("invalid transaction hash: must be 64 hex characters")
	}

	req := map[string]interface{}{
		"command":     "tx",
		"transaction": hash,
		"binary":      false,
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("request failed: %s", resp.Error)
	}

	result := &TxResult{}

	if hash, ok := resp.Result["hash"].(string); ok {
		result.Hash = hash
	}
	if validated, ok := resp.Result["validated"].(bool); ok {
		result.Validated = validated
	}

	// Extract transaction details
	tx := Transaction{}
	if hash, ok := resp.Result["hash"].(string); ok {
		tx.Hash = hash
	}
	if txType, ok := resp.Result["TransactionType"].(string); ok {
		tx.TransactionType = txType
	}
	if account, ok := resp.Result["Account"].(string); ok {
		tx.Account = account
	}
	if dest, ok := resp.Result["Destination"].(string); ok {
		tx.Destination = dest
	}
	if amount := resp.Result["Amount"]; amount != nil {
		tx.Amount = amount
	}
	if fee, ok := resp.Result["Fee"].(string); ok {
		tx.Fee = fee
	}
	if date, ok := resp.Result["date"].(float64); ok {
		tx.Date = int64(date)
	}
	if validated, ok := resp.Result["validated"].(bool); ok {
		tx.Validated = validated
	}

	result.Tx = tx
	result.Status = "validated"
	if !result.Validated {
		result.Status = "pending"
	}

	return result, nil
}

// GetAccountTransactions retrieves recent transactions for an account
func (c *Client) GetAccountTransactions(ctx context.Context, address string, limit int) ([]Transaction, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	// Basic address validation
	if len(address) == 0 || address[0] != 'r' {
		return nil, ErrInvalidAddress
	}

	// Limit bounds
	if limit <= 0 {
		limit = 10
	}
	if limit > 200 {
		limit = 200
	}

	req := map[string]interface{}{
		"command": "account_tx",
		"account": address,
		"limit":   limit,
		"binary":  false,
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("request failed: %s", resp.Error)
	}

	// Extract transactions array
	txsArray, ok := resp.Result["transactions"].([]interface{})
	if !ok {
		return []Transaction{}, nil
	}

	transactions := make([]Transaction, 0, len(txsArray))
	for _, txItem := range txsArray {
		txMap, ok := txItem.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract tx object
		txData, ok := txMap["tx"].(map[string]interface{})
		if !ok {
			continue
		}

		tx := Transaction{}
		if hash, ok := txData["hash"].(string); ok {
			tx.Hash = hash
		}
		if txType, ok := txData["TransactionType"].(string); ok {
			tx.TransactionType = txType
		}
		if account, ok := txData["Account"].(string); ok {
			tx.Account = account
		}
		if dest, ok := txData["Destination"].(string); ok {
			tx.Destination = dest
		}
		if amount := txData["Amount"]; amount != nil {
			tx.Amount = amount
		}
		if fee, ok := txData["Fee"].(string); ok {
			tx.Fee = fee
		}
		if date, ok := txData["date"].(float64); ok {
			tx.Date = int64(date)
		}

		// Get validated status from meta
		if validated, ok := txMap["validated"].(bool); ok {
			tx.Validated = validated
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// SubscribeToAccount subscribes to transaction streams for an account
func (c *Client) SubscribeToAccount(ctx context.Context, address string) (<-chan StreamMessage, error) {
	if !c.IsConnected() {
		return nil, ErrNotConnected
	}

	// Basic address validation
	if len(address) == 0 || address[0] != 'r' {
		return nil, ErrInvalidAddress
	}

	// Create channel for streaming messages
	streamChan := make(chan StreamMessage, 100)

	// Register stream channel
	c.streamsMu.Lock()
	c.streams[address] = streamChan
	c.streamsMu.Unlock()

	// Send subscribe request
	req := map[string]interface{}{
		"command":  "subscribe",
		"accounts": []string{address},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		// Clean up on error
		c.streamsMu.Lock()
		delete(c.streams, address)
		c.streamsMu.Unlock()
		close(streamChan)
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	if resp.Status != "success" {
		// Clean up on error
		c.streamsMu.Lock()
		delete(c.streams, address)
		c.streamsMu.Unlock()
		close(streamChan)
		return nil, fmt.Errorf("subscribe failed: %s", resp.Error)
	}

	return streamChan, nil
}

// Unsubscribe unsubscribes from transaction streams for an account
func (c *Client) Unsubscribe(ctx context.Context, address string) error {
	if !c.IsConnected() {
		return ErrNotConnected
	}

	// Send unsubscribe request
	req := map[string]interface{}{
		"command":  "unsubscribe",
		"accounts": []string{address},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	if resp.Status != "success" {
		return fmt.Errorf("unsubscribe failed: %s", resp.Error)
	}

	// Clean up stream channel
	c.streamsMu.Lock()
	if ch, ok := c.streams[address]; ok {
		close(ch)
		delete(c.streams, address)
	}
	c.streamsMu.Unlock()

	return nil
}

// sendRequest sends a request and waits for the response
func (c *Client) sendRequest(ctx context.Context, req map[string]interface{}) (*Response, error) {
	c.mu.Lock()

	// Assign unique ID to request
	c.nextID++
	id := c.nextID
	req["id"] = id

	// Create response channel
	respChan := make(chan Response, 1)
	c.responses[id] = respChan

	c.mu.Unlock()

	// Clean up on exit
	defer func() {
		c.mu.Lock()
		delete(c.responses, id)
		c.mu.Unlock()
	}()

	// Send request (protect WebSocket writes with separate mutex)
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, ErrNotConnected
	}

	// Serialize WebSocket writes
	c.writeMu.Lock()
	err := conn.WriteJSON(req)
	c.writeMu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response with timeout
	timeout := 10 * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case resp := <-respChan:
		return &resp, nil
	case <-timer.C:
		return nil, ErrTimeout
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.done:
		return nil, ErrNotConnected
	}
}

// readLoop continuously reads messages from the WebSocket
func (c *Client) readLoop() {
	defer func() {
		// Recover from any panics
		if r := recover(); r != nil {
			// Connection was closed, that's okay
		}
	}()

	for {
		// Check if we should exit
		c.mu.RLock()
		conn := c.conn
		done := c.done
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		select {
		case <-done:
			return
		default:
			// Read raw message
			var rawMsg map[string]interface{}
			if err := conn.ReadJSON(&rawMsg); err != nil {
				// Connection closed or error
				return
			}

			// Check if it's a regular response or stream message
			if msgType, ok := rawMsg["type"].(string); ok && msgType == "transaction" {
				// Handle streaming transaction
				c.handleStreamMessage(rawMsg)
			} else if id, ok := rawMsg["id"].(float64); ok {
				// Handle regular response
				resp := Response{
					ID:     int(id),
					Status: getStringField(rawMsg, "status"),
					Type:   getStringField(rawMsg, "type"),
					Error:  getStringField(rawMsg, "error"),
				}
				if result, ok := rawMsg["result"].(map[string]interface{}); ok {
					resp.Result = result
				}

				c.mu.RLock()
				if ch, ok := c.responses[resp.ID]; ok {
					select {
					case ch <- resp:
					default:
					}
				}
				c.mu.RUnlock()
			}
		}
	}
}

// handleStreamMessage processes streaming messages from subscriptions
func (c *Client) handleStreamMessage(rawMsg map[string]interface{}) {
	msg := StreamMessage{
		Type:                getStringField(rawMsg, "type"),
		Validated:           getBoolField(rawMsg, "validated"),
		Status:              getStringField(rawMsg, "status"),
		EngineResult:        getStringField(rawMsg, "engine_result"),
		EngineResultMessage: getStringField(rawMsg, "engine_result_message"),
	}

	// Extract transaction details if present
	if txData, ok := rawMsg["transaction"].(map[string]interface{}); ok {
		tx := &Transaction{
			Hash:            getStringField(txData, "hash"),
			TransactionType: getStringField(txData, "TransactionType"),
			Account:         getStringField(txData, "Account"),
			Destination:     getStringField(txData, "Destination"),
			Fee:             getStringField(txData, "Fee"),
		}
		if amount := txData["Amount"]; amount != nil {
			tx.Amount = amount
		}
		if date, ok := txData["date"].(float64); ok {
			tx.Date = int64(date)
		}
		if seq, ok := txData["Sequence"].(float64); ok {
			tx.Sequence = int64(seq)
		}
		tx.Validated = msg.Validated
		msg.Transaction = tx
	}

	// Route to subscription channels based on account
	if msg.Transaction != nil {
		account := msg.Transaction.Account
		c.streamsMu.RLock()
		if ch, ok := c.streams[account]; ok {
			select {
			case ch <- msg:
			default:
				// Channel full, drop message
			}
		}
		c.streamsMu.RUnlock()
	}
}

// Helper functions to extract fields from raw messages
func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getBoolField(m map[string]interface{}, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

// DropsToXRP converts drops (smallest unit) to XRP
func DropsToXRP(drops string) (string, error) {
	// Parse drops as integer
	var d int64
	_, err := fmt.Sscanf(drops, "%d", &d)
	if err != nil {
		return "", err
	}

	// Validate that drops is non-negative
	if d < 0 {
		return "", fmt.Errorf("drops cannot be negative: %d", d)
	}

	// Convert to XRP (1 XRP = 1,000,000 drops)
	xrp := float64(d) / 1_000_000.0
	return fmt.Sprintf("%.6f", xrp), nil
}
