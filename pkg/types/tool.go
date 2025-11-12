// Package types defines core types and interfaces for the wink CLI agent.
package types

import "context"

// RiskLevel categorizes tool operations by their potential impact
type RiskLevel string

const (
	// RiskLevelReadOnly - Read-only operations (safe)
	RiskLevelReadOnly RiskLevel = "read_only"
	// RiskLevelSafeWrite - Create new files/directories (relatively safe)
	RiskLevelSafeWrite RiskLevel = "safe_write"
	// RiskLevelDangerous - Modify existing files, execute commands, network access
	RiskLevelDangerous RiskLevel = "dangerous"
)

// Tool represents a capability available to the agent
type Tool interface {
	// Name returns the unique identifier for this tool
	Name() string

	// Description returns what this tool does (for LLM and users)
	Description() string

	// ParametersSchema returns JSON Schema defining parameters
	ParametersSchema() map[string]interface{}

	// Validate checks if parameters are valid before execution
	Validate(params map[string]interface{}, workingDir string) error

	// Execute runs the tool and returns result
	Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*ToolResult, error)

	// RequiresApproval returns true if this tool needs user approval
	RequiresApproval() bool

	// RiskLevel returns the risk category of this tool
	RiskLevel() RiskLevel
}

// ToolResult represents the output from executing a tool
type ToolResult struct {
	ToolCallID      string                 `json:"tool_call_id"`
	Success         bool                   `json:"success"`
	Output          string                 `json:"output"`
	Error           string                 `json:"error,omitempty"`
	ExecutionTimeMs int64                  `json:"execution_time_ms"`
	FilesAffected   []string               `json:"files_affected,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
