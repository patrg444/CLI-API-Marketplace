package cmd

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/api-direct/cli/pkg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTargetEnvironment(t *testing.T) {
	tests := []struct {
		name       string
		production bool
		staging    bool
		local      bool
		expected   string
	}{
		{"default", false, false, false, "development"},
		{"production", true, false, false, "production"},
		{"staging", false, true, false, "staging"},
		{"local", false, false, true, "local"},
		{"production overrides", true, true, true, "production"}, // Production has priority
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			envProduction = tt.production
			envStaging = tt.staging
			envLocal = tt.local

			result := getTargetEnvironment()
			assert.Equal(t, tt.expected, result)

			// Reset flags
			envProduction = false
			envStaging = false
			envLocal = false
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		key       string
		sensitive bool
	}{
		{"DATABASE_PASSWORD", true},
		{"API_SECRET", true},
		{"PRIVATE_KEY", true},
		{"ACCESS_TOKEN", true},
		{"AUTH_CREDENTIAL", true},
		{"SECRET_KEY", true},
		{"password", true}, // Case insensitive
		{"LOG_LEVEL", false},
		{"PORT", false},
		{"DEBUG", false},
		{"ENVIRONMENT", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := isSensitiveKey(tt.key)
			assert.Equal(t, tt.sensitive, result)
		})
	}
}

func TestMaskValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "****"},
		{"a", "****"},
		{"ab", "****"},
		{"abc", "****"},
		{"abcd", "****"},
		{"abcde", "ab...de"},
		{"secret123", "se...23"},
		{"verylongsecretvalue", "ve...ue"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := maskValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadDotenvFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name: "basic variables",
			content: `PORT=8080
LOG_LEVEL=debug
DATABASE_URL=postgres://localhost/db`,
			expected: map[string]string{
				"PORT":         "8080",
				"LOG_LEVEL":    "debug",
				"DATABASE_URL": "postgres://localhost/db",
			},
		},
		{
			name: "with comments and empty lines",
			content: `# This is a comment
PORT=8080

# Another comment
LOG_LEVEL=debug
`,
			expected: map[string]string{
				"PORT":      "8080",
				"LOG_LEVEL": "debug",
			},
		},
		{
			name: "quoted values",
			content: `MESSAGE="Hello World"
PATH='/usr/local/bin'
MIXED="value with 'quotes'"`,
			expected: map[string]string{
				"MESSAGE": "Hello World",
				"PATH":    "/usr/local/bin",
				"MIXED":   "value with 'quotes'",
			},
		},
		{
			name: "values with equals",
			content: `CONNECTION_STRING=user=admin;password=secret
EQUATION=a=b+c`,
			expected: map[string]string{
				"CONNECTION_STRING": "user=admin;password=secret",
				"EQUATION":          "a=b+c",
			},
		},
		{
			name: "spaces around equals",
			content: `KEY1 = value1
KEY2= value2
KEY3 =value3`,
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpfile, err := os.CreateTemp("", "test.env")
			require.NoError(t, err)
			defer os.Remove(tmpfile.Name())

			// Write content
			_, err = tmpfile.Write([]byte(tt.content))
			require.NoError(t, err)
			tmpfile.Close()

			// Test reading
			result, err := readDotenvFile(tmpfile.Name())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		_, err := readDotenvFile("/non/existent/file.env")
		assert.Error(t, err)
	})
}

func TestWriteDotenvFile(t *testing.T) {
	tests := []struct {
		name     string
		vars     map[string]string
		validate func(t *testing.T, content string)
	}{
		{
			name: "basic write",
			vars: map[string]string{
				"PORT":      "8080",
				"LOG_LEVEL": "debug",
			},
			validate: func(t *testing.T, content string) {
				assert.Contains(t, content, "PORT=8080")
				assert.Contains(t, content, "LOG_LEVEL=debug")
				assert.Contains(t, content, "# Environment variables for API-Direct")
			},
		},
		{
			name: "sorted keys",
			vars: map[string]string{
				"Z_VAR": "last",
				"A_VAR": "first",
				"M_VAR": "middle",
			},
			validate: func(t *testing.T, content string) {
				lines := strings.Split(content, "\n")
				var varLines []string
				for _, line := range lines {
					if strings.Contains(line, "_VAR=") {
						varLines = append(varLines, line)
					}
				}
				assert.Equal(t, "A_VAR=first", varLines[0])
				assert.Equal(t, "M_VAR=middle", varLines[1])
				assert.Equal(t, "Z_VAR=last", varLines[2])
			},
		},
		{
			name: "quoted values",
			vars: map[string]string{
				"MESSAGE":  "Hello World",
				"COMMAND":  "echo $PATH",
				"SPECIAL":  "value#with#hashes",
				"NORMAL":   "simple",
			},
			validate: func(t *testing.T, content string) {
				assert.Contains(t, content, `MESSAGE="Hello World"`)
				assert.Contains(t, content, `COMMAND="echo $PATH"`)
				assert.Contains(t, content, `SPECIAL="value#with#hashes"`)
				assert.Contains(t, content, "NORMAL=simple")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile := filepath.Join(t.TempDir(), "test.env")

			err := writeDotenvFile(tmpfile, tt.vars)
			require.NoError(t, err)

			// Read back and validate
			content, err := os.ReadFile(tmpfile)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, string(content))
			}

			// Verify it can be read back correctly
			readVars, err := readDotenvFile(tmpfile)
			assert.NoError(t, err)
			assert.Equal(t, tt.vars, readVars)
		})
	}
}

func TestSetLocalEnvVars(t *testing.T) {
	t.Run("create new env file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		
		err := os.Chdir(tempDir)
		require.NoError(t, err)

		vars := map[string]string{
			"NEW_VAR": "value",
			"PORT":    "3000",
		}

		err = setLocalEnvVars(vars)
		assert.NoError(t, err)

		// Verify file was created
		assert.FileExists(t, ".env")

		// Verify content
		readVars, err := readDotenvFile(".env")
		assert.NoError(t, err)
		assert.Equal(t, vars, readVars)
	})

	t.Run("update existing env file", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		
		err := os.Chdir(tempDir)
		require.NoError(t, err)

		// Create initial .env
		initial := map[string]string{
			"EXISTING": "old_value",
			"PORT":     "8080",
		}
		err = writeDotenvFile(".env", initial)
		require.NoError(t, err)

		// Update with new vars
		updates := map[string]string{
			"EXISTING": "new_value", // Update existing
			"NEW_VAR":  "added",     // Add new
		}

		err = setLocalEnvVars(updates)
		assert.NoError(t, err)

		// Verify merged result
		readVars, err := readDotenvFile(".env")
		assert.NoError(t, err)
		assert.Equal(t, "new_value", readVars["EXISTING"])
		assert.Equal(t, "added", readVars["NEW_VAR"])
		assert.Equal(t, "8080", readVars["PORT"]) // Unchanged
	})
}

func TestGetLocalEnvVar(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	// Create .env file
	vars := map[string]string{
		"TEST_VAR": "test_value",
		"PORT":     "8080",
	}
	err = writeDotenvFile(".env", vars)
	require.NoError(t, err)

	tests := []struct {
		key     string
		want    string
		wantErr bool
	}{
		{"TEST_VAR", "test_value", false},
		{"PORT", "8080", false},
		{"NON_EXISTENT", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			value, err := getLocalEnvVar(tt.key)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, value)
			}
		})
	}
}

func TestUnsetLocalEnvVars(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	
	err := os.Chdir(tempDir)
	require.NoError(t, err)

	// Create .env file with multiple vars
	initial := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
		"VAR3": "value3",
		"VAR4": "value4",
	}
	err = writeDotenvFile(".env", initial)
	require.NoError(t, err)

	t.Run("unset existing vars", func(t *testing.T) {
		err = unsetLocalEnvVars([]string{"VAR1", "VAR3"})
		assert.NoError(t, err)

		// Verify removed
		readVars, err := readDotenvFile(".env")
		assert.NoError(t, err)
		assert.NotContains(t, readVars, "VAR1")
		assert.NotContains(t, readVars, "VAR3")
		assert.Contains(t, readVars, "VAR2")
		assert.Contains(t, readVars, "VAR4")
	})

	t.Run("unset non-existent vars", func(t *testing.T) {
		err = unsetLocalEnvVars([]string{"NON_EXISTENT"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "No matching variables")
	})

	t.Run("mixed existing and non-existent", func(t *testing.T) {
		err = unsetLocalEnvVars([]string{"VAR2", "NON_EXISTENT"})
		assert.NoError(t, err) // Should succeed if at least one exists

		readVars, err := readDotenvFile(".env")
		assert.NoError(t, err)
		assert.NotContains(t, readVars, "VAR2")
	})
}

func TestOutputFormatters(t *testing.T) {
	envVars := map[string]map[string]string{
		"development": {
			"LOG_LEVEL": "debug",
			"PORT":      "8080",
		},
		"production": {
			"LOG_LEVEL": "error",
			"PORT":      "8080",
			"API_KEY":   "secret",
		},
	}

	t.Run("JSON output", func(t *testing.T) {
		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputEnvJSON(envVars)
		assert.NoError(t, err)

		w.Close()
		os.Stdout = old

		buf := new(strings.Builder)
		io.Copy(buf, r)
		output := buf.String()

		// Verify it's valid JSON
		var parsed map[string]map[string]string
		err = json.Unmarshal([]byte(output), &parsed)
		assert.NoError(t, err)
		assert.Equal(t, envVars, parsed)
	})

	t.Run("dotenv output", func(t *testing.T) {
		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputDotenv(envVars)
		assert.NoError(t, err)

		w.Close()
		os.Stdout = old

		buf := new(strings.Builder)
		io.Copy(buf, r)
		output := buf.String()

		// Verify format
		assert.Contains(t, output, "# development")
		assert.Contains(t, output, "LOG_LEVEL=debug")
		assert.Contains(t, output, "# production")
		assert.Contains(t, output, "API_KEY=secret")
	})

	t.Run("table output", func(t *testing.T) {
		// Capture output
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := outputTable(envVars)
		assert.NoError(t, err)

		w.Close()
		os.Stdout = old

		buf := new(strings.Builder)
		io.Copy(buf, r)
		output := buf.String()

		// Verify format
		assert.Contains(t, output, "Development environment")
		assert.Contains(t, output, "Production environment")
		assert.Contains(t, output, "LOG_LEVEL")
		// API_KEY should be masked
		assert.NotContains(t, output, "API_KEY     = secret")
		assert.Contains(t, output, "API_KEY")
		assert.Contains(t, output, "se...et") // Masked value
	})
}

func TestGetAPIName(t *testing.T) {
	t.Run("from directory name", func(t *testing.T) {
		tempDir := t.TempDir()
		testDir := filepath.Join(tempDir, "test-api-name")
		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		
		err = os.Chdir(testDir)
		require.NoError(t, err)

		name, err := getAPIName()
		assert.NoError(t, err)
		assert.Equal(t, "test-api-name", name)
	})

	t.Run("from manifest", func(t *testing.T) {
		tempDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		
		err := os.Chdir(tempDir)
		require.NoError(t, err)

		// Create a mock manifest
		manifestContent := `name: my-awesome-api
runtime: python3.11
start_command: python main.py
port: 8080
endpoints:
  - GET /hello`
		
		err = os.WriteFile("apidirect.yaml", []byte(manifestContent), 0644)
		require.NoError(t, err)

		// Verify we're in the right directory
		cwd, _ := os.Getwd()
		assert.Contains(t, cwd, tempDir)
		
		// Verify manifest exists
		_, err = os.Stat("apidirect.yaml")
		assert.NoError(t, err)
		
		// Test manifest loading directly  
		manifestPath, manifestErr := manifest.FindManifest(".")
		assert.NoError(t, manifestErr)
		
		m, loadErr := manifest.Load(manifestPath)
		assert.NoError(t, loadErr)
		assert.Equal(t, "my-awesome-api", m.Name)
		
		name, err := getAPIName()
		assert.NoError(t, err)
		assert.Equal(t, "my-awesome-api", name)
	})
}

func TestGetEnvironmentVars(t *testing.T) {
	t.Run("single environment", func(t *testing.T) {
		vars := getEnvironmentVars("test-api", "production", false)
		assert.Len(t, vars, 1)
		assert.Contains(t, vars, "production")
		assert.Contains(t, vars["production"], "LOG_LEVEL")
		assert.Contains(t, vars["production"], "API_KEY")
	})

	t.Run("all environments", func(t *testing.T) {
		vars := getEnvironmentVars("test-api", "development", true)
		assert.Len(t, vars, 3)
		assert.Contains(t, vars, "development")
		assert.Contains(t, vars, "staging")
		assert.Contains(t, vars, "production")
	})
}

func TestEnvCommandHelpers(t *testing.T) {
	t.Run("parse KEY=VALUE format", func(t *testing.T) {
		args := []string{"KEY1=value1", "KEY2=value2", "KEY3=value with spaces"}
		
		vars := make(map[string]string)
		for _, arg := range args {
			parts := strings.SplitN(arg, "=", 2)
			if len(parts) == 2 {
				vars[parts[0]] = parts[1]
			}
		}

		assert.Equal(t, "value1", vars["KEY1"])
		assert.Equal(t, "value2", vars["KEY2"])
		assert.Equal(t, "value with spaces", vars["KEY3"])
	})

	t.Run("invalid KEY=VALUE format", func(t *testing.T) {
		arg := "INVALID_FORMAT"
		parts := strings.SplitN(arg, "=", 2)
		assert.Len(t, parts, 1)
	})
}