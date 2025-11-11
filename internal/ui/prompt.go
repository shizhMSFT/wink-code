// Package ui handles user interaction prompts
package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
func PromptForApproval(toolName string, params map[string]interface{}) (ApprovalResponse, error) {
	// Display operation details to stderr (keeps stdout clean)
	fmt.Fprintf(os.Stderr, "\n┌─────────────────────────────────────────┐\n")
	fmt.Fprintf(os.Stderr, "│ Tool Approval Required                  │\n")
	fmt.Fprintf(os.Stderr, "├─────────────────────────────────────────┤\n")
	fmt.Fprintf(os.Stderr, "│ Tool: %-34s │\n", toolName)

	// Display parameters
	for key, value := range params {
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 30 {
			valueStr = valueStr[:27] + "..."
		}
		fmt.Fprintf(os.Stderr, "│ %s: %-30s │\n", key, valueStr)
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
