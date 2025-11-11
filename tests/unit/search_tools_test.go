package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// TestFileSearchTool tests file_search with glob patterns
func TestFileSearchTool(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file structure
	testFiles := map[string]string{
		"main.go":           "package main",
		"util.go":           "package util",
		"src/app.go":        "package app",
		"src/test.go":       "package test",
		"docs/README.md":    "# README",
		"docs/GUIDE.md":     "# Guide",
		"test/unit_test.go": "package test",
		"config.json":       "{}",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tool := tools.NewFileSearchTool()
	ctx := context.Background()

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectError bool
		expectFiles []string // Files that should be found
		minMatches  int
	}{
		{
			name: "Find all Go files",
			params: map[string]interface{}{
				"pattern": "*.go",
			},
			expectError: false,
			expectFiles: []string{"main.go", "util.go"},
			minMatches:  2,
		},
		{
			name: "Find Go files recursively",
			params: map[string]interface{}{
				"pattern": "**/*.go",
			},
			expectError: false,
			minMatches:  3, // Only subdirectory .go files (src/*, test/*)
		},
		{
			name: "Find markdown files in docs",
			params: map[string]interface{}{
				"pattern": "docs/*.md",
			},
			expectError: false,
			minMatches:  2,
		},
		{
			name: "Find all test files",
			params: map[string]interface{}{
				"pattern": "**/*test*.go",
			},
			expectError: false,
			minMatches:  2,
		},
		{
			name: "Find with base_path",
			params: map[string]interface{}{
				"pattern":   "*.go",
				"base_path": "src",
			},
			expectError: false,
			minMatches:  2,
		},
		{
			name: "No matches for non-existent pattern",
			params: map[string]interface{}{
				"pattern": "*.xyz",
			},
			expectError: false,
			minMatches:  0,
		},
		{
			name: "Invalid pattern",
			params: map[string]interface{}{
				"pattern": "**/*[.go",
			},
			expectError: true,
		},
		{
			name: "Empty pattern",
			params: map[string]interface{}{
				"pattern": "",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate first
			err := tool.Validate(tt.params, tmpDir)
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Validation failed: %v", err)
			}

			// Execute
			result, err := tool.Execute(ctx, tt.params, tmpDir)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if !result.Success {
				t.Errorf("Expected success but got failure: %s", result.Output)
			}

			// Check metadata
			matches, ok := result.Metadata["matches"].(int)
			if !ok {
				t.Error("Missing or invalid 'matches' metadata")
			}

			if matches < tt.minMatches {
				t.Errorf("Expected at least %d matches, got %d. Output: %s",
					tt.minMatches, matches, result.Output)
			}

			// Check specific files if provided
			for _, expectedFile := range tt.expectFiles {
				if !containsFile(result.Output, expectedFile) {
					t.Errorf("Expected to find file '%s' in output: %s",
						expectedFile, result.Output)
				}
			}
		})
	}
}

// TestGrepSearchTool tests grep_search with text and regex patterns
func TestGrepSearchTool(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files with known content
	testFiles := map[string]string{
		"app.go":       "package main\n// TODO: implement feature\nfunc main() {}",
		"util.go":      "package util\n// TODO: refactor this\nfunc helper() {}",
		"README.md":    "# Project\nTODO: write docs",
		"config.json":  `{"name": "test", "version": "1.0"}`,
		"test_file.go": "package test\n// FIXME: broken test\nfunc Test() {}",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	tool := tools.NewGrepSearchTool()
	ctx := context.Background()

	tests := []struct {
		name           string
		params         map[string]interface{}
		expectError    bool
		minMatches     int
		expectInOutput string
	}{
		{
			name: "Text search for TODO",
			params: map[string]interface{}{
				"pattern": "TODO",
			},
			expectError:    false,
			minMatches:     3,
			expectInOutput: "TODO:",
		},
		{
			name: "Regex search for TODO or FIXME",
			params: map[string]interface{}{
				"pattern":  "TODO|FIXME",
				"is_regex": true,
			},
			expectError: false,
			minMatches:  4,
		},
		{
			name: "Search with file pattern",
			params: map[string]interface{}{
				"pattern":      "TODO",
				"file_pattern": "*.go",
			},
			expectError: false,
			minMatches:  2, // Only in .go files
		},
		{
			name: "Search with max_results limit",
			params: map[string]interface{}{
				"pattern":     "TODO",
				"max_results": 2.0,
			},
			expectError: false,
			minMatches:  2,
		},
		{
			name: "Regex for function definitions",
			params: map[string]interface{}{
				"pattern":  "func [a-zA-Z]+\\(",
				"is_regex": true,
			},
			expectError: false,
			minMatches:  3,
		},
		{
			name: "No matches",
			params: map[string]interface{}{
				"pattern": "NONEXISTENT_STRING",
			},
			expectError: false,
			minMatches:  0,
		},
		{
			name: "Invalid regex",
			params: map[string]interface{}{
				"pattern":  "([unclosed",
				"is_regex": true,
			},
			expectError: true,
		},
		{
			name: "Empty pattern",
			params: map[string]interface{}{
				"pattern": "",
			},
			expectError: true,
		},
		{
			name: "Invalid max_results (too high)",
			params: map[string]interface{}{
				"pattern":     "TODO",
				"max_results": 2000.0,
			},
			expectError: true,
		},
		{
			name: "Invalid max_results (too low)",
			params: map[string]interface{}{
				"pattern":     "TODO",
				"max_results": 0.0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate first
			err := tool.Validate(tt.params, tmpDir)
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Validation failed: %v", err)
			}

			// Execute
			result, err := tool.Execute(ctx, tt.params, tmpDir)
			if err != nil {
				t.Fatalf("Execution failed: %v", err)
			}

			if !result.Success {
				t.Errorf("Expected success but got failure: %s", result.Output)
			}

			// Check metadata
			matches, ok := result.Metadata["total_matches"].(int)
			if !ok {
				t.Error("Missing or invalid 'total_matches' metadata")
			}

			if matches < tt.minMatches {
				t.Errorf("Expected at least %d matches, got %d. Output: %s",
					tt.minMatches, matches, result.Output)
			}

			// Check output content if specified
			if tt.expectInOutput != "" && matches > 0 {
				if !containsFile(result.Output, tt.expectInOutput) {
					t.Errorf("Expected to find '%s' in output: %s",
						tt.expectInOutput, result.Output)
				}
			}
		})
	}
}

// TestSearchToolsRiskLevel verifies risk levels
func TestSearchToolsRiskLevel(t *testing.T) {
	tests := []struct {
		name     string
		tool     types.Tool
		expected types.RiskLevel
	}{
		{
			name:     "file_search is read_only",
			tool:     tools.NewFileSearchTool(),
			expected: types.RiskLevelReadOnly,
		},
		{
			name:     "grep_search is read_only",
			tool:     tools.NewGrepSearchTool(),
			expected: types.RiskLevelReadOnly,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tool.RiskLevel() != tt.expected {
				t.Errorf("Expected risk level %s, got %s",
					tt.expected, tt.tool.RiskLevel())
			}
		})
	}
}

// TestSearchToolsRequireApproval verifies approval requirements
func TestSearchToolsRequireApproval(t *testing.T) {
	tests := []struct {
		name     string
		tool     types.Tool
		expected bool
	}{
		{
			name:     "file_search requires approval",
			tool:     tools.NewFileSearchTool(),
			expected: true,
		},
		{
			name:     "grep_search requires approval",
			tool:     tools.NewGrepSearchTool(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.tool.RequiresApproval() != tt.expected {
				t.Errorf("Expected RequiresApproval=%v, got %v",
					tt.expected, tt.tool.RequiresApproval())
			}
		})
	}
}

// Helper function to check if output contains a string
func containsFile(output, needle string) bool {
	return len(output) > 0 && (output == needle ||
		len(needle) > 0 && len(output) >= len(needle) &&
			containsSubstring(output, needle))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
