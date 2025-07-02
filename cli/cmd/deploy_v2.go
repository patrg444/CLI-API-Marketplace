package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/api-direct/cli/pkg/aws"
	"github.com/api-direct/cli/pkg/config"
	"github.com/api-direct/cli/pkg/errors"
	"github.com/api-direct/cli/pkg/manifest"
	"github.com/api-direct/cli/pkg/orchestrator"
	"github.com/api-direct/cli/pkg/terraform"
	"github.com/spf13/cobra"
)

// This is the new deploy command that uses the manifest system

func runDeployV2(cmd *cobra.Command, args []string) error {
	// Check for demo mode first
	if os.Getenv("APIDIRECT_DEMO_MODE") == "true" {
		return runDemoDeploymentV2(cmd, args)
	}

	// Check authentication
	if !config.IsAuthenticated() {
		err := errors.NewAuthError("You must be authenticated to deploy APIs")
		errors.OutputError(err, outputFormat == "json")
		return fmt.Errorf(err.Error())
	}

	// Find and load manifest
	manifestPath, err := manifest.FindManifest(".")
	if err != nil {
		return fmt.Errorf("no manifest found. Run 'apidirect import' to create one")
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Validate manifest before deployment
	if err := m.Validate(); err != nil {
		return fmt.Errorf("manifest validation failed: %w\nRun 'apidirect validate' for details", err)
	}

	// Override name if provided
	apiName := m.Name
	if len(args) > 0 {
		apiName = args[0]
	}

	// Choose deployment path based on mode
	if hostedMode {
		if outputFormat != "json" {
			fmt.Printf("🚀 Deploying '%s' to hosted infrastructure\n", apiName)
		}
		return deployHostedV2(apiName, m)
	} else {
		if outputFormat != "json" {
			fmt.Printf("🚀 Deploying '%s' to your AWS account\n", apiName)
		}
		return deployBYOAV2(apiName, m)
	}
}

func deployHostedV2(apiName string, m *manifest.Manifest) error {
	if outputFormat != "json" {
		fmt.Println("☁️  Using API-Direct hosted infrastructure...")
		fmt.Printf("📋 Configuration: %s runtime, port %d\n", m.Runtime, m.Port)
	}

	// Check if deployment already exists
	existingDeployment, err := checkExistingDeployment(apiName)
	if err != nil && outputFormat != "json" {
		fmt.Println("⚠️  Could not check for existing deployments")
	}

	if existingDeployment != nil && !forceFlag {
		if !yesFlag && outputFormat != "json" {
			fmt.Printf("\n⚠️  API '%s' already exists. Update it? [y/N]: ", apiName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				return fmt.Errorf("deployment cancelled")
			}
		}

		if outputFormat != "json" {
			fmt.Println("🔄 Updating existing deployment...")
		}
	}

	// Build container image
	if outputFormat != "json" {
		fmt.Println("🐳 Building container image...")
	}

	// Generate Dockerfile if not provided
	var dockerfilePath string
	if m.Files.Dockerfile != "" {
		dockerfilePath = m.Files.Dockerfile
		if outputFormat != "json" {
			fmt.Printf("   Using custom Dockerfile: %s\n", dockerfilePath)
		}
	} else {
		// Generate Dockerfile
		dockerfileContent := m.GenerateDockerfile()
		dockerfilePath = ".apidirect.dockerfile"
		if err := ioutil.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
			return fmt.Errorf("failed to write Dockerfile: %w", err)
		}
		defer os.Remove(dockerfilePath)
		if outputFormat != "json" {
			fmt.Println("   Generated Dockerfile from manifest")
		}
	}

	// Create build context
	buildContext, err := createBuildContextV2(dockerfilePath)
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	defer os.Remove(buildContext)

	// Upload and build
	imageTag := fmt.Sprintf("%s:%d", apiName, time.Now().Unix())
	if outputFormat != "json" {
		fmt.Println("⬆️  Uploading code and building image...")
	}

	buildResult, err := uploadAndBuild(apiName, imageTag, buildContext)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Deploy the built image
	if outputFormat != "json" {
		fmt.Println("🚀 Deploying to platform...")
	}

	deploymentResult, err := deployBuiltImage(apiName, buildResult.ImageTag, m)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Wait for deployment to be ready
	if outputFormat != "json" {
		fmt.Print("⏳ Waiting for deployment to be ready")
	}

	if err := waitForDeploymentReady(deploymentResult.DeploymentID); err != nil {
		return fmt.Errorf("deployment failed to become ready: %w", err)
	}

	if outputFormat != "json" {
		fmt.Println(" ✓")
	}

	// Output results
	if outputFormat == "json" {
		result := map[string]interface{}{
			"success":        true,
			"api_name":       apiName,
			"api_url":        deploymentResult.Endpoint,
			"deployment_id":  deploymentResult.DeploymentID,
			"deployment_type": "hosted",
			"runtime":        m.Runtime,
			"endpoints":      m.Endpoints,
		}

		output, _ := json.Marshal(result)
		fmt.Println(string(output))
	} else {
		fmt.Println("\n✅ Deployment successful!")
		fmt.Printf("🌐 API URL: %s\n", deploymentResult.Endpoint)
		fmt.Printf("🆔 Deployment ID: %s\n", deploymentResult.DeploymentID)
		fmt.Printf("📊 Dashboard: https://console.api-direct.io/apis/%s\n", deploymentResult.DeploymentID)

		if len(m.Endpoints) > 0 {
			fmt.Printf("\n📍 Available endpoints:\n")
			for i, endpoint := range m.Endpoints {
				if i < 5 {
					fmt.Printf("   %s%s\n", deploymentResult.Endpoint, parseEndpointPath(endpoint))
				}
			}
			if len(m.Endpoints) > 5 {
				fmt.Printf("   ... and %d more\n", len(m.Endpoints)-5)
			}
		}

		fmt.Printf("\n🧪 Test your API:\n")
		fmt.Printf("   curl %s%s\n", deploymentResult.Endpoint, m.HealthCheck)

		fmt.Printf("\n📝 Next steps:\n")
		fmt.Printf("   View logs:  apidirect logs %s\n", apiName)
		fmt.Printf("   Update:     apidirect deploy\n")
		fmt.Printf("   Scale:      apidirect scale %s --replicas 3\n", apiName)

		if len(m.Env.Required) > 0 {
			fmt.Printf("\n⚠️  Required environment variables:\n")
			fmt.Printf("   Set these in the dashboard: %s\n", strings.Join(m.Env.Required, ", "))
		}
	}

	return nil
}

func deployBYOAV2(apiName string, m *manifest.Manifest) error {
	// Check prerequisites
	if err := checkBYOAPrerequisites(); err != nil {
		return err
	}

	// Create deployment orchestrator
	deployment, err := orchestrator.NewBYOADeployment(apiName, m)
	if err != nil {
		return fmt.Errorf("failed to initialize deployment: %w", err)
	}
	defer deployment.Cleanup()

	// Prepare deployment environment
	if outputFormat != "json" {
		fmt.Println("🔧 Preparing deployment environment...")
	}
	if err := deployment.Prepare(); err != nil {
		return fmt.Errorf("failed to prepare deployment: %w", err)
	}

	// Create deployment plan
	if err := deployment.Plan(); err != nil {
		return fmt.Errorf("deployment planning failed: %w", err)
	}

	// Ask for confirmation unless --yes flag is set
	if !yesFlag && outputFormat != "json" {
		fmt.Printf("\n⚠️  This will create AWS resources in account %s (region: %s)\n", 
			deployment.AWSAccountID, deployment.AWSRegion)
		fmt.Printf("   Estimated cost: ~$50-300/month depending on usage\n")
		fmt.Printf("\nDo you want to continue? [y/N]: ")
		
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			return fmt.Errorf("deployment cancelled")
		}
	}

	// Execute deployment
	result, err := deployment.Deploy()
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	// Output results
	if outputFormat == "json" {
		output := map[string]interface{}{
			"success":         true,
			"api_name":        apiName,
			"api_url":         result.APIURL,
			"deployment_id":   result.DeploymentID,
			"deployment_type": "byoa",
			"aws_account":     result.AWSAccountID,
			"aws_region":      result.AWSRegion,
			"runtime":         m.Runtime,
			"endpoints":       m.Endpoints,
		}

		jsonOutput, _ := json.Marshal(output)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("\n✅ BYOA Deployment successful!")
		fmt.Printf("🌐 API URL: https://%s\n", result.LoadBalancerDNS)
		fmt.Printf("🆔 Deployment ID: %s\n", result.DeploymentID)
		fmt.Printf("☁️  AWS Account: %s\n", result.AWSAccountID)
		fmt.Printf("📍 AWS Region: %s\n", result.AWSRegion)

		if len(m.Endpoints) > 0 {
			fmt.Printf("\n📍 Available endpoints:\n")
			for i, endpoint := range m.Endpoints {
				if i < 5 {
					fmt.Printf("   https://%s%s\n", result.LoadBalancerDNS, parseEndpointPath(endpoint))
				}
			}
			if len(m.Endpoints) > 5 {
				fmt.Printf("   ... and %d more\n", len(m.Endpoints)-5)
			}
		}

		fmt.Printf("\n🧪 Test your API:\n")
		fmt.Printf("   curl https://%s%s\n", result.LoadBalancerDNS, m.HealthCheck)

		fmt.Printf("\n📝 Next steps:\n")
		fmt.Printf("   1. Update DNS: Point your domain to %s\n", result.LoadBalancerDNS)
		fmt.Printf("   2. Configure SSL: Add certificate to ALB\n")
		fmt.Printf("   3. Set environment variables in AWS Systems Manager\n")
		fmt.Printf("   4. Monitor: Check CloudWatch logs and metrics\n")

		if len(m.Env.Required) > 0 {
			fmt.Printf("\n⚠️  Required environment variables:\n")
			fmt.Printf("   Set these in AWS Systems Manager Parameter Store:\n")
			for _, env := range m.Env.Required {
				fmt.Printf("   - /%s/%s/%s\n", apiName, deployment.Environment, env)
			}
		}

		fmt.Printf("\n💡 Manage your deployment:\n")
		fmt.Printf("   View status:  apidirect status %s\n", apiName)
		fmt.Printf("   View logs:    apidirect logs %s\n", apiName)
		fmt.Printf("   Update:       apidirect deploy\n")
		fmt.Printf("   Destroy:      apidirect destroy %s\n", apiName)
	}

	return nil
}

func checkBYOAPrerequisites() error {
	// Check AWS CLI
	if err := aws.CheckAWSCLI(); err != nil {
		return err
	}

	// Check AWS credentials
	if err := aws.CheckAWSCredentials(); err != nil {
		return err
	}

	// Check Terraform
	if err := terraform.CheckInstalled(); err != nil {
		return err
	}

	// Get and display AWS account info
	info, err := aws.GetCallerIdentity()
	if err != nil {
		return fmt.Errorf("failed to get AWS account info: %w", err)
	}

	if outputFormat != "json" {
		fmt.Printf("🔐 AWS Account: %s\n", info.AccountID)
		fmt.Printf("👤 AWS User: %s\n", info.Arn)
	}

	return nil
}

func createBuildContextV2(dockerfilePath string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "build-context-*.tar.gz")
	if err != nil {
		return "", err
	}
	tmpFile.Close()

	file, err := os.Create(tmpFile.Name())
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk directory and add files
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip files that shouldn't be in the build context
		if shouldSkipInBuildContext(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}
		header.Name = path

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})

	return tmpFile.Name(), err
}

func shouldSkipInBuildContext(path string) bool {
	skipPatterns := []string{
		".git",
		".gitignore", 
		"__pycache__",
		"*.pyc",
		"node_modules",
		".env",       // Never include actual env files
		".venv",
		"venv",
		".DS_Store",
		"*.log",
		".apidirect", // Skip our temp files
		"*.tar.gz",
	}

	for _, pattern := range skipPatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}

type BuildResult struct {
	ImageTag string `json:"image_tag"`
	BuildID  string `json:"build_id"`
}

func uploadAndBuild(apiName, imageTag, buildContextPath string) (*BuildResult, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add build context file
	file, err := os.Open(buildContextPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("context", "build-context.tar.gz")
	if err != nil {
		return nil, err
	}
	io.Copy(part, file)

	// Add metadata
	writer.WriteField("api_name", apiName)
	writer.WriteField("image_tag", imageTag)
	writer.Close()

	// Send request
	url := fmt.Sprintf("%s/hosted/v1/build", cfg.API.BaseURL)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)

	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("build request failed: %s", string(body))
	}

	var result BuildResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

type DeploymentResult struct {
	Endpoint     string `json:"endpoint"`
	DeploymentID string `json:"deployment_id"`
}

func deployBuiltImage(apiName, imageTag string, m *manifest.Manifest) (*DeploymentResult, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Prepare deployment request from manifest
	deployReq := map[string]interface{}{
		"api_name":    apiName,
		"image_tag":   imageTag,
		"runtime":     m.Runtime,
		"port":        m.Port,
		"start_command": m.StartCommand,
		"endpoints":   m.Endpoints,
		"health_check": m.HealthCheck,
		"environment": map[string]interface{}{
			"required": m.Env.Required,
			"optional": m.Env.Optional,
		},
	}

	// Add scaling config if provided
	if m.Scaling != nil {
		deployReq["scaling"] = map[string]interface{}{
			"min_replicas": m.Scaling.Min,
			"max_replicas": m.Scaling.Max,
			"target_cpu":   m.Scaling.TargetCPU,
		}
	} else {
		// Default scaling
		deployReq["scaling"] = map[string]interface{}{
			"min_replicas": 1,
			"max_replicas": 10,
			"target_cpu":   70,
		}
	}

	// Add resource limits if provided
	if m.Resources != nil {
		deployReq["resources"] = map[string]interface{}{
			"memory": m.Resources.Memory,
			"cpu":    m.Resources.CPU,
		}
	} else {
		// Default resources
		deployReq["resources"] = map[string]interface{}{
			"memory": "512Mi",
			"cpu":    "250m",
		}
	}

	body, err := json.Marshal(deployReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/hosted/v1/deploy", cfg.API.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("deployment request failed: %s", string(body))
	}

	var result DeploymentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func checkExistingDeployment(apiName string) (*DeploymentResult, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/hosted/v1/deployments/%s", cfg.API.BaseURL, apiName)
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

	if resp.StatusCode == 404 {
		return nil, nil // No existing deployment
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check existing deployment")
	}

	var result DeploymentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func waitForDeploymentReady(deploymentID string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Poll for up to 5 minutes
	maxAttempts := 60
	for i := 0; i < maxAttempts; i++ {
		status, err := getDeploymentStatusV2(cfg, deploymentID)
		if err != nil {
			return err
		}

		switch status {
		case "ready", "running":
			return nil
		case "failed":
			return fmt.Errorf("deployment failed")
		case "building", "deploying":
			if outputFormat != "json" {
				fmt.Print(".")
			}
			time.Sleep(5 * time.Second)
		default:
			time.Sleep(5 * time.Second)
		}
	}

	return fmt.Errorf("deployment timed out")
}

func getDeploymentStatusV2(cfg *config.Config, deploymentID string) (string, error) {
	url := fmt.Sprintf("%s/hosted/v1/deployments/%s/status", cfg.API.BaseURL, deploymentID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get deployment status")
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Status, nil
}

func parseEndpointPath(endpoint string) string {
	// Extract path from endpoint string like "GET /users"
	parts := strings.Fields(endpoint)
	if len(parts) >= 2 {
		return parts[1]
	}
	return "/"
}

func runDemoDeploymentV2(cmd *cobra.Command, args []string) error {
	// Load manifest for demo
	manifestPath, err := manifest.FindManifest(".")
	if err != nil {
		return fmt.Errorf("no manifest found. Run 'apidirect import' to create one")
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	apiName := m.Name
	if len(args) > 0 {
		apiName = args[0]
	}

	fmt.Printf("🚀 Deploying '%s' (Demo Mode)\n", apiName)
	fmt.Printf("📋 Configuration: %s runtime, port %d\n", m.Runtime, m.Port)

	// Simulate deployment steps
	steps := []struct {
		message  string
		duration time.Duration
	}{
		{"📦 Packaging application...", 2 * time.Second},
		{"🐳 Building container image...", 3 * time.Second},
		{"⬆️  Uploading to API-Direct registry...", 2 * time.Second},
		{"🔧 Configuring auto-scaling groups...", 2 * time.Second},
		{"🔒 Setting up SSL certificates...", 1 * time.Second},
		{"⚡ Starting application instances...", 3 * time.Second},
	}

	for _, step := range steps {
		fmt.Println(step.message)
		time.Sleep(step.duration)
	}

	// Generate demo results
	endpoint := fmt.Sprintf("https://%s-%s.api-direct.io", apiName, generateRandomID())
	deploymentID := fmt.Sprintf("dep_%s", generateRandomID())

	fmt.Println("\n✅ Deployment successful!")
	fmt.Printf("🌐 API URL: %s\n", endpoint)
	fmt.Printf("🆔 Deployment ID: %s\n", deploymentID)
	fmt.Printf("📊 Dashboard: https://console.api-direct.io/apis/%s\n", deploymentID)

	if len(m.Endpoints) > 0 {
		fmt.Printf("\n📍 Available endpoints:\n")
		for i, ep := range m.Endpoints {
			if i < 5 {
				fmt.Printf("   %s%s\n", endpoint, parseEndpointPath(ep))
			}
		}
	}

	return nil
}

func generateRandomID() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 8)
	for i := range id {
		id[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(id)
}