package orchestrator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/api-direct/cli/pkg/aws"
	"github.com/api-direct/cli/pkg/config"
	"github.com/api-direct/cli/pkg/manifest"
	"github.com/api-direct/cli/pkg/terraform"
)

// BYOADeployment manages BYOA deployments
type BYOADeployment struct {
	APIName       string
	Manifest      *manifest.Manifest
	WorkDir       string
	StateBackend  StateBackend
	AWSAccountID  string
	AWSRegion     string
	Environment   string
	OutputWriter  io.Writer
}

// StateBackend configuration for Terraform state
type StateBackend struct {
	Bucket    string
	Key       string
	Region    string
	DynamoDB  string
}

// DeploymentResult contains deployment outputs
type DeploymentResult struct {
	APIURL          string `json:"api_url"`
	LoadBalancerDNS string `json:"load_balancer_dns"`
	DeploymentID    string `json:"deployment_id"`
	AWSRegion       string `json:"aws_region"`
	AWSAccountID    string `json:"aws_account_id"`
	Timestamp       string `json:"timestamp"`
}

// NewBYOADeployment creates a new BYOA deployment
func NewBYOADeployment(apiName string, m *manifest.Manifest) (*BYOADeployment, error) {
	// Get AWS account info
	accountInfo, err := aws.GetCallerIdentity()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS account info: %w", err)
	}

	// Get AWS region
	region, err := aws.GetRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS region: %w", err)
	}

	// Create working directory
	workDir := filepath.Join(os.TempDir(), fmt.Sprintf("apidirect-deploy-%s-%d", apiName, time.Now().Unix()))

	// Determine environment (default to prod)
	environment := "prod"

	return &BYOADeployment{
		APIName:      apiName,
		Manifest:     m,
		WorkDir:      workDir,
		AWSAccountID: accountInfo.AccountID,
		AWSRegion:    region,
		Environment:  environment,
		StateBackend: StateBackend{
			Bucket:   fmt.Sprintf("apidirect-terraform-state-%s", accountInfo.AccountID),
			Key:      fmt.Sprintf("deployments/%s/%s/terraform.tfstate", apiName, environment),
			Region:   region,
			DynamoDB: "apidirect-terraform-locks",
		},
		OutputWriter: os.Stdout,
	}, nil
}

// Prepare prepares the deployment environment
func (d *BYOADeployment) Prepare() error {
	// Create working directory
	if err := os.MkdirAll(d.WorkDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}

	// Copy Terraform modules
	modulesPath := d.getModulesPath()
	if err := terraform.CopyModules(modulesPath, d.WorkDir); err != nil {
		return fmt.Errorf("failed to copy Terraform modules: %w", err)
	}

	// Create backend configuration
	if err := d.createBackendConfig(); err != nil {
		return fmt.Errorf("failed to create backend config: %w", err)
	}

	// Ensure state backend exists
	if err := d.ensureStateBackend(); err != nil {
		return fmt.Errorf("failed to setup state backend: %w", err)
	}

	return nil
}

// Plan creates a deployment plan
func (d *BYOADeployment) Plan() error {
	client := terraform.NewClient(d.WorkDir)
	
	// Set Terraform variables from manifest
	vars := d.getTerraformVars()
	client.SetVars(vars)

	// Initialize Terraform
	fmt.Fprintln(d.OutputWriter, "🔧 Initializing Terraform...")
	if err := client.Init(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Create plan
	fmt.Fprintln(d.OutputWriter, "📋 Creating deployment plan...")
	planFile := filepath.Join(d.WorkDir, "tfplan")
	if err := client.StreamingPlan(planFile, d.OutputWriter); err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	fmt.Fprintln(d.OutputWriter, "\n✅ Deployment plan created successfully")
	return nil
}

// Deploy executes the deployment
func (d *BYOADeployment) Deploy() (*DeploymentResult, error) {
	client := terraform.NewClient(d.WorkDir)
	
	// Set Terraform variables
	vars := d.getTerraformVars()
	client.SetVars(vars)

	// Apply the plan
	fmt.Fprintln(d.OutputWriter, "🚀 Deploying infrastructure...")
	planFile := filepath.Join(d.WorkDir, "tfplan")
	if err := client.StreamingApply(planFile, d.OutputWriter); err != nil {
		return nil, fmt.Errorf("deployment failed: %w", err)
	}

	// Get outputs
	fmt.Fprintln(d.OutputWriter, "\n📊 Retrieving deployment details...")
	outputs, err := client.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment outputs: %w", err)
	}

	// Parse outputs
	result := &DeploymentResult{
		AWSRegion:    d.AWSRegion,
		AWSAccountID: d.AWSAccountID,
		Timestamp:    time.Now().Format(time.RFC3339),
		DeploymentID: fmt.Sprintf("%s-%s-%s", d.APIName, d.Environment, d.AWSAccountID),
	}

	// Extract specific outputs
	if apiURL, ok := outputs["api_url"].(string); ok {
		result.APIURL = apiURL
	}
	if lbDNS, ok := outputs["load_balancer_dns"].(string); ok {
		result.LoadBalancerDNS = lbDNS
	}

	// Save deployment info
	if err := d.saveDeploymentInfo(result); err != nil {
		fmt.Fprintf(d.OutputWriter, "⚠️  Warning: Failed to save deployment info: %v\n", err)
	}

	return result, nil
}

// Cleanup removes temporary files
func (d *BYOADeployment) Cleanup() error {
	if d.WorkDir != "" && strings.Contains(d.WorkDir, "apidirect-deploy") {
		return os.RemoveAll(d.WorkDir)
	}
	return nil
}

// getTerraformVars converts manifest to Terraform variables
func (d *BYOADeployment) getTerraformVars() map[string]interface{} {
	cfg, _ := config.LoadConfig()
	
	vars := map[string]interface{}{
		"project_name":          d.APIName,
		"environment":           d.Environment,
		"aws_region":            d.AWSRegion,
		"owner_email":           cfg.User.Email,
		"api_direct_account_id": "123456789012", // This should come from API-Direct config
		
		// Container configuration from manifest
		"container_image":     d.getContainerImage(),
		"container_port":      d.Manifest.Port,
		"health_check_path":   d.Manifest.HealthCheck,
		
		// Resource configuration
		"cpu":    d.getCPUValue(),
		"memory": d.getMemoryValue(),
		
		// Scaling configuration
		"min_capacity": d.getMinCapacity(),
		"max_capacity": d.getMaxCapacity(),
		
		// Database configuration
		"enable_database": d.shouldEnableDatabase(),
		
		// Tags
		"tags": map[string]string{
			"ManagedBy":   "API-Direct",
			"DeployedBy":  "CLI",
			"Environment": d.Environment,
			"APIName":     d.APIName,
		},
	}

	// Add environment variables if any
	if len(d.Manifest.Env.Required) > 0 || len(d.Manifest.Env.Optional) > 0 {
		envVars := make(map[string]string)
		// Add placeholder values for required env vars
		for _, key := range d.Manifest.Env.Required {
			envVars[key] = fmt.Sprintf("PLACEHOLDER_%s", key)
		}
		vars["environment_variables"] = envVars
	}

	return vars
}

// Helper methods

func (d *BYOADeployment) getModulesPath() string {
	// In production, this would be embedded in the CLI binary
	// For development, use relative path
	execPath, _ := os.Executable()
	baseDir := filepath.Dir(filepath.Dir(execPath))
	return filepath.Join(baseDir, "infrastructure", "deployments", "user-api")
}

func (d *BYOADeployment) createBackendConfig() error {
	backendConfig := fmt.Sprintf(`terraform {
  backend "s3" {
    bucket         = "%s"
    key            = "%s"
    region         = "%s"
    dynamodb_table = "%s"
    encrypt        = true
  }
}`, d.StateBackend.Bucket, d.StateBackend.Key, d.StateBackend.Region, d.StateBackend.DynamoDB)

	backendFile := filepath.Join(d.WorkDir, "backend.tf")
	return os.WriteFile(backendFile, []byte(backendConfig), 0644)
}

func (d *BYOADeployment) ensureStateBackend() error {
	// Create S3 bucket for state
	if err := aws.CreateS3Bucket(d.StateBackend.Bucket, d.StateBackend.Region); err != nil {
		return fmt.Errorf("failed to create state bucket: %w", err)
	}

	// Create DynamoDB table for locking
	if err := aws.CreateDynamoDBTable(d.StateBackend.DynamoDB, d.StateBackend.Region); err != nil {
		return fmt.Errorf("failed to create lock table: %w", err)
	}

	return nil
}

func (d *BYOADeployment) getContainerImage() string {
	// This would be built and pushed to ECR
	// For now, return a placeholder
	return fmt.Sprintf("%s:%s", d.APIName, "latest")
}

func (d *BYOADeployment) getCPUValue() int {
	if d.Manifest.Resources != nil && d.Manifest.Resources.CPU != "" {
		// Parse CPU value (e.g., "250m" -> 256)
		// Simplified for now
		return 256
	}
	return 256 // Default
}

func (d *BYOADeployment) getMemoryValue() int {
	if d.Manifest.Resources != nil && d.Manifest.Resources.Memory != "" {
		// Parse memory value (e.g., "512Mi" -> 512)
		// Simplified for now
		return 512
	}
	return 512 // Default
}

func (d *BYOADeployment) getMinCapacity() int {
	if d.Manifest.Scaling != nil {
		return d.Manifest.Scaling.Min
	}
	return 1
}

func (d *BYOADeployment) getMaxCapacity() int {
	if d.Manifest.Scaling != nil {
		return d.Manifest.Scaling.Max
	}
	return 10
}

func (d *BYOADeployment) shouldEnableDatabase() bool {
	// Check if manifest indicates database requirement
	for _, env := range d.Manifest.Env.Required {
		if strings.Contains(strings.ToUpper(env), "DATABASE") ||
		   strings.Contains(strings.ToUpper(env), "DB_") {
			return true
		}
	}
	return false
}

func (d *BYOADeployment) saveDeploymentInfo(result *DeploymentResult) error {
	// Save deployment info to local config
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Initialize deployments map if needed
	if cfg.Deployments == nil {
		cfg.Deployments = make(map[string]interface{})
	}

	// Save deployment info
	cfg.Deployments[d.APIName] = map[string]interface{}{
		"type":        "byoa",
		"aws_account": d.AWSAccountID,
		"aws_region":  d.AWSRegion,
		"environment": d.Environment,
		"api_url":     result.APIURL,
		"deployed_at": result.Timestamp,
	}

	return config.SaveConfig(cfg)
}