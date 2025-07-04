package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// configMutex protects concurrent access to config file
	configMutex sync.RWMutex
)

// Config represents the CLI configuration
type Config struct {
	Auth        AuthConfig                  `json:"auth"`
	API         APIConfig                   `json:"api"`
	Preferences PreferencesConfig           `json:"preferences"`
	User        UserConfig                  `json:"user,omitempty"`
	Deployments map[string]interface{}      `json:"deployments,omitempty"`
	AuthToken   string                      `json:"-"` // Temporary for backward compatibility
	APIEndpoint string                      `json:"-"` // Temporary for backward compatibility
}

// AuthConfig stores authentication information
type AuthConfig struct {
	AccessToken  string    `json:"access_token,omitempty"`
	IDToken      string    `json:"id_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Username     string    `json:"username,omitempty"`
	Email        string    `json:"email,omitempty"`
}

// APIConfig stores API endpoint configuration
type APIConfig struct {
	BaseURL      string `json:"base_url"`
	Region       string `json:"region"`
	CognitoPool  string `json:"cognito_pool"`
	CognitoClient string `json:"cognito_client"`
}

// PreferencesConfig stores user preferences
type PreferencesConfig struct {
	DefaultRuntime string `json:"default_runtime,omitempty"`
	OutputFormat   string `json:"output_format,omitempty"`
}

// UserConfig stores user information
type UserConfig struct {
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			BaseURL:       "https://api.api-direct.io",
			Region:        "us-east-1",
			CognitoPool:   "", // Will be set from environment or during setup
			CognitoClient: "", // Will be set from environment or during setup
		},
		Preferences: PreferencesConfig{
			DefaultRuntime: "python3.9",
			OutputFormat:   "table",
		},
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".apidirect", "config.json"), nil
}

// Load is an alias for LoadConfig for backward compatibility
func Load() (*Config, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	// Set backward compatibility fields
	cfg.AuthToken = cfg.Auth.AccessToken
	cfg.APIEndpoint = cfg.API.BaseURL
	return cfg, nil
}

// Get returns the current configuration (loads if needed)
func Get() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		// Return default config on error
		return DefaultConfig()
	}
	// Set backward compatibility fields
	cfg.AuthToken = cfg.Auth.AccessToken
	cfg.APIEndpoint = cfg.API.BaseURL
	return cfg
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	// Always acquire read lock first
	configMutex.RLock()
	
	path, err := ConfigPath()
	if err != nil {
		configMutex.RUnlock()
		return nil, err
	}

	// Check if file exists with read lock
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Upgrade to write lock to create default config
		configMutex.RUnlock()
		configMutex.Lock()
		defer configMutex.Unlock()
		
		// Double-check after acquiring write lock
		if _, err := os.Stat(path); os.IsNotExist(err) {
			config := DefaultConfig()
			// Check environment variables for default config
			if envURL := os.Getenv("APIDIRECT_API_ENDPOINT"); envURL != "" {
				config.API.BaseURL = envURL
			}
			if envToken := os.Getenv("APIDIRECT_AUTH_TOKEN"); envToken != "" {
				config.Auth.AccessToken = envToken
			}
			if err := saveConfigLocked(config); err != nil {
				return nil, fmt.Errorf("failed to save default config: %w", err)
			}
			return config, nil
		}
		// File was created by another goroutine, fall through to read it
		// Note: we hold the write lock but that's OK for reading
	} else {
		// File exists, keep read lock
		defer configMutex.RUnlock()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults for any missing fields
	defaultConfig := DefaultConfig()
	if config.API.BaseURL == "" {
		// Check environment variable first
		if envURL := os.Getenv("APIDIRECT_API_ENDPOINT"); envURL != "" {
			config.API.BaseURL = envURL
		} else {
			config.API.BaseURL = defaultConfig.API.BaseURL
		}
	}
	if config.API.Region == "" {
		config.API.Region = defaultConfig.API.Region
	}
	if config.Preferences.DefaultRuntime == "" {
		config.Preferences.DefaultRuntime = defaultConfig.Preferences.DefaultRuntime
	}
	if config.Preferences.OutputFormat == "" {
		config.Preferences.OutputFormat = defaultConfig.Preferences.OutputFormat
	}
	
	// Check for auth token environment variable
	if envToken := os.Getenv("APIDIRECT_AUTH_TOKEN"); envToken != "" && config.Auth.AccessToken == "" {
		config.Auth.AccessToken = envToken
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()
	
	return saveConfigLocked(config)
}

// saveConfigLocked saves the configuration to disk without locking (caller must hold lock)
func saveConfigLocked(config *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// UpdateAuth updates the authentication configuration
func UpdateAuth(auth AuthConfig) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.Auth = auth
	
	// Also update user info if available
	if auth.Email != "" {
		config.User.Email = auth.Email
	}
	if auth.Username != "" {
		config.User.Username = auth.Username
	}
	
	return SaveConfig(config)
}

// ClearAuth clears the authentication configuration
func ClearAuth() error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}

	config.Auth = AuthConfig{}
	return SaveConfig(config)
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated() bool {
	config, err := LoadConfig()
	if err != nil {
		return false
	}

	// Check if we have tokens and they haven't expired
	if config.Auth.AccessToken == "" {
		return false
	}

	// Check expiration
	if !config.Auth.ExpiresAt.IsZero() && time.Now().After(config.Auth.ExpiresAt) {
		return false
	}

	return true
}

// GetProjectConfig loads the project configuration from apidirect.yaml
type ProjectConfig struct {
	Name      string              `yaml:"name"`
	Runtime   string              `yaml:"runtime"`
	Endpoints []EndpointConfig    `yaml:"endpoints"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

type EndpointConfig struct {
	Path    string `yaml:"path"`
	Method  string `yaml:"method"`
	Handler string `yaml:"handler"`
}
