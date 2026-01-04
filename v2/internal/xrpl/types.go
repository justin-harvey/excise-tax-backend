package xrpl

// AccountInfo represents basic account information from XRPL
type AccountInfo struct {
	Account     string `json:"Account"`
	Balance     string `json:"Balance"` // In drops (1 XRP = 1,000,000 drops)
	Sequence    int64  `json:"Sequence"`
	OwnerCount  int    `json:"OwnerCount"`
	PreviousTxn string `json:"PreviousTxnID"`
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

// TxResult represents the result of a transaction query
type TxResult struct {
	Hash      string      `json:"hash"`
	Validated bool        `json:"validated"`
	Status    string      `json:"status"`
	Tx        Transaction `json:"transaction"`
}
