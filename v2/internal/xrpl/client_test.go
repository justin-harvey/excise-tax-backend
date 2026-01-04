package xrpl

import (
	"testing"
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
