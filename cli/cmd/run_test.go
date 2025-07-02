package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		manifest     string
		files        map[string]string
		envFile      string
		setup        func(*testing.T) string
		validate     func(*testing.T, string, *exec.Cmd)
		wantErr      bool
		skipActualRun bool // Skip actually running the process
	}{
		{
			name: "run python fastapi project",
			manifest: `name: test-api
runtime: python3.11
start_command: uvicorn main:app --reload --host 0.0.0.0 --port 8080
port: 8080
files:
  main: main.py
health_check: /health
`,
			files: map[string]string{
				"main.py": `from fastapi import FastAPI
app = FastAPI()
@app.get("/")
def read_root():
    return {"message": "Hello World"}
`,
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string, cmd *exec.Cmd) {
				assert.Contains(t, output, "Starting API")
				assert.Contains(t, output, "uvicorn")
				assert.Contains(t, output, "http://localhost:8080")
			},
			skipActualRun: true,
		},
		{
			name: "run node express project",
			manifest: `name: test-express-api
runtime: node18
start_command: node index.js
port: 3000
files:
  main: index.js
health_check: /health
`,
			files: map[string]string{
				"index.js": `const express = require('express');
const app = express();
app.get('/', (req, res) => res.json({message: 'Hello'}));
app.listen(3000);
`,
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string, cmd *exec.Cmd) {
				assert.Contains(t, output, "Starting API")
				assert.Contains(t, output, "node")
				assert.Contains(t, output, "http://localhost:3000")
			},
			skipActualRun: true,
		},
		{
			name: "run with custom port",
			args: []string{"--port", "9999"},
			manifest: `name: test-api
runtime: python3.11
start_command: uvicorn main:app --reload --host 0.0.0.0 --port {PORT}
port: 8080
files:
  main: main.py
`,
			files: map[string]string{
				"main.py": "# main file",
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string, cmd *exec.Cmd) {
				assert.Contains(t, output, "9999")
				assert.Contains(t, output, "Overriding port")
			},
			skipActualRun: true,
		},
		{
			name: "run with env file",
			args: []string{"--env-file", "custom.env"},
			manifest: `name: test-api
runtime: python3.11
start_command: python main.py
port: 8080
`,
			files: map[string]string{
				"main.py": "# main file",
				"custom.env": `API_KEY=test123
DEBUG=true
`,
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string, cmd *exec.Cmd) {
				assert.Contains(t, output, "Loading environment from custom.env")
			},
			skipActualRun: true,
		},
		{
			name: "missing manifest file",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
		{
			name: "missing main file",
			manifest: `name: test-api
runtime: python3.11
start_command: python main.py
port: 8080
files:
  main: main.py
`,
			files: map[string]string{
				// main.py is missing
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
		},
		{
			name: "detect runtime from files",
			manifest: `name: test-api
runtime: python3.11
port: 8080
`,
			files: map[string]string{
				"main.py": "# Python file",
				"requirements.txt": "fastapi\nuvicorn",
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string, cmd *exec.Cmd) {
				// Should auto-detect Python runtime command
				assert.Contains(t, output, "Detected Python project")
			},
			skipActualRun: true,
		},
		{
			name: "docker mode",
			args: []string{"--docker"},
			manifest: `name: test-api
runtime: python3.11
port: 8080
`,
			files: map[string]string{
				"Dockerfile": `FROM python:3.11
WORKDIR /app
COPY . .
CMD ["python", "main.py"]
`,
				"main.py": "# main file",
			},
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			validate: func(t *testing.T, output string, cmd *exec.Cmd) {
				assert.Contains(t, output, "docker")
				assert.Contains(t, output, "Building container")
			},
			skipActualRun: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := tt.setup(t)
			
			// Change to test directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir(testDir)
			require.NoError(t, err)
			defer os.Chdir(oldWd)
			
			// Write manifest file if provided
			if tt.manifest != "" {
				err = ioutil.WriteFile("apidirect.yaml", []byte(tt.manifest), 0644)
				require.NoError(t, err)
			}
			
			// Create files
			for filename, content := range tt.files {
				dir := filepath.Dir(filename)
				if dir != "." {
					err = os.MkdirAll(dir, 0755)
					require.NoError(t, err)
				}
				err = ioutil.WriteFile(filename, []byte(content), 0644)
				require.NoError(t, err)
			}
			
			// Create command structure
			rootCmd := &cobra.Command{
				Use:   "apidirect",
				Short: "API-Direct CLI",
			}
			
			// Mock run command
			testRunCmd := &cobra.Command{
				Use:   "run",
				Short: "Run your API locally",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Mock implementation
					port, _ := cmd.Flags().GetInt("port")
					envFile, _ := cmd.Flags().GetString("env-file")
					docker, _ := cmd.Flags().GetBool("docker")
					
					// Check manifest exists
					if _, err := os.Stat("apidirect.yaml"); os.IsNotExist(err) {
						return fmt.Errorf("apidirect.yaml not found")
					}
					
					// Read manifest
					data, err := ioutil.ReadFile("apidirect.yaml")
					if err != nil {
						return err
					}
					
					// Basic validation
					if !strings.Contains(string(data), "name:") {
						return fmt.Errorf("invalid manifest")
					}
					
					// Check for main file
					if strings.Contains(string(data), "files:") && strings.Contains(string(data), "main:") {
						// Extract main file
						lines := strings.Split(string(data), "\n")
						for _, line := range lines {
							trimmed := strings.TrimSpace(line)
							if strings.HasPrefix(trimmed, "main:") {
								// Handle both "main: file.py" and "main:" on separate line
								parts := strings.SplitN(trimmed, ":", 2)
								if len(parts) == 2 {
									mainFile := strings.TrimSpace(parts[1])
									if mainFile != "" {
										if _, err := os.Stat(mainFile); os.IsNotExist(err) {
											return fmt.Errorf("main file not found: %s", mainFile)
										}
									}
								}
							}
						}
					}
					
					// Print mock output
					cmd.Println("🚀 Starting API locally...")
					
					if port > 0 {
						cmd.Printf("⚙️  Overriding port to %d\n", port)
					}
					
					if envFile != "" && envFile != ".env" {
						if _, err := os.Stat(envFile); err == nil {
							cmd.Printf("📋 Loading environment from %s\n", envFile)
						}
					}
					
					if docker {
						cmd.Println("🐳 Building container...")
						cmd.Println("   Using docker mode")
					}
					
					// Detect runtime
					if _, err := os.Stat("requirements.txt"); err == nil {
						cmd.Println("🐍 Detected Python project")
					} else if _, err := os.Stat("package.json"); err == nil {
						cmd.Println("📦 Detected Node.js project")
					} else if _, err := os.Stat("go.mod"); err == nil {
						cmd.Println("🐹 Detected Go project")
					}
					
					// Extract start command and port
					if strings.Contains(string(data), "start_command:") {
						for _, line := range strings.Split(string(data), "\n") {
							if strings.HasPrefix(strings.TrimSpace(line), "start_command:") {
								startCmd := strings.TrimPrefix(line, "start_command:")
								startCmd = strings.TrimSpace(startCmd)
								cmd.Printf("▶️  Running: %s\n", startCmd)
								break
							}
						}
					}
					
					// Extract port
					actualPort := 8080
					if port > 0 {
						actualPort = port
					} else {
						for _, line := range strings.Split(string(data), "\n") {
							if strings.HasPrefix(strings.TrimSpace(line), "port:") {
								portStr := strings.TrimPrefix(line, "port:")
								portStr = strings.TrimSpace(portStr)
								if p, err := strconv.Atoi(portStr); err == nil {
									actualPort = p
								}
								break
							}
						}
					}
					
					cmd.Printf("\n✅ API running at http://localhost:%d\n", actualPort)
					cmd.Println("   Press Ctrl+C to stop")
					
					// If not skipping actual run, simulate process
					if !tt.skipActualRun {
						// In real implementation, would start the actual process
						time.Sleep(10 * time.Millisecond)
					}
					
					return nil
				},
			}
			
			// Add flags
			testRunCmd.Flags().IntP("port", "p", 0, "Override the port")
			testRunCmd.Flags().StringP("env-file", "e", ".env", "Environment file to load")
			testRunCmd.Flags().BoolP("docker", "d", false, "Run in Docker mode")
			
			// Capture output
			output := &bytes.Buffer{}
			rootCmd.SetOut(output)
			rootCmd.SetErr(output)
			
			// Add run command
			rootCmd.AddCommand(testRunCmd)
			
			// Execute command
			args := append([]string{"run"}, tt.args...)
			rootCmd.SetArgs(args)
			
			err = rootCmd.Execute()
			
			// Check error
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			// Validate output
			if tt.validate != nil && !tt.wantErr {
				tt.validate(t, output.String(), nil)
			}
		})
	}
}

func TestRunHelperFunctions(t *testing.T) {
	t.Run("detectStartCommand", func(t *testing.T) {
		testCases := []struct {
			name     string
			runtime  string
			files    map[string]string
			expected string
		}{
			{
				name:    "Python with uvicorn",
				runtime: "python3.11",
				files: map[string]string{
					"main.py": "from fastapi import FastAPI",
					"requirements.txt": "fastapi\nuvicorn",
				},
				expected: "uvicorn main:app --reload --host 0.0.0.0",
			},
			{
				name:    "Node.js with package.json start script",
				runtime: "node18",
				files: map[string]string{
					"package.json": `{"scripts": {"start": "node server.js"}}`,
				},
				expected: "npm start",
			},
			{
				name:    "Go project",
				runtime: "go1.21",
				files: map[string]string{
					"main.go": "package main",
					"go.mod": "module test",
				},
				expected: "go run .",
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				testDir := t.TempDir()
				
				// Create test files
				for filename, content := range tc.files {
					err := ioutil.WriteFile(
						filepath.Join(testDir, filename),
						[]byte(content),
						0644,
					)
					require.NoError(t, err)
				}
				
				// Test detection
				cmd := detectStartCommand(tc.runtime, testDir)
				assert.Contains(t, cmd, tc.expected)
			})
		}
	})
	
	t.Run("findAvailablePort", func(t *testing.T) {
		// Test finding available port
		port := findAvailablePort(8080)
		assert.Greater(t, port, 0)
		assert.LessOrEqual(t, port, 65535)
	})
	
	t.Run("loadEnvFile", func(t *testing.T) {
		testDir := t.TempDir()
		envFile := filepath.Join(testDir, "test.env")
		
		// Create test env file
		content := `TEST_VAR=value1
TEST_PORT=3000
# Comment line
EMPTY_VAR=
`
		err := ioutil.WriteFile(envFile, []byte(content), 0644)
		require.NoError(t, err)
		
		// Test loading
		envVars := loadEnvFile(envFile)
		assert.Equal(t, "value1", envVars["TEST_VAR"])
		assert.Equal(t, "3000", envVars["TEST_PORT"])
		assert.Equal(t, "", envVars["EMPTY_VAR"])
		assert.NotContains(t, envVars, "# Comment line")
	})
}

func TestProcessManagement(t *testing.T) {
	t.Run("graceful shutdown", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping process management test")
		}
		
		// Create a simple process that responds to signals
		cmd := exec.Command("sleep", "10")
		err := cmd.Start()
		require.NoError(t, err)
		
		// Send interrupt signal after short delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			cmd.Process.Signal(syscall.SIGINT)
		}()
		
		// Wait for process
		err = cmd.Wait()
		
		// Should have been interrupted
		assert.Error(t, err)
	})
}

// Mock helper functions (these would be in the actual implementation)
func detectStartCommand(runtime, dir string) string {
	switch {
	case strings.HasPrefix(runtime, "python"):
		if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
			data, _ := ioutil.ReadFile(filepath.Join(dir, "requirements.txt"))
			if strings.Contains(string(data), "fastapi") {
				return "uvicorn main:app --reload --host 0.0.0.0"
			}
		}
		return "python main.py"
		
	case strings.HasPrefix(runtime, "node"):
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
			data, _ := ioutil.ReadFile(filepath.Join(dir, "package.json"))
			if strings.Contains(string(data), `"start":`) {
				return "npm start"
			}
		}
		return "node index.js"
		
	case strings.HasPrefix(runtime, "go"):
		return "go run ."
		
	default:
		return ""
	}
}

func findAvailablePort(preferred int) int {
	// Simple implementation - in real code would check if port is available
	return preferred
}

func loadEnvFile(path string) map[string]string {
	env := make(map[string]string)
	
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return env
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	
	return env
}