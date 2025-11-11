// Package types defines configuration types
package types

// OutputFormat specifies the output format
type OutputFormat string

const (
	// OutputFormatHuman - Human-readable output
	OutputFormatHuman OutputFormat = "human"
	// OutputFormatJSON - JSON output
	OutputFormatJSON OutputFormat = "json"
)

// Config represents user configuration and preferences
type Config struct {
	ConfigVersion      string         `json:"config_version"`
	DefaultModel       string         `json:"default_model"`
	OllamaBaseURL      string         `json:"ollama_base_url"`
	APITimeoutSeconds  int            `json:"api_timeout_seconds"`
	MaxSessionMessages int            `json:"max_session_messages"`
	AutoApprovalRules  []ApprovalRule `json:"auto_approval_rules"`
	OutputFormat       OutputFormat   `json:"output_format"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ConfigVersion:      "1.0",
		DefaultModel:       "qwen3:8b",
		OllamaBaseURL:      "http://localhost:11434",
		APITimeoutSeconds:  30,
		MaxSessionMessages: 100,
		AutoApprovalRules:  []ApprovalRule{},
		OutputFormat:       OutputFormatHuman,
	}
}
