package main

import (
	"encoding/json"
	"testing"
)

func TestConstants(t *testing.T) {
	// Verify exit codes
	if exitOK != 0 {
		t.Errorf("exitOK = %d, want 0", exitOK)
	}
	if exitError != 1 {
		t.Errorf("exitError = %d, want 1", exitError)
	}
	if exitUsage != 2 {
		t.Errorf("exitUsage = %d, want 2", exitUsage)
	}
}

func TestNetworkURLs(t *testing.T) {
	// Verify network URLs are defined
	if testnetURL == "" {
		t.Error("testnetURL is empty")
	}
	if mainnetURL == "" {
		t.Error("mainnetURL is empty")
	}

	// Verify URLs have correct scheme
	if len(testnetURL) < 6 || testnetURL[:6] != "wss://" {
		t.Errorf("testnetURL should start with wss://, got %s", testnetURL)
	}
	if len(mainnetURL) < 6 || mainnetURL[:6] != "wss://" {
		t.Errorf("mainnetURL should start with wss://, got %s", mainnetURL)
	}
}

func TestJSONMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "simple object",
			data: map[string]interface{}{
				"test": "value",
				"num":  123,
			},
			wantErr: false,
		},
		{
			name:    "array",
			data:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "nil",
			data:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling used by printJSON
			_, err := json.MarshalIndent(tt.data, "", "  ")
			if (err != nil) != tt.wantErr {
				t.Errorf("json.MarshalIndent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
