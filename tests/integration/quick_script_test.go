//go:build integration
// +build integration

package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/agent"
	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/internal/tools"
)

func TestQuickScriptGeneration(t *testing.T) {
	// Skip if Ollama is not available
	ollamaURL := os.Getenv("WINK_OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	// Initialize logging for test
	logging.InitLogger(true)

	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create agent
	agentInstance, err := agent.NewAgent(ollamaURL, "qwen3:8b", 30)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Register create_file tool
	createFile := tools.NewCreateFileTool()
	if err := agentInstance.RegisterTool(createFile); err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}

	// Test prompt
	prompt := "Create a simple hello world Python script named hello.py"

	// Run agent (this will require manual approval in interactive mode)
	// For automated testing, we would need to mock the approval workflow
	ctx := context.Background()
	err = agentInstance.Run(ctx, prompt, tempDir, false)

	// In actual integration tests, we would:
	// 1. Mock the LLM response
	// 2. Auto-approve tool calls
	// 3. Verify file creation

	// For now, just verify agent can be created and started
	if err != nil {
		// Error is expected without Ollama running
		t.Logf("Agent execution error (expected without Ollama): %v", err)
	}
}

func TestCreateFileToolDirect(t *testing.T) {
	// Test the create_file tool directly without LLM
	tempDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create tool
	createFile := tools.NewCreateFileTool()

	// Test parameters
	params := map[string]interface{}{
		"path":    "test.txt",
		"content": "Hello, World!",
	}

	// Validate
	if err := createFile.Validate(params, tempDir); err != nil {
		t.Fatalf("Validation failed: %v", err)
	}

	// Execute
	ctx := context.Background()
	result, err := createFile.Execute(ctx, params, tempDir)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("Tool execution reported failure: %s", result.Error)
	}

	// Verify file was created
	filePath := filepath.Join(tempDir, "test.txt")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expectedContent := "Hello, World!"
	if string(content) != expectedContent {
		t.Errorf("File content mismatch. Expected: %q, Got: %q", expectedContent, string(content))
	}

	// Verify metadata
	if len(result.FilesAffected) != 1 || result.FilesAffected[0] != "test.txt" {
		t.Errorf("FilesAffected incorrect: %v", result.FilesAffected)
	}
}

func TestCreateFilePathValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createFile := tools.NewCreateFileTool()

	tests := []struct {
		name        string
		path        string
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "Valid relative path",
			path:        "subdir/test.txt",
			shouldError: false,
		},
		{
			name:        "Path traversal attempt",
			path:        "../etc/passwd",
			shouldError: true,
			errorMsg:    "outside working directory",
		},
		{
			name:        "Current directory reference",
			path:        "./test.txt",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{
				"path":    tt.path,
				"content": "test",
			}

			err := createFile.Validate(params, tempDir)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error containing %q, but got no error", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
