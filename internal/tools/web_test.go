package tools_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// TestFetchWebpageTool tests web page fetching
func TestFetchWebpageTool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := tools.NewFetchWebpageTool()
	ctx := context.Background()

	tests := []struct {
		name          string
		params        map[string]interface{}
		expectError   bool
		skipExecution bool // Skip actual execution (for network tests)
	}{
		{
			name: "Valid HTTP URL",
			params: map[string]interface{}{
				"url": "http://example.com",
			},
			expectError:   false,
			skipExecution: true, // Would need network
		},
		{
			name: "Valid HTTPS URL",
			params: map[string]interface{}{
				"url": "https://example.com",
			},
			expectError:   false,
			skipExecution: true,
		},
		{
			name: "URL with custom timeout",
			params: map[string]interface{}{
				"url":             "https://example.com",
				"timeout_seconds": 5.0,
			},
			expectError:   false,
			skipExecution: true,
		},
		{
			name: "Invalid scheme (ftp)",
			params: map[string]interface{}{
				"url": "ftp://example.com",
			},
			expectError: true,
		},
		{
			name: "Invalid scheme (file)",
			params: map[string]interface{}{
				"url": "file:///etc/passwd",
			},
			expectError: true,
		},
		{
			name: "Empty URL",
			params: map[string]interface{}{
				"url": "",
			},
			expectError: true,
		},
		{
			name: "Missing URL",
			params: map[string]interface{}{
				"timeout_seconds": 10.0,
			},
			expectError: true,
		},
		{
			name: "Invalid timeout (too low)",
			params: map[string]interface{}{
				"url":             "https://example.com",
				"timeout_seconds": 0.0,
			},
			expectError: true,
		},
		{
			name: "Invalid timeout (too high)",
			params: map[string]interface{}{
				"url":             "https://example.com",
				"timeout_seconds": 100.0,
			},
			expectError: true,
		},
		{
			name: "Malformed URL",
			params: map[string]interface{}{
				"url": "not a url",
			},
			expectError: true,
		},
		{
			name: "URL without host",
			params: map[string]interface{}{
				"url": "http://",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate
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

			// Skip execution for network tests
			if tt.skipExecution {
				t.Skip("Skipping network execution in unit tests")
			}

			// Execute (only for non-network tests)
			result, err := tool.Execute(ctx, tt.params, tmpDir)
			if err != nil && !tt.expectError {
				t.Errorf("Unexpected execution error: %v", err)
			}

			// Basic result checks
			if result != nil {
				if result.Metadata == nil {
					t.Error("Expected metadata in result")
				}
			}
		})
	}
}

// TestFetchWebpageToolRiskLevel verifies risk level
func TestFetchWebpageToolRiskLevel(t *testing.T) {
	tool := tools.NewFetchWebpageTool()

	if tool.RiskLevel() != types.RiskLevelDangerous {
		t.Errorf("Expected risk level dangerous, got %s", tool.RiskLevel())
	}

	if !tool.RequiresApproval() {
		t.Error("fetch_webpage should require approval")
	}
}

// TestFetchWebpageURLParsing tests URL validation edge cases
func TestFetchWebpageURLParsing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := tools.NewFetchWebpageTool()

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "HTTP with port",
			url:         "http://example.com:8080",
			expectError: false,
		},
		{
			name:        "HTTPS with path",
			url:         "https://example.com/path/to/resource",
			expectError: false,
		},
		{
			name:        "HTTPS with query params",
			url:         "https://example.com/api?key=value&foo=bar",
			expectError: false,
		},
		{
			name:        "HTTPS with fragment",
			url:         "https://example.com/page#section",
			expectError: false,
		},
		{
			name:        "URL with credentials (not recommended but valid)",
			url:         "https://user:pass@example.com",
			expectError: false,
		},
		{
			name:        "Invalid: relative URL",
			url:         "/path/to/resource",
			expectError: true,
		},
		{
			name:        "Invalid: no scheme",
			url:         "example.com",
			expectError: true,
		},
		{
			name:        "Invalid: javascript protocol",
			url:         "javascript:alert(1)",
			expectError: true,
		},
		{
			name:        "Invalid: data URL",
			url:         "data:text/html,<h1>Test</h1>",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tool.Validate(map[string]interface{}{
				"url": tt.url,
			}, tmpDir)

			if tt.expectError && err == nil {
				t.Errorf("Expected validation error for URL: %s", tt.url)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected validation error for URL %s: %v", tt.url, err)
			}
		})
	}
}

// TestFetchWebpageParameters tests parameter schema
func TestFetchWebpageParameters(t *testing.T) {
	tool := tools.NewFetchWebpageTool()

	schema := tool.ParametersSchema()
	if schema == nil {
		t.Fatal("Expected parameters schema")
	}

	// Check required fields
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties in schema")
	}

	// Check url property
	if _, ok := props["url"]; !ok {
		t.Error("Expected url property in schema")
	}

	// Check timeout_seconds property
	if _, ok := props["timeout_seconds"]; !ok {
		t.Error("Expected timeout_seconds property in schema")
	}

	// Check required array
	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required array in schema")
	}

	foundURL := false
	for _, r := range required {
		if r == "url" {
			foundURL = true
		}
	}
	if !foundURL {
		t.Error("Expected url to be required")
	}
}

// TestFetchWebpageDescription tests tool description
func TestFetchWebpageDescription(t *testing.T) {
	tool := tools.NewFetchWebpageTool()

	name := tool.Name()
	if name != "fetch_webpage" {
		t.Errorf("Expected name 'fetch_webpage', got '%s'", name)
	}

	desc := tool.Description()
	if desc == "" {
		t.Error("Expected non-empty description")
	}

	// Description should mention http/https
	if !strings.Contains(strings.ToLower(desc), "http") {
		t.Error("Description should mention http/https")
	}
}
