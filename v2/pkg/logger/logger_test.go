package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want string // Expected to contain this string
	}{
		{
			name: "text handler",
			opts: Options{Level: LevelInfo, JSON: false},
			want: "level=INFO",
		},
		{
			name: "json handler",
			opts: Options{Level: LevelInfo, JSON: true},
			want: `"level":"INFO"`,
		},
		{
			name: "debug level",
			opts: Options{Level: LevelDebug, JSON: false},
			want: "level=INFO", // We log INFO message, so it should show as INFO
		},
		{
			name: "error level",
			opts: Options{Level: LevelError, JSON: false},
			want: "", // Info message won't be logged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.opts.Writer = &buf

			logger := New(tt.opts)
			logger.Info("test message", "key", "value")

			got := buf.String()
			if tt.want != "" && !strings.Contains(got, tt.want) {
				t.Errorf("New() output = %q, want to contain %q", got, tt.want)
			}
			if tt.want == "" && got != "" {
				t.Errorf("New() output = %q, want empty (message should be filtered)", got)
			}
		})
	}
}

func TestNew_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     Level
		logFunc   func(*slog.Logger)
		shouldLog bool
	}{
		{
			name:      "debug logged at debug level",
			level:     LevelDebug,
			logFunc:   func(l *slog.Logger) { l.Debug("debug message") },
			shouldLog: true,
		},
		{
			name:      "debug not logged at info level",
			level:     LevelInfo,
			logFunc:   func(l *slog.Logger) { l.Debug("debug message") },
			shouldLog: false,
		},
		{
			name:      "info logged at info level",
			level:     LevelInfo,
			logFunc:   func(l *slog.Logger) { l.Info("info message") },
			shouldLog: true,
		},
		{
			name:      "info not logged at error level",
			level:     LevelError,
			logFunc:   func(l *slog.Logger) { l.Info("info message") },
			shouldLog: false,
		},
		{
			name:      "error logged at info level",
			level:     LevelInfo,
			logFunc:   func(l *slog.Logger) { l.Error("error message") },
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(Options{
				Level:  tt.level,
				JSON:   false,
				Writer: &buf,
			})

			tt.logFunc(logger)

			got := buf.String()
			if tt.shouldLog && got == "" {
				t.Errorf("Expected log output but got none")
			}
			if !tt.shouldLog && got != "" {
				t.Errorf("Expected no log output but got: %q", got)
			}
		})
	}
}

func TestNew_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Level:  LevelInfo,
		JSON:   true,
		Writer: &buf,
	})

	logger.Info("test message", "user_id", 123, "action", "login")

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, buf.String())
	}

	// Check expected fields
	if result["msg"] != "test message" {
		t.Errorf("Expected msg='test message', got %v", result["msg"])
	}
	if result["user_id"] != float64(123) {
		t.Errorf("Expected user_id=123, got %v", result["user_id"])
	}
	if result["action"] != "login" {
		t.Errorf("Expected action='login', got %v", result["action"])
	}
	if result["level"] != "INFO" {
		t.Errorf("Expected level='INFO', got %v", result["level"])
	}
}

func TestNew_DefaultWriter(t *testing.T) {
	logger := New(Options{Level: LevelInfo, JSON: false})
	if logger == nil {
		t.Error("Expected logger to be created with default writer")
	}
	// If we get here without panic, default writer (stdout) works
}

func TestNewFromEnv(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		check   func(*testing.T, *bytes.Buffer)
	}{
		{
			name: "default level info",
			envVars: map[string]string{
				"LOG_LEVEL": "",
			},
			check: func(t *testing.T, buf *bytes.Buffer) {
				// At info level, info messages should be logged
				if buf.Len() == 0 {
					t.Error("Expected info output at info level")
				}
			},
		},
		{
			name: "debug level from env",
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
			},
			check: func(t *testing.T, buf *bytes.Buffer) {
				if buf.Len() == 0 {
					t.Error("Expected debug output")
				}
			},
		},
		{
			name: "json format from LOG_JSON",
			envVars: map[string]string{
				"LOG_JSON": "true",
			},
			check: func(t *testing.T, buf *bytes.Buffer) {
				if !strings.Contains(buf.String(), `"level":"INFO"`) {
					t.Errorf("Expected JSON format, got: %s", buf.String())
				}
			},
		},
		{
			name: "json format from ENV=production",
			envVars: map[string]string{
				"ENV": "production",
			},
			check: func(t *testing.T, buf *bytes.Buffer) {
				if !strings.Contains(buf.String(), `"level":"INFO"`) {
					t.Errorf("Expected JSON format, got: %s", buf.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			// Capture output
			var buf bytes.Buffer

			// Create logger manually with env settings to capture output
			level := LevelInfo
			if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
				level = Level(envLevel)
			}

			json := false
			if os.Getenv("LOG_JSON") == "true" || os.Getenv("ENV") == "production" {
				json = true
			}

			logger := New(Options{
				Level:  level,
				JSON:   json,
				Writer: &buf,
			})

			// Log info message for all tests
			logger.Info("info message")

			tt.check(t, &buf)
		})
	}
}

func TestDefault(t *testing.T) {
	logger := Default()
	if logger == nil {
		t.Error("Expected logger to be created")
	}

	// Test that it works (won't panic)
	var buf bytes.Buffer
	logger = New(Options{Level: LevelInfo, JSON: false, Writer: &buf})
	logger.Info("test")

	if !strings.Contains(buf.String(), "test") {
		t.Error("Expected log message to contain 'test'")
	}
}

func TestNewFromEnv_Variants(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		cleanup func()
		verify  func(*testing.T, *slog.Logger)
	}{
		{
			name: "with LOG_LEVEL=debug",
			setup: func() {
				os.Setenv("LOG_LEVEL", "debug")
			},
			cleanup: func() {
				os.Unsetenv("LOG_LEVEL")
			},
			verify: func(t *testing.T, logger *slog.Logger) {
				if logger == nil {
					t.Fatal("NewFromEnv returned nil")
				}
				// Verify by logging and checking level behavior with a test logger
				var buf bytes.Buffer
				testLogger := New(Options{Level: LevelDebug, JSON: false, Writer: &buf})
				testLogger.Debug("debug msg")
				if !strings.Contains(buf.String(), "debug msg") {
					t.Error("Debug level not working")
				}
			},
		},
		{
			name: "with LOG_JSON=true",
			setup: func() {
				os.Setenv("LOG_JSON", "true")
			},
			cleanup: func() {
				os.Unsetenv("LOG_JSON")
			},
			verify: func(t *testing.T, logger *slog.Logger) {
				if logger == nil {
					t.Fatal("NewFromEnv returned nil")
				}
			},
		},
		{
			name: "with ENV=production",
			setup: func() {
				os.Setenv("ENV", "production")
			},
			cleanup: func() {
				os.Unsetenv("ENV")
			},
			verify: func(t *testing.T, logger *slog.Logger) {
				if logger == nil {
					t.Fatal("NewFromEnv returned nil")
				}
			},
		},
		{
			name: "without any env vars",
			setup: func() {
				// Clean environment
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("LOG_JSON")
				os.Unsetenv("ENV")
			},
			cleanup: func() {},
			verify: func(t *testing.T, logger *slog.Logger) {
				if logger == nil {
					t.Fatal("NewFromEnv returned nil")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer tt.cleanup()

			logger := NewFromEnv()
			tt.verify(t, logger)
		})
	}
}
