package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// CommandHistory tracks executed commands in the session
type CommandHistory struct {
	mu       sync.RWMutex
	commands []ExecutedCommand
}

// ExecutedCommand represents a command that was run
type ExecutedCommand struct {
	Command    string
	ExitCode   int
	Stdout     string
	Stderr     string
	ExecutedAt time.Time
}

var (
	// Global command history for the session
	commandHistory = &CommandHistory{
		commands: make([]ExecutedCommand, 0),
	}
)

// AddCommand adds a command to history
func (h *CommandHistory) AddCommand(cmd ExecutedCommand) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.commands = append(h.commands, cmd)
}

// GetLast returns the last command, or nil if none
func (h *CommandHistory) GetLast() *ExecutedCommand {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.commands) == 0 {
		return nil
	}
	return &h.commands[len(h.commands)-1]
}

// RunInTerminalTool implements run_in_terminal
type RunInTerminalTool struct{}

// NewRunInTerminalTool creates a new run_in_terminal tool
func NewRunInTerminalTool() *RunInTerminalTool {
	return &RunInTerminalTool{}
}

func (t *RunInTerminalTool) Name() string {
	return "run_in_terminal"
}

func (t *RunInTerminalTool) Description() string {
	return "Execute a shell command in the terminal"
}

func (t *RunInTerminalTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Shell command to execute",
			},
			"timeout_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Timeout in seconds (default: 30)",
				"default":     30,
			},
		},
		"required": []string{"command"},
	}
}

func (t *RunInTerminalTool) Validate(params map[string]interface{}, workingDir string) error {
	command, ok := params["command"].(string)
	if !ok || strings.TrimSpace(command) == "" {
		return fmt.Errorf("command parameter is required and must be a non-empty string")
	}

	// Validate timeout if provided
	if timeout, ok := params["timeout_seconds"].(float64); ok {
		if timeout < 1 || timeout > 300 {
			return fmt.Errorf("timeout_seconds must be between 1 and 300")
		}
	}

	return nil
}

func (t *RunInTerminalTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	command := strings.TrimSpace(params["command"].(string))
	timeoutSeconds := 30
	if timeout, ok := params["timeout_seconds"].(float64); ok {
		timeoutSeconds = int(timeout)
	}

	// Detect shell
	shell, shellArgs := detectShell()
	logging.Debug("run_in_terminal: shell=%s command=%s timeout=%d", shell, command, timeoutSeconds)

	// Create command with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Build command with shell
	args := append(shellArgs, command)
	cmd := exec.CommandContext(cmdCtx, shell, args...)
	cmd.Dir = workingDir

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	executionTime := time.Since(startTime).Milliseconds()

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if cmdCtx.Err() == context.DeadlineExceeded {
			// Timeout
			return &types.ToolResult{
				Success:         false,
				Output:          fmt.Sprintf("Command timed out after %d seconds", timeoutSeconds),
				Error:           fmt.Sprintf("timeout after %d seconds", timeoutSeconds),
				ExecutionTimeMs: executionTime,
			}, fmt.Errorf("command timed out")
		} else {
			// Other error (e.g., command not found)
			return &types.ToolResult{
				Success:         false,
				Output:          fmt.Sprintf("Command failed: %v", err),
				Error:           err.Error(),
				ExecutionTimeMs: executionTime,
			}, err
		}
	}

	// Limit output size (100KB each)
	const maxOutputSize = 100 * 1024
	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	if len(stdoutStr) > maxOutputSize {
		stdoutStr = stdoutStr[:maxOutputSize] + "\n... (output truncated)"
	}
	if len(stderrStr) > maxOutputSize {
		stderrStr = stderrStr[:maxOutputSize] + "\n... (output truncated)"
	}

	// Add to command history
	commandHistory.AddCommand(ExecutedCommand{
		Command:    command,
		ExitCode:   exitCode,
		Stdout:     stdoutStr,
		Stderr:     stderrStr,
		ExecutedAt: startTime,
	})

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Command: %s\n", command))
	output.WriteString(fmt.Sprintf("Exit code: %d\n", exitCode))

	if len(stdoutStr) > 0 {
		output.WriteString("\nStdout:\n")
		output.WriteString(stdoutStr)
	}

	if len(stderrStr) > 0 {
		output.WriteString("\nStderr:\n")
		output.WriteString(stderrStr)
	}

	success := exitCode == 0

	logging.Debug("run_in_terminal: exit_code=%d stdout_bytes=%d stderr_bytes=%d time=%dms",
		exitCode, len(stdoutStr), len(stderrStr), executionTime)

	return &types.ToolResult{
		Success:         success,
		Output:          output.String(),
		ExecutionTimeMs: executionTime,
		Metadata: map[string]interface{}{
			"exit_code":    exitCode,
			"stdout_lines": len(strings.Split(stdoutStr, "\n")),
			"stderr_lines": len(strings.Split(stderrStr, "\n")),
			"command":      command,
		},
	}, nil
}

func (t *RunInTerminalTool) RequiresApproval() bool {
	return true
}

func (t *RunInTerminalTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelDangerous
}

// detectShell returns the appropriate shell and its command flag
func detectShell() (string, []string) {
	switch runtime.GOOS {
	case "windows":
		// Check for PowerShell first, fallback to cmd
		if _, err := exec.LookPath("pwsh.exe"); err == nil {
			return "pwsh.exe", []string{"-NoProfile", "-Command"}
		}
		if _, err := exec.LookPath("powershell.exe"); err == nil {
			return "powershell.exe", []string{"-NoProfile", "-Command"}
		}
		return "cmd.exe", []string{"/C"}
	default:
		// Unix-like systems
		return "sh", []string{"-c"}
	}
}

// TerminalLastCommandTool implements terminal_last_command
type TerminalLastCommandTool struct{}

// NewTerminalLastCommandTool creates a new terminal_last_command tool
func NewTerminalLastCommandTool() *TerminalLastCommandTool {
	return &TerminalLastCommandTool{}
}

func (t *TerminalLastCommandTool) Name() string {
	return "terminal_last_command"
}

func (t *TerminalLastCommandTool) Description() string {
	return "Retrieve the last shell command that was executed"
}

func (t *TerminalLastCommandTool) ParametersSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *TerminalLastCommandTool) Validate(params map[string]interface{}, workingDir string) error {
	return nil // No parameters to validate
}

func (t *TerminalLastCommandTool) Execute(ctx context.Context, params map[string]interface{}, workingDir string) (*types.ToolResult, error) {
	startTime := time.Now()

	lastCmd := commandHistory.GetLast()
	if lastCmd == nil {
		return &types.ToolResult{
			Success:         false,
			Output:          "No previous command in this session",
			Error:           "no command history",
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, fmt.Errorf("no command history")
	}

	// Format output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Last command: %s\n", lastCmd.Command))
	output.WriteString(fmt.Sprintf("Executed at: %s\n", lastCmd.ExecutedAt.Format(time.RFC3339)))
	output.WriteString(fmt.Sprintf("Exit code: %d\n", lastCmd.ExitCode))

	if len(lastCmd.Stdout) > 0 {
		output.WriteString("\nStdout:\n")
		output.WriteString(lastCmd.Stdout)
	}

	if len(lastCmd.Stderr) > 0 {
		output.WriteString("\nStderr:\n")
		output.WriteString(lastCmd.Stderr)
	}

	return &types.ToolResult{
		Success:         true,
		Output:          output.String(),
		ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		Metadata: map[string]interface{}{
			"command":     lastCmd.Command,
			"exit_code":   lastCmd.ExitCode,
			"executed_at": lastCmd.ExecutedAt.Format(time.RFC3339),
		},
	}, nil
}

func (t *TerminalLastCommandTool) RequiresApproval() bool {
	return false // Read-only, no approval needed
}

func (t *TerminalLastCommandTool) RiskLevel() types.RiskLevel {
	return types.RiskLevelReadOnly
}
