package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/api-direct/cli/pkg/config"
	"github.com/api-direct/cli/pkg/manifest"
	"github.com/spf13/cobra"
)

var (
	scaleMin      int
	scaleMax      int
	scaleCPU      int
	scaleMemory   string
	scaleDown     bool
	scaleToZero   bool
	autoScale     bool
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:   "scale [api-name] [options]",
	Short: "Scale your API deployment",
	Long: `Scale your API deployment by adjusting the number of replicas or resource limits.

Examples:
  apidirect scale my-api --replicas 5          # Scale to exactly 5 replicas
  apidirect scale my-api --min 2 --max 10      # Set auto-scaling range
  apidirect scale my-api --cpu 80              # Set CPU threshold for auto-scaling
  apidirect scale my-api --memory 1Gi          # Increase memory limit
  apidirect scale my-api --down                # Scale down to minimum replicas
  apidirect scale --replicas 3                 # Scale current API (from manifest)`,
	RunE: runScale,
}

func init() {
	rootCmd.AddCommand(scaleCmd)
	
	scaleCmd.Flags().IntVarP(&deployReplicas, "replicas", "r", 0, "Set exact number of replicas")
	scaleCmd.Flags().IntVar(&scaleMin, "min", 0, "Minimum replicas for auto-scaling")
	scaleCmd.Flags().IntVar(&scaleMax, "max", 0, "Maximum replicas for auto-scaling")
	scaleCmd.Flags().IntVar(&scaleCPU, "cpu", 0, "Target CPU percentage for auto-scaling (1-100)")
	scaleCmd.Flags().StringVar(&scaleMemory, "memory", "", "Memory limit per replica (e.g., 512Mi, 1Gi)")
	scaleCmd.Flags().BoolVar(&scaleDown, "down", false, "Scale down to minimum replicas")
	scaleCmd.Flags().BoolVar(&scaleToZero, "zero", false, "Scale to zero replicas (pause API)")
	scaleCmd.Flags().BoolVar(&autoScale, "auto", false, "Enable auto-scaling with smart defaults")
}

func runScale(cmd *cobra.Command, args []string) error {
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

	// Validate inputs
	if err := validateScaleInputs(); err != nil {
		return err
	}

	// Get current status first
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	currentStatus, err := getDeploymentScale(cfg, apiName)
	if err != nil {
		printWarning("Could not fetch current scale status")
	} else {
		displayCurrentScale(apiName, currentStatus)
	}

	// Determine what scaling operation to perform
	scaleRequest := buildScaleRequest(currentStatus)
	
	if scaleRequest == nil {
		return fmt.Errorf("no scaling changes specified")
	}

	// Confirm dangerous operations
	if scaleToZero {
		fmt.Print("\n⚠️  Scaling to zero will make your API unavailable. Continue? [y/N]: ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			return fmt.Errorf("scale operation cancelled")
		}
	}

	// Apply scaling changes
	fmt.Printf("\n🔄 Applying scale changes to '%s'...\n", apiName)
	
	result, err := applyScale(cfg, apiName, scaleRequest)
	if err != nil {
		return fmt.Errorf("failed to scale: %w", err)
	}

	// Display results
	displayScaleResult(result)
	
	return nil
}

func validateScaleInputs() error {
	// Validate replica counts
	if deployReplicas < 0 {
		return fmt.Errorf("replicas cannot be negative")
	}
	
	if scaleMin < 0 || scaleMax < 0 {
		return fmt.Errorf("min/max replicas cannot be negative")
	}
	
	if scaleMin > 0 && scaleMax > 0 && scaleMin > scaleMax {
		return fmt.Errorf("min replicas (%d) cannot be greater than max replicas (%d)", scaleMin, scaleMax)
	}
	
	// Validate CPU threshold
	if scaleCPU != 0 && (scaleCPU < 1 || scaleCPU > 100) {
		return fmt.Errorf("CPU threshold must be between 1 and 100")
	}
	
	// Validate memory format
	if scaleMemory != "" {
		if !isValidResourceString(scaleMemory) {
			return fmt.Errorf("invalid memory format: %s (use format like 512Mi, 1Gi)", scaleMemory)
		}
	}
	
	// Check conflicting options
	if scaleToZero && deployReplicas > 0 {
		return fmt.Errorf("cannot use --zero with --replicas")
	}
	
	if scaleDown && deployReplicas > 0 {
		return fmt.Errorf("cannot use --down with --replicas")
	}
	
	return nil
}

type DeploymentScale struct {
	CurrentReplicas int               `json:"current_replicas"`
	DesiredReplicas int               `json:"desired_replicas"`
	MinReplicas     int               `json:"min_replicas"`
	MaxReplicas     int               `json:"max_replicas"`
	AutoScaling     bool              `json:"auto_scaling"`
	CPUThreshold    int               `json:"cpu_threshold"`
	Resources       ResourceAllocation `json:"resources"`
	Status          string            `json:"status"`
	LastScaled      time.Time         `json:"last_scaled"`
}

type ResourceAllocation struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

type ScaleRequest struct {
	Replicas    *int               `json:"replicas,omitempty"`
	MinReplicas *int               `json:"min_replicas,omitempty"`
	MaxReplicas *int               `json:"max_replicas,omitempty"`
	AutoScaling *bool              `json:"auto_scaling,omitempty"`
	CPUTarget   *int               `json:"cpu_target,omitempty"`
	Resources   *ResourceAllocation `json:"resources,omitempty"`
}

func getDeploymentScale(cfg *config.Config, apiName string) (*DeploymentScale, error) {
	url := fmt.Sprintf("%s/deployment/v1/scale/%s", cfg.API.BaseURL, apiName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get scale info: %s", resp.Status)
	}

	var scale DeploymentScale
	if err := json.NewDecoder(resp.Body).Decode(&scale); err != nil {
		return nil, err
	}

	return &scale, nil
}

func displayCurrentScale(apiName string, scale *DeploymentScale) {
	fmt.Printf("\n📊 Current scale for '%s':\n", apiName)
	fmt.Println(strings.Repeat("─", 50))
	
	// Replica info
	if scale.AutoScaling {
		fmt.Printf("🔄 Auto-scaling: ENABLED\n")
		fmt.Printf("   Range: %d - %d replicas\n", scale.MinReplicas, scale.MaxReplicas)
		fmt.Printf("   Current: %d replica(s)\n", scale.CurrentReplicas)
		fmt.Printf("   CPU Target: %d%%\n", scale.CPUThreshold)
	} else {
		fmt.Printf("🔢 Fixed replicas: %d\n", scale.CurrentReplicas)
	}
	
	// Resource info
	fmt.Printf("\n💾 Resources per replica:\n")
	fmt.Printf("   Memory: %s\n", scale.Resources.Memory)
	fmt.Printf("   CPU: %s\n", scale.Resources.CPU)
	
	// Status
	statusIcon := "✅"
	if scale.Status != "healthy" {
		statusIcon = "⚠️"
	}
	fmt.Printf("\n%s Status: %s\n", statusIcon, scale.Status)
	
	if !scale.LastScaled.IsZero() {
		fmt.Printf("⏰ Last scaled: %s\n", scale.LastScaled.Format("2006-01-02 15:04:05"))
	}
}

func buildScaleRequest(current *DeploymentScale) *ScaleRequest {
	req := &ScaleRequest{}
	hasChanges := false

	// Handle special cases first
	if scaleToZero {
		zero := 0
		req.Replicas = &zero
		return req
	}

	if scaleDown && current != nil {
		min := current.MinReplicas
		if min == 0 {
			min = 1
		}
		req.Replicas = &min
		return req
	}

	// Handle auto-scaling setup
	if autoScale {
		enabled := true
		req.AutoScaling = &enabled
		
		// Set smart defaults if not specified
		if scaleMin == 0 && scaleMax == 0 {
			min := 2
			max := 10
			req.MinReplicas = &min
			req.MaxReplicas = &max
		}
		if scaleCPU == 0 {
			cpu := 70
			req.CPUTarget = &cpu
		}
		hasChanges = true
	}

	// Handle exact replica count
	if deployReplicas > 0 {
		req.Replicas = &deployReplicas
		hasChanges = true
		
		// Disable auto-scaling when setting exact count
		disabled := false
		req.AutoScaling = &disabled
	}

	// Handle auto-scaling parameters
	if scaleMin > 0 {
		req.MinReplicas = &scaleMin
		hasChanges = true
	}
	
	if scaleMax > 0 {
		req.MaxReplicas = &scaleMax
		hasChanges = true
	}
	
	if scaleCPU > 0 {
		req.CPUTarget = &scaleCPU
		hasChanges = true
	}

	// Handle resource changes
	if scaleMemory != "" {
		if req.Resources == nil {
			req.Resources = &ResourceAllocation{}
			if current != nil {
				req.Resources.CPU = current.Resources.CPU
			}
		}
		req.Resources.Memory = scaleMemory
		hasChanges = true
	}

	if !hasChanges {
		return nil
	}

	return req
}

func applyScale(cfg *config.Config, apiName string, req *ScaleRequest) (*DeploymentScale, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/deployment/v1/scale/%s", cfg.API.BaseURL, apiName)
	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("scale request failed: %s - %s", resp.Status, string(body))
	}

	var result DeploymentScale
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func displayScaleResult(scale *DeploymentScale) {
	fmt.Println("\n✅ Scale operation completed successfully!")
	
	// Show what changed
	fmt.Println("\n📋 New configuration:")
	
	if scale.AutoScaling {
		fmt.Printf("   🔄 Auto-scaling: %d - %d replicas\n", scale.MinReplicas, scale.MaxReplicas)
		fmt.Printf("   📊 CPU target: %d%%\n", scale.CPUThreshold)
	} else {
		fmt.Printf("   🔢 Fixed replicas: %d\n", scale.DesiredReplicas)
	}
	
	if scale.Resources.Memory != "" {
		fmt.Printf("   💾 Memory: %s per replica\n", scale.Resources.Memory)
	}
	
	// Deployment progress
	if scale.CurrentReplicas != scale.DesiredReplicas {
		fmt.Printf("\n⏳ Scaling in progress: %d → %d replicas\n", 
			scale.CurrentReplicas, scale.DesiredReplicas)
		fmt.Println("💡 Use 'apidirect status' to monitor progress")
	} else {
		fmt.Printf("\n✅ Currently running: %d replica(s)\n", scale.CurrentReplicas)
	}
	
	// Tips based on configuration
	if scale.MinReplicas == 0 && scale.CurrentReplicas == 0 {
		fmt.Println("\n⚠️  API is scaled to zero - it will start when first request arrives")
		fmt.Println("💡 First request may take longer due to cold start")
	}
	
	if scale.AutoScaling && scale.MaxReplicas > 10 {
		fmt.Println("\n💰 Note: Higher replica counts may increase costs during traffic spikes")
	}
}

func isValidResourceString(resource string) bool {
	// Simple validation for Kubernetes-style resource strings
	validSuffixes := []string{"Ki", "Mi", "Gi", "Ti", "m", "k", "M", "G", "T"}
	
	for _, suffix := range validSuffixes {
		if strings.HasSuffix(resource, suffix) {
			numberPart := strings.TrimSuffix(resource, suffix)
			if _, err := strconv.ParseFloat(numberPart, 64); err == nil {
				return true
			}
		}
	}
	
	// Also allow plain numbers (interpreted as bytes or millicores)
	if _, err := strconv.ParseFloat(resource, 64); err == nil {
		return true
	}
	
	return false
}

// Demo mode functions
func getDemoScale(apiName string) *DeploymentScale {
	return &DeploymentScale{
		CurrentReplicas: 3,
		DesiredReplicas: 3,
		MinReplicas:     1,
		MaxReplicas:     10,
		AutoScaling:     true,
		CPUThreshold:    70,
		Resources: ResourceAllocation{
			Memory: "512Mi",
			CPU:    "250m",
		},
		Status:     "healthy",
		LastScaled: time.Now().Add(-2 * time.Hour),
	}
}

func applyDemoScale(apiName string, req *ScaleRequest) *DeploymentScale {
	// Simulate scaling
	result := getDemoScale(apiName)
	
	if req.Replicas != nil {
		result.DesiredReplicas = *req.Replicas
		result.AutoScaling = false
	}
	
	if req.MinReplicas != nil {
		result.MinReplicas = *req.MinReplicas
		result.AutoScaling = true
	}
	
	if req.MaxReplicas != nil {
		result.MaxReplicas = *req.MaxReplicas
		result.AutoScaling = true
	}
	
	if req.CPUTarget != nil {
		result.CPUThreshold = *req.CPUTarget
	}
	
	if req.Resources != nil && req.Resources.Memory != "" {
		result.Resources.Memory = req.Resources.Memory
	}
	
	result.LastScaled = time.Now()
	
	return result
}