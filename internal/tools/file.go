// Package tools implements the create_file tool
package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		}, fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(resolvedPath, []byte(content), 0644); err != nil {
		errMsg := fmt.Sprintf("failed to write file '%s': %v", path, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("failed to write file '%s': %w", path, err)
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

// ReadFileTool implements the read_file tool
type ReadFileTool struct{}

// NewReadFileTool creates a new read_file tool instance
func NewReadFileTool() *ReadFileTool {
	return &ReadFileTool{}
}

// Name returns the tool name
func (r *ReadFileTool) Name() string {
	return "read_file"
}

// Description returns the tool description
func (r *ReadFileTool) Description() string {
	return "Read the contents of a file, optionally specifying a line range"
}

// ParametersSchema returns the JSON schema for parameters
func (r *ReadFileTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the file to read",
			},
			"start_line": map[string]interface{}{
				"type":        "integer",
				"description": "Starting line number (1-indexed, optional)",
			},
			"end_line": map[string]interface{}{
				"type":        "integer",
				"description": "Ending line number (1-indexed, inclusive, optional)",
			},
		},
		"required": []string{"path"},
	}
}

// Validate checks if parameters are valid
func (r *ReadFileTool) Validate(params map[string]interface{}, workingDir string) error {
	// Check required parameters
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	// Validate path is within working directory and exists
	resolvedPath, err := ResolvePath(workingDir, path)
	if err != nil {
		return err
	}

	// Check if file exists
	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file '%s' not found", path)
		}
		return fmt.Errorf("cannot access file '%s': %w", path, err)
	}

	// Check if it's a regular file
	if fileInfo.IsDir() {
		return fmt.Errorf("path '%s' is a directory, not a file", path)
	}

	// Validate line range if specified
	if startLine, ok := params["start_line"].(float64); ok {
		if startLine < 1 {
			return fmt.Errorf("start_line must be positive, got %.0f", startLine)
		}

		if endLine, ok := params["end_line"].(float64); ok {
			if endLine < startLine {
				return fmt.Errorf("end_line (%.0f) must be >= start_line (%.0f)", endLine, startLine)
			}
		}
	}

	return nil
}

// Execute reads the file
func (r *ReadFileTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
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

	// Read file
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read file '%s': %v", path, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("failed to read file '%s': %w", path, err)
	}

	// Get file info
	fileSize := int64(len(content))

	// Check size limit
	if fileSize > maxFileSize {
		logging.Warn("File exceeds size limit, truncating",
			"path", path,
			"size_bytes", fileSize,
			"limit_bytes", maxFileSize,
		)
		content = content[:maxFileSize]
	}

	// Handle line range if specified
	output := string(content)
	totalLines := len(strings.Split(output, "\n"))
	linesReturned := totalLines

	if startLine, ok := params["start_line"].(float64); ok {
		lines := strings.Split(output, "\n")
		start := int(startLine) - 1 // Convert to 0-indexed
		end := len(lines)

		if endLine, ok := params["end_line"].(float64); ok {
			end = int(endLine)
			if end > len(lines) {
				end = len(lines)
			}
		}

		if start >= len(lines) {
			return &types.ToolResult{
				Success:         false,
				Error:           fmt.Sprintf("line range invalid - file has %d lines, requested start line %d", len(lines), int(startLine)),
				ExecutionTimeMs: time.Since(startTime).Milliseconds(),
			}, fmt.Errorf("line range invalid")
		}

		if start < 0 {
			start = 0
		}

		lines = lines[start:end]
		output = strings.Join(lines, "\n")
		linesReturned = len(lines)

		logging.Debug("Reading file with line range",
			"path", path,
			"start_line", int(startLine),
			"end_line", end,
			"lines_returned", linesReturned,
		)
	}

	executionTime := time.Since(startTime).Milliseconds()

	logging.Info("File read",
		"path", SanitizePathForDisplay(workingDir, resolvedPath),
		"size_bytes", fileSize,
		"lines", linesReturned,
		"execution_time_ms", executionTime,
	)

	// Format output
	var outputMsg string
	if linesReturned < totalLines {
		outputMsg = fmt.Sprintf("Contents of %s (lines %d-%d):\n%s",
			path,
			int(params["start_line"].(float64)),
			int(params["start_line"].(float64))+linesReturned-1,
			output)
	} else {
		outputMsg = fmt.Sprintf("Contents of %s:\n%s", path, output)
	}

	return &types.ToolResult{
		Success:         true,
		Output:          outputMsg,
		ExecutionTimeMs: executionTime,
		FilesAffected:   []string{},
		Metadata: map[string]interface{}{
			"total_lines":     totalLines,
			"lines_returned":  linesReturned,
			"file_size_bytes": fileSize,
		},
	}, nil
}

// RequiresApproval returns true as file reading requires approval
func (r *ReadFileTool) RequiresApproval() bool {
	return true
}

// RiskLevel returns the risk level for this tool
func (r *ReadFileTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelReadOnly
}

// ReplaceStringInFileTool implements the replace_string_in_file tool
type ReplaceStringInFileTool struct{}

// NewReplaceStringInFileTool creates a new replace_string_in_file tool instance
func NewReplaceStringInFileTool() *ReplaceStringInFileTool {
	return &ReplaceStringInFileTool{}
}

// Name returns the tool name
func (t *ReplaceStringInFileTool) Name() string {
	return "replace_string_in_file"
}

// Description returns the tool description
func (t *ReplaceStringInFileTool) Description() string {
	return "Replace a specific string in a file with new content. Only replaces the first occurrence."
}

// ParametersSchema returns the JSON schema for parameters
func (t *ReplaceStringInFileTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Relative path to the file to modify",
			},
			"old_string": map[string]interface{}{
				"type":        "string",
				"description": "Exact string to find and replace",
			},
			"new_string": map[string]interface{}{
				"type":        "string",
				"description": "String to replace with",
			},
		},
		"required": []string{"path", "old_string", "new_string"},
	}
}

// Validate checks if parameters are valid
func (t *ReplaceStringInFileTool) Validate(params map[string]interface{}, workingDir string) error {
	// Check required parameters
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a non-empty string")
	}

	oldString, ok := params["old_string"].(string)
	if !ok || oldString == "" {
		return fmt.Errorf("old_string parameter is required and must be a non-empty string")
	}

	_, ok = params["new_string"].(string)
	if !ok {
		return fmt.Errorf("new_string parameter is required and must be a string")
	}

	// Validate path is within working directory and exists
	resolvedPath, err := ResolvePath(workingDir, path)
	if err != nil {
		return err
	}

	// Check if file exists
	fileInfo, err := os.Stat(resolvedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file '%s' not found", path)
		}
		return fmt.Errorf("cannot access file '%s': %w", path, err)
	}

	// Check if it's a regular file
	if fileInfo.IsDir() {
		return fmt.Errorf("path '%s' is a directory, not a file", path)
	}

	return nil
}

// Execute replaces the string in the file
func (t *ReplaceStringInFileTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	// Extract parameters
	path := params["path"].(string)
	oldString := params["old_string"].(string)
	newString := params["new_string"].(string)

	// Resolve path
	resolvedPath, err := ResolvePath(workingDir, path)
	if err != nil {
		return &types.ToolResult{
			Success:         false,
			Error:           err.Error(),
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, err
	}

	// Read file
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read file '%s': %v", path, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("failed to read file '%s': %w", path, err)
	}

	contentStr := string(content)

	// Check if old string exists
	if !strings.Contains(contentStr, oldString) {
		errMsg := fmt.Sprintf("string '%s' not found in file", oldString)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf(errMsg)
	}

	// Count occurrences
	occurrences := strings.Count(contentStr, oldString)

	// Replace only the first occurrence
	newContent := strings.Replace(contentStr, oldString, newString, 1)

	// Write file
	if err := os.WriteFile(resolvedPath, []byte(newContent), 0644); err != nil {
		errMsg := fmt.Sprintf("failed to write file '%s': %v", path, err)
		return &types.ToolResult{
			Success:         false,
			Error:           errMsg,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("failed to write file '%s': %w", path, err)
	}

	executionTime := time.Since(startTime).Milliseconds()

	// Determine which line was changed (approximate)
	lines := strings.Split(contentStr, "\n")
	changedLine := 0
	charsProcessed := 0
	for i, line := range lines {
		if strings.Contains(line, oldString) {
			changedLine = i + 1 // 1-indexed
			break
		}
		charsProcessed += len(line) + 1 // +1 for newline
		if charsProcessed >= len(contentStr) {
			break
		}
	}

	logging.Info("String replaced in file",
		"path", SanitizePathForDisplay(workingDir, resolvedPath),
		"occurrences_found", occurrences,
		"line_changed", changedLine,
		"execution_time_ms", executionTime,
	)

	// Create output message
	output := fmt.Sprintf("Replaced 1 occurrence in %s", path)
	if occurrences > 1 {
		output = fmt.Sprintf("Replaced 1 occurrence in %s (found %d total occurrences, replaced only the first at line %d)",
			path, occurrences, changedLine)
	}

	return &types.ToolResult{
		Success:         true,
		Output:          output,
		ExecutionTimeMs: executionTime,
		FilesAffected:   []string{path},
		Metadata: map[string]interface{}{
			"occurrences_found":    occurrences,
			"occurrences_replaced": 1,
			"lines_changed":        []int{changedLine},
		},
	}, nil
}

// RequiresApproval returns true as file modification requires approval
func (t *ReplaceStringInFileTool) RequiresApproval() bool {
	return true
}

// RiskLevel returns the risk level for this tool
func (t *ReplaceStringInFileTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelDangerous
}
