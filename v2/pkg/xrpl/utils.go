package xrpl

import "fmt"

// DropsToXRP converts drops (smallest unit) to XRP.
// 1 XRP = 1,000,000 drops
func DropsToXRP(drops string) (string, error) {
	// Parse drops as integer
	var d int64
	_, err := fmt.Sscanf(drops, "%d", &d)
	if err != nil {
		return "", fmt.Errorf("invalid drops format: %w", err)
	}

	// Validate that drops is non-negative
	if d < 0 {
		return "", fmt.Errorf("drops cannot be negative: %d", d)
	}

	// Convert to XRP (1 XRP = 1,000,000 drops)
	xrp := float64(d) / 1_000_000.0
	return fmt.Sprintf("%.6f", xrp), nil
}

// XRPToDrops converts XRP to drops.
// 1 XRP = 1,000,000 drops
func XRPToDrops(xrp float64) (string, error) {
	if xrp < 0 {
		return "", fmt.Errorf("XRP cannot be negative: %f", xrp)
	}

	drops := int64(xrp * 1_000_000.0)
	return fmt.Sprintf("%d", drops), nil
}

// IsValidAddress performs basic validation on an XRPL address.
// XRPL addresses are base58-encoded and start with 'r'.
func IsValidAddress(address string) bool {
	// Basic validation: address should start with 'r' and be reasonable length
	// Full validation requires base58 decode and checksum verification
	return len(address) >= 25 && len(address) <= 35 && address[0] == 'r'
}

// IsValidTxHash performs basic validation on a transaction hash.
// Transaction hashes are 64 hex characters.
func IsValidTxHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}

	// Check all characters are hex
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}
