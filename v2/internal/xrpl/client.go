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
	nextID    int
	responses map[int]chan Response
	done      chan struct{}
}

// New creates a new XRPL client with the given WebSocket URL
func New(url string) *Client {
	return &Client{
		url:       url,
		responses: make(map[int]chan Response),
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
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	close(c.done)
	err := c.conn.Close()
	c.conn = nil
	
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
		"command": "account_info",
		"account": address,
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

	// Send request
	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return nil, ErrNotConnected
	}

	if err := conn.WriteJSON(req); err != nil {
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
	for {
		select {
		case <-c.done:
			return
		default:
			var resp Response
			
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()
			
			if conn == nil {
				return
			}

			if err := conn.ReadJSON(&resp); err != nil {
				// Connection closed or error
				return
			}

			// Route response to waiting request
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

// DropsToXRP converts drops (smallest unit) to XRP
func DropsToXRP(drops string) (string, error) {
	// Parse drops as integer
	var d int64
	_, err := fmt.Sscanf(drops, "%d", &d)
	if err != nil {
		return "", err
	}
	
	// Convert to XRP (1 XRP = 1,000,000 drops)
	xrp := float64(d) / 1_000_000.0
	return fmt.Sprintf("%.6f", xrp), nil
}
