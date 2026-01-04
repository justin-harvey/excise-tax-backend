package xrpl

// AccountInfo represents basic account information from XRPL
type AccountInfo struct {
	Account           string `json:"Account"`
	Balance           int64  `json:"Balance"` // In drops (1 XRP = 1,000,000 drops)
	Sequence          int64  `json:"Sequence"`
	OwnerCount        int    `json:"OwnerCount"`
	Flags             int64  `json:"Flags,omitempty"`
	PreviousTxnID     string `json:"PreviousTxnID,omitempty"`
	PreviousTxnLgrSeq int64  `json:"PreviousTxnLgrSeq,omitempty"`
}

// Transaction represents an XRPL transaction
type Transaction struct {
	Hash            string                 `json:"hash"`
	LedgerIndex     int64                  `json:"ledger_index"`
	Date            int64                  `json:"date"` // Ripple epoch timestamp
	TransactionType string                 `json:"TransactionType"`
	Account         string                 `json:"Account"`
	Destination     string                 `json:"Destination,omitempty"`
	Amount          interface{}            `json:"Amount,omitempty"` // Can be string (XRP) or object (IOU)
	Fee             string                 `json:"Fee"`
	Sequence        int64                  `json:"Sequence,omitempty"`
	Meta            map[string]interface{} `json:"meta,omitempty"`
	Validated       bool                   `json:"validated"`
}

// TxResult represents the result of a transaction query
type TxResult struct {
	Hash        string      `json:"hash"`
	Validated   bool        `json:"validated"`
	Status      string      `json:"status"` // "validated", "pending", or "failed"
	LedgerIndex int64       `json:"ledger_index,omitempty"`
	Tx          Transaction `json:"transaction"`
}

// StreamMessage represents a streaming message from XRPL subscriptions
type StreamMessage struct {
	Type                string                 `json:"type"`
	Transaction         *Transaction           `json:"transaction,omitempty"`
	Meta                map[string]interface{} `json:"meta,omitempty"`
	LedgerIndex         int64                  `json:"ledger_index,omitempty"`
	LedgerHash          string                 `json:"ledger_hash,omitempty"`
	LedgerTime          int64                  `json:"ledger_time,omitempty"`
	Validated           bool                   `json:"validated"`
	Status              string                 `json:"status,omitempty"`
	EngineResult        string                 `json:"engine_result,omitempty"`
	EngineResultMessage string                 `json:"engine_result_message,omitempty"`
}

// Response represents a generic XRPL response
type Response struct {
	ID     int                    `json:"id"`
	Status string                 `json:"status"`
	Type   string                 `json:"type"`
	Result map[string]interface{} `json:"result"`
	Error  string                 `json:"error,omitempty"`
}

// Request represents a generic XRPL request
type Request struct {
	ID      int                    `json:"id"`
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:",inline"`
}
