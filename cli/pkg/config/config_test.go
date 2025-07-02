package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	assert.NotNil(t, cfg)
	assert.Equal(t, "https://api.api-direct.io", cfg.API.BaseURL)
	assert.Equal(t, "us-east-1", cfg.API.Region)
	assert.Equal(t, "python3.9", cfg.Preferences.DefaultRuntime)
	assert.Equal(t, "table", cfg.Preferences.OutputFormat)
}

func TestConfigPath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE") // Windows
	}
	
	// Test with custom HOME
	testHome := t.TempDir()
	os.Setenv("HOME", testHome)
	defer os.Setenv("HOME", originalHome)
	
	path, err := ConfigPath()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(testHome, ".apidirect", "config.json"), path)
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		setupFunc   func(t *testing.T, configPath string)
		validate    func(t *testing.T, cfg *Config, err error)
	}{
		{
			name: "load valid config",
			configData: `{
  "auth": {
    "access_token": "test-token",
    "refresh_token": "refresh-token",
    "expires_at": "2030-01-01T00:00:00Z",
    "email": "test@example.com"
  },
  "api": {
    "base_url": "https://custom.api.com",
    "region": "eu-west-1"
  },
  "preferences": {
    "default_runtime": "node18",
    "output_format": "json"
  },
  "user": {
    "email": "test@example.com",
    "username": "testuser"
  }
}`,
			validate: func(t *testing.T, cfg *Config, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, "test-token", cfg.Auth.AccessToken)
				assert.Equal(t, "https://custom.api.com", cfg.API.BaseURL)
				assert.Equal(t, "node18", cfg.Preferences.DefaultRuntime)
				assert.Equal(t, "test@example.com", cfg.User.Email)
			},
		},
		{
			name: "create default config if not exists",
			setupFunc: func(t *testing.T, configPath string) {
				// Ensure config doesn't exist
				os.Remove(configPath)
			},
			validate: func(t *testing.T, cfg *Config, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, "https://api.api-direct.io", cfg.API.BaseURL)
				assert.Equal(t, "python3.9", cfg.Preferences.DefaultRuntime)
				
				// Check file was created
				home := os.Getenv("HOME")
				configPath := filepath.Join(home, ".apidirect", "config.json")
				_, statErr := os.Stat(configPath)
				assert.NoError(t, statErr)
			},
		},
		{
			name: "merge with defaults for missing fields",
			configData: `{
  "auth": {
    "access_token": "token"
  }
}`,
			validate: func(t *testing.T, cfg *Config, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "token", cfg.Auth.AccessToken)
				// Should have default values
				assert.Equal(t, "https://api.api-direct.io", cfg.API.BaseURL)
				assert.Equal(t, "python3.9", cfg.Preferences.DefaultRuntime)
			},
		},
		{
			name: "invalid json",
			configData: `{invalid json}`,
			validate: func(t *testing.T, cfg *Config, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "failed to parse config file")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := t.TempDir()
			os.Setenv("HOME", testDir)
			defer os.Unsetenv("HOME")
			
			configPath := filepath.Join(testDir, ".apidirect", "config.json")
			
			// Create config directory
			os.MkdirAll(filepath.Dir(configPath), 0755)
			
			// Write config data if provided
			if tt.configData != "" {
				err := ioutil.WriteFile(configPath, []byte(tt.configData), 0600)
				require.NoError(t, err)
			}
			
			// Run setup function if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, configPath)
			}
			
			// Load config
			cfg, err := LoadConfig()
			
			// Validate
			tt.validate(t, cfg, err)
		})
	}
}

func TestSaveConfig(t *testing.T) {
	// Setup test directory
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	defer os.Unsetenv("HOME")
	
	cfg := &Config{
		Auth: AuthConfig{
			AccessToken:  "test-token",
			RefreshToken: "refresh-token",
			ExpiresAt:    time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
			Email:        "test@example.com",
		},
		API: APIConfig{
			BaseURL: "https://test.api.com",
			Region:  "us-west-2",
		},
		Preferences: PreferencesConfig{
			DefaultRuntime: "go1.21",
			OutputFormat:   "yaml",
		},
		User: UserConfig{
			Email:    "test@example.com",
			Username: "testuser",
			UserID:   "user-123",
		},
		Deployments: map[string]interface{}{
			"test-api": map[string]interface{}{
				"id":     "dep-123",
				"status": "running",
			},
		},
	}
	
	// Save config
	err := SaveConfig(cfg)
	assert.NoError(t, err)
	
	// Check file exists
	configPath := filepath.Join(testDir, ".apidirect", "config.json")
	assert.FileExists(t, configPath)
	
	// Check permissions
	info, err := os.Stat(configPath)
	assert.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	
	// Load and verify
	data, err := ioutil.ReadFile(configPath)
	assert.NoError(t, err)
	
	var loaded Config
	err = json.Unmarshal(data, &loaded)
	assert.NoError(t, err)
	
	assert.Equal(t, cfg.Auth.AccessToken, loaded.Auth.AccessToken)
	assert.Equal(t, cfg.API.BaseURL, loaded.API.BaseURL)
	assert.Equal(t, cfg.Preferences.DefaultRuntime, loaded.Preferences.DefaultRuntime)
	assert.Equal(t, cfg.User.Email, loaded.User.Email)
	assert.NotNil(t, loaded.Deployments["test-api"])
}

func TestLoadAndGet(t *testing.T) {
	// Setup test directory
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	defer os.Unsetenv("HOME")
	
	// Create config
	configPath := filepath.Join(testDir, ".apidirect", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	
	configData := `{
  "auth": {
    "access_token": "legacy-token"
  },
  "api": {
    "base_url": "https://legacy.api.com"
  }
}`
	err := ioutil.WriteFile(configPath, []byte(configData), 0600)
	require.NoError(t, err)
	
	// Test Load() - backward compatibility
	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, "legacy-token", cfg.AuthToken) // Backward compatibility field
	assert.Equal(t, "https://legacy.api.com", cfg.APIEndpoint) // Backward compatibility field
	assert.Equal(t, "legacy-token", cfg.Auth.AccessToken)
	
	// Test Get() - returns config without error
	cfg2 := Get()
	assert.NotNil(t, cfg2)
	assert.Equal(t, "legacy-token", cfg2.AuthToken)
	assert.Equal(t, "https://legacy.api.com", cfg2.APIEndpoint)
}

func TestGetWithError(t *testing.T) {
	// Setup invalid HOME to cause error
	os.Setenv("HOME", "/invalid/path/that/does/not/exist")
	defer os.Unsetenv("HOME")
	
	// Get() should return default config on error
	cfg := Get()
	assert.NotNil(t, cfg)
	assert.Equal(t, "https://api.api-direct.io", cfg.API.BaseURL)
	assert.Equal(t, "python3.9", cfg.Preferences.DefaultRuntime)
}

func TestUpdateAuth(t *testing.T) {
	// Setup test directory
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	defer os.Unsetenv("HOME")
	
	// Create initial config
	initialCfg := DefaultConfig()
	err := SaveConfig(initialCfg)
	require.NoError(t, err)
	
	// Update auth
	newAuth := AuthConfig{
		AccessToken:  "new-token",
		RefreshToken: "new-refresh",
		ExpiresAt:    time.Now().Add(time.Hour),
		Email:        "new@example.com",
		Username:     "newuser",
	}
	
	err = UpdateAuth(newAuth)
	assert.NoError(t, err)
	
	// Load and verify
	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "new-token", cfg.Auth.AccessToken)
	assert.Equal(t, "new@example.com", cfg.Auth.Email)
	assert.Equal(t, "new@example.com", cfg.User.Email)
	assert.Equal(t, "newuser", cfg.User.Username)
}

func TestClearAuth(t *testing.T) {
	// Setup test directory
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	defer os.Unsetenv("HOME")
	
	// Create config with auth
	cfg := &Config{
		Auth: AuthConfig{
			AccessToken: "token-to-clear",
			Email:       "test@example.com",
		},
		API: APIConfig{
			BaseURL: "https://api.test.com",
		},
	}
	err := SaveConfig(cfg)
	require.NoError(t, err)
	
	// Clear auth
	err = ClearAuth()
	assert.NoError(t, err)
	
	// Load and verify
	loaded, err := LoadConfig()
	assert.NoError(t, err)
	assert.Empty(t, loaded.Auth.AccessToken)
	assert.Empty(t, loaded.Auth.Email)
	// Other config should remain
	assert.Equal(t, "https://api.test.com", loaded.API.BaseURL)
}

func TestIsAuthenticated(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, testDir string)
		expected   bool
	}{
		{
			name: "authenticated with valid token",
			setupFunc: func(t *testing.T, testDir string) {
				cfg := &Config{
					Auth: AuthConfig{
						AccessToken: "valid-token",
						ExpiresAt:   time.Now().Add(time.Hour),
					},
				}
				SaveConfig(cfg)
			},
			expected: true,
		},
		{
			name: "not authenticated - no token",
			setupFunc: func(t *testing.T, testDir string) {
				cfg := DefaultConfig()
				SaveConfig(cfg)
			},
			expected: false,
		},
		{
			name: "not authenticated - expired token",
			setupFunc: func(t *testing.T, testDir string) {
				cfg := &Config{
					Auth: AuthConfig{
						AccessToken: "expired-token",
						ExpiresAt:   time.Now().Add(-time.Hour),
					},
				}
				SaveConfig(cfg)
			},
			expected: false,
		},
		{
			name: "authenticated - no expiry set",
			setupFunc: func(t *testing.T, testDir string) {
				cfg := &Config{
					Auth: AuthConfig{
						AccessToken: "token-no-expiry",
					},
				}
				SaveConfig(cfg)
			},
			expected: true,
		},
		{
			name: "not authenticated - config error",
			setupFunc: func(t *testing.T, testDir string) {
				// Create invalid config
				configPath := filepath.Join(testDir, ".apidirect", "config.json")
				os.MkdirAll(filepath.Dir(configPath), 0755)
				ioutil.WriteFile(configPath, []byte("invalid json"), 0600)
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test directory
			testDir := t.TempDir()
			os.Setenv("HOME", testDir)
			defer os.Unsetenv("HOME")
			
			// Run setup
			tt.setupFunc(t, testDir)
			
			// Test
			result := IsAuthenticated()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigConcurrency(t *testing.T) {
	// Test sequential access patterns (the config package doesn't have mutex protection)
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	defer os.Unsetenv("HOME")
	
	// Create initial config
	cfg := DefaultConfig()
	err := SaveConfig(cfg)
	require.NoError(t, err)
	
	// Test multiple sequential reads
	for i := 0; i < 5; i++ {
		loaded, err := LoadConfig()
		assert.NoError(t, err)
		assert.NotNil(t, loaded)
	}
	
	// Test multiple sequential writes
	for i := 0; i < 5; i++ {
		auth := AuthConfig{
			AccessToken: fmt.Sprintf("token-%d", i),
		}
		err := UpdateAuth(auth)
		assert.NoError(t, err)
	}
	
	// Verify final state
	final, err := LoadConfig()
	assert.NoError(t, err)
	assert.Equal(t, "token-4", final.Auth.AccessToken)
}

func TestConfigMigration(t *testing.T) {
	// Test handling of old config formats
	testDir := t.TempDir()
	os.Setenv("HOME", testDir)
	defer os.Unsetenv("HOME")
	
	// Create old format config
	oldConfig := `{
  "auth_token": "old-token",
  "api_endpoint": "https://old.api.com"
}`
	
	configPath := filepath.Join(testDir, ".apidirect", "config.json")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	err := ioutil.WriteFile(configPath, []byte(oldConfig), 0600)
	require.NoError(t, err)
	
	// Load should handle gracefully
	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	// Should have defaults since old format won't parse correctly
	assert.Equal(t, "https://api.api-direct.io", cfg.API.BaseURL)
}