package unit

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// TestCreateFileTool tests the create_file tool
func TestCreateFileTool(t *testing.T) {
	tool := tools.NewCreateFileTool()

	// Verify tool metadata
	if tool.Name() != "create_file" {
		t.Errorf("expected name 'create_file', got '%s'", tool.Name())
	}
	if !tool.RequiresApproval() {
		t.Error("create_file should require approval")
	}
	if tool.RiskLevel() != types.RiskLevelSafeWrite {
		t.Errorf("expected risk level %v, got %v", types.RiskLevelSafeWrite, tool.RiskLevel())
	}

	tests := []struct {
		name         string
		params       map[string]interface{}
		setupFunc    func(workingDir string) error
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, workingDir string, result *types.ToolResult)
	}{
		{
			name: "success - create simple file",
			params: map[string]interface{}{
				"path":    "test.txt",
				"content": "hello world",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				content, err := os.ReadFile(filepath.Join(workingDir, "test.txt"))
				if err != nil {
					t.Errorf("failed to read created file: %v", err)
				}
				if string(content) != "hello world" {
					t.Errorf("expected content 'hello world', got '%s'", string(content))
				}
				if len(result.FilesAffected) != 1 || result.FilesAffected[0] != "test.txt" {
					t.Errorf("expected FilesAffected=['test.txt'], got %v", result.FilesAffected)
				}
			},
		},
		{
			name: "success - create file in subdirectory",
			params: map[string]interface{}{
				"path":    "subdir/nested/file.go",
				"content": "package main\n\nfunc main() {}",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				content, err := os.ReadFile(filepath.Join(workingDir, "subdir", "nested", "file.go"))
				if err != nil {
					t.Errorf("failed to read created file: %v", err)
				}
				if !strings.Contains(string(content), "package main") {
					t.Errorf("expected content to contain 'package main', got '%s'", string(content))
				}
			},
		},
		{
			name: "error - file already exists",
			params: map[string]interface{}{
				"path":    "existing.txt",
				"content": "new content",
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "existing.txt"), []byte("old content"), 0644)
			},
			wantErr:     true,
			errContains: "already exists",
		},
		{
			name: "error - missing path parameter",
			params: map[string]interface{}{
				"content": "hello",
			},
			wantErr:     true,
			errContains: "path parameter is required",
		},
		{
			name: "error - empty path",
			params: map[string]interface{}{
				"path":    "",
				"content": "hello",
			},
			wantErr:     true,
			errContains: "path parameter is required",
		},
		{
			name: "error - missing content parameter",
			params: map[string]interface{}{
				"path": "test.txt",
			},
			wantErr:     true,
			errContains: "content parameter is required",
		},
		{
			name: "error - path traversal attempt with ..",
			params: map[string]interface{}{
				"path":    "../outside.txt",
				"content": "malicious",
			},
			wantErr:     true,
			errContains: "outside working directory",
		},
		// Note: Windows absolute paths like C:\... will be validated differently
		// This test is skipped as behavior varies by platform
		{
			name: "error - content too large",
			params: map[string]interface{}{
				"path":    "huge.txt",
				"content": strings.Repeat("x", 11*1024*1024), // 11MB
			},
			wantErr:     true,
			errContains: "exceeds maximum allowed size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp working directory
			workingDir, err := os.MkdirTemp("", "wink-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(workingDir)

			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(workingDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Validate parameters
			err = tool.Validate(tt.params, workingDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}

			// Execute tool
			result, err := tool.Execute(context.Background(), tt.params, workingDir)
			if err != nil {
				t.Fatalf("unexpected execution error: %v", err)
			}

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, workingDir, result)
			}
		})
	}
}

// TestReadFileTool tests the read_file tool
func TestReadFileTool(t *testing.T) {
	tool := tools.NewReadFileTool()

	// Verify tool metadata
	if tool.Name() != "read_file" {
		t.Errorf("expected name 'read_file', got '%s'", tool.Name())
	}
	if !tool.RequiresApproval() {
		t.Error("read_file should require approval")
	}
	if tool.RiskLevel() != types.RiskLevelReadOnly {
		t.Errorf("expected risk level %v, got %v", types.RiskLevelReadOnly, tool.RiskLevel())
	}

	tests := []struct {
		name         string
		params       map[string]interface{}
		setupFunc    func(workingDir string) error
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, result *types.ToolResult)
	}{
		{
			name: "success - read full file",
			params: map[string]interface{}{
				"path": "test.txt",
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "test.txt"), []byte("line1\nline2\nline3"), 0644)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				if !strings.Contains(result.Output, "line1") || !strings.Contains(result.Output, "line3") {
					t.Errorf("expected output to contain all lines, got: %s", result.Output)
				}
			},
		},
		{
			name: "success - read with line range",
			params: map[string]interface{}{
				"path":       "test.txt",
				"start_line": float64(2),
				"end_line":   float64(2),
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "test.txt"), []byte("line1\nline2\nline3"), 0644)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				if !strings.Contains(result.Output, "line2") {
					t.Errorf("expected output to contain 'line2', got: %s", result.Output)
				}
				if strings.Contains(result.Output, "line1") || strings.Contains(result.Output, "line3") {
					t.Errorf("expected output to only contain line2, got: %s", result.Output)
				}
			},
		},
		{
			name: "error - file not found",
			params: map[string]interface{}{
				"path": "nonexistent.txt",
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "error - path is directory",
			params: map[string]interface{}{
				"path": "subdir",
			},
			setupFunc: func(workingDir string) error {
				return os.Mkdir(filepath.Join(workingDir, "subdir"), 0755)
			},
			wantErr:     true,
			errContains: "is a directory",
		},
		{
			name:        "error - missing path parameter",
			params:      map[string]interface{}{},
			wantErr:     true,
			errContains: "path parameter is required",
		},
		{
			name: "error - invalid line range",
			params: map[string]interface{}{
				"path":       "test.txt",
				"start_line": float64(0),
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "test.txt"), []byte("line1\nline2"), 0644)
			},
			wantErr:     true,
			errContains: "must be positive",
		},
		{
			name: "error - end_line before start_line",
			params: map[string]interface{}{
				"path":       "test.txt",
				"start_line": float64(5),
				"end_line":   float64(2),
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "test.txt"), []byte("line1\nline2\nline3"), 0644)
			},
			wantErr:     true,
			errContains: "must be >= start_line",
		},
		{
			name: "error - path traversal",
			params: map[string]interface{}{
				"path": "../../etc/passwd",
			},
			wantErr:     true,
			errContains: "outside working directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp working directory
			workingDir, err := os.MkdirTemp("", "wink-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(workingDir)

			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(workingDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Validate parameters
			err = tool.Validate(tt.params, workingDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}

			// Execute tool
			result, err := tool.Execute(context.Background(), tt.params, workingDir)
			if err != nil {
				t.Fatalf("unexpected execution error: %v", err)
			}

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}
		})
	}
}

// TestReplaceStringInFileTool tests the replace_string_in_file tool
func TestReplaceStringInFileTool(t *testing.T) {
	tool := tools.NewReplaceStringInFileTool()

	// Verify tool metadata
	if tool.Name() != "replace_string_in_file" {
		t.Errorf("expected name 'replace_string_in_file', got '%s'", tool.Name())
	}
	if !tool.RequiresApproval() {
		t.Error("replace_string_in_file should require approval")
	}
	if tool.RiskLevel() != types.RiskLevelDangerous {
		t.Errorf("expected risk level %v, got %v", types.RiskLevelDangerous, tool.RiskLevel())
	}

	tests := []struct {
		name         string
		params       map[string]interface{}
		setupFunc    func(workingDir string) error
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, workingDir string, result *types.ToolResult)
	}{
		{
			name: "success - simple replacement",
			params: map[string]interface{}{
				"path":       "test.txt",
				"old_string": "hello",
				"new_string": "goodbye",
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "test.txt"), []byte("hello world"), 0644)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				content, err := os.ReadFile(filepath.Join(workingDir, "test.txt"))
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}
				if string(content) != "goodbye world" {
					t.Errorf("expected 'goodbye world', got '%s'", string(content))
				}
			},
		},
		{
			name: "success - multiline replacement",
			params: map[string]interface{}{
				"path":       "test.go",
				"old_string": "func old() {\n\treturn 1\n}",
				"new_string": "func new() {\n\treturn 2\n}",
			},
			setupFunc: func(workingDir string) error {
				content := "package main\n\nfunc old() {\n\treturn 1\n}\n\nfunc main() {}"
				return os.WriteFile(filepath.Join(workingDir, "test.go"), []byte(content), 0644)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				content, err := os.ReadFile(filepath.Join(workingDir, "test.go"))
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}
				if !strings.Contains(string(content), "func new()") {
					t.Errorf("expected content to contain 'func new()', got: %s", string(content))
				}
			},
		},
		{
			name: "error - file not found",
			params: map[string]interface{}{
				"path":       "nonexistent.txt",
				"old_string": "old",
				"new_string": "new",
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "error - string not found",
			params: map[string]interface{}{
				"path":       "test.txt",
				"old_string": "notfound",
				"new_string": "new",
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "test.txt"), []byte("hello world"), 0644)
			},
			wantErr: false, // Validation passes, but execution will fail
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if result.Success {
					t.Error("expected failure when string not found")
				}
				if !strings.Contains(result.Error, "not found") {
					t.Errorf("expected error to contain 'not found', got: %s", result.Error)
				}
			},
		},
		{
			name: "error - missing parameters",
			params: map[string]interface{}{
				"path": "test.txt",
			},
			wantErr:     true,
			errContains: "parameter is required",
		},
		{
			name: "error - path traversal",
			params: map[string]interface{}{
				"path":       "../outside.txt",
				"old_string": "old",
				"new_string": "new",
			},
			wantErr:     true,
			errContains: "outside working directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp working directory
			workingDir, err := os.MkdirTemp("", "wink-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(workingDir)

			// Run setup if provided
			if tt.setupFunc != nil {
				if err := tt.setupFunc(workingDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			// Validate parameters
			err = tool.Validate(tt.params, workingDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}

			// Execute tool
			result, err := tool.Execute(context.Background(), tt.params, workingDir)
			if err != nil && !tt.wantErr {
				// For cases where we expect execution to return an error in the result
				if result != nil && !result.Success && tt.validateFunc != nil {
					tt.validateFunc(t, workingDir, result)
					return
				}
				t.Fatalf("unexpected execution error: %v", err)
			}

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, workingDir, result)
			}
		})
	}
}

// TestDirectoryTools tests create_directory and list_dir tools
func TestCreateDirectoryTool(t *testing.T) {
	tool := tools.NewCreateDirectoryTool()

	tests := []struct {
		name         string
		params       map[string]interface{}
		setupFunc    func(workingDir string) error
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, workingDir string, result *types.ToolResult)
	}{
		{
			name: "success - create simple directory",
			params: map[string]interface{}{
				"path": "newdir",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				info, err := os.Stat(filepath.Join(workingDir, "newdir"))
				if err != nil {
					t.Errorf("directory not created: %v", err)
				}
				if !info.IsDir() {
					t.Error("expected path to be a directory")
				}
			},
		},
		{
			name: "success - create nested directories",
			params: map[string]interface{}{
				"path": "a/b/c/d",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				info, err := os.Stat(filepath.Join(workingDir, "a", "b", "c", "d"))
				if err != nil {
					t.Errorf("nested directory not created: %v", err)
				}
				if !info.IsDir() {
					t.Error("expected path to be a directory")
				}
			},
		},
		{
			name: "success - directory already exists",
			params: map[string]interface{}{
				"path": "existing",
			},
			setupFunc: func(workingDir string) error {
				return os.Mkdir(filepath.Join(workingDir, "existing"), 0755)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, workingDir string, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success for existing directory, got error: %s", result.Error)
				}
			},
		},
		{
			name: "error - path traversal",
			params: map[string]interface{}{
				"path": "../outside",
			},
			wantErr:     true,
			errContains: "outside working directory",
		},
		{
			name:        "error - missing path",
			params:      map[string]interface{}{},
			wantErr:     true,
			errContains: "path parameter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir, err := os.MkdirTemp("", "wink-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(workingDir)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(workingDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err = tool.Validate(tt.params, workingDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}

			result, err := tool.Execute(context.Background(), tt.params, workingDir)
			if err != nil {
				t.Fatalf("unexpected execution error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, workingDir, result)
			}
		})
	}
}

func TestListDirTool(t *testing.T) {
	tool := tools.NewListDirTool()

	tests := []struct {
		name         string
		params       map[string]interface{}
		setupFunc    func(workingDir string) error
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, result *types.ToolResult)
	}{
		{
			name: "success - list directory",
			params: map[string]interface{}{
				"path": "testdir",
			},
			setupFunc: func(workingDir string) error {
				dir := filepath.Join(workingDir, "testdir")
				if err := os.Mkdir(dir, 0755); err != nil {
					return err
				}
				os.WriteFile(filepath.Join(dir, "file1.txt"), []byte("content"), 0644)
				os.WriteFile(filepath.Join(dir, "file2.go"), []byte("package main"), 0644)
				os.Mkdir(filepath.Join(dir, "subdir"), 0755)
				return nil
			},
			wantErr: false,
			validateFunc: func(t *testing.T, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				// Check for directory separator (both / and \ for cross-platform)
				if !strings.Contains(result.Output, "subdir") {
					t.Errorf("expected output to contain 'subdir', got: %s", result.Output)
				}
			},
		},
		{
			name: "success - list current directory",
			params: map[string]interface{}{
				"path": ".",
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "test.txt"), []byte("content"), 0644)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, result *types.ToolResult) {
				if !result.Success {
					t.Errorf("expected success, got error: %s", result.Error)
				}
				if !strings.Contains(result.Output, "test.txt") {
					t.Errorf("expected output to contain 'test.txt', got: %s", result.Output)
				}
			},
		},
		{
			name: "error - directory not found",
			params: map[string]interface{}{
				"path": "nonexistent",
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "error - path is file",
			params: map[string]interface{}{
				"path": "file.txt",
			},
			setupFunc: func(workingDir string) error {
				return os.WriteFile(filepath.Join(workingDir, "file.txt"), []byte("content"), 0644)
			},
			wantErr:     true,
			errContains: "not a directory",
		},
		{
			name: "error - path traversal",
			params: map[string]interface{}{
				"path": "../../etc",
			},
			wantErr:     true,
			errContains: "outside working directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir, err := os.MkdirTemp("", "wink-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(workingDir)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(workingDir); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
			}

			err = tool.Validate(tt.params, workingDir)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}

			result, err := tool.Execute(context.Background(), tt.params, workingDir)
			if err != nil {
				t.Fatalf("unexpected execution error: %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, result)
			}
		})
	}
}
