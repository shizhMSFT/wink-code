// Package main is the entry point for the wink CLI
package main

import (
	"fmt"
	"os"

	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/spf13/cobra"
)

var (
	// Version is set during build
	Version = "dev"
	// BuildTime is set during build
	BuildTime = "unknown"
)

var (
	promptFlag   string
	modelFlag    string
	continueFlag bool
	debugFlag    bool
)

func main() {
	// Root command
	rootCmd := &cobra.Command{
		Use:   "wink",
		Short: "AI coding agent for quick script generation",
		Long: `Wink is a lightweight CLI coding agent that connects to local LLMs 
(via Ollama) for rapid script generation and coding assistance.

It provides file operations, code search, command execution, and web 
integration with a safe approval workflow.`,
		Version: fmt.Sprintf("%s (built %s)", Version, BuildTime),
		RunE:    run,
	}

	// Flags
	rootCmd.Flags().StringVarP(&promptFlag, "prompt", "p", "", "Natural language prompt (required)")
	rootCmd.Flags().StringVarP(&modelFlag, "model", "m", "qwen3:8b", "LLM model to use")
	rootCmd.Flags().BoolVar(&continueFlag, "continue", false, "Continue previous session")
	rootCmd.Flags().BoolVarP(&debugFlag, "debug", "d", false, "Enable verbose debug logging")

	// Mark prompt as required (unless --continue is used)
	rootCmd.MarkFlagRequired("prompt")

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Initialize logger
	logging.InitLogger(debugFlag)

	logging.Info("Wink CLI starting", "version", Version)

	// Validate flags
	if !continueFlag && promptFlag == "" {
		return fmt.Errorf("--prompt/-p is required unless --continue is specified")
	}

	// Get working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	logging.Debug("Working directory", "path", workingDir)

	// TODO: Implement agent logic
	// This is where we'll:
	// 1. Load/create session
	// 2. Initialize LLM client
	// 3. Register tools
	// 4. Run agent orchestration loop
	// 5. Handle approvals and tool execution

	fmt.Fprintln(os.Stderr, "Wink CLI is not yet fully implemented.")
	fmt.Fprintf(os.Stderr, "Prompt: %s\n", promptFlag)
	fmt.Fprintf(os.Stderr, "Model: %s\n", modelFlag)
	fmt.Fprintf(os.Stderr, "Continue: %v\n", continueFlag)
	fmt.Fprintf(os.Stderr, "Debug: %v\n", debugFlag)
	fmt.Fprintf(os.Stderr, "Working Dir: %s\n", workingDir)

	return nil
}
