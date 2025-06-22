package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/api-direct/cli/pkg/manifest"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	statusJSON     bool
	statusDetailed bool
	statusWatch    bool
	statusInterval int
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [api-name]",
	Short: "Show deployment status and health",
	Long: `Show detailed status information about your deployed API including
health, performance metrics, and current configuration.

Examples:
  apidirect status                  # Status of current API (from manifest)
  apidirect status my-api           # Status of specific API
  apidirect status --detailed       # Show detailed metrics
  apidirect status --watch          # Watch status updates live
  apidirect status --json           # Output as JSON`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
	
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output as JSON")
	statusCmd.Flags().BoolVarP(&statusDetailed, "detailed", "d", false, "Show detailed information")
	statusCmd.Flags().BoolVarP(&statusWatch, "watch", "w", false, "Watch status updates")
	statusCmd.Flags().IntVar(&statusInterval, "interval", 5, "Watch interval in seconds")
}

type DeploymentStatus struct {
	APIName      string                 `json:"api_name"`
	Status       string                 `json:"status"`
	Health       HealthStatus           `json:"health"`
	Deployment   DeploymentInfo         `json:"deployment"`
	Scale        ScaleInfo              `json:"scale"`
	Endpoints    []EndpointStatus       `json:"endpoints"`
	Metrics      PerformanceMetrics     `json:"metrics"`
	Resources    ResourceUsage          `json:"resources"`
	LastUpdated  time.Time              `json:"last_updated"`
}

type HealthStatus struct {
	Overall    string            `json:"overall"`
	Replicas   []ReplicaHealth   `json:"replicas"`
	LastCheck  time.Time         `json:"last_check"`
}

type ReplicaHealth struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Uptime    string    `json:"uptime"`
	LastSeen  time.Time `json:"last_seen"`
}

type DeploymentInfo struct {
	ID           string    `json:"id"`
	Version      string    `json:"version"`
	Environment  string    `json:"environment"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	URL          string    `json:"url"`
	CustomDomain string    `json:"custom_domain,omitempty"`
}

type ScaleInfo struct {
	CurrentReplicas int  `json:"current_replicas"`
	DesiredReplicas int  `json:"desired_replicas"`
	MinReplicas     int  `json:"min_replicas"`
	MaxReplicas     int  `json:"max_replicas"`
	AutoScaling     bool `json:"auto_scaling"`
}

type EndpointStatus struct {
	Path         string  `json:"path"`
	Method       string  `json:"method"`
	Status       string  `json:"status"`
	ResponseTime float64 `json:"response_time_ms"`
	LastTested   time.Time `json:"last_tested"`
}

type PerformanceMetrics struct {
	RequestsPerSecond float64 `json:"requests_per_second"`
	AverageLatency    float64 `json:"average_latency_ms"`
	P95Latency        float64 `json:"p95_latency_ms"`
	P99Latency        float64 `json:"p99_latency_ms"`
	ErrorRate         float64 `json:"error_rate_percent"`
	Throughput        string  `json:"throughput"`
}

type ResourceUsage struct {
	CPUUsage    float64 `json:"cpu_usage_percent"`
	MemoryUsage float64 `json:"memory_usage_percent"`
	MemoryMB    int     `json:"memory_mb"`
	DiskUsage   float64 `json:"disk_usage_percent"`
	NetworkIn   string  `json:"network_in"`
	NetworkOut  string  `json:"network_out"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Check authentication
	if !config.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'apidirect login' first")
	}

	// Get API name
	var apiName string
	if len(args) > 0 {
		apiName = args[0]
	} else {
		// Try to get from manifest
		manifestPath, err := manifest.FindManifest(".")
		if err != nil {
			return fmt.Errorf("no API name provided and no manifest found")
		}
		
		m, err := manifest.Load(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}
		
		apiName = m.Name
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	if statusWatch {
		return watchStatus(cfg, apiName)
	}

	// Get status once
	status, err := getDeploymentStatusInfo(cfg, apiName, statusDetailed)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	if statusJSON {
		return outputStatusJSON(status)
	}

	displayStatus(status, statusDetailed)
	return nil
}

func getDeploymentStatusInfo(cfg *config.Config, apiName string, detailed bool) (*DeploymentStatus, error) {
	url := fmt.Sprintf("%s/deployment/v1/status/%s", cfg.API.BaseURL, apiName)
	if detailed {
		url += "?detailed=true"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("API '%s' not found", apiName)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get status: %s - %s", resp.Status, string(body))
	}

	var status DeploymentStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	// For demo mode, enhance with mock data
	if os.Getenv("APIDIRECT_DEMO_MODE") == "true" {
		enhanceStatusForDemo(&status)
	}

	return &status, nil
}

func displayStatus(status *DeploymentStatus, detailed bool) {
	// Header
	fmt.Printf("\n🚀 %s\n", color.CyanString(status.APIName))
	fmt.Println(strings.Repeat("═", 50))

	// Overall status
	statusIcon, statusColor := getStatusIcon(status.Status)
	fmt.Printf("\n%s Status: %s\n", statusIcon, statusColor(strings.ToUpper(status.Status)))

	// Deployment info
	fmt.Printf("\n📋 Deployment\n")
	fmt.Printf("   ID:          %s\n", status.Deployment.ID)
	fmt.Printf("   Version:     %s\n", status.Deployment.Version)
	fmt.Printf("   Environment: %s\n", status.Deployment.Environment)
	fmt.Printf("   URL:         %s\n", color.BlueString(status.Deployment.URL))
	if status.Deployment.CustomDomain != "" {
		fmt.Printf("   Domain:      %s\n", color.BlueString(status.Deployment.CustomDomain))
	}
	fmt.Printf("   Deployed:    %s\n", formatDuration(time.Since(status.Deployment.CreatedAt)))

	// Health status
	fmt.Printf("\n💚 Health\n")
	healthIcon, healthColor := getHealthIcon(status.Health.Overall)
	fmt.Printf("   Overall:     %s %s\n", healthIcon, healthColor(status.Health.Overall))
	
	if len(status.Health.Replicas) > 0 {
		fmt.Printf("   Replicas:    ")
		for i, replica := range status.Health.Replicas {
			if i > 0 {
				fmt.Print(", ")
			}
			replicaIcon := "✅"
			if replica.Status != "healthy" {
				replicaIcon = "❌"
			}
			fmt.Printf("%s %s", replicaIcon, replica.ID[:8])
		}
		fmt.Println()
	}

	// Scale info
	fmt.Printf("\n📊 Scale\n")
	if status.Scale.AutoScaling {
		fmt.Printf("   Mode:        Auto-scaling\n")
		fmt.Printf("   Range:       %d - %d replicas\n", status.Scale.MinReplicas, status.Scale.MaxReplicas)
		fmt.Printf("   Current:     %d / %d replicas\n", status.Scale.CurrentReplicas, status.Scale.DesiredReplicas)
	} else {
		fmt.Printf("   Mode:        Fixed\n")
		fmt.Printf("   Replicas:    %d\n", status.Scale.CurrentReplicas)
	}

	// Performance metrics
	if status.Metrics.RequestsPerSecond > 0 || detailed {
		fmt.Printf("\n⚡ Performance\n")
		fmt.Printf("   Requests/sec: %.1f\n", status.Metrics.RequestsPerSecond)
		fmt.Printf("   Avg latency:  %.1fms\n", status.Metrics.AverageLatency)
		
		if detailed {
			fmt.Printf("   P95 latency:  %.1fms\n", status.Metrics.P95Latency)
			fmt.Printf("   P99 latency:  %.1fms\n", status.Metrics.P99Latency)
			fmt.Printf("   Error rate:   %.2f%%\n", status.Metrics.ErrorRate)
			if status.Metrics.Throughput != "" {
				fmt.Printf("   Throughput:   %s\n", status.Metrics.Throughput)
			}
		}
	}

	// Resource usage
	if detailed {
		fmt.Printf("\n💾 Resources\n")
		fmt.Printf("   CPU:         %.1f%%\n", status.Resources.CPUUsage)
		fmt.Printf("   Memory:      %.1f%% (%dMB)\n", status.Resources.MemoryUsage, status.Resources.MemoryMB)
		if status.Resources.NetworkIn != "" {
			fmt.Printf("   Network In:  %s\n", status.Resources.NetworkIn)
			fmt.Printf("   Network Out: %s\n", status.Resources.NetworkOut)
		}
	}

	// Endpoint status
	if len(status.Endpoints) > 0 && detailed {
		fmt.Printf("\n🔗 Endpoints\n")
		maxPath := 0
		for _, ep := range status.Endpoints {
			path := fmt.Sprintf("%s %s", ep.Method, ep.Path)
			if len(path) > maxPath {
				maxPath = len(path)
			}
		}
		
		for _, ep := range status.Endpoints {
			path := fmt.Sprintf("%s %s", ep.Method, ep.Path)
			statusIcon := "✅"
			if ep.Status != "healthy" {
				statusIcon = "❌"
			}
			fmt.Printf("   %s %-*s  %.0fms\n", statusIcon, maxPath, path, ep.ResponseTime)
		}
	}

	// Footer
	fmt.Printf("\n⏰ Last updated: %s\n", status.LastUpdated.Format("15:04:05"))
}

func watchStatus(cfg *config.Config, apiName string) error {
	fmt.Printf("👁  Watching status for '%s' (press Ctrl+C to stop)\n", apiName)
	
	ticker := time.NewTicker(time.Duration(statusInterval) * time.Second)
	defer ticker.Stop()
	
	// Set up interrupt handler
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	
	// Clear screen function
	clearScreen := func() {
		fmt.Print("\033[H\033[2J")
	}
	
	// Initial display
	status, err := getDeploymentStatusInfo(cfg, apiName, statusDetailed)
	if err != nil {
		return err
	}
	clearScreen()
	displayStatus(status, statusDetailed)
	
	// Watch loop
	for {
		select {
		case <-ticker.C:
			status, err := getDeploymentStatusInfo(cfg, apiName, statusDetailed)
			if err != nil {
				printError(fmt.Sprintf("Failed to get status: %v", err))
				continue
			}
			clearScreen()
			displayStatus(status, statusDetailed)
			
		case <-interrupt:
			fmt.Println("\n\n🛑 Stopped watching")
			return nil
		}
	}
}

func outputStatusJSON(status *DeploymentStatus) error {
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

func getStatusIcon(status string) (string, func(format string, a ...interface{}) string) {
	switch strings.ToLower(status) {
	case "running", "healthy", "active":
		return "🟢", color.GreenString
	case "deploying", "updating", "scaling":
		return "🔄", color.YellowString
	case "stopped", "paused", "idle":
		return "⏸️ ", color.HiBlackString
	case "error", "failed", "unhealthy":
		return "🔴", color.RedString
	default:
		return "⚪", color.WhiteString
	}
}

func getHealthIcon(health string) (string, func(format string, a ...interface{}) string) {
	switch strings.ToLower(health) {
	case "healthy":
		return "💚", color.GreenString
	case "degraded":
		return "💛", color.YellowString
	case "unhealthy":
		return "❤️ ", color.RedString
	default:
		return "🤍", color.WhiteString
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds ago", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	} else {
		days := int(d.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	}
}

func enhanceStatusForDemo(status *DeploymentStatus) {
	// Add realistic demo data
	status.Status = "running"
	status.Health.Overall = "healthy"
	
	// Add some replicas
	if len(status.Health.Replicas) == 0 {
		status.Health.Replicas = []ReplicaHealth{
			{
				ID:       "api-abc123",
				Status:   "healthy",
				Uptime:   "2h 45m",
				LastSeen: time.Now(),
			},
			{
				ID:       "api-def456",
				Status:   "healthy",
				Uptime:   "2h 45m",
				LastSeen: time.Now(),
			},
		}
	}
	
	// Add performance metrics
	status.Metrics = PerformanceMetrics{
		RequestsPerSecond: 127.3,
		AverageLatency:    23.5,
		P95Latency:        45.2,
		P99Latency:        89.7,
		ErrorRate:         0.02,
		Throughput:        "1.2MB/s",
	}
	
	// Add resource usage
	status.Resources = ResourceUsage{
		CPUUsage:    35.7,
		MemoryUsage: 62.3,
		MemoryMB:    319,
		NetworkIn:   "125KB/s",
		NetworkOut:  "1.1MB/s",
	}
	
	// Add endpoint status
	if len(status.Endpoints) == 0 {
		status.Endpoints = []EndpointStatus{
			{
				Path:         "/health",
				Method:       "GET",
				Status:       "healthy",
				ResponseTime: 5.2,
				LastTested:   time.Now().Add(-30 * time.Second),
			},
			{
				Path:         "/api/users",
				Method:       "GET",
				Status:       "healthy",
				ResponseTime: 18.7,
				LastTested:   time.Now().Add(-30 * time.Second),
			},
			{
				Path:         "/api/users",
				Method:       "POST",
				Status:       "healthy",
				ResponseTime: 42.3,
				LastTested:   time.Now().Add(-30 * time.Second),
			},
		}
	}
}