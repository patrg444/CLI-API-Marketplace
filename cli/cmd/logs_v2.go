package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net/url"
	
	"github.com/api-direct/cli/pkg/config"
	"github.com/api-direct/cli/pkg/manifest"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Enhanced logs command with better features

var (
	logsFilter   string
	logsSince    string
	logsJSON     bool
	logsNoColor  bool
	logsReplica  string
	logsLevel    string
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	ReplicaID   string                 `json:"replica_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

func initLogsV2() {
	// Enhanced flags
	logsCmd.Flags().StringVar(&logsFilter, "filter", "", "Filter logs by text")
	logsCmd.Flags().StringVar(&logsSince, "since", "1h", "Show logs since (e.g., 10m, 1h, 24h)")
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, "Output logs as JSON")
	logsCmd.Flags().BoolVar(&logsNoColor, "no-color", false, "Disable colored output")
	logsCmd.Flags().StringVar(&logsReplica, "replica", "", "Show logs from specific replica")
	logsCmd.Flags().StringVar(&logsLevel, "level", "", "Filter by log level (error, warn, info, debug)")
}

func runLogsV2(cmd *cobra.Command, args []string) error {
	// Check authentication
	if !config.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'apidirect login' first")
	}

	// Get API name - support both explicit and manifest-based
	var apiName string
	if len(args) > 0 {
		apiName = args[0]
	} else {
		// Try to get from manifest
		if manifestPath, err := manifest.FindManifest("."); err == nil {
			if m, err := manifest.Load(manifestPath); err == nil {
				apiName = m.Name
			}
		}
		
		if apiName == "" {
			return fmt.Errorf("no API name provided and no manifest found")
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse time duration
	duration, err := parseDuration(logsSince)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	// Set up colored output
	if logsNoColor || logsJSON {
		color.NoColor = true
	}

	printInfo(fmt.Sprintf("📋 Fetching logs for '%s'", apiName))
	
	if logsFilter != "" {
		fmt.Printf("🔍 Filter: %s\n", logsFilter)
	}
	if logsLevel != "" {
		fmt.Printf("📊 Level: %s\n", logsLevel)
	}
	if logsReplica != "" {
		fmt.Printf("🖥️  Replica: %s\n", logsReplica)
	}
	
	fmt.Printf("⏰ Since: %s ago\n", logsSince)
	fmt.Println(strings.Repeat("─", 80))

	if logsFollow || logsLive {
		return streamLogsV2(cfg, apiName, duration)
	}
	
	return fetchLogsV2(cfg, apiName, duration)
}

func streamLogsV2(cfg *config.Config, apiName string, since time.Duration) error {
	// Set up interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Build query parameters
	params := buildLogParams(since)
	
	// Create SSE request
	url := fmt.Sprintf("%s/logs/v2/stream/%s?%s", cfg.API.BaseURL, apiName, params.Encode())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	// Create HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to stream logs: %s - %s", resp.Status, string(body))
	}

	fmt.Println("🔄 Streaming logs (press Ctrl+C to stop)...")
	fmt.Println()

	// Read SSE stream
	reader := bufio.NewReader(resp.Body)
	done := make(chan bool)
	// lastEventID := "" // Will be used for reconnection support

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					printError(fmt.Sprintf("Stream error: %v", err))
				}
				done <- true
				return
			}

			line = strings.TrimSpace(line)

			// Parse SSE format
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					continue
				}

				// Parse log entry
				var entry LogEntry
				if err := json.Unmarshal([]byte(data), &entry); err == nil {
					displayLogEntry(entry)
				} else {
					// Fallback to plain text
					fmt.Println(data)
				}
			} else if strings.HasPrefix(line, "id: ") {
				_ = strings.TrimPrefix(line, "id: ") // lastEventID for future reconnection support
			} else if strings.HasPrefix(line, ":") {
				// Server heartbeat comment
				continue
			}
		}
	}()

	// Wait for interrupt or stream end
	select {
	case <-interrupt:
		fmt.Println("\n\n🛑 Stopping log stream...")
		return nil
	case <-done:
		return nil
	}
}

func fetchLogsV2(cfg *config.Config, apiName string, since time.Duration) error {
	// Build query parameters
	params := buildLogParams(since)
	if logsTail > 0 {
		params.Add("limit", fmt.Sprintf("%d", logsTail))
	}

	// Fetch logs
	url := fmt.Sprintf("%s/logs/v2/fetch/%s?%s", cfg.API.BaseURL, apiName, params.Encode())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to fetch logs: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var result struct {
		Logs       []LogEntry `json:"logs"`
		TotalCount int        `json:"total_count"`
		HasMore    bool       `json:"has_more"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse logs: %w", err)
	}

	// Display logs
	if logsJSON {
		// Output as JSON array
		output, _ := json.MarshalIndent(result.Logs, "", "  ")
		fmt.Println(string(output))
	} else {
		// Display formatted logs
		for _, entry := range result.Logs {
			displayLogEntry(entry)
		}

		// Summary
		fmt.Println(strings.Repeat("─", 80))
		fmt.Printf("📊 Displayed %d of %d total logs\n", len(result.Logs), result.TotalCount)
		
		if result.HasMore {
			fmt.Println("💡 Use --tail to see more logs or --follow to stream")
		}
	}

	return nil
}

func displayLogEntry(entry LogEntry) {
	if logsJSON {
		output, _ := json.Marshal(entry)
		fmt.Println(string(output))
		return
	}

	// Format timestamp
	timestamp := entry.Timestamp.Format("15:04:05.000")
	
	// Color based on level
	var levelColor func(format string, a ...interface{}) string
	// var levelIcon string // Can be used for icon-based display
	
	switch strings.ToLower(entry.Level) {
	case "error", "fatal":
		levelColor = color.RedString
		// levelIcon = "❌"
	case "warn", "warning":
		levelColor = color.YellowString
		// levelIcon = "⚠️ "
	case "info":
		levelColor = color.CyanString
		// levelIcon = "ℹ️ "
	case "debug":
		levelColor = color.WhiteString
		// levelIcon = "🔍"
	default:
		levelColor = color.WhiteString
		// levelIcon = "  "
	}

	// Build output line
	var output strings.Builder
	
	// Timestamp
	output.WriteString(color.HiBlackString("[%s]", timestamp))
	output.WriteString(" ")
	
	// Level
	if entry.Level != "" {
		output.WriteString(levelColor("%-5s", strings.ToUpper(entry.Level)))
		output.WriteString(" ")
	}
	
	// Replica ID
	if entry.ReplicaID != "" && (logsReplica == "" || logsReplica == entry.ReplicaID) {
		output.WriteString(color.MagentaString("[%s]", entry.ReplicaID))
		output.WriteString(" ")
	}
	
	// Request ID
	if entry.RequestID != "" {
		output.WriteString(color.HiBlackString("{%s}", entry.RequestID[:8]))
		output.WriteString(" ")
	}
	
	// Message
	message := entry.Message
	if logsFilter != "" && !strings.Contains(strings.ToLower(message), strings.ToLower(logsFilter)) {
		return // Skip if doesn't match filter
	}
	
	// Highlight filter term if present
	if logsFilter != "" {
		message = highlightTerm(message, logsFilter)
	}
	
	output.WriteString(message)
	
	// Additional fields
	if len(entry.Fields) > 0 && verbose {
		output.WriteString(" ")
		fields, _ := json.Marshal(entry.Fields)
		output.WriteString(color.HiBlackString(string(fields)))
	}
	
	fmt.Println(output.String())
}

func buildLogParams(since time.Duration) *url.Values {
	params := &url.Values{}
	
	// Time range
	startTime := time.Now().Add(-since)
	params.Add("start_time", startTime.Format(time.RFC3339))
	
	// Filters
	if logsFilter != "" {
		params.Add("filter", logsFilter)
	}
	if logsLevel != "" {
		params.Add("level", logsLevel)
	}
	if logsReplica != "" {
		params.Add("replica", logsReplica)
	}
	
	return params
}

func parseDuration(s string) (time.Duration, error) {
	// Handle common formats like "10m", "1h", "24h", "7d"
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		var d int
		if _, err := fmt.Sscanf(days, "%d", &d); err != nil {
			return 0, err
		}
		return time.Duration(d) * 24 * time.Hour, nil
	}
	
	return time.ParseDuration(s)
}

func highlightTerm(text, term string) string {
	// Case-insensitive highlight
	lower := strings.ToLower(text)
	lowerTerm := strings.ToLower(term)
	
	index := strings.Index(lower, lowerTerm)
	if index == -1 {
		return text
	}
	
	// Highlight the matching portion
	before := text[:index]
	match := text[index : index+len(term)]
	after := text[index+len(term):]
	
	return before + color.HiYellowString(match) + after
}

// Mock log entries for demo mode
func generateMockLogs(apiName string, count int) []LogEntry {
	levels := []string{"info", "warn", "error", "debug"}
	messages := []string{
		"Server started on port 8080",
		"Connected to database successfully",
		"Request received: GET /api/users",
		"User authentication successful",
		"Cache miss for key: user_123",
		"Database query took 45ms",
		"Error: Connection timeout to external API",
		"Warning: Memory usage above 80%",
		"Deployed new version: v1.2.3",
		"Health check passed",
		"Rate limit exceeded for IP: 192.168.1.1",
		"Background job completed successfully",
	}
	
	replicas := []string{"api-abc123", "api-def456", "api-ghi789"}
	
	logs := make([]LogEntry, count)
	now := time.Now()
	
	for i := 0; i < count; i++ {
		// Generate timestamps going backwards
		timestamp := now.Add(-time.Duration(count-i) * time.Minute)
		
		logs[i] = LogEntry{
			Timestamp: timestamp,
			Level:     levels[i%len(levels)],
			Message:   messages[i%len(messages)],
			Source:    apiName,
			ReplicaID: replicas[i%len(replicas)],
			RequestID: fmt.Sprintf("req_%d", timestamp.Unix()),
			Fields: map[string]interface{}{
				"duration_ms": 100 + i*10,
				"status_code": 200,
			},
		}
	}
	
	return logs
}