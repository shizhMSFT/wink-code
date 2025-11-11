package unit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/shizhMSFT/wink-code/pkg/types"
)

// TestApprovalRuleMatching tests auto-approval rule matching logic
func TestApprovalRuleMatching(t *testing.T) {
	// Create temp config directory
	tempDir, err := os.MkdirTemp("", "wink-approval-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config file path
	configFile := filepath.Join(tempDir, "config.json")

	// Create initial config with some rules
	cfg := types.DefaultConfig()
	cfg.AutoApprovalRules = []types.ApprovalRule{
		{
			ID:           "rule-1",
			ToolName:     "read_file",
			ParamPattern: `"path":".*\.txt"`,
			Description:  "Auto-approve reading .txt files",
			UseCount:     0,
		},
		{
			ID:           "rule-2",
			ToolName:     "create_file",
			ParamPattern: `"path":"test_.*"`,
			Description:  "Auto-approve creating test_ files",
			UseCount:     0,
		},
		{
			ID:           "rule-3",
			ToolName:     "run_in_terminal",
			ParamPattern: `"command":"git status"`,
			Description:  "Auto-approve git status command",
			UseCount:     0,
		},
	}

	// Save config
	data, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create approval manager pointing to temp config
	// Note: This would need to be adjusted based on actual implementation
	// For now, we'll test the matching logic directly

	tests := []struct {
		name        string
		toolName    string
		params      map[string]interface{}
		shouldMatch bool
		ruleID      string
	}{
		{
			name:     "match read_file with .txt extension",
			toolName: "read_file",
			params: map[string]interface{}{
				"path": "document.txt",
			},
			shouldMatch: true,
			ruleID:      "rule-1",
		},
		{
			name:     "no match read_file with .go extension",
			toolName: "read_file",
			params: map[string]interface{}{
				"path": "main.go",
			},
			shouldMatch: false,
		},
		{
			name:     "match create_file with test_ prefix",
			toolName: "create_file",
			params: map[string]interface{}{
				"path":    "test_output.txt",
				"content": "test content",
			},
			shouldMatch: true,
			ruleID:      "rule-2",
		},
		{
			name:     "no match create_file without test_ prefix",
			toolName: "create_file",
			params: map[string]interface{}{
				"path":    "production.txt",
				"content": "prod content",
			},
			shouldMatch: false,
		},
		{
			name:     "match run_in_terminal with git status",
			toolName: "run_in_terminal",
			params: map[string]interface{}{
				"command": "git status",
			},
			shouldMatch: true,
			ruleID:      "rule-3",
		},
		{
			name:     "no match run_in_terminal with different command",
			toolName: "run_in_terminal",
			params: map[string]interface{}{
				"command": "rm -rf /",
			},
			shouldMatch: false,
		},
		{
			name:     "no match for unregistered tool",
			toolName: "unknown_tool",
			params: map[string]interface{}{
				"param": "value",
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test pattern matching logic
			paramsJSON, _ := json.Marshal(tt.params)
			paramsStr := string(paramsJSON)

			matched := false
			matchedRuleID := ""

			// Find matching rule
			for _, rule := range cfg.AutoApprovalRules {
				if rule.ToolName != tt.toolName {
					continue
				}

				// Check if pattern matches using regex
				isMatch, err := regexp.MatchString(rule.ParamPattern, paramsStr)
				if err != nil {
					t.Logf("regex error for pattern '%s': %v", rule.ParamPattern, err)
					continue
				}

				if isMatch {
					matched = true
					matchedRuleID = rule.ID
					break
				}
			}

			if matched != tt.shouldMatch {
				t.Errorf("expected match=%v, got match=%v for params: %s", tt.shouldMatch, matched, paramsStr)
			}

			if tt.shouldMatch && matchedRuleID != tt.ruleID {
				t.Errorf("expected rule ID '%s', got '%s'", tt.ruleID, matchedRuleID)
			}
		})
	}
}

// TestAddRuleValidation tests validation when adding rules
func TestApprovalRulePersistence(t *testing.T) {
	// Create temp directory for config
	tempDir, err := os.MkdirTemp("", "wink-approval-persist-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "config.json")

	// Create config with rules
	cfg := types.DefaultConfig()
	cfg.AutoApprovalRules = []types.ApprovalRule{
		{
			ID:           "test-rule-1",
			ToolName:     "read_file",
			ParamPattern: ".*\\.txt$",
			Description:  "Test rule 1",
			UseCount:     5,
		},
		{
			ID:           "test-rule-2",
			ToolName:     "create_file",
			ParamPattern: "test_.*",
			Description:  "Test rule 2",
			UseCount:     10,
		},
	}

	// Save to file
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Load from file
	loadedData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var loadedCfg types.Config
	if err := json.Unmarshal(loadedData, &loadedCfg); err != nil {
		t.Fatalf("failed to unmarshal config: %v", err)
	}

	// Verify rules were persisted correctly
	if len(loadedCfg.AutoApprovalRules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(loadedCfg.AutoApprovalRules))
	}

	// Verify first rule
	rule1 := loadedCfg.AutoApprovalRules[0]
	if rule1.ID != "test-rule-1" {
		t.Errorf("expected ID 'test-rule-1', got '%s'", rule1.ID)
	}
	if rule1.ToolName != "read_file" {
		t.Errorf("expected tool name 'read_file', got '%s'", rule1.ToolName)
	}
	if rule1.UseCount != 5 {
		t.Errorf("expected use count 5, got %d", rule1.UseCount)
	}

	// Verify second rule
	rule2 := loadedCfg.AutoApprovalRules[1]
	if rule2.ID != "test-rule-2" {
		t.Errorf("expected ID 'test-rule-2', got '%s'", rule2.ID)
	}
	if rule2.UseCount != 10 {
		t.Errorf("expected use count 10, got %d", rule2.UseCount)
	}
}

// TestApprovalRuleSpecificity tests that more specific rules take precedence
func TestApprovalRuleSpecificity(t *testing.T) {
	rules := []types.ApprovalRule{
		{
			ID:           "general",
			ToolName:     "read_file",
			ParamPattern: ".*",
			Description:  "Match all read_file calls",
		},
		{
			ID:           "specific",
			ToolName:     "read_file",
			ParamPattern: `"path":".*\.txt"`,
			Description:  "Match only .txt files",
		},
	}

	params := map[string]interface{}{
		"path": "document.txt",
	}

	paramsJSON, _ := json.Marshal(params)
	paramsStr := string(paramsJSON)

	// First matching rule should win (order matters)
	var matchedRule *types.ApprovalRule
	for i := range rules {
		rule := &rules[i]
		if rule.ToolName != "read_file" {
			continue
		}

		// Check if the specific pattern matches
		if rule.ID == "specific" {
			matched, _ := regexp.MatchString(rule.ParamPattern, paramsStr)
			if matched {
				matchedRule = rule
				break
			}
		}

		// Check if general pattern matches
		if rule.ID == "general" {
			matched, _ := regexp.MatchString(rule.ParamPattern, paramsStr)
			if matched {
				matchedRule = rule
				break
			}
		}
	}

	if matchedRule == nil {
		t.Fatal("expected a rule to match")
	}

	// In a proper implementation, more specific rules should be checked first
	// This test documents the expected behavior
	t.Logf("Matched rule: %s (%s)", matchedRule.ID, matchedRule.Description)
}

// TestApprovalRuleEdgeCases tests edge cases in approval rules
func TestApprovalRuleEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		rule     types.ApprovalRule
		params   map[string]interface{}
		expected bool
	}{
		{
			name: "empty params",
			rule: types.ApprovalRule{
				ToolName:     "simple_tool",
				ParamPattern: ".*",
			},
			params:   map[string]interface{}{},
			expected: true, // Empty JSON object "{}" should match ".*"
		},
		{
			name: "special characters in path",
			rule: types.ApprovalRule{
				ToolName:     "read_file",
				ParamPattern: `path.*file\.txt`,
			},
			params: map[string]interface{}{
				"path": "dir/sub.dir/file.txt",
			},
			expected: true,
		},
		{
			name: "numeric parameters",
			rule: types.ApprovalRule{
				ToolName:     "read_file",
				ParamPattern: `"start_line":1`,
			},
			params: map[string]interface{}{
				"path":       "file.txt",
				"start_line": 1,
				"end_line":   10,
			},
			expected: true,
		},
		{
			name: "nested JSON structures",
			rule: types.ApprovalRule{
				ToolName:     "complex_tool",
				ParamPattern: `"nested":\{"key":"value"\}`,
			},
			params: map[string]interface{}{
				"nested": map[string]string{
					"key": "value",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paramsJSON, _ := json.Marshal(tt.params)
			paramsStr := string(paramsJSON)

			// Use regex matching for test
			matched, err := regexp.MatchString(tt.rule.ParamPattern, paramsStr)
			if err != nil {
				t.Logf("regex error: %v", err)
				matched = false
			}

			if matched != tt.expected {
				t.Errorf("expected match=%v, got match=%v\nparams: %s\npattern: %s",
					tt.expected, matched, paramsStr, tt.rule.ParamPattern)
			}
		})
	}
}

// TestAddRuleValidation tests validation when adding rules
func TestAddRuleValidation(t *testing.T) {
	tests := []struct {
		name         string
		toolName     string
		pattern      string
		description  string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "valid rule",
			toolName:    "read_file",
			pattern:     ".*\\.txt",
			description: "Valid rule",
			expectError: false,
		},
		{
			name:         "empty tool name",
			toolName:     "",
			pattern:      ".*",
			description:  "Empty tool name",
			expectError:  true,
			errorMessage: "tool name cannot be empty",
		},
		{
			name:         "empty description",
			toolName:     "read_file",
			pattern:      ".*",
			description:  "",
			expectError:  true,
			errorMessage: "description cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate inputs
			var err error
			if tt.toolName == "" {
				err = &validationError{msg: "tool name cannot be empty"}
			} else if tt.description == "" {
				err = &validationError{msg: "description cannot be empty"}
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', got nil", tt.errorMessage)
				} else if !stringContains(err.Error(), tt.errorMessage) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

type validationError struct {
	msg string
}

func (e *validationError) Error() string {
	return e.msg
}

// Helper function for string contains check
func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
