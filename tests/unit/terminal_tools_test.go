package unit

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// TestRunInTerminalTool tests command execution
func TestRunInTerminalTool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := tools.NewRunInTerminalTool()
	ctx := context.Background()

	// Platform-specific test commands
	var echoCmd, failCmd, timeoutCmd string
	if runtime.GOOS == "windows" {
		echoCmd = "echo Hello World"
		failCmd = "exit 1"
		timeoutCmd = "timeout /t 2 /nobreak"
	} else {
		echoCmd = "echo Hello World"
		failCmd = "exit 1"
		timeoutCmd = "sleep 2"
	}

	tests := []struct {
		name          string
		params        map[string]interface{}
		expectError   bool
		expectSuccess bool
		checkOutput   string
	}{
		{
			name: "Simple echo command",
			params: map[string]interface{}{
				"command": echoCmd,
			},
			expectError:   false,
			expectSuccess: true,
			checkOutput:   "Hello World",
		},
		{
			name: "Command with exit code 1",
			params: map[string]interface{}{
				"command": failCmd,
			},
			expectError:   false,
			expectSuccess: false, // exit code 1 = not successful
		},
		{
			name: "Command with timeout (should timeout)",
			params: map[string]interface{}{
				"command":         timeoutCmd,
				"timeout_seconds": 1.0,
			},
			expectError:   true, // Timeout is an error
			expectSuccess: false,
		},
		{
			name: "Empty command",
			params: map[string]interface{}{
				"command": "",
			},
			expectError: true,
		},
		{
			name: "Invalid timeout (too low)",
			params: map[string]interface{}{
				"command":         echoCmd,
				"timeout_seconds": 0.0,
			},
			expectError: true,
		},
		{
			name: "Invalid timeout (too high)",
			params: map[string]interface{}{
				"command":         echoCmd,
				"timeout_seconds": 400.0,
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
					// Check if error happens during execution
					result, execErr := tool.Execute(ctx, tt.params, tmpDir)
					if execErr == nil && result.Success {
						t.Error("Expected error but got success")
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("Validation failed: %v", err)
			}

			// Execute
			result, err := tool.Execute(ctx, tt.params, tmpDir)

			// Check success expectation
			if tt.expectSuccess && !result.Success {
				t.Errorf("Expected success but got failure: %s", result.Output)
			}
			if !tt.expectSuccess && result.Success {
				t.Errorf("Expected failure but got success: %s", result.Output)
			}

			// Check output contains expected string
			if tt.checkOutput != "" && !strings.Contains(result.Output, tt.checkOutput) {
				t.Errorf("Expected output to contain '%s', got: %s",
					tt.checkOutput, result.Output)
			}

			// Check metadata
			if result.Metadata != nil {
				if _, ok := result.Metadata["exit_code"]; !ok {
					t.Error("Missing exit_code in metadata")
				}
				if _, ok := result.Metadata["command"]; !ok {
					t.Error("Missing command in metadata")
				}
			}
		})
	}
}

// TestTerminalLastCommandTool tests last command retrieval
func TestTerminalLastCommandTool(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()

	// First, run a command to populate history
	runTool := tools.NewRunInTerminalTool()
	lastTool := tools.NewTerminalLastCommandTool()

	var testCmd string
	if runtime.GOOS == "windows" {
		testCmd = "echo Test Command"
	} else {
		testCmd = "echo Test Command"
	}

	t.Run("No command history initially", func(t *testing.T) {
		// Note: This might fail if other tests already ran commands
		// We can't easily reset global history, so just check structure
		result, _ := lastTool.Execute(ctx, map[string]interface{}{}, tmpDir)

		// Should have metadata structure even if empty
		if result.Metadata == nil {
			// It's ok if there's no history yet
			if !strings.Contains(result.Output, "No previous command") {
				// If there IS history, verify it has the right fields
				if _, ok := result.Metadata["command"]; !ok {
					t.Error("If history exists, should have command in metadata")
				}
			}
		}
	})

	t.Run("After running a command", func(t *testing.T) {
		// Execute a command
		_, err := runTool.Execute(ctx, map[string]interface{}{
			"command": testCmd,
		}, tmpDir)
		if err != nil {
			t.Fatalf("Failed to run command: %v", err)
		}

		// Get last command
		result, err := lastTool.Execute(ctx, map[string]interface{}{}, tmpDir)
		if err != nil {
			t.Fatalf("Failed to get last command: %v", err)
		}

		if !result.Success {
			t.Errorf("Expected success, got: %s", result.Output)
		}

		// Check that output contains the command
		if !strings.Contains(result.Output, testCmd) {
			t.Errorf("Expected output to contain '%s', got: %s",
				testCmd, result.Output)
		}

		// Check metadata
		if result.Metadata == nil {
			t.Fatal("Missing metadata")
		}

		cmd, ok := result.Metadata["command"].(string)
		if !ok {
			t.Error("Missing or invalid command in metadata")
		}
		if cmd != testCmd {
			t.Errorf("Expected command '%s', got '%s'", testCmd, cmd)
		}

		exitCode, ok := result.Metadata["exit_code"].(int)
		if !ok {
			t.Error("Missing or invalid exit_code in metadata")
		}
		if exitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", exitCode)
		}
	})
}

// TestTerminalToolsRiskLevel verifies risk levels
func TestTerminalToolsRiskLevel(t *testing.T) {
	tests := []struct {
		name     string
		tool     types.Tool
		expected types.RiskLevel
	}{
		{
			name:     "run_in_terminal is dangerous",
			tool:     tools.NewRunInTerminalTool(),
			expected: types.RiskLevelDangerous,
		},
		{
			name:     "terminal_last_command is read_only",
			tool:     tools.NewTerminalLastCommandTool(),
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

// TestTerminalToolsRequireApproval verifies approval requirements
func TestTerminalToolsRequireApproval(t *testing.T) {
	tests := []struct {
		name     string
		tool     types.Tool
		expected bool
	}{
		{
			name:     "run_in_terminal requires approval",
			tool:     tools.NewRunInTerminalTool(),
			expected: true,
		},
		{
			name:     "terminal_last_command does not require approval",
			tool:     tools.NewTerminalLastCommandTool(),
			expected: false,
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

// TestCommandOutputLimits tests output size limiting
func TestCommandOutputLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping output limit test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := tools.NewRunInTerminalTool()
	ctx := context.Background()

	// Generate large output (platform-specific)
	var largeOutputCmd string
	if runtime.GOOS == "windows" {
		// Print many lines
		largeOutputCmd = "for /L %i in (1,1,10000) do @echo Line %i"
	} else {
		largeOutputCmd = "for i in {1..10000}; do echo Line $i; done"
	}

	result, err := tool.Execute(ctx, map[string]interface{}{
		"command": largeOutputCmd,
	}, tmpDir)

	if err != nil {
		t.Fatalf("Command execution failed: %v", err)
	}

	// Output should be truncated (100KB limit)
	const maxExpectedSize = 110 * 1024 // 100KB + some overhead
	if len(result.Output) > maxExpectedSize {
		t.Errorf("Output size %d exceeds expected max %d",
			len(result.Output), maxExpectedSize)
	}

	// Should contain truncation message if truncated
	if len(result.Output) > 100*1024 {
		if !strings.Contains(result.Output, "truncated") {
			t.Error("Large output should contain truncation notice")
		}
	}
}
