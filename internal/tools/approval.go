// Package tools implements the approval workflow
package tools

import (
	"encoding/json"
	"fmt"

	"github.com/shizhMSFT/wink-code/internal/config"
	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/internal/ui"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// ApprovalWorkflow handles tool execution approval
type ApprovalWorkflow struct {
	approvalManager *config.ApprovalManager
}

// NewApprovalWorkflow creates a new approval workflow
func NewApprovalWorkflow() (*ApprovalWorkflow, error) {
	approvalManager, err := config.NewApprovalManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize approval manager: %w", err)
	}

	return &ApprovalWorkflow{
		approvalManager: approvalManager,
	}, nil
}

// CheckApproval checks if a tool call should be approved
// Returns: (approved bool, autoApproved bool, ruleDescription string, error)
func (aw *ApprovalWorkflow) CheckApproval(toolName string, params map[string]interface{}, tool types.Tool) (bool, bool, string, error) {
	// Check auto-approval rules
	rule, err := aw.approvalManager.MatchRule(toolName, params)
	if err != nil {
		logging.Warn("Failed to check auto-approval rules", "error", err)
		// Continue to manual approval on error
	}

	if rule != nil {
		// Auto-approved
		logging.Debug("Tool call auto-approved",
			"tool", toolName,
			"rule_id", rule.ID,
			"rule_description", rule.Description,
		)
		return true, true, rule.Description, nil
	}

	// Prompt user for approval
	response, err := ui.PromptForApproval(toolName, params, tool)
	if err != nil {
		return false, false, "", fmt.Errorf("failed to get approval: %w", err)
	}

	switch response {
	case ui.ApprovalResponseYes:
		logging.Debug("Tool call manually approved", "tool", toolName)
		return true, false, "", nil

	case ui.ApprovalResponseNo:
		logging.Debug("Tool call rejected", "tool", toolName)
		return false, false, "", nil

	case ui.ApprovalResponseAlways:
		// Create auto-approval rule
		if err := aw.CreateAutoApprovalRule(toolName, params); err != nil {
			logging.Warn("Failed to create auto-approval rule", "error", err)
			// Still approve this time even if rule creation failed
		}
		logging.Debug("Tool call approved and rule created", "tool", toolName)
		return true, false, "newly created rule", nil

	default:
		return false, false, "", fmt.Errorf("unexpected approval response: %v", response)
	}
}

// CreateAutoApprovalRule creates an auto-approval rule from a tool call
func (aw *ApprovalWorkflow) CreateAutoApprovalRule(toolName string, params map[string]interface{}) error {
	// Generate regex pattern from params
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}

	// Escape special regex characters in the JSON string
	pattern := escapeRegex(string(paramsJSON))

	// Create description
	description := fmt.Sprintf("Auto-approve %s with params: %s", toolName, truncate(string(paramsJSON), 50))

	// Add rule
	rule, err := aw.approvalManager.AddRule(toolName, pattern, description)
	if err != nil {
		return fmt.Errorf("failed to add rule: %w", err)
	}

	logging.Info("Auto-approval rule created",
		"rule_id", rule.ID,
		"tool", toolName,
		"description", description,
	)

	ui.PrintSuccess(fmt.Sprintf("Created auto-approval rule: %s", description))

	return nil
}

// escapeRegex escapes special regex characters
func escapeRegex(s string) string {
	// Escape common regex special characters
	replacements := map[string]string{
		"\\": "\\\\",
		".":  "\\.",
		"*":  "\\*",
		"+":  "\\+",
		"?":  "\\?",
		"^":  "\\^",
		"$":  "\\$",
		"[":  "\\[",
		"]":  "\\]",
		"{":  "\\{",
		"}":  "\\}",
		"(":  "\\(",
		")":  "\\)",
		"|":  "\\|",
	}

	result := s
	for old, new := range replacements {
		result = replaceAll(result, old, new)
	}
	return result
}

// replaceAll replaces all occurrences of old with new in s
func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); i++ {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old) - 1
		} else {
			result += string(s[i])
		}
	}
	return result
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
