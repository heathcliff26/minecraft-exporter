package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/heathcliff26/minecraft-exporter/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(t, err, "Should create request")

	rr := httptest.NewRecorder()

	ServerRootHandler(rr, req)

	assert := assert.New(t)

	assert.Equal(http.StatusOK, rr.Code)
	body := rr.Body.String()
	assert.Contains(body, "<html>")
	assert.Contains(body, "</html>")
	assert.Contains(body, "<a href='/metrics'>")
}

func TestHandleVersionFlag(t *testing.T) {
	tests := []struct {
		name           string
		showVersionVal bool
		expected       bool
	}{
		{
			name:           "Version flag not set",
			showVersionVal: false,
			expected:       false,
		},
		{
			name:           "Version flag set",
			showVersionVal: true,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original value and restore it after the test
			originalShowVersion := showVersion
			defer func() { showVersion = originalShowVersion }()

			// Capture stdout to verify version output when flag is set
			originalStdout := os.Stdout
			if tt.showVersionVal {
				r, w, _ := os.Pipe()
				os.Stdout = w
				defer func() {
					w.Close()
					os.Stdout = originalStdout
				}()

				showVersion = tt.showVersionVal
				result := handleVersionFlag()

				w.Close()
				var buf bytes.Buffer
				buf.ReadFrom(r)
				output := buf.String()

				assert.Equal(t, tt.expected, result)
				assert.NotEmpty(t, output, "Should print version when flag is set")
			} else {
				showVersion = tt.showVersionVal
				result := handleVersionFlag()
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestLoadConfigWithLogging(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		env         bool
		expectError bool
	}{
		{
			name:        "Empty config path loads default config",
			configPath:  "",
			env:         false,
			expectError: false,
		},
		{
			name:        "Invalid config path returns error",
			configPath:  "/nonexistent/path/config.yaml",
			env:         false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadConfigWithLogging(tt.configPath, tt.env)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				// Verify we get a valid config with default values
				assert.Equal(t, config.DEFAULT_PORT, cfg.Port)
				assert.Equal(t, config.DEFAULT_WORLD_DIR, cfg.WorldDir)
			}
		})
	}
}

func TestSetupPrometheusRegistry(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Config
		expectError bool
	}{
		{
			name: "Basic config without RCON",
			config: config.Config{
				WorldDir:      "/tmp/test-world",
				ReduceMetrics: false,
				RCON: config.RCONConfig{
					Enable: false,
				},
			},
			expectError: true, // Will fail because /tmp/test-world doesn't exist
		},
		{
			name: "Config with invalid world directory",
			config: config.Config{
				WorldDir:      "/nonexistent/world",
				ReduceMetrics: false,
				RCON: config.RCONConfig{
					Enable: false,
				},
			},
			expectError: true,
		},
		{
			name: "Config with temporary directory as world (will fail because it lacks minecraft structure)",
			config: config.Config{
				WorldDir:      t.TempDir(),
				ReduceMetrics: true,
				RCON: config.RCONConfig{
					Enable: false,
				},
			},
			expectError: true, // Temp dir doesn't have minecraft world structure
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, err := setupPrometheusRegistry(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, reg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reg)
			}
		})
	}
}

func TestSetupRemoteWrite(t *testing.T) {
	tests := []struct {
		name        string
		config      config.Config
		expectError bool
	}{
		{
			name: "Remote write disabled",
			config: config.Config{
				Remote: config.RemoteConfig{
					Enable: false,
				},
			},
			expectError: false,
		},
		{
			name: "Remote write enabled with empty URL",
			config: config.Config{
				Remote: config.RemoteConfig{
					Enable:   true,
					URL:      "", // Empty URL should cause error
					Instance: "test-instance",
				},
			},
			expectError: true,
		},
		{
			name: "Remote write enabled with username but no password",
			config: config.Config{
				Remote: config.RemoteConfig{
					Enable:   true,
					URL:      "http://localhost:9090/api/v1/write",
					Instance: "test-instance",
					Username: "user", // Username without password should cause error
				},
				Interval: config.Duration(30 * time.Second),
			},
			expectError: true, // Should fail because password is required when username is provided
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal registry for testing
			reg := prometheus.NewRegistry()

			cleanup, err := setupRemoteWrite(tt.config, reg)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cleanup)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cleanup)
				// Test that cleanup function doesn't panic
				assert.NotPanics(t, func() { cleanup() })
			}
		})
	}
}
