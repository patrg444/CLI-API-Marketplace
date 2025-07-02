package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/api-direct/cli/pkg/manifest"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	runPort      int
	runEnvFile   string
	watchMode    bool
	buildOnly    bool
	dockerMode   bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run your API locally for development",
	Long: `Run your API locally using the configuration from apidirect.yaml.
This command will:
- Load your manifest configuration
- Set up environment variables
- Start your API server locally
- Provide live reload in watch mode

Perfect for testing before deployment.`,
	RunE: runLocal,
}

func init() {
	rootCmd.AddCommand(runCmd)
	
	runCmd.Flags().IntVarP(&runPort, "port", "p", 0, "Override the port (default: from manifest)")
	runCmd.Flags().StringVarP(&runEnvFile, "env-file", "e", ".env", "Environment file to load")
	runCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "Enable auto-reload on file changes")
	runCmd.Flags().BoolVar(&buildOnly, "build-only", false, "Build but don't run (Docker mode)")
	runCmd.Flags().BoolVar(&dockerMode, "docker", false, "Run using Docker")
}

func runLocal(cmd *cobra.Command, args []string) error {
	// Load manifest
	manifestPath, err := manifest.FindManifest(".")
	if err != nil {
		return fmt.Errorf("no manifest found. Run 'apidirect import' first")
	}

	m, err := manifest.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Override port if specified
	if runPort == 0 {
		runPort = m.Port
	}

	// Display startup info
	fmt.Println(color.CyanString("🚀 Starting %s locally", m.Name))
	fmt.Printf("📋 Runtime: %s\n", m.Runtime)
	fmt.Printf("🔌 Port: %d\n", runPort)
	
	// Load environment variables
	if err := loadEnvironment(runEnvFile, m); err != nil {
		printWarning(fmt.Sprintf("Could not load environment file: %v", err))
	}

	// Run based on mode
	if dockerMode {
		return runInDocker(m)
	}

	// Check dependencies
	if err := checkDependencies(m); err != nil {
		return err
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run the server
	if watchMode {
		fmt.Println("👁  Watch mode enabled - will restart on file changes")
		return runWithWatch(ctx, m, sigChan)
	}

	return runDirect(ctx, m, sigChan)
}

func loadEnvironment(envFile string, m *manifest.Manifest) error {
	// First, set optional environment variables from manifest
	for key, value := range m.Env.Optional {
		os.Setenv(key, value)
	}

	// Then load from .env file if it exists
	if _, err := os.Stat(envFile); err == nil {
		file, err := os.Open(envFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := strings.TrimSpace(scanner.Text())
			
			// Skip comments and empty lines
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse KEY=VALUE
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				// Remove quotes if present
				value = strings.Trim(value, `"'`)
				
				os.Setenv(key, value)
				if verbose {
					fmt.Printf("   Set %s from %s\n", key, envFile)
				}
			}
		}
		
		fmt.Printf("✅ Loaded environment from %s\n", envFile)
	}

	// Check for required variables
	var missing []string
	for _, required := range m.Env.Required {
		if os.Getenv(required) == "" {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		printWarning(fmt.Sprintf("Missing required environment variables: %s", strings.Join(missing, ", ")))
		fmt.Println("💡 Tip: Create a .env file or set them in your shell")
	}

	// Override PORT for local development
	os.Setenv("PORT", strconv.Itoa(runPort))

	return nil
}

func checkDependencies(m *manifest.Manifest) error {
	fmt.Print("🔍 Checking dependencies... ")

	switch {
	case strings.HasPrefix(m.Runtime, "python"):
		// Check Python
		if _, err := exec.LookPath("python3"); err != nil {
			if _, err := exec.LookPath("python"); err != nil {
				return fmt.Errorf("Python not found. Please install Python %s", m.Runtime)
			}
		}
		
		// Check if dependencies are installed
		if m.Files.Requirements != "" {
			// Simple check - see if common packages are importable
			checkCmd := exec.Command("python3", "-c", "import sys")
			if err := checkCmd.Run(); err != nil {
				fmt.Println("\n⚠️  Dependencies might not be installed")
				fmt.Printf("💡 Run: pip install -r %s\n", m.Files.Requirements)
			}
		}

	case strings.HasPrefix(m.Runtime, "node"):
		// Check Node.js
		if _, err := exec.LookPath("node"); err != nil {
			return fmt.Errorf("Node.js not found. Please install Node.js %s", m.Runtime)
		}

		// Check if node_modules exists
		if _, err := os.Stat("node_modules"); os.IsNotExist(err) {
			fmt.Println("\n⚠️  Dependencies not installed")
			fmt.Println("💡 Run: npm install")
			return fmt.Errorf("missing dependencies")
		}

	case strings.HasPrefix(m.Runtime, "go"):
		// Check Go
		if _, err := exec.LookPath("go"); err != nil {
			return fmt.Errorf("Go not found. Please install Go %s", m.Runtime)
		}

	case m.Runtime == "docker":
		// Check Docker
		if _, err := exec.LookPath("docker"); err != nil {
			return fmt.Errorf("Docker not found. Please install Docker")
		}
	}

	fmt.Println("✓")
	return nil
}

func runDirect(ctx context.Context, m *manifest.Manifest, sigChan chan os.Signal) error {
	// Parse and prepare the start command
	cmdParts := parseCommand(m.StartCommand)
	if len(cmdParts) == 0 {
		return fmt.Errorf("invalid start command")
	}

	// Create command
	cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set working directory if main file is in a subdirectory
	if m.Files.Main != "" && strings.Contains(m.Files.Main, "/") {
		cmd.Dir = filepath.Dir(m.Files.Main)
	}

	// Start the server
	fmt.Printf("\n▶️  Running: %s\n", m.StartCommand)
	fmt.Println(strings.Repeat("─", 50))

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Show access info
	fmt.Printf("\n🌐 Local API running at:\n")
	fmt.Printf("   • http://localhost:%d\n", runPort)
	if m.HealthCheck != "" {
		fmt.Printf("   • http://localhost:%d%s (health check)\n", runPort, m.HealthCheck)
	}
	fmt.Printf("\n📝 Press Ctrl+C to stop\n\n")

	// Wait for shutdown signal or process exit
	go func() {
		<-sigChan
		fmt.Println("\n\n🛑 Shutting down...")
		cmd.Process.Signal(os.Interrupt)
	}()

	// Wait for process to exit
	err := cmd.Wait()
	if err != nil && !strings.Contains(err.Error(), "interrupt") {
		return fmt.Errorf("server exited with error: %w", err)
	}

	fmt.Println("✅ Server stopped")
	return nil
}

func runWithWatch(ctx context.Context, m *manifest.Manifest, sigChan chan os.Signal) error {
	// This is a simplified watch mode
	// In production, you'd use a proper file watcher like fsnotify
	
	restartChan := make(chan bool)
	var cmd *exec.Cmd
	
	// Start server
	startServer := func() {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Signal(os.Interrupt)
			cmd.Wait()
		}

		cmdParts := parseCommand(m.StartCommand)
		cmd = exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		
		if err := cmd.Start(); err != nil {
			printError(fmt.Sprintf("Failed to start: %v", err))
			return
		}

		fmt.Println("🔄 Server restarted")
	}

	// Initial start
	startServer()

	// Simple file watcher (checks every 2 seconds)
	go func() {
		lastModTime := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				// Check if any relevant files changed
				changed := false
				err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return nil
					}
					
					// Skip non-relevant files
					if shouldSkipForWatch(path) {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}

					// Check modification time
					if info.ModTime().After(lastModTime) {
						changed = true
						return filepath.SkipDir
					}
					
					return nil
				})

				if err == nil && changed {
					lastModTime = time.Now()
					fmt.Println("\n📝 File change detected")
					restartChan <- true
				}
			}
		}
	}()

	// Handle restarts and shutdown
	for {
		select {
		case <-sigChan:
			if cmd != nil && cmd.Process != nil {
				cmd.Process.Signal(os.Interrupt)
			}
			return nil
		case <-restartChan:
			startServer()
		}
	}
}

func runInDocker(m *manifest.Manifest) error {
	fmt.Println("🐳 Running in Docker mode")

	// Generate Dockerfile if needed
	dockerfilePath := m.Files.Dockerfile
	if dockerfilePath == "" {
		dockerfilePath = "Dockerfile.dev"
		dockerfileContent := m.GenerateDockerfile()
		if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
			return fmt.Errorf("failed to write Dockerfile: %w", err)
		}
		defer os.Remove(dockerfilePath)
		fmt.Println("📄 Generated development Dockerfile")
	}

	// Build image
	imageName := fmt.Sprintf("%s:dev", m.Name)
	fmt.Printf("🔨 Building image: %s\n", imageName)
	
	buildCmd := exec.Command("docker", "build", "-t", imageName, "-f", dockerfilePath, ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	if buildOnly {
		fmt.Println("✅ Docker image built successfully")
		return nil
	}

	// Run container
	fmt.Printf("🚀 Starting container on port %d\n", runPort)
	
	runArgs := []string{
		"run",
		"--rm",
		"-it",
		"-p", fmt.Sprintf("%d:%d", runPort, m.Port),
		"--name", fmt.Sprintf("%s-dev", m.Name),
	}

	// Add environment variables
	if _, err := os.Stat(runEnvFile); err == nil {
		runArgs = append(runArgs, "--env-file", runEnvFile)
	}

	// Add volume mount for development
	cwd, _ := os.Getwd()
	runArgs = append(runArgs, "-v", fmt.Sprintf("%s:/app", cwd))

	runArgs = append(runArgs, imageName)

	runCmd := exec.Command("docker", runArgs...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin

	fmt.Printf("\n🌐 API running at: http://localhost:%d\n", runPort)
	fmt.Println("📝 Press Ctrl+C to stop")

	return runCmd.Run()
}

func parseCommand(command string) []string {
	// Simple command parsing - handles quotes
	var parts []string
	var current string
	inQuotes := false
	
	for _, char := range command {
		switch char {
		case '"', '\'':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes && current != "" {
				parts = append(parts, current)
				current = ""
			} else {
				current += string(char)
			}
		default:
			current += string(char)
		}
	}
	
	if current != "" {
		parts = append(parts, current)
	}
	
	return parts
}

func shouldSkipForWatch(path string) bool {
	// Skip directories and files that shouldn't trigger restart
	skipPatterns := []string{
		".git", ".idea", ".vscode", "__pycache__", 
		"node_modules", ".env", "*.log", "*.pyc",
		".DS_Store", "venv", ".venv",
	}
	
	for _, pattern := range skipPatterns {
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
		if strings.Contains(path, pattern) {
			return true
		}
	}
	
	// Only watch source files
	ext := filepath.Ext(path)
	watchExts := []string{".py", ".js", ".go", ".rb", ".java", ".ts", ".jsx", ".tsx"}
	for _, watchExt := range watchExts {
		if ext == watchExt {
			return false
		}
	}
	
	return true
}

func tryOpenBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	if err != nil {
		printWarning("Could not open browser automatically")
	}
}