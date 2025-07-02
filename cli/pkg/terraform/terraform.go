package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Client provides an interface for Terraform operations
type Client struct {
	workDir string
	vars    map[string]interface{}
}

// NewClient creates a new Terraform client
func NewClient(workDir string) *Client {
	return &Client{
		workDir: workDir,
		vars:    make(map[string]interface{}),
	}
}

// SetVar sets a Terraform variable
func (c *Client) SetVar(key string, value interface{}) {
	c.vars[key] = value
}

// SetVars sets multiple Terraform variables
func (c *Client) SetVars(vars map[string]interface{}) {
	for k, v := range vars {
		c.vars[k] = v
	}
}

// Init initializes Terraform in the working directory
func (c *Client) Init() error {
	cmd := exec.Command("terraform", "init", "-backend=true", "-input=false")
	cmd.Dir = c.workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %w\n%s", err, string(output))
	}
	
	return nil
}

// Plan creates a Terraform plan
func (c *Client) Plan(planFile string) error {
	args := []string{"plan", "-input=false", "-out=" + planFile}
	
	// Add variables
	for k, v := range c.vars {
		args = append(args, fmt.Sprintf("-var=%s=%v", k, v))
	}
	
	cmd := exec.Command("terraform", args...)
	cmd.Dir = c.workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform plan failed: %w\n%s", err, string(output))
	}
	
	return nil
}

// Apply applies a Terraform plan
func (c *Client) Apply(planFile string) error {
	cmd := exec.Command("terraform", "apply", "-input=false", "-auto-approve", planFile)
	cmd.Dir = c.workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w\n%s", err, string(output))
	}
	
	return nil
}

// Destroy destroys Terraform-managed infrastructure
func (c *Client) Destroy() error {
	args := []string{"destroy", "-input=false", "-auto-approve"}
	
	// Add variables for destroy
	for k, v := range c.vars {
		args = append(args, fmt.Sprintf("-var=%s=%v", k, v))
	}
	
	cmd := exec.Command("terraform", args...)
	cmd.Dir = c.workDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w\n%s", err, string(output))
	}
	
	return nil
}

// Output retrieves Terraform outputs
func (c *Client) Output() (map[string]interface{}, error) {
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = c.workDir
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("terraform output failed: %w", err)
	}
	
	var outputs map[string]interface{}
	if err := json.Unmarshal(output, &outputs); err != nil {
		return nil, fmt.Errorf("failed to parse terraform output: %w", err)
	}
	
	// Extract values from output format
	result := make(map[string]interface{})
	for k, v := range outputs {
		if m, ok := v.(map[string]interface{}); ok {
			if val, exists := m["value"]; exists {
				result[k] = val
			}
		}
	}
	
	return result, nil
}

// CheckInstalled checks if Terraform is installed
func CheckInstalled() error {
	cmd := exec.Command("terraform", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform not found. Please install terraform: https://www.terraform.io/downloads")
	}
	return nil
}

// StreamingApply applies Terraform with streaming output
func (c *Client) StreamingApply(planFile string, outputWriter io.Writer) error {
	cmd := exec.Command("terraform", "apply", "-input=false", "-auto-approve", planFile)
	cmd.Dir = c.workDir
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}
	
	return nil
}

// StreamingPlan creates a plan with streaming output
func (c *Client) StreamingPlan(planFile string, outputWriter io.Writer) error {
	args := []string{"plan", "-input=false", "-out=" + planFile}
	
	// Add variables
	for k, v := range c.vars {
		args = append(args, fmt.Sprintf("-var=%s=%v", k, v))
	}
	
	cmd := exec.Command("terraform", args...)
	cmd.Dir = c.workDir
	cmd.Stdout = outputWriter
	cmd.Stderr = outputWriter
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}
	
	return nil
}

// CopyModules copies Terraform modules to a working directory
func CopyModules(sourceDir, destDir string) error {
	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Walk through source directory and copy files
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Calculate relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		
		// Calculate destination path
		destPath := filepath.Join(destDir, relPath)
		
		if info.IsDir() {
			// Skip .terraform directories
			if filepath.Base(path) == ".terraform" {
				return filepath.SkipDir
			}
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		}
		
		// Copy file
		return copyFile(path, destPath)
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// WriteVarsFile writes variables to a tfvars file
func WriteVarsFile(filename string, vars map[string]interface{}) error {
	var buffer bytes.Buffer
	
	for k, v := range vars {
		switch val := v.(type) {
		case string:
			fmt.Fprintf(&buffer, "%s = %q\n", k, val)
		case []string:
			fmt.Fprintf(&buffer, "%s = [", k)
			for i, s := range val {
				if i > 0 {
					buffer.WriteString(", ")
				}
				fmt.Fprintf(&buffer, "%q", s)
			}
			buffer.WriteString("]\n")
		case map[string]string:
			fmt.Fprintf(&buffer, "%s = {\n", k)
			for mk, mv := range val {
				fmt.Fprintf(&buffer, "  %s = %q\n", mk, mv)
			}
			buffer.WriteString("}\n")
		default:
			fmt.Fprintf(&buffer, "%s = %v\n", k, val)
		}
	}
	
	return os.WriteFile(filename, buffer.Bytes(), 0644)
}