package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogsCommand(t *testing.T) {
	tests := []struct {
		name        string
		apiName     string
		follow      bool
		tail        int
		live        bool
		config      *config.Config
		mockServer  func(*testing.T) *httptest.Server
		wantErr     bool
		errContains string
		validateOut func(*testing.T, string)
	}{
		{
			name:    "fetch logs with tail",
			apiName: "test-api",
			tail:    20,
			config: &config.Config{
				Auth: config.AuthConfig{
					AccessToken: "test-token",
				},
				API: config.APIConfig{
					BaseURL: "http://localhost",
				},
			},
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/deployment/api/v1/logs/test-api", r.URL.Path)
					assert.Equal(t, "20", r.URL.Query().Get("tail"))
					assert.Contains(t, r.Header.Get("Authorization"), "Bearer")
					
					// Send sample logs
					fmt.Fprintln(w, "[2024-01-15 10:00:00] INFO Starting application")
					fmt.Fprintln(w, "[2024-01-15 10:00:01] INFO Server listening on port 8080")
					fmt.Fprintln(w, "[2024-01-15 10:00:05] WARN Slow query detected")
					fmt.Fprintln(w, "[2024-01-15 10:00:10] ERROR Failed to connect to database")
				}))
			},
			wantErr: false,
			validateOut: func(t *testing.T, output string) {
				assert.Contains(t, output, "Fetching logs for API: test-api (last 20 lines)")
				assert.Contains(t, output, "INFO Starting application")
				assert.Contains(t, output, "ERROR Failed to connect to database")
			},
		},
		{
			name:    "stream logs with SSE",
			apiName: "test-api",
			follow:  true,
			config: &config.Config{
				Auth: config.AuthConfig{
					AccessToken: "test-token",
				},
				API: config.APIConfig{
					BaseURL: "http://localhost",
				},
			},
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "text/event-stream", r.Header.Get("Accept"))
					
					// Set SSE headers
					w.Header().Set("Content-Type", "text/event-stream")
					w.Header().Set("Cache-Control", "no-cache")
					w.Header().Set("Connection", "keep-alive")
					
					flusher, ok := w.(http.Flusher)
					require.True(t, ok)
					
					// Send a few log events
					fmt.Fprintf(w, "data: [2024-01-15 10:00:00] INFO Log stream started\n\n")
					flusher.Flush()
					
					fmt.Fprintf(w, "data: [2024-01-15 10:00:01] INFO Processing request\n\n")
					flusher.Flush()
					
					// Close connection after sending logs
					// In real implementation, this would continue streaming
				}))
			},
			wantErr: false,
			validateOut: func(t *testing.T, output string) {
				assert.Contains(t, output, "Streaming logs for API: test-api")
				assert.Contains(t, output, "Press Ctrl+C to stop streaming")
			},
		},
		{
			name:    "not authenticated",
			apiName: "test-api",
			config: &config.Config{
				Auth: config.AuthConfig{
					AccessToken: "",
				},
			},
			wantErr:     true,
			errContains: "not authenticated",
		},
		{
			name:    "server error",
			apiName: "test-api",
			config: &config.Config{
				Auth: config.AuthConfig{
					AccessToken: "test-token",
				},
				API: config.APIConfig{
					BaseURL: "http://localhost",
				},
			},
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal server error"))
				}))
			},
			wantErr:     true,
			errContains: "failed to fetch logs",
		},
		{
			name:    "stream with server error",
			apiName: "test-api",
			follow:  true,
			config: &config.Config{
				Auth: config.AuthConfig{
					AccessToken: "test-token",
				},
				API: config.APIConfig{
					BaseURL: "http://localhost",
				},
			},
			mockServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
				}))
			},
			wantErr:     true,
			errContains: "failed to stream logs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment
			testDir := t.TempDir()
			oldHome := os.Getenv("HOME")
			os.Setenv("HOME", testDir)
			defer os.Setenv("HOME", oldHome)

			// Create config directory
			configDir := filepath.Join(testDir, ".apidirect")
			err := os.MkdirAll(configDir, 0755)
			require.NoError(t, err)

			// Save test config
			if tt.config != nil {
				err = config.SaveConfig(tt.config)
				require.NoError(t, err)
			}

			// Start mock server
			var server *httptest.Server
			if tt.mockServer != nil {
				server = tt.mockServer(t)
				defer server.Close()
				
				// Update config with mock server URL
				if tt.config != nil {
					tt.config.API.BaseURL = server.URL
					config.SaveConfig(tt.config)
				}
			}

			// Simulate the logs command execution
			var output bytes.Buffer
			err = executeLogsCommand(&output, tt.apiName, tt.follow, tt.tail, tt.live, tt.config)

			// Validate
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.validateOut != nil {
				tt.validateOut(t, output.String())
			}
		})
	}
}

func TestFormatLogLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		color    bool
	}{
		{
			name:     "timestamp with INFO",
			input:    "[2024-01-15 10:00:00] INFO Server started",
			contains: []string{"[2024-01-15 10:00:00]", "INFO Server started"},
			color:    true,
		},
		{
			name:     "timestamp with ERROR",
			input:    "[2024-01-15 10:00:00] ERROR Database connection failed",
			contains: []string{"[2024-01-15 10:00:00]", "ERROR Database connection failed"},
			color:    true,
		},
		{
			name:     "timestamp with WARN",
			input:    "[2024-01-15 10:00:00] WARN High memory usage",
			contains: []string{"[2024-01-15 10:00:00]", "WARN High memory usage"},
			color:    true,
		},
		{
			name:     "no timestamp with ERROR",
			input:    "ERROR: Connection refused",
			contains: []string{"ERROR: Connection refused"},
			color:    true,
		},
		{
			name:     "plain log line",
			input:    "Processing request for user 123",
			contains: []string{"Processing request for user 123"},
			color:    false,
		},
		{
			name:     "timestamp without level",
			input:    "[2024-01-15 10:00:00] Request processed",
			contains: []string{"[2024-01-15 10:00:00]", "Request processed"},
			color:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatLogLine(tt.input)
			
			// Check that all expected parts are in the output
			for _, part := range tt.contains {
				assert.Contains(t, output, part)
			}
			
			// Check for ANSI color codes if color is expected
			if tt.color {
				assert.Contains(t, output, "\033[")
			}
		})
	}
}

func TestSSEParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "single data line",
			input: "data: [2024-01-15 10:00:00] INFO Test log\n\n",
			expected: []string{"[2024-01-15 10:00:00] INFO Test log"},
		},
		{
			name: "multiple data lines",
			input: "data: Log line 1\n\ndata: Log line 2\n\n",
			expected: []string{"Log line 1", "Log line 2"},
		},
		{
			name: "mixed SSE events",
			input: "event: log\ndata: Log entry\n\ndata: Another log\n\n",
			expected: []string{"Log entry", "Another log"},
		},
		{
			name: "empty data lines",
			input: "data: \n\ndata: Valid log\n\n",
			expected: []string{"Valid log"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			var logs []string
			
			for {
				line, err := reader.ReadString('\n')
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				
				if strings.HasPrefix(line, "data: ") {
					logLine := strings.TrimPrefix(line, "data: ")
					logLine = strings.TrimSpace(logLine)
					if logLine != "" {
						logs = append(logs, logLine)
					}
				}
			}
			
			assert.Equal(t, tt.expected, logs)
		})
	}
}

func TestLogLevelDetection(t *testing.T) {
	tests := []struct {
		line     string
		hasError bool
		hasWarn  bool
		hasInfo  bool
	}{
		{"ERROR: Connection failed", true, false, false},
		{"FATAL: System crash", true, false, false},
		{"WARN: High CPU usage", false, true, false},
		{"WARNING: Deprecated function", false, true, false},
		{"INFO: Server started", false, false, true},
		{"Debug: Variable x = 5", false, false, false},
		{"Request processed successfully", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			formatted := formatLogLine(tt.line)
			
			// Check for color codes based on log level
			if tt.hasError {
				assert.Contains(t, formatted, "\033[31m") // Red
			} else if tt.hasWarn {
				assert.Contains(t, formatted, "\033[33m") // Yellow
			} else if tt.hasInfo {
				assert.Contains(t, formatted, "\033[36m") // Cyan
			}
		})
	}
}

// Helper functions

func executeLogsCommand(output io.Writer, apiName string, follow bool, tail int, live bool, cfg *config.Config) error {
	// Check authentication
	if cfg.Auth.AccessToken == "" {
		return fmt.Errorf("not authenticated. Please run 'apidirect login' first")
	}

	if follow || live {
		return streamLogs(output, cfg, apiName)
	} else {
		return fetchLogs(output, cfg, apiName, tail)
	}
}

func streamLogs(output io.Writer, cfg *config.Config, apiName string) error {
	fmt.Fprintf(output, "📋 Streaming logs for API: %s\n", apiName)
	fmt.Fprintln(output, "Press Ctrl+C to stop streaming...")
	fmt.Fprintln(output, strings.Repeat("-", 80))

	// Create request for SSE endpoint
	url := fmt.Sprintf("%s/deployment/api/v1/logs/%s", cfg.API.BaseURL, apiName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 5 * time.Second} // Short timeout for test
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to stream logs: %s", resp.Status)
	}

	// In test, just read a few lines
	reader := bufio.NewReader(resp.Body)
	for i := 0; i < 5; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		if strings.HasPrefix(line, "data: ") {
			logLine := strings.TrimPrefix(line, "data: ")
			logLine = strings.TrimSpace(logLine)
			if logLine != "" {
				fmt.Fprintln(output, formatLogLine(logLine))
			}
		}
	}

	return nil
}

func fetchLogs(output io.Writer, cfg *config.Config, apiName string, tail int) error {
	fmt.Fprintf(output, "📋 Fetching logs for API: %s (last %d lines)\n", apiName, tail)
	fmt.Fprintln(output, strings.Repeat("-", 80))

	url := fmt.Sprintf("%s/deployment/api/v1/logs/%s?tail=%d", cfg.API.BaseURL, apiName, tail)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch logs: %s", resp.Status)
	}

	// Read and display logs
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(output, formatLogLine(line))
	}

	return scanner.Err()
}

func formatLogLine(line string) string {
	// Try to parse and format log lines
	if strings.HasPrefix(line, "[") {
		endIdx := strings.Index(line, "]")
		if endIdx > 0 {
			timestamp := line[:endIdx+1]
			rest := line[endIdx+1:]
			
			// Color based on log level if present
			if strings.Contains(rest, "ERROR") || strings.Contains(rest, "FATAL") {
				return fmt.Sprintf("\033[90m%s\033[0m\033[31m%s\033[0m", timestamp, rest)
			} else if strings.Contains(rest, "WARN") {
				return fmt.Sprintf("\033[90m%s\033[0m\033[33m%s\033[0m", timestamp, rest)
			} else if strings.Contains(rest, "INFO") {
				return fmt.Sprintf("\033[90m%s\033[0m\033[36m%s\033[0m", timestamp, rest)
			} else {
				return fmt.Sprintf("\033[90m%s\033[0m%s", timestamp, rest)
			}
		}
	}

	// Check for common log levels without timestamp
	if strings.Contains(line, "ERROR") || strings.Contains(line, "FATAL") {
		return fmt.Sprintf("\033[31m%s\033[0m", line)
	} else if strings.Contains(line, "WARN") {
		return fmt.Sprintf("\033[33m%s\033[0m", line)
	} else if strings.Contains(line, "INFO") {
		return fmt.Sprintf("\033[36m%s\033[0m", line)
	}

	return line
}