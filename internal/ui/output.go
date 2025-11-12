// Package ui handles output formatting
package ui

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shizhMSFT/wink-code/pkg/types"
)

// Formatter handles output formatting
type Formatter struct {
	format types.OutputFormat
}

// NewFormatter creates a new output formatter
func NewFormatter(format types.OutputFormat) *Formatter {
	return &Formatter{format: format}
}

// FormatToolResult formats a tool result for display
func (f *Formatter) FormatToolResult(result *types.ToolResult) string {
	if f.format == types.OutputFormatJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		return string(data)
	}

	// Human-readable format
	if result.Success {
		return fmt.Sprintf("✓ %s", result.Output)
	}
	return fmt.Sprintf("✗ Error: %s", result.Error)
}

// FormatMessage formats an assistant message for display
func (f *Formatter) FormatMessage(content string) string {
	if f.format == types.OutputFormatJSON {
		data := map[string]string{"type": "message", "content": content}
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonData)
	}

	return content
}

// FormatAutoApproval formats an auto-approval notification
func (f *Formatter) FormatAutoApproval(toolName string, ruleDescription string) string {
	if f.format == types.OutputFormatJSON {
		data := map[string]string{
			"type": "auto_approval",
			"tool": toolName,
			"rule": ruleDescription,
		}
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonData)
	}

	return fmt.Sprintf("⚡ Auto-approved: %s (rule: %s)", toolName, ruleDescription)
}

// FormatSessionInfo formats session information
func (f *Formatter) FormatSessionInfo(sessionID string, messageCount int) string {
	if f.format == types.OutputFormatJSON {
		data := map[string]interface{}{
			"type":          "session_info",
			"session_id":    sessionID,
			"message_count": messageCount,
		}
		jsonData, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonData)
	}

	return fmt.Sprintf("Session: %s (%d messages)\nUse 'wink --continue' to resume this session.",
		sessionID[:8], messageCount)
}

// PrintOutput prints output to stdout (for piping)
func PrintOutput(content string) {
	fmt.Fprintln(os.Stdout, content)
}

// PrintInfo prints informational messages to stderr
func PrintInfo(message string) {
	fmt.Fprintln(os.Stderr, message)
}

// PrintSuccess prints success messages to stderr
func PrintSuccess(message string) {
	fmt.Fprintf(os.Stderr, "✓ %s\n", message)
}

// PrintWarning prints warning messages to stderr
func PrintWarning(message string) {
	fmt.Fprintf(os.Stderr, "⚠ %s\n", message)
}
