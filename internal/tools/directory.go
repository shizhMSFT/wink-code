// Package tools implements directory operation tools
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

const (
	maxDirEntries = 1000 // Maximum number of directory entries to return
)

// CreateDirectoryTool implements the create_directory tool
type CreateDirectoryTool struct{}

// NewCreateDirectoryTool creates a new create_directory tool instance
func NewCreateDirectoryTool() *CreateDirectoryTool {
	return &CreateDirectoryTool{}
}

// Name returns the tool name
func (t *CreateDirectoryTool) Name() string {
	return "create_directory"
}

// Description returns the tool description
func (t *CreateDirectoryTool) Description() string {
	return "Create a directory structure recursively"
}

// ParametersSchema returns the JSON schema for parameters
func (t *CreateDirectoryTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the directory to create",
			},
		},
		"required": []string{"path"},
	}
}

// Validate checks if parameters are valid
func (t *CreateDirectoryTool) Validate(params map[string]interface{}, workingDir string) error {
	// Check required parameters
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	// Validate path is within working directory
	_, err := ResolvePath(workingDir, path)
	if err != nil {
		return err
	}

	return nil
}

// Execute creates the directory
func (t *CreateDirectoryTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	// Extract parameters
	path := params["path"].(string)

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
	logging.Debug("Creating directory",
		"path", SanitizePathForDisplay(workingDir, resolvedPath),
	)

	// Create directory with parents
	if err := os.MkdirAll(resolvedPath, 0755); err != nil {
		errMsg := fmt.Sprintf("failed to create directory '%s': %v", path, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("failed to create directory '%s': %w", path, err)
	}

	executionTime := time.Since(startTime).Milliseconds()

	logging.Info("Directory created",
		"path", SanitizePathForDisplay(workingDir, resolvedPath),
		"execution_time_ms", executionTime,
	)

	return &types.ToolResult{
		Success:         true,
		Output:          fmt.Sprintf("Created directory: %s", path),
		ExecutionTimeMs: executionTime,
		FilesAffected:   []string{path},
		Metadata: map[string]interface{}{
			"path": path,
		},
	}, nil
}

// RequiresApproval returns true as directory creation requires approval
func (t *CreateDirectoryTool) RequiresApproval() bool {
	return true
}

// RiskLevel returns the risk level for this tool
func (t *CreateDirectoryTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelSafeWrite
}

// ListDirTool implements the list_dir tool
type ListDirTool struct{}

// NewListDirTool creates a new list_dir tool instance
func NewListDirTool() *ListDirTool {
	return &ListDirTool{}
}

// Name returns the tool name
func (l *ListDirTool) Name() string {
	return "list_dir"
}

// Description returns the tool description
func (l *ListDirTool) Description() string {
	return "List contents of a directory"
}

// ParametersSchema returns the JSON schema for parameters
func (l *ListDirTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to directory (default: current directory)",
				"default":     ".",
			},
		},
	}
}

// Validate checks if parameters are valid
func (l *ListDirTool) Validate(params map[string]interface{}, workingDir string) error {
	// Get path parameter or use default
	path := "."
	if p, ok := params["path"].(string); ok && p != "" {
		path = p
	}

	// Validate path is within working directory
	resolvedPath, err := ResolvePath(workingDir, path)
	if err != nil {
		return err
	}

	// Check if path exists
	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path '%s' not found", path)
		}
		return fmt.Errorf("cannot access path '%s': %w", path, err)
	}

	// Check if it's a directory
	if !fileInfo.IsDir() {
		return fmt.Errorf("path '%s' is not a directory", path)
	}

	return nil
}

// Execute lists the directory
func (l *ListDirTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	// Get path parameter or use default
	path := "."
	if p, ok := params["path"].(string); ok && p != "" {
		path = p
	}

	// Resolve path
	resolvedPath, err := ResolvePath(workingDir, path)
	if err != nil {
		return &types.ToolResult{
			Success:         false,
			Error:           err.Error(),
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, err
	}

	// Read directory
	entries, err := os.ReadDir(resolvedPath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read directory '%s': %v", path, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("failed to read directory '%s': %w", path, err)
	}

	// Count files and directories
	fileCount := 0
	dirCount := 0
	var names []string

	for _, entry := range entries {
		if len(names) >= maxDirEntries {
			break
		}

		name := entry.Name()
		if entry.IsDir() {
			name += string(filepath.Separator)
			dirCount++
		} else {
			fileCount++
		}
		names = append(names, name)
	}

	// Sort entries
	sort.Strings(names)

	// Format output
	output := fmt.Sprintf("Contents of %s:\n", path)
	for _, name := range names {
		output += fmt.Sprintf("  %s\n", name)
	}

	if len(entries) > maxDirEntries {
		output += fmt.Sprintf("  ... (%d more entries not shown)\n", len(entries)-maxDirEntries)
	}

	executionTime := time.Since(startTime).Milliseconds()

	logging.Info("Directory listed",
		"path", SanitizePathForDisplay(workingDir, resolvedPath),
		"total_entries", len(entries),
		"files", fileCount,
		"directories", dirCount,
		"execution_time_ms", executionTime,
	)

	return &types.ToolResult{
		Success:         true,
		Output:          output,
		ExecutionTimeMs: executionTime,
		FilesAffected:   []string{},
		Metadata: map[string]interface{}{
			"total_entries": len(entries),
			"files":         fileCount,
			"directories":   dirCount,
		},
	}, nil
}

// RequiresApproval returns true as directory listing requires approval
func (l *ListDirTool) RequiresApproval() bool {
	return true
}

// RiskLevel returns the risk level for this tool
func (l *ListDirTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelReadOnly
}
