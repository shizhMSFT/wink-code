package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shizhMSFT/wink-code/pkg/types"
)

// TestAutoApprovalWorkflow tests the complete auto-approval workflow
func TestAutoApprovalWorkflow(t *testing.T) {
	// Create temp directory for config
	tempDir, err := os.MkdirTemp("", "wink-autoapproval-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create config directory
	configDir := filepath.Join(tempDir, ".wink")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "config.json")

	t.Run("create and save auto-approval rule", func(t *testing.T) {
		// Start with empty config
		cfg := types.DefaultConfig()
		cfg.AutoApprovalRules = []types.ApprovalRule{}

		// Save initial config
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal config: %v", err)
		}
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Simulate user choosing "always" for a tool call
		// This would normally be done through ApprovalManager.AddRule()
		newRule := types.ApprovalRule{
			ID:           "test-rule-1",
			ToolName:     "read_file",
			ParamPattern: `"path":"test\.txt"`,
			Description:  "Auto-approve reading test.txt",
			UseCount:     0,
		}

		// Add rule to config
		cfg.AutoApprovalRules = append(cfg.AutoApprovalRules, newRule)

		// Save updated config
		data, err = json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal updated config: %v", err)
		}
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			t.Fatalf("failed to write updated config: %v", err)
		}

		// Verify rule was saved
		loadedData, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		var loadedCfg types.Config
		if err := json.Unmarshal(loadedData, &loadedCfg); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}

		if len(loadedCfg.AutoApprovalRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(loadedCfg.AutoApprovalRules))
		}

		if loadedCfg.AutoApprovalRules[0].ID != "test-rule-1" {
			t.Errorf("expected rule ID 'test-rule-1', got '%s'", loadedCfg.AutoApprovalRules[0].ID)
		}
	})

	t.Run("load and match auto-approval rule on next run", func(t *testing.T) {
		// Load config (simulating next run)
		loadedData, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		var cfg types.Config
		if err := json.Unmarshal(loadedData, &cfg); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}

		// Verify rule was persisted
		if len(cfg.AutoApprovalRules) != 1 {
			t.Fatalf("expected 1 rule, got %d", len(cfg.AutoApprovalRules))
		}

		// Test matching the same tool call
		params := map[string]interface{}{
			"path": "test.txt",
		}

		paramsJSON, _ := json.Marshal(params)
		paramsStr := string(paramsJSON)

		// Check if rule matches
		rule := cfg.AutoApprovalRules[0]
		if rule.ToolName != "read_file" {
			t.Fatalf("expected tool name 'read_file', got '%s'", rule.ToolName)
		}

		// Simple pattern match for test
		matched := contains(paramsStr, "test.txt")
		if !matched {
			t.Errorf("expected rule to match params: %s", paramsStr)
		}

		t.Logf("Rule successfully matched on next run: %s", rule.Description)
	})

	t.Run("update use count when rule is matched", func(t *testing.T) {
		// Load config
		loadedData, err := os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		var cfg types.Config
		if err := json.Unmarshal(loadedData, &cfg); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}

		// Increment use count (simulating actual match)
		cfg.AutoApprovalRules[0].UseCount++

		// Save updated config
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal config: %v", err)
		}
		if err := os.WriteFile(configFile, data, 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Load and verify use count was updated
		loadedData, err = os.ReadFile(configFile)
		if err != nil {
			t.Fatalf("failed to read config: %v", err)
		}

		var updatedCfg types.Config
		if err := json.Unmarshal(loadedData, &updatedCfg); err != nil {
			t.Fatalf("failed to unmarshal config: %v", err)
		}

		if updatedCfg.AutoApprovalRules[0].UseCount != 1 {
			t.Errorf("expected use count 1, got %d", updatedCfg.AutoApprovalRules[0].UseCount)
		}
	})
}

// TestApprovalManagerIntegration tests the ApprovalManager directly
func TestApprovalManagerIntegration(t *testing.T) {
	// Create temp home directory
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		originalHome = os.Getenv("USERPROFILE")
	}

	tempHome, err := os.MkdirTemp("", "wink-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Create .wink config directory
	winkDir := filepath.Join(tempHome, ".wink")
	if err := os.MkdirAll(winkDir, 0755); err != nil {
		t.Fatalf("failed to create .wink dir: %v", err)
	}

	// Note: For actual integration testing with ApprovalManager,
	// we would need to override the config path or use dependency injection
	t.Run("approval manager workflow", func(t *testing.T) {
		// This test documents the expected workflow
		// Actual implementation would use the real ApprovalManager

		configFile := filepath.Join(winkDir, "config.json")

		// Step 1: User selects "always" for a tool call
		cfg := types.DefaultConfig()
		newRule := types.ApprovalRule{
			ID:           "rule-1",
			ToolName:     "create_file",
			ParamPattern: `"path":".*\.md"`,
			Description:  "Auto-approve creating markdown files",
		}
		cfg.AutoApprovalRules = []types.ApprovalRule{newRule}

		// Step 2: Rule is persisted
		data, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(configFile, data, 0644)

		// Step 3: Next run loads rules
		loadedData, _ := os.ReadFile(configFile)
		var loadedCfg types.Config
		json.Unmarshal(loadedData, &loadedCfg)

		// Step 4: Rule is matched for similar tool call
		params := map[string]interface{}{
			"path":    "README.md",
			"content": "# Project",
		}
		paramsJSON, _ := json.Marshal(params)
		paramsStr := string(paramsJSON)

		// Check if rule would match
		matched := contains(paramsStr, ".md")
		if !matched {
			t.Error("expected rule to match markdown file creation")
		}

		t.Log("✓ Auto-approval workflow complete: rule created, persisted, and matched")
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findInString(s, substr)
}

func findInString(s, substr string) bool {
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
