package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
)

// TestFileOperationsWorkflow tests multi-operation file workflows
func TestFileOperationsWorkflow(t *testing.T) {
	// Create temp working directory
	workingDir, err := os.MkdirTemp("", "wink-integration-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(workingDir)

	// Create tool registry
	registry := tools.NewRegistry()

	// Register file tools
	if err := registry.Register(tools.NewCreateFileTool()); err != nil {
		t.Fatalf("failed to register create_file: %v", err)
	}
	if err := registry.Register(tools.NewReadFileTool()); err != nil {
		t.Fatalf("failed to register read_file: %v", err)
	}
	if err := registry.Register(tools.NewReplaceStringInFileTool()); err != nil {
		t.Fatalf("failed to register replace_string_in_file: %v", err)
	}
	if err := registry.Register(tools.NewCreateDirectoryTool()); err != nil {
		t.Fatalf("failed to register create_directory: %v", err)
	}
	if err := registry.Register(tools.NewListDirTool()); err != nil {
		t.Fatalf("failed to register list_dir: %v", err)
	}

	tests := []struct {
		name       string
		operations []struct {
			toolName string
			params   map[string]interface{}
			approve  bool // simulate user approval
		}
		validateFunc func(t *testing.T, workingDir string)
	}{
		{
			name: "create and read file workflow",
			operations: []struct {
				toolName string
				params   map[string]interface{}
				approve  bool
			}{
				{
					toolName: "create_file",
					params: map[string]interface{}{
						"path":    "test.txt",
						"content": "hello world",
					},
					approve: true,
				},
				{
					toolName: "read_file",
					params: map[string]interface{}{
						"path": "test.txt",
					},
					approve: true,
				},
			},
			validateFunc: func(t *testing.T, workingDir string) {
				content, err := os.ReadFile(filepath.Join(workingDir, "test.txt"))
				if err != nil {
					t.Errorf("failed to read created file: %v", err)
				}
				if string(content) != "hello world" {
					t.Errorf("expected 'hello world', got '%s'", string(content))
				}
			},
		},
		{
			name: "create directory and create file in it",
			operations: []struct {
				toolName string
				params   map[string]interface{}
				approve  bool
			}{
				{
					toolName: "create_directory",
					params: map[string]interface{}{
						"path": "subdir",
					},
					approve: true,
				},
				{
					toolName: "create_file",
					params: map[string]interface{}{
						"path":    "subdir/file.txt",
						"content": "nested content",
					},
					approve: true,
				},
				{
					toolName: "list_dir",
					params: map[string]interface{}{
						"path": "subdir",
					},
					approve: true,
				},
			},
			validateFunc: func(t *testing.T, workingDir string) {
				content, err := os.ReadFile(filepath.Join(workingDir, "subdir", "file.txt"))
				if err != nil {
					t.Errorf("failed to read nested file: %v", err)
				}
				if string(content) != "nested content" {
					t.Errorf("expected 'nested content', got '%s'", string(content))
				}
			},
		},
		{
			name: "create, read, and modify file",
			operations: []struct {
				toolName string
				params   map[string]interface{}
				approve  bool
			}{
				{
					toolName: "create_file",
					params: map[string]interface{}{
						"path":    "script.py",
						"content": "def old_function():\n    return 1",
					},
					approve: true,
				},
				{
					toolName: "replace_string_in_file",
					params: map[string]interface{}{
						"path":       "script.py",
						"old_string": "old_function",
						"new_string": "new_function",
					},
					approve: true,
				},
				{
					toolName: "read_file",
					params: map[string]interface{}{
						"path": "script.py",
					},
					approve: true,
				},
			},
			validateFunc: func(t *testing.T, workingDir string) {
				content, err := os.ReadFile(filepath.Join(workingDir, "script.py"))
				if err != nil {
					t.Errorf("failed to read modified file: %v", err)
				}
				if !strings.Contains(string(content), "new_function") {
					t.Errorf("expected content to contain 'new_function', got: %s", string(content))
				}
				if strings.Contains(string(content), "old_function") {
					t.Errorf("expected 'old_function' to be replaced, but still found it in: %s", string(content))
				}
			},
		},
		{
			name: "create multiple files and list directory",
			operations: []struct {
				toolName string
				params   map[string]interface{}
				approve  bool
			}{
				{
					toolName: "create_directory",
					params: map[string]interface{}{
						"path": "project",
					},
					approve: true,
				},
				{
					toolName: "create_file",
					params: map[string]interface{}{
						"path":    "project/main.go",
						"content": "package main\n\nfunc main() {}",
					},
					approve: true,
				},
				{
					toolName: "create_file",
					params: map[string]interface{}{
						"path":    "project/README.md",
						"content": "# Project",
					},
					approve: true,
				},
				{
					toolName: "list_dir",
					params: map[string]interface{}{
						"path": "project",
					},
					approve: true,
				},
			},
			validateFunc: func(t *testing.T, workingDir string) {
				// Check both files exist
				if _, err := os.Stat(filepath.Join(workingDir, "project", "main.go")); err != nil {
					t.Errorf("main.go not found: %v", err)
				}
				if _, err := os.Stat(filepath.Join(workingDir, "project", "README.md")); err != nil {
					t.Errorf("README.md not found: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean working directory between test cases
			entries, _ := os.ReadDir(workingDir)
			for _, entry := range entries {
				os.RemoveAll(filepath.Join(workingDir, entry.Name()))
			}

			// Execute each operation in sequence
			for i, op := range tt.operations {
				tool, err := registry.Get(op.toolName)
				if err != nil {
					t.Fatalf("operation %d: failed to get tool '%s': %v", i, op.toolName, err)
				}

				// Validate
				if err := tool.Validate(op.params, workingDir); err != nil {
					t.Fatalf("operation %d: validation failed for %s: %v", i, op.toolName, err)
				}

				// For tools requiring approval, test the approval check
				if tool.RequiresApproval() {
					if !op.approve {
						t.Logf("operation %d: %s would require approval (simulated rejection)", i, op.toolName)
						continue
					}
					t.Logf("operation %d: %s approved (simulated)", i, op.toolName)
				}

				// Execute
				result, err := tool.Execute(context.Background(), op.params, workingDir)
				if err != nil {
					t.Fatalf("operation %d: execution failed for %s: %v", i, op.toolName, err)
				}

				if !result.Success {
					t.Fatalf("operation %d: %s failed: %s", i, op.toolName, result.Error)
				}

				t.Logf("operation %d: %s succeeded - %s", i, op.toolName, result.Output)
			}

			// Validate final state
			if tt.validateFunc != nil {
				tt.validateFunc(t, workingDir)
			}
		})
	}
}
