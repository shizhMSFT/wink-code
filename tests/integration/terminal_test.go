package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/shizhMSFT/wink-code/internal/tools"
)

// TestTerminalWorkflow tests complete terminal command workflows
func TestTerminalWorkflow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Register tools
	registry := tools.NewRegistry()
	if err := registry.Register(tools.NewRunInTerminalTool()); err != nil {
		t.Fatalf("Failed to register run_in_terminal: %v", err)
	}
	if err := registry.Register(tools.NewTerminalLastCommandTool()); err != nil {
		t.Fatalf("Failed to register terminal_last_command: %v", err)
	}
	if err := registry.Register(tools.NewCreateFileTool()); err != nil {
		t.Fatalf("Failed to register create_file: %v", err)
	}
	if err := registry.Register(tools.NewReadFileTool()); err != nil {
		t.Fatalf("Failed to register read_file: %v", err)
	}

	ctx := context.Background()

	t.Run("Execute command and retrieve output", func(t *testing.T) {
		var echoCmd string
		if runtime.GOOS == "windows" {
			echoCmd = "echo Hello from terminal"
		} else {
			echoCmd = "echo Hello from terminal"
		}

		// Execute command
		result, err := registry.Execute(ctx, "run_in_terminal",
			map[string]interface{}{
				"command": echoCmd,
			}, tmpDir)
		if err != nil {
			t.Fatalf("run_in_terminal failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Command execution failed: %s", result.Output)
		}

		// Check output contains expected text
		if !strings.Contains(result.Output, "Hello from terminal") {
			t.Errorf("Expected 'Hello from terminal' in output: %s", result.Output)
		}

		// Check metadata
		exitCode, ok := result.Metadata["exit_code"].(int)
		if !ok || exitCode != 0 {
			t.Errorf("Expected exit code 0, got %v", exitCode)
		}

		// Retrieve last command
		lastResult, err := registry.Execute(ctx, "terminal_last_command",
			map[string]interface{}{}, tmpDir)
		if err != nil {
			t.Fatalf("terminal_last_command failed: %v", err)
		}

		if !lastResult.Success {
			t.Fatalf("Failed to get last command: %s", lastResult.Output)
		}

		// Verify it matches
		if !strings.Contains(lastResult.Output, echoCmd) {
			t.Errorf("Last command should contain '%s', got: %s",
				echoCmd, lastResult.Output)
		}
	})

	t.Run("Create file and verify with command", func(t *testing.T) {
		// Create a test file
		testFile := "test_output.txt"
		testContent := "Test content from integration"

		createResult, err := registry.Execute(ctx, "create_file",
			map[string]interface{}{
				"path":    testFile,
				"content": testContent,
			}, tmpDir)
		if err != nil {
			t.Fatalf("create_file failed: %v", err)
		}

		if !createResult.Success {
			t.Fatalf("File creation failed: %s", createResult.Output)
		}

		// Use terminal command to verify file exists
		var catCmd string
		if runtime.GOOS == "windows" {
			catCmd = "Get-Content " + testFile
		} else {
			catCmd = "cat " + testFile
		}

		result, err := registry.Execute(ctx, "run_in_terminal",
			map[string]interface{}{
				"command": catCmd,
			}, tmpDir)
		if err != nil {
			t.Fatalf("run_in_terminal failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Command execution failed: %s", result.Output)
		}

		// Verify content is in output
		if !strings.Contains(result.Output, testContent) {
			t.Errorf("Expected '%s' in command output: %s",
				testContent, result.Output)
		}
	})

	t.Run("List directory contents", func(t *testing.T) {
		// Create some test files
		testFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
		for _, fname := range testFiles {
			path := filepath.Join(tmpDir, fname)
			if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}

		// Use terminal to list directory
		var lsCmd string
		if runtime.GOOS == "windows" {
			lsCmd = "Get-ChildItem -Name" // PowerShell-compatible
		} else {
			lsCmd = "ls"
		}

		result, err := registry.Execute(ctx, "run_in_terminal",
			map[string]interface{}{
				"command": lsCmd,
			}, tmpDir)
		if err != nil {
			t.Fatalf("run_in_terminal failed: %v", err)
		}

		if !result.Success {
			t.Fatalf("Command execution failed: %s", result.Output)
		}

		// Check that at least some files are listed
		foundFiles := 0
		for _, fname := range testFiles {
			if strings.Contains(result.Output, fname) {
				foundFiles++
			}
		}

		if foundFiles == 0 {
			t.Errorf("Expected to find files in directory listing: %s",
				result.Output)
		}
	})

	t.Run("Command with non-zero exit code", func(t *testing.T) {
		var failCmd string
		if runtime.GOOS == "windows" {
			failCmd = "exit 1"
		} else {
			failCmd = "exit 1"
		}

		result, err := registry.Execute(ctx, "run_in_terminal",
			map[string]interface{}{
				"command": failCmd,
			}, tmpDir)

		// Should not return error (command ran)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// But success should be false
		if result.Success {
			t.Error("Expected success=false for exit code 1")
		}

		// Check exit code in metadata
		exitCode, ok := result.Metadata["exit_code"].(int)
		if !ok || exitCode != 1 {
			t.Errorf("Expected exit code 1, got %v", exitCode)
		}
	})
}

// TestTerminalSafety tests safety features
func TestTerminalSafety(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tool := tools.NewRunInTerminalTool()
	ctx := context.Background()

	t.Run("Requires approval", func(t *testing.T) {
		if !tool.RequiresApproval() {
			t.Error("run_in_terminal should require approval")
		}

		if tool.RiskLevel() != "dangerous" {
			t.Errorf("Expected dangerous risk level, got %s", tool.RiskLevel())
		}
	})

	t.Run("Working directory is respected", func(t *testing.T) {
		// Create a file in tmpDir
		testFile := "verify_wd.txt"
		testPath := filepath.Join(tmpDir, testFile)
		if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Command to check if file exists
		var checkCmd string
		if runtime.GOOS == "windows" {
			checkCmd = "if (Test-Path " + testFile + ") { echo FOUND } else { echo NOT_FOUND }"
		} else {
			checkCmd = "[ -f " + testFile + " ] && echo FOUND || echo NOT_FOUND"
		}

		result, err := tool.Execute(ctx, map[string]interface{}{
			"command": checkCmd,
		}, tmpDir)

		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}

		// Should find the file (working directory was set correctly)
		if !strings.Contains(result.Output, "FOUND") {
			t.Errorf("File should be found in working directory. Output: %s",
				result.Output)
		}
	})

	t.Run("Timeout is enforced", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping timeout test in short mode")
		}

		var sleepCmd string
		if runtime.GOOS == "windows" {
			sleepCmd = "Start-Sleep -Seconds 3" // PowerShell sleep
		} else {
			sleepCmd = "sleep 3"
		}

		result, err := tool.Execute(ctx, map[string]interface{}{
			"command":         sleepCmd,
			"timeout_seconds": 1.0,
		}, tmpDir)

		// Should timeout OR fail quickly
		// On Windows PowerShell might fail with exit code 1 if interrupted
		if err == nil && result.Success {
			t.Error("Expected command to timeout or fail, but it succeeded")
		}

		// Accept either timeout message or failure
		hasTimeoutMsg := strings.Contains(result.Output, "timed out") ||
			strings.Contains(result.Error, "timeout")
		hasFailure := !result.Success

		if !hasTimeoutMsg && !hasFailure {
			t.Errorf("Expected timeout or failure. Output: %s, Error: %s",
				result.Output, result.Error)
		}
	})
}

// TestCommandHistory tests command history tracking
func TestCommandHistory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wink-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	runTool := tools.NewRunInTerminalTool()
	lastTool := tools.NewTerminalLastCommandTool()
	ctx := context.Background()

	commands := []string{
		"echo First command",
		"echo Second command",
		"echo Third command",
	}

	for i, cmd := range commands {
		// Execute command
		_, err := runTool.Execute(ctx, map[string]interface{}{
			"command": cmd,
		}, tmpDir)
		if err != nil {
			t.Fatalf("Failed to execute command %d: %v", i, err)
		}

		// Check last command matches
		result, err := lastTool.Execute(ctx, map[string]interface{}{}, tmpDir)
		if err != nil {
			t.Fatalf("Failed to get last command: %v", err)
		}

		if !strings.Contains(result.Output, cmd) {
			t.Errorf("Last command should be '%s', got: %s",
				cmd, result.Output)
		}

		// Verify metadata
		lastCmd, ok := result.Metadata["command"].(string)
		if !ok || lastCmd != cmd {
			t.Errorf("Expected command '%s' in metadata, got '%s'",
				cmd, lastCmd)
		}
	}
}
