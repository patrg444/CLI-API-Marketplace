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

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"github.com/api-direct/cli/pkg/config"
)

var (
	deployVersion string
	deployReplicas int
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [api-name]",
	Short: "Deploy your API to the API-Direct platform",
	Long: `Deploy your API to the API-Direct platform. This command packages your code,
uploads it to the platform, and creates a live endpoint for your API.

If no API name is provided, it will use the name from apidirect.yaml in the current directory.`,
	RunE: runDeploy,
}

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().StringVar(&deployVersion, "version", "", "Version label for this deployment")
	deployCmd.Flags().IntVar(&deployReplicas, "replicas", 1, "Number of replicas to deploy")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// Check authentication
	if !config.IsAuthenticated() {
		return fmt.Errorf("not authenticated. Please run 'apidirect login' first")
	}

	// Load project configuration
	projectConfig, err := loadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	// Determine API name
	apiName := projectConfig.Name
	if len(args) > 0 {
		apiName = args[0]
	}

	fmt.Printf("🚀 Deploying API: %s\n", apiName)

	// Validate project
	if err := validateProject(projectConfig); err != nil {
		return fmt.Errorf("project validation failed: %w", err)
	}

	// Package code
	fmt.Println("📦 Packaging code...")
	packagePath, err := packageCode(apiName)
	if err != nil {
		return fmt.Errorf("failed to package code: %w", err)
	}
	defer os.Remove(packagePath)

	// Upload code
	fmt.Println("⬆️  Uploading code...")
	version, err := uploadCode(apiName, packagePath, projectConfig.Runtime)
	if err != nil {
		return fmt.Errorf("failed to upload code: %w", err)
	}

	// Deploy to platform
	fmt.Println("🔧 Deploying to platform...")
	endpoint, err := deployToplatform(apiName, version, projectConfig)
	if err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}

	// Wait for deployment to be ready
	fmt.Println("⏳ Waiting for deployment to be ready...")
	if err := waitForDeployment(apiName); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	fmt.Println("\n✅ Deployment successful!")
	fmt.Printf("🌐 Your API is available at: %s\n", endpoint)
	fmt.Printf("\nTest your API:\n")
	fmt.Printf("  curl %s/hello\n", endpoint)
	fmt.Printf("\nView logs:\n")
	fmt.Printf("  apidirect logs %s\n", apiName)

	return nil
}

func loadProjectConfig() (*config.ProjectConfig, error) {
	data, err := ioutil.ReadFile("apidirect.yaml")
	if err != nil {
		return nil, fmt.Errorf("apidirect.yaml not found. Run 'apidirect init' to create a project")
	}

	var projectConfig config.ProjectConfig
	if err := yaml.Unmarshal(data, &projectConfig); err != nil {
		return nil, fmt.Errorf("invalid apidirect.yaml: %w", err)
	}

	return &projectConfig, nil
}

func validateProject(projectConfig *config.ProjectConfig) error {
	// Check if main file exists
	var mainFile string
	switch {
	case strings.HasPrefix(projectConfig.Runtime, "python"):
		mainFile = "main.py"
	case strings.HasPrefix(projectConfig.Runtime, "node"):
		mainFile = "main.js"
	default:
		return fmt.Errorf("unsupported runtime: %s", projectConfig.Runtime)
	}

	if _, err := os.Stat(mainFile); err != nil {
		return fmt.Errorf("main file %s not found", mainFile)
	}

	// Validate endpoints
	if len(projectConfig.Endpoints) == 0 {
		return fmt.Errorf("no endpoints defined in apidirect.yaml")
	}

	return nil
}

func packageCode(apiName string) (string, error) {
	// Create temporary file for package
	tmpFile, err := ioutil.TempFile("", fmt.Sprintf("%s-*.tar.gz", apiName))
	if err != nil {
		return "", err
	}
	tmpFile.Close()

	// Create gzip writer
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

		// Skip certain files/directories
		if shouldSkipFile(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, path)
		if err != nil {
			return err
		}
		header.Name = path

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
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

	if err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func shouldSkipFile(path string) bool {
	// Skip common directories and files that shouldn't be deployed
	skipPatterns := []string{
		".git",
		".gitignore",
		"__pycache__",
		"*.pyc",
		"node_modules",
		".env",
		".venv",
		"venv",
		".DS_Store",
		"*.log",
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

func uploadCode(apiName, packagePath, runtime string) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", err
	}

	// Prepare multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	file, err := os.Open(packagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("code", filepath.Base(packagePath))
	if err != nil {
		return "", err
	}
	io.Copy(part, file)

	// Add runtime
	writer.WriteField("runtime", runtime)

	// Close writer
	writer.Close()

	// Create request
	url := fmt.Sprintf("%s/storage/api/v1/upload/%s", cfg.API.BaseURL, apiName)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)

	// Send request
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", string(body))
	}

	// Parse response
	var result struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Version, nil
}

func deployToplatform(apiName, version string, projectConfig *config.ProjectConfig) (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", err
	}

	// Prepare deployment request
	deployReq := map[string]interface{}{
		"api_id":   apiName,
		"version":  version,
		"runtime":  projectConfig.Runtime,
		"code_url": fmt.Sprintf("s3://code-storage/%s/%s", apiName, version),
		"environment": projectConfig.Environment,
		"replicas": deployReplicas,
		"resources": map[string]string{
			"cpu_request":    "100m",
			"cpu_limit":      "500m",
			"memory_request": "128Mi",
			"memory_limit":   "512Mi",
		},
	}

	body, err := json.Marshal(deployReq)
	if err != nil {
		return "", err
	}

	// Create request
	url := fmt.Sprintf("%s/deployment/api/v1/deploy/%s", cfg.API.BaseURL, apiName)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Auth.AccessToken)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("deployment failed: %s", string(body))
	}

	// Parse response
	var result struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Endpoint, nil
}

func waitForDeployment(apiName string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	// Poll deployment status
	for i := 0; i < 60; i++ { // Max 5 minutes
		status, err := getDeploymentStatus(cfg, apiName)
		if err != nil {
			return err
		}

		switch status {
		case "running":
			return nil
		case "failed":
			return fmt.Errorf("deployment failed")
		default:
			fmt.Print(".")
			time.Sleep(5 * time.Second)
		}
	}

	return fmt.Errorf("deployment timeout")
}

func getDeploymentStatus(cfg *config.Config, apiName string) (string, error) {
	url := fmt.Sprintf("%s/deployment/api/v1/status/%s", cfg.API.BaseURL, apiName)
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
		return "", fmt.Errorf("failed to get status")
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Status, nil
}
