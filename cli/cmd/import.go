package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/api-direct/cli/pkg/detector"
	"github.com/api-direct/cli/pkg/manifest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	autoMode bool
	outputFile string
	skipConfirmation bool
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import [path]",
	Short: "Import an existing API project",
	Long: `Import an existing API project by analyzing its structure and generating
an apidirect.yaml manifest file. This command will scan your project, detect
the framework and configuration, and create a deployment manifest.

The import process will:
1. Scan your project structure
2. Detect runtime and framework
3. Find entry points and ports
4. Generate a manifest file
5. Ask for confirmation before saving`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
	
	importCmd.Flags().BoolVar(&autoMode, "auto", false, "Run in automatic mode with defaults")
	importCmd.Flags().StringVarP(&outputFile, "output", "o", "apidirect.yaml", "Output manifest file name")
	importCmd.Flags().BoolVarP(&skipConfirmation, "yes", "y", false, "Skip confirmation prompt")
}

func runImport(cmd *cobra.Command, args []string) error {
	// Determine project path
	projectPath := "."
	if len(args) > 0 {
		projectPath = args[0]
	}
	
	// Validate project path exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", projectPath)
	}
	
	fmt.Printf("🔍 Scanning project structure in %s...\n", projectPath)
	
	// Run detection
	detection, err := detector.AnalyzeProject(projectPath)
	if err != nil {
		return fmt.Errorf("failed to analyze project: %w", err)
	}
	
	// Display detection results
	displayDetectionResults(detection)
	
	// Generate manifest
	manifest, err := generateManifest(detection, projectPath)
	if err != nil {
		return fmt.Errorf("failed to generate manifest: %w", err)
	}
	
	// Show manifest and get confirmation
	if !skipConfirmation {
		confirmed, err := showManifestAndConfirm(manifest)
		if err != nil {
			return err
		}
		
		if !confirmed {
			return handleManifestRejection(manifest, projectPath)
		}
	}
	
	// Save manifest
	manifestPath := filepath.Join(projectPath, outputFile)
	if err := saveManifest(manifest, manifestPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}
	
	fmt.Printf("\n✅ Manifest saved to %s\n", manifestPath)
	fmt.Println("🚀 Ready to deploy! Run: apidirect deploy")
	fmt.Println("💡 Tip: Run 'apidirect validate' to check your manifest")
	
	return nil
}

func displayDetectionResults(d *detector.ProjectDetection) {
	if d.Language != "" {
		fmt.Printf("📦 Detected: %s project\n", d.Language)
	}
	if d.Framework != "" {
		fmt.Printf("🚀 Found: %s framework\n", d.Framework)
	}
	if d.RequirementsFile != "" {
		fmt.Printf("📄 Located: %s\n", d.RequirementsFile)
	}
	if len(d.Endpoints) > 0 {
		fmt.Printf("🔧 Discovered: %d API endpoints\n", len(d.Endpoints))
	}
}

func generateManifest(d *detector.ProjectDetection, projectPath string) (*manifest.Manifest, error) {
	// Get project name from directory
	projectName := filepath.Base(projectPath)
	if projectName == "." {
		cwd, _ := os.Getwd()
		projectName = filepath.Base(cwd)
	}
	
	m := &manifest.Manifest{
		Name:    strings.ToLower(strings.ReplaceAll(projectName, " ", "-")),
		Runtime: d.Runtime,
		StartCommand: d.StartCommand,
		Port:    d.Port,
		Files: manifest.FileRefs{
			Main:         d.MainFile,
			Requirements: d.RequirementsFile,
			EnvExample:   d.EnvFile,
		},
		Endpoints:    convertEndpoints(d.Endpoints),
		Env:          manifest.EnvironmentVars{
			Required: d.Environment.Required,
			Optional: d.Environment.Optional,
		},
		HealthCheck:  d.HealthCheck,
	}
	
	// Add helpful comments
	m.Comments = map[string]string{
		"start_command": "How to start your server (PLEASE VERIFY!)",
		"port":          "The port your application listens on",
		"health_check":  "Endpoint for health monitoring",
	}
	
	return m, nil
}

func convertEndpoints(detected []detector.Endpoint) []string {
	endpoints := make([]string, len(detected))
	for i, ep := range detected {
		endpoints[i] = fmt.Sprintf("%s %s", ep.Method, ep.Path)
	}
	return endpoints
}

func showManifestAndConfirm(m *manifest.Manifest) (bool, error) {
	fmt.Println("\n✅ Generated apidirect.yaml based on analysis:")
	fmt.Println(strings.Repeat("━", 50))
	
	// Display the manifest in YAML format with comments
	manifestYAML, err := generateYAMLWithComments(m)
	if err != nil {
		return false, err
	}
	
	fmt.Println(manifestYAML)
	fmt.Println(strings.Repeat("━", 50))
	
	// Ask for confirmation
	fmt.Print("\n📝 Does this look correct? [Y/n/e]: ")
	
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	switch response {
	case "", "y", "yes":
		return true, nil
	case "e", "edit":
		return false, fmt.Errorf("edit requested")
	default:
		return false, nil
	}
}

func generateYAMLWithComments(m *manifest.Manifest) (string, error) {
	var sb strings.Builder
	
	// Add header comments
	sb.WriteString(fmt.Sprintf("# apidirect.yaml\n"))
	sb.WriteString(fmt.Sprintf("# Auto-generated on %s\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString("# PLEASE REVIEW: These are our best guesses!\n\n")
	
	// Name and runtime
	sb.WriteString(fmt.Sprintf("name: %s\n", m.Name))
	sb.WriteString(fmt.Sprintf("runtime: %s\n\n", m.Runtime))
	
	// Start command with comment
	sb.WriteString("# How to start your server (PLEASE VERIFY!)\n")
	sb.WriteString(fmt.Sprintf("start_command: \"%s\"\n\n", m.StartCommand))
	
	// Port
	sb.WriteString("# Where your server listens\n")
	sb.WriteString(fmt.Sprintf("port: %d\n\n", m.Port))
	
	// Files
	sb.WriteString("# Your application files\n")
	sb.WriteString("files:\n")
	if m.Files.Main != "" {
		sb.WriteString(fmt.Sprintf("  main: %s\n", m.Files.Main))
	}
	if m.Files.Requirements != "" {
		sb.WriteString(fmt.Sprintf("  requirements: %s\n", m.Files.Requirements))
	}
	if m.Files.EnvExample != "" {
		sb.WriteString(fmt.Sprintf("  env_example: %s\n", m.Files.EnvExample))
	}
	sb.WriteString("\n")
	
	// Endpoints
	if len(m.Endpoints) > 0 {
		sb.WriteString("# Detected endpoints\n")
		sb.WriteString("endpoints:\n")
		maxShow := 5
		for i, ep := range m.Endpoints {
			if i < maxShow {
				sb.WriteString(fmt.Sprintf("  - %s\n", ep))
			}
		}
		if len(m.Endpoints) > maxShow {
			sb.WriteString(fmt.Sprintf("  # ... (showing first %d of %d detected)\n", maxShow, len(m.Endpoints)))
		}
		sb.WriteString("\n")
	}
	
	// Environment variables
	if m.Env.Required != nil || m.Env.Optional != nil {
		sb.WriteString("# Environment variables\n")
		sb.WriteString("env:\n")
		if len(m.Env.Required) > 0 {
			sb.WriteString(fmt.Sprintf("  required: %v\n", m.Env.Required))
		}
		if len(m.Env.Optional) > 0 {
			sb.WriteString("  optional:\n")
			for k, v := range m.Env.Optional {
				sb.WriteString(fmt.Sprintf("    %s: %s\n", k, v))
			}
		}
		sb.WriteString("\n")
	}
	
	// Health check
	if m.HealthCheck != "" {
		sb.WriteString("# Health check for monitoring\n")
		sb.WriteString(fmt.Sprintf("health_check: %s\n", m.HealthCheck))
	}
	
	return sb.String(), nil
}

func handleManifestRejection(m *manifest.Manifest, projectPath string) error {
	fmt.Println("\n📝 No problem! Let me help you fix it.")
	fmt.Println("\nWhat would you like to do?")
	fmt.Println("  1) Edit in your default editor")
	fmt.Println("  2) Fix specific field")
	fmt.Println("  3) Start over with different settings")
	fmt.Println("  4) View documentation")
	fmt.Print("\nChoose [1-4]: ")
	
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	switch choice {
	case "1":
		return editInDefaultEditor(m, projectPath)
	case "2":
		return quickEditMode(m, projectPath)
	case "3":
		fmt.Println("\n🔄 Starting over...")
		return fmt.Errorf("restart requested")
	case "4":
		fmt.Println("\n📚 Opening documentation: https://docs.api-direct.io/manifest")
		return fmt.Errorf("documentation requested")
	default:
		return fmt.Errorf("invalid choice")
	}
}

func editInDefaultEditor(m *manifest.Manifest, projectPath string) error {
	// Save manifest to temporary file
	tempFile := filepath.Join(projectPath, outputFile+".tmp")
	if err := saveManifest(m, tempFile); err != nil {
		return err
	}
	defer os.Remove(tempFile)
	
	// Determine editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // fallback
		if _, err := exec.LookPath("code"); err == nil {
			editor = "code"
		}
	}
	
	fmt.Printf("\n🔧 Opening %s in %s...\n", outputFile, editor)
	cmd := exec.Command(editor, tempFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	
	// Move temp file to final location
	finalPath := filepath.Join(projectPath, outputFile)
	if err := os.Rename(tempFile, finalPath); err != nil {
		return fmt.Errorf("failed to save edited file: %w", err)
	}
	
	fmt.Printf("\n✅ Manifest saved to %s\n", finalPath)
	fmt.Println("💡 Tip: Run 'apidirect validate' to check syntax")
	
	return nil
}

func quickEditMode(m *manifest.Manifest, projectPath string) error {
	fmt.Println("\n🔧 Quick edit mode. What needs fixing?")
	fmt.Printf("  1) start_command (currently: %s)\n", m.StartCommand)
	fmt.Printf("  2) port (currently: %d)\n", m.Port)
	fmt.Printf("  3) main file (currently: %s)\n", m.Files.Main)
	fmt.Println("  4) Other field...")
	fmt.Print("\nChoose field to edit [1-4]: ")
	
	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)
	
	switch choice {
	case "1":
		fmt.Print("\nEnter new start_command: ")
		newCmd, _ := reader.ReadString('\n')
		oldCmd := m.StartCommand
		m.StartCommand = strings.TrimSpace(newCmd)
		fmt.Printf("\n✅ Updated! Here's the change:\n")
		fmt.Printf("- start_command: \"%s\"\n", oldCmd)
		fmt.Printf("+ start_command: \"%s\"\n", m.StartCommand)
		
	case "2":
		fmt.Print("\nEnter new port: ")
		var newPort int
		fmt.Scanln(&newPort)
		oldPort := m.Port
		m.Port = newPort
		fmt.Printf("\n✅ Updated! Here's the change:\n")
		fmt.Printf("- port: %d\n", oldPort)
		fmt.Printf("+ port: %d\n", m.Port)
		
	case "3":
		fmt.Print("\nEnter new main file path: ")
		newMain, _ := reader.ReadString('\n')
		oldMain := m.Files.Main
		m.Files.Main = strings.TrimSpace(newMain)
		fmt.Printf("\n✅ Updated! Here's the change:\n")
		fmt.Printf("- main: %s\n", oldMain)
		fmt.Printf("+ main: %s\n", m.Files.Main)
		
	default:
		return fmt.Errorf("invalid choice")
	}
	
	// Ask if more edits needed
	fmt.Print("\nAnything else to fix? [y/N]: ")
	more, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(more)) == "y" {
		return quickEditMode(m, projectPath)
	}
	
	// Save the manifest
	manifestPath := filepath.Join(projectPath, outputFile)
	if err := saveManifest(m, manifestPath); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}
	
	fmt.Printf("\n✅ Manifest saved! Run 'apidirect validate' to verify.\n")
	return nil
}

func saveManifest(m *manifest.Manifest, path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, data, 0644)
}