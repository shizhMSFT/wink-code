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
	timeoutFlag  int
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
	rootCmd.Flags().IntVar(&timeoutFlag, "timeout", 30, "LLM API timeout in seconds (default: 30s, min: 5s)")

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

	// Determine timeout with precedence: flag > env > default
	timeoutSeconds := timeoutFlag
	if envTimeout := os.Getenv("WINK_TIMEOUT"); envTimeout != "" && timeoutFlag == 30 {
		// Only use env var if flag was not explicitly set (still at default)
		var timeout int
		_, err := fmt.Sscanf(envTimeout, "%d", &timeout)
		if err == nil && timeout >= 5 {
			timeoutSeconds = timeout
		}
	}

	// Validate timeout
	if timeoutSeconds < 5 {
		return fmt.Errorf("timeout must be at least 5 seconds, got %d", timeoutSeconds)
	}
	if timeoutSeconds > 300 {
		logging.Warn("Timeout is very high", "timeout", timeoutSeconds, "recommended_max", 300)
	}

	logging.Debug("Configuration", "model", model, "timeout", timeoutSeconds, "ollama_url", ollamaURL)

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

	// Register read_file tool
	readFile := tools.NewReadFileTool()
	if err := a.RegisterTool(readFile); err != nil {
		return fmt.Errorf("failed to register read_file tool: %w", err)
	}

	// Register replace_string_in_file tool
	replaceString := tools.NewReplaceStringInFileTool()
	if err := a.RegisterTool(replaceString); err != nil {
		return fmt.Errorf("failed to register replace_string_in_file tool: %w", err)
	}

	// Register create_directory tool
	createDir := tools.NewCreateDirectoryTool()
	if err := a.RegisterTool(createDir); err != nil {
		return fmt.Errorf("failed to register create_directory tool: %w", err)
	}

	// Register list_dir tool
	listDir := tools.NewListDirTool()
	if err := a.RegisterTool(listDir); err != nil {
		return fmt.Errorf("failed to register list_dir tool: %w", err)
	}

	// Register file_search tool
	fileSearch := tools.NewFileSearchTool()
	if err := a.RegisterTool(fileSearch); err != nil {
		return fmt.Errorf("failed to register file_search tool: %w", err)
	}

	// Register grep_search tool
	grepSearch := tools.NewGrepSearchTool()
	if err := a.RegisterTool(grepSearch); err != nil {
		return fmt.Errorf("failed to register grep_search tool: %w", err)
	}

	logging.Debug("Registered tools", "count", 7)

	return nil
}
