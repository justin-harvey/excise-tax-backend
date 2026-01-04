package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Verify XRPL defaults
	if cfg.XRPL.Network != "testnet" {
		t.Errorf("Expected default network 'testnet', got %s", cfg.XRPL.Network)
	}
	if time.Duration(cfg.XRPL.Timeout) != 10*time.Second {
		t.Errorf("Expected default timeout 10s, got %v", cfg.XRPL.Timeout)
	}

	// Verify server defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	// Verify logging defaults
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Logging.Level)
	}

	// Verify environment default
	if cfg.Env != "development" {
		t.Errorf("Expected default env 'development', got %s", cfg.Env)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			cfg:     Default(),
			wantErr: false,
		},
		{
			name: "invalid network",
			cfg: Config{
				XRPL:    XRPLConfig{Network: "invalid", URL: "wss://test", Timeout: Duration(10 * time.Second)},
				Server:  ServerConfig{Port: 8080, ReadTimeout: Duration(10 * time.Second), WriteTimeout: Duration(10 * time.Second), IdleTimeout: Duration(60 * time.Second)},
				Logging: LogConfig{Level: "info"},
				Env:     "development",
			},
			wantErr: true,
			errMsg:  "invalid XRPL network",
		},
		{
			name: "missing URL",
			cfg: Config{
				XRPL:    XRPLConfig{Network: "testnet", URL: "", Timeout: Duration(10 * time.Second)},
				Server:  ServerConfig{Port: 8080, ReadTimeout: Duration(10 * time.Second), WriteTimeout: Duration(10 * time.Second), IdleTimeout: Duration(60 * time.Second)},
				Logging: LogConfig{Level: "info"},
				Env:     "development",
			},
			wantErr: true,
			errMsg:  "URL is required",
		},
		{
			name: "invalid port low",
			cfg: Config{
				XRPL:    XRPLConfig{Network: "testnet", URL: "wss://test", Timeout: Duration(10 * time.Second)},
				Server:  ServerConfig{Port: 0, ReadTimeout: Duration(10 * time.Second), WriteTimeout: Duration(10 * time.Second), IdleTimeout: Duration(60 * time.Second)},
				Logging: LogConfig{Level: "info"},
				Env:     "development",
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid port high",
			cfg: Config{
				XRPL:    XRPLConfig{Network: "testnet", URL: "wss://test", Timeout: Duration(10 * time.Second)},
				Server:  ServerConfig{Port: 99999, ReadTimeout: Duration(10 * time.Second), WriteTimeout: Duration(10 * time.Second), IdleTimeout: Duration(60 * time.Second)},
				Logging: LogConfig{Level: "info"},
				Env:     "development",
			},
			wantErr: true,
			errMsg:  "invalid server port",
		},
		{
			name: "invalid log level",
			cfg: Config{
				XRPL:    XRPLConfig{Network: "testnet", URL: "wss://test", Timeout: Duration(10 * time.Second)},
				Server:  ServerConfig{Port: 8080, ReadTimeout: Duration(10 * time.Second), WriteTimeout: Duration(10 * time.Second), IdleTimeout: Duration(60 * time.Second)},
				Logging: LogConfig{Level: "invalid"},
				Env:     "development",
			},
			wantErr: true,
			errMsg:  "invalid log level",
		},
		{
			name: "invalid environment",
			cfg: Config{
				XRPL:    XRPLConfig{Network: "testnet", URL: "wss://test", Timeout: Duration(10 * time.Second)},
				Server:  ServerConfig{Port: 8080, ReadTimeout: Duration(10 * time.Second), WriteTimeout: Duration(10 * time.Second), IdleTimeout: Duration(60 * time.Second)},
				Logging: LogConfig{Level: "info"},
				Env:     "invalid",
			},
			wantErr: true,
			errMsg:  "invalid environment",
		},
		{
			name: "negative timeout",
			cfg: Config{
				XRPL:    XRPLConfig{Network: "testnet", URL: "wss://test", Timeout: Duration(-1 * time.Second)},
				Server:  ServerConfig{Port: 8080, ReadTimeout: Duration(10 * time.Second), WriteTimeout: Duration(10 * time.Second), IdleTimeout: Duration(60 * time.Second)},
				Logging: LogConfig{Level: "info"},
				Env:     "development",
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestLoadFromFile_JSON(t *testing.T) {
	// Create temp JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
		"xrpl": {
			"network": "mainnet",
			"url": "wss://xrplcluster.com",
			"timeout": "15s"
		},
		"server": {
			"port": 9090
		},
		"logging": {
			"level": "debug",
			"json": true
		},
		"env": "production"
	}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(jsonFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded values
	if cfg.XRPL.Network != "mainnet" {
		t.Errorf("Expected network 'mainnet', got %s", cfg.XRPL.Network)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", cfg.Logging.Level)
	}
	if !cfg.Logging.JSON {
		t.Error("Expected JSON logging to be true")
	}
	if cfg.Env != "production" {
		t.Errorf("Expected env 'production', got %s", cfg.Env)
	}
}

func TestLoadFromFile_YAML(t *testing.T) {
	// Create temp YAML file
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
xrpl:
  network: mainnet
  url: wss://xrplcluster.com
  timeout: 15s
server:
  port: 9090
logging:
  level: debug
  json: true
env: production
`

	if err := os.WriteFile(yamlFile, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(yamlFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded values
	if cfg.XRPL.Network != "mainnet" {
		t.Errorf("Expected network 'mainnet', got %s", cfg.XRPL.Network)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("XRPL_NETWORK", "mainnet")
	os.Setenv("XRPL_URL", "wss://custom.url")
	os.Setenv("SERVER_PORT", "9999")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ENV", "production")
	defer func() {
		os.Unsetenv("XRPL_NETWORK")
		os.Unsetenv("XRPL_URL")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("ENV")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify environment variables override defaults
	if cfg.XRPL.Network != "mainnet" {
		t.Errorf("Expected network 'mainnet', got %s", cfg.XRPL.Network)
	}
	if cfg.XRPL.URL != "wss://custom.url" {
		t.Errorf("Expected custom URL, got %s", cfg.XRPL.URL)
	}
	if cfg.Server.Port != 9999 {
		t.Errorf("Expected port 9999, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", cfg.Logging.Level)
	}
	if cfg.Env != "production" {
		t.Errorf("Expected env 'production', got %s", cfg.Env)
	}
}

func TestLoad_Priority(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
		"xrpl": {
			"network": "mainnet"
		},
		"server": {
			"port": 9090
		}
	}`

	if err := os.WriteFile(configFile, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Set environment variable (should override file)
	os.Setenv("SERVER_PORT", "7777")
	defer os.Unsetenv("SERVER_PORT")

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify priority: env var > file > defaults
	if cfg.XRPL.Network != "mainnet" {
		t.Errorf("Expected network from file 'mainnet', got %s", cfg.XRPL.Network)
	}
	if cfg.Server.Port != 7777 {
		t.Errorf("Expected port from env 7777, got %d", cfg.Server.Port)
	}
	// Default value should be used for unset fields
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Logging.Level)
	}
}

func TestLoadFromFile_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "config.txt")

	if err := os.WriteFile(txtFile, []byte("invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(txtFile)
	if err == nil {
		t.Error("Expected error for invalid file format")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(jsonFile, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(jsonFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadFromFile_NonExistent(t *testing.T) {
	_, err := Load("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
