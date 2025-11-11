// Package config handles configuration loading and saving
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shizhMSFT/wink-code/pkg/types"
	"github.com/spf13/viper"
)

const (
	configDir  = ".wink"
	configFile = "config.json"
)

// Manager handles configuration operations
type Manager struct {
	config     *types.Config
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	// Determine config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configDir, configFile)

	// Create config directory if it doesn't exist
	configDirPath := filepath.Dir(configPath)
	if err := os.MkdirAll(configDirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	manager := &Manager{
		configPath: configPath,
	}

	// Load or create default config
	if err := manager.Load(); err != nil {
		return nil, err
	}

	return manager, nil
}

// Load reads configuration from file or creates default
func (m *Manager) Load() error {
	// Check if config file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Create default config
		m.config = types.DefaultConfig()
		return m.Save()
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	m.config = &types.Config{}
	if err := json.Unmarshal(data, m.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate
	if err := m.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	return nil
}

// Save writes configuration to file
func (m *Manager) Save() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if configuration is valid
func (m *Manager) Validate() error {
	if m.config.DefaultModel == "" {
		return fmt.Errorf("default_model cannot be empty")
	}
	if m.config.OllamaBaseURL == "" {
		return fmt.Errorf("ollama_base_url cannot be empty")
	}
	if m.config.APITimeoutSeconds < 5 || m.config.APITimeoutSeconds > 300 {
		return fmt.Errorf("api_timeout_seconds must be between 5 and 300")
	}
	if m.config.MaxSessionMessages < 10 || m.config.MaxSessionMessages > 1000 {
		return fmt.Errorf("max_session_messages must be between 10 and 1000")
	}
	return nil
}

// Get returns the current configuration
func (m *Manager) Get() *types.Config {
	return m.config
}

// Update updates configuration and saves to file
func (m *Manager) Update(config *types.Config) error {
	m.config = config
	if err := m.Validate(); err != nil {
		return err
	}
	return m.Save()
}

// LoadWithViper loads config using viper (supports env vars and CLI flags)
func LoadWithViper() (*types.Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	// Add config paths
	homeDir, _ := os.UserHomeDir()
	viper.AddConfigPath(filepath.Join(homeDir, configDir))
	viper.AddConfigPath(".")

	// Set defaults
	defaultCfg := types.DefaultConfig()
	viper.SetDefault("default_model", defaultCfg.DefaultModel)
	viper.SetDefault("ollama_base_url", defaultCfg.OllamaBaseURL)
	viper.SetDefault("api_timeout_seconds", defaultCfg.APITimeoutSeconds)
	viper.SetDefault("max_session_messages", defaultCfg.MaxSessionMessages)
	viper.SetDefault("output_format", defaultCfg.OutputFormat)

	// Environment variables
	viper.SetEnvPrefix("WINK")
	viper.AutomaticEnv()

	// Read config file (optional - use defaults if not found)
	_ = viper.ReadInConfig()

	config := &types.Config{
		ConfigVersion:      defaultCfg.ConfigVersion,
		DefaultModel:       viper.GetString("default_model"),
		OllamaBaseURL:      viper.GetString("ollama_base_url"),
		APITimeoutSeconds:  viper.GetInt("api_timeout_seconds"),
		MaxSessionMessages: viper.GetInt("max_session_messages"),
		OutputFormat:       types.OutputFormat(viper.GetString("output_format")),
		AutoApprovalRules:  []types.ApprovalRule{},
	}

	return config, nil
}
