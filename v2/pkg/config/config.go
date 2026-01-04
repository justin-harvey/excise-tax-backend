package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a custom type that supports unmarshaling from strings
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(dur)
		return nil
	default:
		return fmt.Errorf("invalid duration type: %T", v)
	}
}

// UnmarshalYAML implements yaml.Unmarshaler
func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// Config holds all application configuration
type Config struct {
	// XRPL settings
	XRPL XRPLConfig `json:"xrpl" yaml:"xrpl"`

	// Server settings
	Server ServerConfig `json:"server" yaml:"server"`

	// Logging settings
	Logging LogConfig `json:"logging" yaml:"logging"`

	// Environment
	Env string `json:"env" yaml:"env"`
}

// XRPLConfig holds XRPL-specific configuration
type XRPLConfig struct {
	Network string   `json:"network" yaml:"network"`
	URL     string   `json:"url" yaml:"url"`
	Timeout Duration `json:"timeout" yaml:"timeout"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int      `json:"port" yaml:"port"`
	ReadTimeout  Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  Duration `json:"idle_timeout" yaml:"idle_timeout"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level string `json:"level" yaml:"level"`
	JSON  bool   `json:"json" yaml:"json"`
}

// Default returns a Config with sensible default values
func Default() Config {
	return Config{
		XRPL: XRPLConfig{
			Network: "testnet",
			URL:     "wss://s.altnet.rippletest.net:51233",
			Timeout: Duration(10 * time.Second),
		},
		Server: ServerConfig{
			Port:         8080,
			ReadTimeout:  Duration(10 * time.Second),
			WriteTimeout: Duration(10 * time.Second),
			IdleTimeout:  Duration(60 * time.Second),
		},
		Logging: LogConfig{
			Level: "info",
			JSON:  false,
		},
		Env: "development",
	}
}

// Load loads configuration from file, environment variables, and defaults
// Priority: env vars > file > defaults
func Load(filePath string) (Config, error) {
	// Start with defaults
	cfg := Default()

	// Load from file if provided
	if filePath != "" {
		fileCfg, err := loadFromFile(filePath)
		if err != nil {
			return Config{}, fmt.Errorf("failed to load config file: %w", err)
		}
		merge(&cfg, &fileCfg)
	}

	// Override with environment variables
	loadFromEnv(&cfg)

	// Validate
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads config from a JSON or YAML file
func loadFromFile(filePath string) (Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	ext := filepath.Ext(filePath)

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("failed to parse JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("failed to parse YAML: %w", err)
		}
	default:
		return Config{}, fmt.Errorf("unsupported config file format: %s", ext)
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(cfg *Config) {
	// XRPL
	if v := os.Getenv("XRPL_NETWORK"); v != "" {
		cfg.XRPL.Network = v
	}
	if v := os.Getenv("XRPL_URL"); v != "" {
		cfg.XRPL.URL = v
	}
	if v := os.Getenv("XRPL_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.XRPL.Timeout = Duration(d)
		}
	}

	// Server
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("SERVER_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ReadTimeout = Duration(d)
		}
	}
	if v := os.Getenv("SERVER_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.WriteTimeout = Duration(d)
		}
	}
	if v := os.Getenv("SERVER_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.IdleTimeout = Duration(d)
		}
	}

	// Logging
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
	if v := os.Getenv("LOG_JSON"); v == "true" {
		cfg.Logging.JSON = true
	}

	// Environment
	if v := os.Getenv("ENV"); v != "" {
		cfg.Env = v
	}
}

// merge merges non-zero values from src into dst
func merge(dst, src *Config) {
	// XRPL
	if src.XRPL.Network != "" {
		dst.XRPL.Network = src.XRPL.Network
	}
	if src.XRPL.URL != "" {
		dst.XRPL.URL = src.XRPL.URL
	}
	if src.XRPL.Timeout != 0 {
		dst.XRPL.Timeout = src.XRPL.Timeout
	}

	// Server
	if src.Server.Port != 0 {
		dst.Server.Port = src.Server.Port
	}
	if src.Server.ReadTimeout != 0 {
		dst.Server.ReadTimeout = src.Server.ReadTimeout
	}
	if src.Server.WriteTimeout != 0 {
		dst.Server.WriteTimeout = src.Server.WriteTimeout
	}
	if src.Server.IdleTimeout != 0 {
		dst.Server.IdleTimeout = src.Server.IdleTimeout
	}

	// Logging
	if src.Logging.Level != "" {
		dst.Logging.Level = src.Logging.Level
	}
	if src.Logging.JSON {
		dst.Logging.JSON = true
	}

	// Environment
	if src.Env != "" {
		dst.Env = src.Env
	}
}

// Validate checks if the configuration is valid
func (c Config) Validate() error {
	// Validate XRPL network
	if c.XRPL.Network != "testnet" && c.XRPL.Network != "mainnet" && c.XRPL.Network != "devnet" {
		return fmt.Errorf("invalid XRPL network: %s (must be 'testnet', 'mainnet', or 'devnet')", c.XRPL.Network)
	}

	// Validate XRPL URL
	if c.XRPL.URL == "" {
		return errors.New("XRPL URL is required")
	}

	// Validate server port
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", c.Server.Port)
	}

	// Validate timeouts
	if time.Duration(c.XRPL.Timeout) <= 0 {
		return errors.New("XRPL timeout must be positive")
	}
	if time.Duration(c.Server.ReadTimeout) <= 0 {
		return errors.New("server read timeout must be positive")
	}
	if time.Duration(c.Server.WriteTimeout) <= 0 {
		return errors.New("server write timeout must be positive")
	}
	if time.Duration(c.Server.IdleTimeout) <= 0 {
		return errors.New("server idle timeout must be positive")
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be 'debug', 'info', 'warn', or 'error')", c.Logging.Level)
	}

	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[c.Env] {
		return fmt.Errorf("invalid environment: %s (must be 'development', 'staging', or 'production')", c.Env)
	}

	return nil
}
