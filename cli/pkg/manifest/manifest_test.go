package manifest

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManifestLoad(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		errContains string
		validate    func(*testing.T, *Manifest)
	}{
		{
			name: "valid manifest",
			content: `name: test-api
runtime: python3.11
start_command: uvicorn main:app --host 0.0.0.0 --port 8080
port: 8080
files:
  main: main.py
  requirements: requirements.txt
health_check: /health
endpoints:
  - GET /
  - POST /users
env:
  required:
    - DATABASE_URL
  optional:
    DEBUG: "false"
`,
			validate: func(t *testing.T, m *Manifest) {
				assert.Equal(t, "test-api", m.Name)
				assert.Equal(t, "python3.11", m.Runtime)
				assert.Equal(t, 8080, m.Port)
				assert.Equal(t, "uvicorn main:app --host 0.0.0.0 --port 8080", m.StartCommand)
				assert.Equal(t, "/health", m.HealthCheck)
				assert.Equal(t, "main.py", m.Files.Main)
				assert.Equal(t, "requirements.txt", m.Files.Requirements)
				assert.Len(t, m.Endpoints, 2)
				assert.Contains(t, m.Env.Required, "DATABASE_URL")
				assert.Equal(t, "false", m.Env.Optional["DEBUG"])
			},
		},
		{
			name: "manifest with scaling",
			content: `name: scaled-api
runtime: node18
start_command: node server.js
port: 3000
scaling:
  min: 2
  max: 10
  target_cpu: 70
`,
			validate: func(t *testing.T, m *Manifest) {
				require.NotNil(t, m.Scaling)
				assert.Equal(t, 2, m.Scaling.Min)
				assert.Equal(t, 10, m.Scaling.Max)
				assert.Equal(t, 70, m.Scaling.TargetCPU)
			},
		},
		{
			name: "manifest with resources",
			content: `name: resource-api
runtime: go1.21
start_command: ./main
port: 8080
resources:
  memory: 512Mi
  cpu: 500m
`,
			validate: func(t *testing.T, m *Manifest) {
				require.NotNil(t, m.Resources)
				assert.Equal(t, "512Mi", m.Resources.Memory)
				assert.Equal(t, "500m", m.Resources.CPU)
			},
		},
		{
			name: "invalid yaml",
			content: `name: test-api
  runtime: python3.11
 port: invalid
`,
			wantErr:     true,
			errContains: "failed to parse manifest",
		},
		{
			name: "missing required fields",
			content: `name: test-api
port: 8080
`,
			wantErr:     true,
			errContains: "missing 'runtime' field",
		},
		{
			name: "invalid runtime",
			content: `name: test-api
runtime: cobol85
start_command: run program
port: 8080
`,
			wantErr:     true,
			errContains: "unsupported runtime: cobol85",
		},
		{
			name: "invalid port",
			content: `name: test-api
runtime: python3.11
start_command: python app.py
port: 99999
`,
			wantErr:     true,
			errContains: "invalid port: 99999",
		},
		{
			name: "invalid name format",
			content: `name: Test_API
runtime: python3.11
start_command: python app.py
port: 8080
`,
			wantErr:     true,
			errContains: "invalid name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpdir, err := ioutil.TempDir("", "manifest-test-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpdir)

			// Change to temp directory for file validation
			oldWd, _ := os.Getwd()
			os.Chdir(tmpdir)
			defer os.Chdir(oldWd)

			// Create manifest file
			manifestPath := filepath.Join(tmpdir, "apidirect.yaml")
			err = ioutil.WriteFile(manifestPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Create any referenced files if this is a valid test case
			if !tt.wantErr && strings.Contains(tt.content, "main:") {
				// Parse files from content (simple approach for testing)
				if strings.Contains(tt.content, "main: main.py") {
					ioutil.WriteFile("main.py", []byte("# test"), 0644)
				}
				if strings.Contains(tt.content, "requirements: requirements.txt") {
					ioutil.WriteFile("requirements.txt", []byte("fastapi"), 0644)
				}
			}

			// Load manifest
			manifest, err := Load("apidirect.yaml")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				if tt.validate != nil && manifest != nil {
					tt.validate(t, manifest)
				}
			}
		})
	}
}

func TestManifestSave(t *testing.T) {
	manifest := &Manifest{
		Name:         "test-api",
		Runtime:      "python3.11",
		StartCommand: "python app.py",
		Port:         8080,
		Files: FileRefs{
			Main:         "app.py",
			Requirements: "requirements.txt",
		},
		Endpoints: []string{"GET /", "POST /users"},
		Env: EnvironmentVars{
			Required: []string{"API_KEY"},
			Optional: map[string]string{
				"DEBUG": "false",
			},
		},
		HealthCheck: "/health",
	}

	// Create temp directory
	tmpdir, err := ioutil.TempDir("", "manifest-save-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpdir)
	defer os.Chdir(oldWd)

	// Create referenced files
	ioutil.WriteFile("app.py", []byte("# test"), 0644)
	ioutil.WriteFile("requirements.txt", []byte("fastapi"), 0644)

	// Save manifest
	err = manifest.Save("apidirect.yaml")
	assert.NoError(t, err)

	// Load it back
	loaded, err := Load("apidirect.yaml")
	assert.NoError(t, err)

	// Compare
	assert.Equal(t, manifest.Name, loaded.Name)
	assert.Equal(t, manifest.Runtime, loaded.Runtime)
	assert.Equal(t, manifest.Port, loaded.Port)
	assert.Equal(t, manifest.Files.Main, loaded.Files.Main)
	assert.Equal(t, manifest.Endpoints, loaded.Endpoints)
	assert.Equal(t, manifest.Env.Required, loaded.Env.Required)
	assert.Equal(t, manifest.Env.Optional["DEBUG"], loaded.Env.Optional["DEBUG"])
}

func TestManifestValidate(t *testing.T) {
	tests := []struct {
		name        string
		manifest    *Manifest
		setupFiles  map[string]string
		wantErr     bool
		errContains []string
	}{
		{
			name: "valid manifest",
			manifest: &Manifest{
				Name:         "test-api",
				Runtime:      "python3.11",
				StartCommand: "python app.py",
				Port:         8080,
			},
			wantErr: false,
		},
		{
			name: "multiple validation errors",
			manifest: &Manifest{
				Name:    "",
				Runtime: "invalid",
				Port:    99999,
			},
			wantErr: true,
			errContains: []string{
				"missing 'name' field",
				"unsupported runtime",
				"invalid port",
				"missing 'start_command' field",
			},
		},
		{
			name: "invalid endpoints",
			manifest: &Manifest{
				Name:         "test-api",
				Runtime:      "python3.11",
				StartCommand: "python app.py",
				Port:         8080,
				Endpoints: []string{
					"GET /valid",
					"INVALID",
					"POST",
					"DELETE users", // missing leading /
				},
			},
			wantErr: true,
			errContains: []string{
				"invalid endpoint format: INVALID",
				"invalid endpoint format: POST",
				"invalid endpoint format: DELETE users",
			},
		},
		{
			name: "invalid scaling config",
			manifest: &Manifest{
				Name:         "test-api",
				Runtime:      "python3.11",
				StartCommand: "python app.py",
				Port:         8080,
				Scaling: &ScalingConfig{
					Min:       -1,
					Max:       -2,
					TargetCPU: 150,
				},
			},
			wantErr: true,
			errContains: []string{
				"scaling.min must be >= 0",
				"scaling.max must be >= scaling.min",
				"scaling.target_cpu must be between 0 and 100",
			},
		},
		{
			name: "invalid resource limits",
			manifest: &Manifest{
				Name:         "test-api",
				Runtime:      "python3.11",
				StartCommand: "python app.py",
				Port:         8080,
				Resources: &ResourceLimits{
					Memory: "invalid-memory",
					CPU:    "not-a-cpu-value",
				},
			},
			wantErr: true,
			errContains: []string{
				"invalid memory format",
				"invalid cpu format",
			},
		},
		{
			name: "missing files",
			manifest: &Manifest{
				Name:         "test-api",
				Runtime:      "python3.11",
				StartCommand: "python app.py",
				Port:         8080,
				Files: FileRefs{
					Main:         "app.py",
					Requirements: "requirements.txt",
					Dockerfile:   "Dockerfile",
				},
			},
			wantErr:     true,
			errContains: []string{"missing files"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for file tests
			if tt.setupFiles != nil || tt.manifest.Files.Main != "" {
				tmpdir, err := ioutil.TempDir("", "manifest-test-*")
				require.NoError(t, err)
				defer os.RemoveAll(tmpdir)

				oldWd, _ := os.Getwd()
				os.Chdir(tmpdir)
				defer os.Chdir(oldWd)

				// Create setup files
				for path, content := range tt.setupFiles {
					dir := filepath.Dir(path)
					if dir != "." {
						os.MkdirAll(dir, 0755)
					}
					ioutil.WriteFile(path, []byte(content), 0644)
				}
			}

			err := tt.manifest.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				for _, contains := range tt.errContains {
					assert.Contains(t, err.Error(), contains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindManifest(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		expected  string
		wantErr   bool
	}{
		{
			name:     "find apidirect.yaml",
			files:    []string{"apidirect.yaml"},
			expected: "apidirect.yaml",
		},
		{
			name:     "find apidirect.yml",
			files:    []string{"apidirect.yml"},
			expected: "apidirect.yml",
		},
		{
			name:     "find hidden .apidirect.yaml",
			files:    []string{".apidirect.yaml"},
			expected: ".apidirect.yaml",
		},
		{
			name:     "prefer non-hidden over hidden",
			files:    []string{".apidirect.yaml", "apidirect.yaml"},
			expected: "apidirect.yaml",
		},
		{
			name:    "no manifest found",
			files:   []string{"other.yaml", "config.yml"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpdir, err := ioutil.TempDir("", "manifest-find-*")
			require.NoError(t, err)
			defer os.RemoveAll(tmpdir)

			// Create files
			for _, file := range tt.files {
				path := filepath.Join(tmpdir, file)
				ioutil.WriteFile(path, []byte("name: test"), 0644)
			}

			// Find manifest
			found, err := FindManifest(tmpdir)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, filepath.Join(tmpdir, tt.expected), found)
			}
		})
	}
}

func TestGenerateDockerfile(t *testing.T) {
	tests := []struct {
		name     string
		manifest *Manifest
		contains []string
	}{
		{
			name: "Python Dockerfile",
			manifest: &Manifest{
				Name:         "python-api",
				Runtime:      "python3.11",
				StartCommand: "uvicorn main:app --host 0.0.0.0 --port 8080",
				Port:         8080,
				Files: FileRefs{
					Requirements: "requirements.txt",
				},
				HealthCheck: "/health",
				Env: EnvironmentVars{
					Optional: map[string]string{
						"DEBUG": "false",
					},
				},
			},
			contains: []string{
				"FROM python:3.11-slim",
				"COPY requirements.txt",
				"RUN pip install --no-cache-dir -r requirements.txt",
				"EXPOSE 8080",
				"ENV DEBUG=false",
				"HEALTHCHECK",
				"CMD [\"uvicorn\", \"main:app\", \"--host\", \"0.0.0.0\", \"--port\", \"8080\"]",
				"USER apiuser",
			},
		},
		{
			name: "Node.js Dockerfile",
			manifest: &Manifest{
				Name:         "node-api",
				Runtime:      "node18",
				StartCommand: "node server.js",
				Port:         3000,
			},
			contains: []string{
				"FROM node:18-alpine",
				"COPY package*.json",
				"RUN npm ci --only=production",
				"EXPOSE 3000",
				"CMD [\"node\", \"server.js\"]",
			},
		},
		{
			name: "Go Dockerfile",
			manifest: &Manifest{
				Name:         "go-api",
				Runtime:      "go1.21",
				StartCommand: "./main",
				Port:         8080,
			},
			contains: []string{
				"FROM golang:1.21-alpine",
				"COPY go.* ./",
				"RUN go mod download",
				"CMD [\"./main\"]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerfile := tt.manifest.GenerateDockerfile()

			for _, expected := range tt.contains {
				assert.Contains(t, dockerfile, expected)
			}

			// Common checks
			assert.Contains(t, dockerfile, "WORKDIR /app")
			assert.Contains(t, dockerfile, "COPY . .")
			assert.Contains(t, dockerfile, "RUN useradd -m -u 1001 apiuser")
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("isValidName", func(t *testing.T) {
		testCases := []struct {
			name  string
			valid bool
		}{
			{"test-api", true},
			{"my-awesome-api-123", true},
			{"a", false}, // too short, needs at least 2 chars
			{"api", true},
			{"Test-API", false},       // uppercase
			{"test_api", false},       // underscore
			{"test-api-", false},      // ends with hyphen
			{"-test-api", false},      // starts with hyphen
			{"test.api", false},       // contains dot
			{"1test-api", false},      // starts with number
			{strings.Repeat("a", 64), false}, // too long
		}

		for _, tc := range testCases {
			result := isValidName(tc.name)
			assert.Equal(t, tc.valid, result, "name: %s", tc.name)
		}
	})

	t.Run("isValidRuntime", func(t *testing.T) {
		testCases := []struct {
			runtime string
			valid   bool
		}{
			{"python3.11", true},
			{"node18", true},
			{"go1.21", true},
			{"ruby3.2", true},
			{"java17", true},
			{"dotnet8", true},
			{"php8.2", true},
			{"docker", true},
			{"python2.7", false},
			{"node14", false},
			{"rust", false},
			{"", false},
		}

		for _, tc := range testCases {
			result := isValidRuntime(tc.runtime)
			assert.Equal(t, tc.valid, result, "runtime: %s", tc.runtime)
		}
	})

	t.Run("isValidEndpoint", func(t *testing.T) {
		testCases := []struct {
			endpoint string
			valid    bool
		}{
			{"GET /", true},
			{"POST /users", true},
			{"PUT /users/123", true},
			{"DELETE /users/{id}", true},
			{"PATCH /items", true},
			{"HEAD /status", true},
			{"OPTIONS /api", true},
			{"GET", false},           // missing path
			{"/users", false},        // missing method
			{"INVALID /path", false}, // invalid method
			{"GET users", false},     // path doesn't start with /
			{"", false},
		}

		for _, tc := range testCases {
			result := isValidEndpoint(tc.endpoint)
			assert.Equal(t, tc.valid, result, "endpoint: %s", tc.endpoint)
		}
	})

	t.Run("validateResourceString", func(t *testing.T) {
		testCases := []struct {
			resource string
			valid    bool
		}{
			{"512Mi", true},
			{"1Gi", true},
			{"2G", true},
			{"100m", true},
			{"0.5", true},
			{"1.5Gi", true},
			{"", true}, // empty is valid
			{"invalid", false},
			{"1 Gi", false}, // space
			{"-100m", false}, // negative
		}

		for _, tc := range testCases {
			err := validateResourceString(tc.resource, "test")
			if tc.valid {
				assert.NoError(t, err, "resource: %s", tc.resource)
			} else {
				assert.Error(t, err, "resource: %s", tc.resource)
			}
		}
	})

	t.Run("getBaseImage", func(t *testing.T) {
		testCases := []struct {
			runtime  string
			expected string
		}{
			{"python3.11", "python:3.11-slim"},
			{"node18", "node:18-alpine"},
			{"go1.21", "golang:1.21-alpine"},
			{"ruby3.2", "ruby:3.2-slim"},
			{"java17", "openjdk:17-slim"},
			{"dotnet8", "mcr.microsoft.com/dotnet/aspnet:8"},
			{"php8.2", "php:8.2-apache"},
			{"unknown", "ubuntu:22.04"},
		}

		for _, tc := range testCases {
			result := getBaseImage(tc.runtime)
			assert.Equal(t, tc.expected, result, "runtime: %s", tc.runtime)
		}
	})

	t.Run("shellToDockerCmd", func(t *testing.T) {
		testCases := []struct {
			command  string
			expected string
		}{
			{"python app.py", `["python", "app.py"]`},
			{"node server.js", `["node", "server.js"]`},
			{"./main", `["./main"]`},
			{"uvicorn main:app --host 0.0.0.0", `["uvicorn", "main:app", "--host", "0.0.0.0"]`},
		}

		for _, tc := range testCases {
			result := shellToDockerCmd(tc.command)
			assert.Equal(t, tc.expected, result, "command: %s", tc.command)
		}
	})
}