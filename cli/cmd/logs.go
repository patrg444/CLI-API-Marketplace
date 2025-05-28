package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/api-direct/cli/pkg/config"
)

var (
	logsFollow bool
	logsTail   int
	logsLive   bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs <api-name>",
	Short: "View logs from your deployed API",
	Long: `View logs from your deployed API. You can stream logs in real-time
or view historical logs.

Examples:
  apidirect logs my-api              # View recent logs
  apidirect logs my-api --follow     # Stream logs in real-time
  apidirect logs my-api --tail 100   # View last 100 log lines`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output (like tail -f)")
	logsCmd.Flags().IntVarP(&logsTail, "tail", "n", 50, "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVar(&logsLive, "live", false, "Stream logs live from the deployment")
}

func runLogs(cmd *cobra.Command, args []string) error {
	// Check authentication
	if !config.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'apidirect login' first")
	}

	apiName := args[0]

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if logsFollow || logsLive {
		return streamLogs(cfg, apiName)
	} else {
		return fetchLogs(cfg, apiName)
	}
}

func streamLogs(cfg *config.Config, apiName string) error {
	fmt.Printf("📋 Streaming logs for API: %s\n", apiName)
	fmt.Println("Press Ctrl+C to stop streaming...")
	fmt.Println(strings.Repeat("-", 80))

	// Set up interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Create request for SSE endpoint
	url := fmt.Sprintf("%s/deployment/api/v1/logs/%s", cfg.API.BaseURL, apiName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	req.Header.Set("Accept", "text/event-stream")

	// Create HTTP client with no timeout for streaming
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to stream logs: %s", resp.Status)
	}

	// Read SSE stream
	reader := bufio.NewReader(resp.Body)
	done := make(chan bool)

	go func() {
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Printf("\nError reading stream: %v\n", err)
				}
				done <- true
				return
			}

			// Parse SSE format
			if strings.HasPrefix(line, "data: ") {
				logLine := strings.TrimPrefix(line, "data: ")
				logLine = strings.TrimSpace(logLine)
				if logLine != "" {
					// Format timestamp if present
					formattedLine := formatLogLine(logLine)
					fmt.Println(formattedLine)
				}
			}
		}
	}()

	// Wait for interrupt or stream end
	select {
	case <-interrupt:
		fmt.Println("\n\nStopping log stream...")
		return nil
	case <-done:
		return nil
	}
}

func fetchLogs(cfg *config.Config, apiName string) error {
	fmt.Printf("📋 Fetching logs for API: %s (last %d lines)\n", apiName, logsTail)
	fmt.Println(strings.Repeat("-", 80))

	// For now, we'll use a simple HTTP endpoint
	// In production, this would fetch from CloudWatch or similar
	url := fmt.Sprintf("%s/deployment/api/v1/logs/%s?tail=%d", cfg.API.BaseURL, apiName, logsTail)
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
		formattedLine := formatLogLine(line)
		fmt.Println(formattedLine)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}

func formatLogLine(line string) string {
	// Try to parse and format log lines
	// This is a simple implementation - could be enhanced with structured log parsing

	// If line starts with timestamp in brackets, color it
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
