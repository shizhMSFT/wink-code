// Package main is the entry point for the wink CLI
package main

import (
	"fmt"
	"os"

	"github.com/shizhMSFT/wink-code/internal/agent"
	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/internal/tools"
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

	// Get configuration (use defaults for now, TODO: load from config file)
	ollamaURL := os.Getenv("WINK_OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	model := modelFlag
	if envModel := os.Getenv("WINK_MODEL"); envModel != "" {
		model = envModel
	}

	timeoutSeconds := 30

	// Create agent
	agentInstance, err := agent.NewAgent(ollamaURL, model, timeoutSeconds)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	// Register tools
	if err := registerTools(agentInstance); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	// Run agent
	ctx := cmd.Context()
	if err := agentInstance.Run(ctx, promptFlag, workingDir, continueFlag); err != nil {
		return fmt.Errorf("agent execution failed: %w", err)
	}

	return nil
}

// registerTools registers all available tools with the agent
func registerTools(a *agent.Agent) error {
	// Register create_file tool
	createFile := tools.NewCreateFileTool()
	if err := a.RegisterTool(createFile); err != nil {
		return fmt.Errorf("failed to register create_file tool: %w", err)
	}

	logging.Debug("Registered tools", "count", 1)

	return nil
}
