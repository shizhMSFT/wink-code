// Package tools implements the create_file tool
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

const (
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

// CreateFileTool implements the create_file tool
type CreateFileTool struct{}

// NewCreateFileTool creates a new create_file tool instance
func NewCreateFileTool() *CreateFileTool {
	return &CreateFileTool{}
}

// Name returns the tool name
func (t *CreateFileTool) Name() string {
	return "create_file"
}

// Description returns the tool description
func (t *CreateFileTool) Description() string {
	return "Create a new file with specified content. The file must not already exist."
}

// ParametersSchema returns the JSON schema for parameters
func (t *CreateFileTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the file to create",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write to the file",
			},
		},
		"required": []string{"path", "content"},
	}
}

// Validate checks if parameters are valid
func (t *CreateFileTool) Validate(params map[string]interface{}, workingDir string) error {
	// Check required parameters
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	content, ok := params["content"].(string)
	if !ok {
		return fmt.Errorf("content parameter is required and must be a string")
	}

	// Check content size
	if len(content) > maxFileSize {
		return fmt.Errorf("content size (%d bytes) exceeds maximum allowed size (%d bytes)",
			len(content), maxFileSize)
	}

	// Validate path is within working directory
	resolvedPath, err := ResolvePath(workingDir, path)
	if err != nil {
		return err
	}

	// Check if file already exists
	if _, err := os.Stat(resolvedPath); err == nil {
		return fmt.Errorf("file '%s' already exists. Use replace_string_in_file to modify", path)
	}

	return nil
}

// Execute creates the file
func (t *CreateFileTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	// Extract parameters
	path := params["path"].(string)
	content := params["content"].(string)

	// Resolve path
	resolvedPath, err := ResolvePath(workingDir, path)
	if err != nil {
		return &types.ToolResult{
			Success:         false,
			Error:           err.Error(),
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, err
	}

	// Log operation
	logging.Debug("Creating file",
		"path", SanitizePathForDisplay(workingDir, resolvedPath),
		"size_bytes", len(content),
	)

	// Create parent directories if needed
	dir := filepath.Dir(resolvedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to create directory '%s': %v", dir, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf(errMsg)
	}

	// Write file
	if err := os.WriteFile(resolvedPath, []byte(content), 0644); err != nil {
		errMsg := fmt.Sprintf("failed to write file '%s': %v", path, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf(errMsg)
	}

	// Get file info for result
	fileInfo, _ := os.Stat(resolvedPath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	}

	executionTime := time.Since(startTime).Milliseconds()

	logging.Info("File created",
		"path", SanitizePathForDisplay(workingDir, resolvedPath),
		"size_bytes", fileSize,
		"execution_time_ms", executionTime,
	)

	return &types.ToolResult{
		Success:         true,
		Output:          fmt.Sprintf("Created file: %s (%d bytes)", path, fileSize),
		ExecutionTimeMs: executionTime,
		FilesAffected:   []string{path},
		Metadata: map[string]interface{}{
			"size_bytes": fileSize,
			"path":       path,
		},
	}, nil
}

// RequiresApproval returns true as file creation requires approval
func (t *CreateFileTool) RequiresApproval() bool {
	return true
}

// RiskLevel returns the risk level for this tool
func (t *CreateFileTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelSafeWrite
}
