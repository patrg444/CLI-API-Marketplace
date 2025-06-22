package manifest

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest represents the apidirect.yaml configuration
type Manifest struct {
	Name         string           `yaml:"name"`
	Runtime      string           `yaml:"runtime"`
	StartCommand string           `yaml:"start_command"`
	Port         int              `yaml:"port"`
	Files        FileRefs         `yaml:"files,omitempty"`
	Endpoints    []string         `yaml:"endpoints,omitempty"`
	Env          EnvironmentVars  `yaml:"env,omitempty"`
	HealthCheck  string           `yaml:"health_check,omitempty"`
	Scaling      *ScalingConfig   `yaml:"scaling,omitempty"`
	Resources    *ResourceLimits  `yaml:"resources,omitempty"`
	
	// Not serialized to YAML
	Comments     map[string]string `yaml:"-"`
}

// FileRefs specifies important file locations
type FileRefs struct {
	Main         string `yaml:"main,omitempty"`
	Requirements string `yaml:"requirements,omitempty"`
	EnvExample   string `yaml:"env_example,omitempty"`
	Dockerfile   string `yaml:"dockerfile,omitempty"`
}

// EnvironmentVars holds environment variable configuration
type EnvironmentVars struct {
	Required []string          `yaml:"required,omitempty"`
	Optional map[string]string `yaml:"optional,omitempty"`
}

// ScalingConfig defines auto-scaling parameters
type ScalingConfig struct {
	Min       int `yaml:"min,omitempty"`
	Max       int `yaml:"max,omitempty"`
	TargetCPU int `yaml:"target_cpu,omitempty"`
}

// ResourceLimits defines resource constraints
type ResourceLimits struct {
	Memory string `yaml:"memory,omitempty"`
	CPU    string `yaml:"cpu,omitempty"`
}

// Load reads and parses a manifest file
func Load(path string) (*Manifest, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	
	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	
	// Validate after loading
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	
	return &manifest, nil
}

// Save writes the manifest to a file
func (m *Manifest) Save(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	return ioutil.WriteFile(path, data, 0644)
}

// Validate checks if the manifest has all required fields and valid values
func (m *Manifest) Validate() error {
	var errors []string
	
	// Required fields
	if m.Name == "" {
		errors = append(errors, "missing 'name' field")
	} else if !isValidName(m.Name) {
		errors = append(errors, "invalid name: must contain only lowercase letters, numbers, and hyphens")
	}
	
	if m.Runtime == "" {
		errors = append(errors, "missing 'runtime' field")
	} else if !isValidRuntime(m.Runtime) {
		errors = append(errors, fmt.Sprintf("unsupported runtime: %s", m.Runtime))
	}
	
	if m.StartCommand == "" {
		errors = append(errors, "missing 'start_command' field")
	}
	
	if m.Port == 0 {
		errors = append(errors, "missing 'port' field")
	} else if m.Port < 1 || m.Port > 65535 {
		errors = append(errors, fmt.Sprintf("invalid port: %d (must be 1-65535)", m.Port))
	}
	
	// Validate file references
	if err := m.validateFiles(); err != nil {
		errors = append(errors, err.Error())
	}
	
	// Validate endpoints format
	for _, endpoint := range m.Endpoints {
		if !isValidEndpoint(endpoint) {
			errors = append(errors, fmt.Sprintf("invalid endpoint format: %s", endpoint))
		}
	}
	
	// Validate resource limits
	if m.Resources != nil {
		if err := validateResourceString(m.Resources.Memory, "memory"); err != nil {
			errors = append(errors, err.Error())
		}
		if err := validateResourceString(m.Resources.CPU, "cpu"); err != nil {
			errors = append(errors, err.Error())
		}
	}
	
	// Validate scaling config
	if m.Scaling != nil {
		if m.Scaling.Min < 0 {
			errors = append(errors, "scaling.min must be >= 0")
		}
		if m.Scaling.Max < m.Scaling.Min {
			errors = append(errors, "scaling.max must be >= scaling.min")
		}
		if m.Scaling.TargetCPU < 0 || m.Scaling.TargetCPU > 100 {
			errors = append(errors, "scaling.target_cpu must be between 0 and 100")
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}
	
	return nil
}

// validateFiles checks if referenced files exist
func (m *Manifest) validateFiles() error {
	// Note: This should be called from the project directory
	var missing []string
	
	if m.Files.Main != "" && !fileExists(m.Files.Main) {
		missing = append(missing, fmt.Sprintf("main file not found: %s", m.Files.Main))
	}
	
	if m.Files.Requirements != "" && !fileExists(m.Files.Requirements) {
		missing = append(missing, fmt.Sprintf("requirements file not found: %s", m.Files.Requirements))
	}
	
	if m.Files.Dockerfile != "" && !fileExists(m.Files.Dockerfile) {
		missing = append(missing, fmt.Sprintf("Dockerfile not found: %s", m.Files.Dockerfile))
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing files: %s", strings.Join(missing, ", "))
	}
	
	return nil
}

// FindManifest looks for a manifest file in the given directory
func FindManifest(dir string) (string, error) {
	candidates := []string{
		"apidirect.yaml",
		"apidirect.yml",
		".apidirect.yaml",
		".apidirect.yml",
	}
	
	for _, candidate := range candidates {
		path := filepath.Join(dir, candidate)
		if fileExists(path) {
			return path, nil
		}
	}
	
	return "", fmt.Errorf("no manifest file found in %s", dir)
}

// GenerateDockerfile creates a Dockerfile based on the manifest
func (m *Manifest) GenerateDockerfile() string {
	var dockerfile strings.Builder
	
	// Base image based on runtime
	baseImage := getBaseImage(m.Runtime)
	dockerfile.WriteString(fmt.Sprintf("FROM %s\n\n", baseImage))
	
	dockerfile.WriteString("WORKDIR /app\n\n")
	
	// Copy and install dependencies first (for better caching)
	switch {
	case strings.HasPrefix(m.Runtime, "python"):
		if m.Files.Requirements != "" {
			dockerfile.WriteString(fmt.Sprintf("COPY %s .\n", m.Files.Requirements))
			dockerfile.WriteString("RUN pip install --no-cache-dir -r requirements.txt\n\n")
		}
	case strings.HasPrefix(m.Runtime, "node"):
		dockerfile.WriteString("COPY package*.json ./\n")
		dockerfile.WriteString("RUN npm ci --only=production\n\n")
	case strings.HasPrefix(m.Runtime, "go"):
		dockerfile.WriteString("COPY go.* ./\n")
		dockerfile.WriteString("RUN go mod download\n\n")
	}
	
	// Copy application code
	dockerfile.WriteString("COPY . .\n\n")
	
	// Create non-root user
	dockerfile.WriteString("RUN useradd -m -u 1001 apiuser && chown -R apiuser:apiuser /app\n")
	dockerfile.WriteString("USER apiuser\n\n")
	
	// Expose port
	dockerfile.WriteString(fmt.Sprintf("EXPOSE %d\n\n", m.Port))
	
	// Set environment variables
	if len(m.Env.Optional) > 0 {
		for key, value := range m.Env.Optional {
			dockerfile.WriteString(fmt.Sprintf("ENV %s=%s\n", key, value))
		}
		dockerfile.WriteString("\n")
	}
	
	// Health check
	if m.HealthCheck != "" {
		dockerfile.WriteString(fmt.Sprintf("HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \\\n"))
		dockerfile.WriteString(fmt.Sprintf("  CMD curl -f http://localhost:%d%s || exit 1\n\n", m.Port, m.HealthCheck))
	}
	
	// Start command
	dockerfile.WriteString(fmt.Sprintf("CMD %s\n", shellToDockerCmd(m.StartCommand)))
	
	return dockerfile.String()
}

// Helper functions

func isValidName(name string) bool {
	// Must be lowercase letters, numbers, and hyphens only
	// Must start with a letter
	// Must not end with a hyphen
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9-]*[a-z0-9]$`, name)
	return matched && len(name) <= 63
}

func isValidRuntime(runtime string) bool {
	validRuntimes := []string{
		"python3.9", "python3.10", "python3.11", "python3.12",
		"node16", "node18", "node20",
		"go1.19", "go1.20", "go1.21",
		"ruby3.0", "ruby3.1", "ruby3.2",
		"java11", "java17", "java21",
		"dotnet6", "dotnet7", "dotnet8",
		"php8.0", "php8.1", "php8.2",
		"docker", // Custom Dockerfile
	}
	
	for _, valid := range validRuntimes {
		if runtime == valid {
			return true
		}
	}
	return false
}

func isValidEndpoint(endpoint string) bool {
	// Should be in format "METHOD /path"
	parts := strings.Fields(endpoint)
	if len(parts) != 2 {
		return false
	}
	
	method := parts[0]
	path := parts[1]
	
	// Validate method
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	methodValid := false
	for _, m := range validMethods {
		if method == m {
			methodValid = true
			break
		}
	}
	if !methodValid {
		return false
	}
	
	// Validate path starts with /
	return strings.HasPrefix(path, "/")
}

func validateResourceString(resource, name string) error {
	if resource == "" {
		return nil
	}
	
	// Check format like "512Mi", "1Gi", "100m", "0.5"
	matched, _ := regexp.MatchString(`^\d+(\.\d+)?[A-Za-z]*$`, resource)
	if !matched {
		return fmt.Errorf("invalid %s format: %s", name, resource)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func getBaseImage(runtime string) string {
	switch {
	case strings.HasPrefix(runtime, "python"):
		version := strings.TrimPrefix(runtime, "python")
		return fmt.Sprintf("python:%s-slim", version)
	case strings.HasPrefix(runtime, "node"):
		version := strings.TrimPrefix(runtime, "node")
		return fmt.Sprintf("node:%s-alpine", version)
	case strings.HasPrefix(runtime, "go"):
		version := strings.TrimPrefix(runtime, "go")
		return fmt.Sprintf("golang:%s-alpine", version)
	case strings.HasPrefix(runtime, "ruby"):
		version := strings.TrimPrefix(runtime, "ruby")
		return fmt.Sprintf("ruby:%s-slim", version)
	case strings.HasPrefix(runtime, "java"):
		version := strings.TrimPrefix(runtime, "java")
		return fmt.Sprintf("openjdk:%s-slim", version)
	case strings.HasPrefix(runtime, "dotnet"):
		version := strings.TrimPrefix(runtime, "dotnet")
		return fmt.Sprintf("mcr.microsoft.com/dotnet/aspnet:%s", version)
	case strings.HasPrefix(runtime, "php"):
		version := strings.TrimPrefix(runtime, "php")
		return fmt.Sprintf("php:%s-apache", version)
	default:
		return "ubuntu:22.04"
	}
}

func shellToDockerCmd(command string) string {
	// Convert shell command to Dockerfile CMD format
	// This is a simple implementation - could be enhanced
	if strings.Contains(command, " ") {
		// Split and format as JSON array
		parts := strings.Fields(command)
		quotedParts := make([]string, len(parts))
		for i, part := range parts {
			quotedParts[i] = fmt.Sprintf(`"%s"`, part)
		}
		return fmt.Sprintf("[%s]", strings.Join(quotedParts, ", "))
	}
	return fmt.Sprintf(`["%s"]`, command)
}