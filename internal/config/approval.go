// Package config handles auto-approval rule management
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// ApprovalManager manages auto-approval rules
type ApprovalManager struct {
	rules      []types.ApprovalRule
	configPath string
}

// NewApprovalManager creates a new approval manager
func NewApprovalManager() (*ApprovalManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, configDir, configFile)

	manager := &ApprovalManager{
		configPath: configPath,
		rules:      []types.ApprovalRule{},
	}

	// Load existing rules
	if err := manager.Load(); err != nil {
		return nil, err
	}

	return manager, nil
}

// Load reads approval rules from config file
func (a *ApprovalManager) Load() error {
	// Check if config file exists
	if _, err := os.Stat(a.configPath); os.IsNotExist(err) {
		return nil // No rules yet
	}

	// Read config file
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config types.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	a.rules = config.AutoApprovalRules
	return nil
}

// Save writes approval rules to config file
func (a *ApprovalManager) Save() error {
	// Read existing config
	var config types.Config
	if data, err := os.ReadFile(a.configPath); err == nil {
		_ = json.Unmarshal(data, &config)
	} else {
		config = *types.DefaultConfig()
	}

	// Update rules
	config.AutoApprovalRules = a.rules

	// Write back
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(a.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddRule adds a new auto-approval rule
func (a *ApprovalManager) AddRule(toolName string, paramPattern string, description string) (*types.ApprovalRule, error) {
	// Validate regex pattern
	if _, err := regexp.Compile(paramPattern); err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// Create rule
	rule := types.ApprovalRule{
		ID:           uuid.New().String(),
		ToolName:     toolName,
		ParamPattern: paramPattern,
		Description:  description,
		CreatedAt:    time.Now(),
		UseCount:     0,
	}

	a.rules = append(a.rules, rule)

	// Save to disk
	if err := a.Save(); err != nil {
		return nil, err
	}

	return &rule, nil
}

// MatchRule checks if a tool call matches any auto-approval rule
func (a *ApprovalManager) MatchRule(toolName string, params map[string]interface{}) (*types.ApprovalRule, error) {
	// Serialize params to JSON string for regex matching
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	paramsStr := string(paramsJSON)

	// Check each rule
	for i := range a.rules {
		rule := &a.rules[i]

		// Must match tool name
		if rule.ToolName != toolName {
			continue
		}

		// Check regex match
		matched, err := regexp.MatchString(rule.ParamPattern, paramsStr)
		if err != nil {
			continue // Skip invalid regex
		}

		if matched {
			// Update usage stats
			rule.UseCount++
			rule.LastUsedAt = time.Now()
			_ = a.Save() // Best effort save

			return rule, nil
		}
	}

	return nil, nil // No match
}

// GetRules returns all approval rules
func (a *ApprovalManager) GetRules() []types.ApprovalRule {
	return a.rules
}

// RemoveRule removes a rule by ID
func (a *ApprovalManager) RemoveRule(ruleID string) error {
	for i, rule := range a.rules {
		if rule.ID == ruleID {
			a.rules = append(a.rules[:i], a.rules[i+1:]...)
			return a.Save()
		}
	}
	return fmt.Errorf("rule not found: %s", ruleID)
}
