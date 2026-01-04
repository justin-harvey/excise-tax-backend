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
			name:    "1 XRP",
			drops:   "1000000",
			want:    "1.000000",
			wantErr: false,
		},
		{
			name:    "0.5 XRP",
			drops:   "500000",
			want:    "0.500000",
			wantErr: false,
		},
		{
			name:    "0 XRP",
			drops:   "0",
			want:    "0.000000",
			wantErr: false,
		},
		{
			name:    "invalid format",
			drops:   "abc",
			want:    "",
			wantErr: true,
		},
		{
			name:    "negative drops",
			drops:   "-1000000",
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

func TestXRPToDrops(t *testing.T) {
	tests := []struct {
		name    string
		xrp     float64
		want    string
		wantErr bool
	}{
		{
			name:    "1 XRP",
			xrp:     1.0,
			want:    "1000000",
			wantErr: false,
		},
		{
			name:    "0.5 XRP",
			xrp:     0.5,
			want:    "500000",
			wantErr: false,
		},
		{
			name:    "0 XRP",
			xrp:     0.0,
			want:    "0",
			wantErr: false,
		},
		{
			name:    "negative XRP",
			xrp:     -1.0,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := XRPToDrops(tt.xrp)
			if (err != nil) != tt.wantErr {
				t.Errorf("XRPToDrops() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("XRPToDrops() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "valid testnet address",
			address: "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY",
			want:    true,
		},
		{
			name:    "valid mainnet address",
			address: "rN7n7otQDd6FczFgLdlqtyMVrn3HMfXdY2",
			want:    true,
		},
		{
			name:    "empty address",
			address: "",
			want:    false,
		},
		{
			name:    "too short",
			address: "r123",
			want:    false,
		},
		{
			name:    "doesn't start with r",
			address: "xPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY",
			want:    false,
		},
		{
			name:    "too long",
			address: "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDYtoolong",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidAddress(tt.address); got != tt.want {
				t.Errorf("IsValidAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidTxHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{
			name: "valid hash",
			hash: "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF",
			want: true,
		},
		{
			name: "valid lowercase hash",
			hash: "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			want: true,
		},
		{
			name: "too short",
			hash: "1234567890ABCDEF",
			want: false,
		},
		{
			name: "too long",
			hash: "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF00",
			want: false,
		},
		{
			name: "invalid characters",
			hash: "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEZ",
			want: false,
		},
		{
			name: "empty",
			hash: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidTxHash(tt.hash); got != tt.want {
				t.Errorf("IsValidTxHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkDropsToXRP(b *testing.B) {
	drops := "1000000"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DropsToXRP(drops)
	}
}

func BenchmarkXRPToDrops(b *testing.B) {
	xrp := 1.0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = XRPToDrops(xrp)
	}
}

func BenchmarkIsValidAddress(b *testing.B) {
	address := "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidAddress(address)
	}
}

func BenchmarkIsValidTxHash(b *testing.B) {
	hash := "1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsValidTxHash(hash)
	}
}
