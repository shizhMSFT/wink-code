// Package ui handles user interaction prompts
package ui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/shizhMSFT/wink-code/pkg/types"
)

// ApprovalResponse represents the user's approval decision
type ApprovalResponse string

const (
	// ApprovalResponseYes - Approve once
	ApprovalResponseYes ApprovalResponse = "yes"
	// ApprovalResponseNo - Reject
	ApprovalResponseNo ApprovalResponse = "no"
	// ApprovalResponseAlways - Approve and create rule
	ApprovalResponseAlways ApprovalResponse = "always"
)

// PromptForApproval asks the user to approve a tool operation
func PromptForApproval(toolName string, params map[string]interface{}, tool types.Tool) (ApprovalResponse, error) {
	// Display operation details to stderr (keeps stdout clean)
	fmt.Fprintf(os.Stderr, "\n┌─────────────────────────────────────────┐\n")
	fmt.Fprintf(os.Stderr, "│ Tool Approval Required                  │\n")
	fmt.Fprintf(os.Stderr, "├─────────────────────────────────────────┤\n")
	fmt.Fprintf(os.Stderr, "│ Tool: %-34s │\n", toolName)

	// Show risk level with color coding
	riskLevel := tool.RiskLevel()
	riskStr := formatRiskLevel(riskLevel)
	fmt.Fprintf(os.Stderr, "│ Risk Level: %-28s │\n", riskStr)

	fmt.Fprintf(os.Stderr, "├─────────────────────────────────────────┤\n")
	fmt.Fprintf(os.Stderr, "│ Parameters:                             │\n")

	// Display parameters more nicely
	for key, value := range params {
		valueStr := formatParamValue(value)
		lines := splitIntoLines(fmt.Sprintf("%s: %s", key, valueStr), 37)
		for i, line := range lines {
			if i == 0 {
				fmt.Fprintf(os.Stderr, "│   %-37s │\n", line)
			} else {
				fmt.Fprintf(os.Stderr, "│     %-35s │\n", line)
			}
		}
	}

	// Show files affected if path parameter exists
	if path, ok := params["path"].(string); ok {
		fmt.Fprintf(os.Stderr, "├─────────────────────────────────────────┤\n")
		fmt.Fprintf(os.Stderr, "│ Files affected: %-23s │\n", truncate(path, 23))
	}

	fmt.Fprintf(os.Stderr, "└─────────────────────────────────────────┘\n")
	fmt.Fprintf(os.Stderr, "\nApprove this operation?\n")
	fmt.Fprintf(os.Stderr, "  (y)es    - Approve once\n")
	fmt.Fprintf(os.Stderr, "  (n)o     - Reject\n")
	fmt.Fprintf(os.Stderr, "  (a)lways - Approve and auto-approve similar operations\n")
	fmt.Fprintf(os.Stderr, "\nYour choice: ")

	// Read user input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return ApprovalResponseNo, fmt.Errorf("failed to read input: %w", err)
	}

	// Parse response
	input = strings.TrimSpace(strings.ToLower(input))
	switch input {
	case "y", "yes":
		return ApprovalResponseYes, nil
	case "n", "no":
		return ApprovalResponseNo, nil
	case "a", "always":
		return ApprovalResponseAlways, nil
	default:
		fmt.Fprintf(os.Stderr, "Invalid response. Defaulting to 'no'.\n")
		return ApprovalResponseNo, nil
	}
}

// formatRiskLevel formats risk level with appropriate label
func formatRiskLevel(level types.RiskLevel) string {
	switch level {
	case types.RiskLevelReadOnly:
		return "read_only"
	case types.RiskLevelSafeWrite:
		return "safe_write"
	case types.RiskLevelDangerous:
		return "dangerous"
	default:
		return "unknown"
	}
}

// formatParamValue formats a parameter value for display
func formatParamValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if len(v) > 100 {
			return v[:97] + "..."
		}
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		// For complex types, use JSON
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		str := string(data)
		if len(str) > 100 {
			return str[:97] + "..."
		}
		return str
	}
}

// splitIntoLines splits a string into lines of maximum width
func splitIntoLines(s string, maxWidth int) []string {
	if len(s) <= maxWidth {
		return []string{s}
	}

	var lines []string
	for len(s) > 0 {
		if len(s) <= maxWidth {
			lines = append(lines, s)
			break
		}
		lines = append(lines, s[:maxWidth])
		s = s[maxWidth:]
	}
	return lines
}

// truncate truncates a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// PromptYesNo asks a simple yes/no question
func PromptYesNo(question string) bool {
	fmt.Fprintf(os.Stderr, "%s (y/n): ", question)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// DisplayMessage shows a message to the user (to stderr)
func DisplayMessage(message string) {
	fmt.Fprintln(os.Stderr, message)
}

// DisplayError shows an error message to the user (to stderr)
func DisplayError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}
