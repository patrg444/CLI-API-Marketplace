package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/api-direct/cli/pkg/aws"
	"github.com/api-direct/cli/pkg/config"
	"github.com/api-direct/cli/pkg/terraform"
	"github.com/spf13/cobra"
)

var (
	forceDestroy bool
)

var destroyCmd = &cobra.Command{
	Use:   "destroy [API_NAME]",
	Short: "Destroy a BYOA deployment and all associated AWS resources",
	Long: `Destroy removes all AWS resources created for a BYOA deployment.

This command will:
- Remove all AWS resources (ALB, ECS, RDS, VPC, etc.)
- Delete the deployment from local configuration
- Clean up Terraform state

WARNING: This action is irreversible!`,
	Example: `  # Destroy a deployment
  apidirect destroy my-api
  
  # Force destroy without confirmation
  apidirect destroy my-api --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDestroy,
}

func init() {
	rootCmd.AddCommand(destroyCmd)
	
	destroyCmd.Flags().BoolVarP(&forceDestroy, "force", "f", false, "Force destroy without confirmation")
	destroyCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation prompt")
}

func runDestroy(cmd *cobra.Command, args []string) error {
	apiName := args[0]
	
	// Check if deployment exists in config
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	
	deploymentInfo, exists := getDeploymentInfo(cfg, apiName)
	if !exists {
		return fmt.Errorf("deployment '%s' not found", apiName)
	}
	
	// Check if it's a BYOA deployment
	deploymentType, _ := deploymentInfo["type"].(string)
	if deploymentType != "byoa" {
		return fmt.Errorf("deployment '%s' is not a BYOA deployment", apiName)
	}
	
	// Get deployment details
	awsAccount, _ := deploymentInfo["aws_account"].(string)
	awsRegion, _ := deploymentInfo["aws_region"].(string)
	environment, _ := deploymentInfo["environment"].(string)
	
	if environment == "" {
		environment = "prod"
	}
	
	// Check AWS credentials
	if err := aws.CheckAWSCLI(); err != nil {
		return err
	}
	
	if err := aws.CheckAWSCredentials(); err != nil {
		return err
	}
	
	// Verify we're using the correct AWS account
	accountInfo, err := aws.GetCallerIdentity()
	if err != nil {
		return fmt.Errorf("failed to get AWS account info: %w", err)
	}
	
	if accountInfo.AccountID != awsAccount {
		return fmt.Errorf("current AWS account (%s) doesn't match deployment account (%s)", 
			accountInfo.AccountID, awsAccount)
	}
	
	// Confirmation prompt
	if !forceDestroy && !yesFlag {
		fmt.Printf("⚠️  WARNING: This will destroy ALL resources for '%s'\n", apiName)
		fmt.Printf("   AWS Account: %s\n", awsAccount)
		fmt.Printf("   AWS Region: %s\n", awsRegion)
		fmt.Printf("   Environment: %s\n", environment)
		fmt.Printf("\n   Resources to be destroyed:\n")
		fmt.Printf("   - Application Load Balancer\n")
		fmt.Printf("   - ECS Fargate Service and Tasks\n")
		fmt.Printf("   - RDS Database (if enabled)\n")
		fmt.Printf("   - VPC and all networking components\n")
		fmt.Printf("   - IAM roles and policies\n")
		fmt.Printf("   - CloudWatch logs and metrics\n")
		fmt.Printf("\n   This action is IRREVERSIBLE!\n")
		fmt.Printf("\nType the API name to confirm destruction: ")
		
		var response string
		fmt.Scanln(&response)
		if response != apiName {
			return fmt.Errorf("Destruction cancelled")
		}
	}
	
	// Create working directory
	workDir := filepath.Join(os.TempDir(), fmt.Sprintf("apidirect-destroy-%s-%d", apiName, os.Getpid()))
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}
	defer os.RemoveAll(workDir)
	
	// Copy Terraform modules
	modulesPath := getModulesPath()
	if err := terraform.CopyModules(modulesPath, workDir); err != nil {
		return fmt.Errorf("failed to copy Terraform modules: %w", err)
	}
	
	// Create backend configuration
	backendConfig := fmt.Sprintf(`terraform {
  backend "s3" {
    bucket         = "apidirect-terraform-state-%s"
    key            = "deployments/%s/%s/terraform.tfstate"
    region         = "%s"
    dynamodb_table = "apidirect-terraform-locks"
    encrypt        = true
  }
}`, awsAccount, apiName, environment, awsRegion)
	
	backendFile := filepath.Join(workDir, "backend.tf")
	if err := os.WriteFile(backendFile, []byte(backendConfig), 0644); err != nil {
		return fmt.Errorf("failed to create backend config: %w", err)
	}
	
	// Initialize Terraform
	fmt.Println("🔧 Initializing Terraform...")
	tfClient := terraform.NewClient(workDir)
	if err := tfClient.Init(); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	
	// Set minimal variables for destroy
	tfClient.SetVars(map[string]interface{}{
		"project_name":          apiName,
		"environment":           environment,
		"aws_region":            awsRegion,
		"owner_email":           "destroy@api-direct.io", // Placeholder
		"api_direct_account_id": "123456789012",          // Placeholder
		
		// These are required but won't be used during destroy
		"container_image":   "placeholder",
		"container_port":    8080,
		"health_check_path": "/",
	})
	
	// Execute destroy
	fmt.Println("💥 Destroying infrastructure...")
	fmt.Println("   This may take 5-10 minutes...")
	
	if err := tfClient.Destroy(); err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}
	
	// Remove from config
	delete(cfg.Deployments, apiName)
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("⚠️  Warning: Failed to update config: %v\n", err)
	}
	
	// Clean up S3 state if bucket is empty
	// This is optional and we ignore errors
	stateBucket := fmt.Sprintf("apidirect-terraform-state-%s", awsAccount)
	_ = cleanupEmptyStateBucket(stateBucket, awsRegion)
	
	fmt.Println("\n✅ Deployment destroyed successfully!")
	fmt.Printf("   All AWS resources for '%s' have been removed.\n", apiName)
	
	return nil
}

func getDeploymentInfo(cfg *config.Config, apiName string) (map[string]interface{}, bool) {
	if cfg.Deployments == nil {
		return nil, false
	}
	
	if deployment, ok := cfg.Deployments[apiName]; ok {
		if deployMap, ok := deployment.(map[string]interface{}); ok {
			return deployMap, true
		}
	}
	
	return nil, false
}

func getModulesPath() string {
	// In production, this would be embedded in the CLI binary
	// For development, use relative path
	execPath, _ := os.Executable()
	baseDir := filepath.Dir(filepath.Dir(execPath))
	return filepath.Join(baseDir, "infrastructure", "deployments", "user-api")
}

func cleanupEmptyStateBucket(bucketName, region string) error {
	// Check if bucket is empty
	listCmd := exec.Command("aws", "s3", "ls", fmt.Sprintf("s3://%s", bucketName), "--region", region)
	output, err := listCmd.Output()
	if err != nil {
		return err // Bucket might not exist
	}
	
	// If bucket has content, don't delete it
	if strings.TrimSpace(string(output)) != "" {
		return nil
	}
	
	// Delete empty bucket
	deleteCmd := exec.Command("aws", "s3", "rb", fmt.Sprintf("s3://%s", bucketName), "--region", region)
	return deleteCmd.Run()
}