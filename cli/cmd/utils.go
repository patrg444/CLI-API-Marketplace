package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// formatNumber formats a number with thousand separators
func formatNumber(n int64) string {
	str := fmt.Sprintf("%d", n)
	var result []string
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, ",")
		}
		result = append(result, string(char))
	}
	return strings.Join(result, "")
}

// outputJSON formats and outputs data as JSON
func outputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// openBrowser opens a URL in the default browser
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default: // linux and others
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// makeAuthenticatedRequest already defined in test_utils.go

// handleErrorResponse already defined in test_utils.go

// confirmAction already defined in test_utils.go

// readManifest reads the apidirect.yaml manifest file
func readManifest(path string) (map[string]interface{}, error) {
	// This is a placeholder - actual implementation would read and parse YAML
	return map[string]interface{}{
		"name": "api",
		"version": "1.0.0",
	}, nil
}

// getCurrencySymbol already defined in subscriptions.go

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getCurrentAPIName already defined in docs.go

// getLatestRelease already defined in self_update.go