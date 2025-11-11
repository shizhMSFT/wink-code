// Package agent implements the core agent orchestration
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shizhMSFT/wink-code/internal/llm"
	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/internal/tools"
	"github.com/shizhMSFT/wink-code/internal/ui"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// Agent orchestrates the interaction between user, LLM, and tools
type Agent struct {
	llmClient        *llm.Client
	toolRegistry     *tools.Registry
	approvalWorkflow *tools.ApprovalWorkflow
	sessionManager   *SessionManager
	contextManager   *ContextManager
	baseURL          string
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// NewAgent creates a new agent instance
func NewAgent(baseURL, model string, timeoutSeconds int) (*Agent, error) {
	// Initialize components
	llmClient := llm.NewClient(baseURL, model, timeoutSeconds)

	toolRegistry := tools.NewRegistry()

	approvalWorkflow, err := tools.NewApprovalWorkflow()
	if err != nil {
		return nil, fmt.Errorf("failed to create approval workflow: %w", err)
	}

	sessionManager, err := NewSessionManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	contextManager := NewContextManager(100) // Max 100 messages

	return &Agent{
		llmClient:        llmClient,
		toolRegistry:     toolRegistry,
		approvalWorkflow: approvalWorkflow,
		sessionManager:   sessionManager,
		contextManager:   contextManager,
		baseURL:          baseURL,
	}, nil
}

// RegisterTool registers a tool with the agent
func (a *Agent) RegisterTool(tool types.Tool) error {
	return a.toolRegistry.Register(tool)
}

// Run executes the agent with a user prompt
func (a *Agent) Run(ctx context.Context, prompt string, workingDir string, continueSession bool) error {
	// Load or create session
	var session *types.Session
	var err error

	if continueSession {
		session, err = a.sessionManager.GetLatest()
		if err != nil {
			return fmt.Errorf("failed to load previous session: %w", err)
		}
		logging.Info("Continuing session", "session_id", session.ID)
		ui.PrintInfo(fmt.Sprintf("Continuing session: %s", session.ID[:8]))
	} else {
		session, err = a.sessionManager.Create(workingDir, a.llmClient.Model())
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		logging.Info("Created new session", "session_id", session.ID)
	}

	// Add user message
	userMessage := types.Message{
		Role:      types.MessageRoleUser,
		Content:   prompt,
		Timestamp: time.Now(),
	}
	a.contextManager.AddMessage(session, userMessage)

	// Agent loop
	maxIterations := 10 // Prevent infinite loops
	for iteration := 0; iteration < maxIterations; iteration++ {
		logging.Debug("Agent iteration", "iteration", iteration)

		// Get available tools
		availableTools := a.toolRegistry.GetAll()

		// Call LLM
		response, err := a.llmClient.ChatCompletion(ctx, session.Messages, availableTools)
		if err != nil {
			// User-friendly error messages for common issues
			if contains(err.Error(), "connection refused") || contains(err.Error(), "no such host") {
				return fmt.Errorf("unable to connect to LLM server at %s. Please ensure Ollama is running with 'ollama serve'",
					a.baseURL)
			}
			if contains(err.Error(), "model") && contains(err.Error(), "not found") {
				return fmt.Errorf("model '%s' not found. Try pulling it with: ollama pull %s",
					a.Model(), a.Model())
			}
			if contains(err.Error(), "timeout") || contains(err.Error(), "deadline exceeded") {
				return fmt.Errorf("LLM request timed out after 30 seconds. The server may be overloaded or the request too complex")
			}
			return fmt.Errorf("LLM request failed: %w\n\nTry:\n  - Ensure Ollama is running: ollama serve\n  - Check model is available: ollama list\n  - Use --debug flag for detailed logs", err)
		}

		// Check if we have a response
		if len(response.Choices) == 0 {
			return fmt.Errorf("no response from LLM")
		}

		choice := response.Choices[0]

		// Add assistant message
		assistantMessage := types.Message{
			Role:      types.MessageRoleAssistant,
			Content:   choice.Message.Content,
			Timestamp: time.Now(),
			ToolCalls: []types.ToolCall{},
		}

		// Check for tool calls
		if len(choice.Message.ToolCalls) > 0 {
			// Process tool calls
			for _, toolCall := range choice.Message.ToolCalls {
				// Parse parameters
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
					logging.Error("Failed to parse tool parameters", "error", err)
					continue
				}

				// Add to message
				assistantMessage.ToolCalls = append(assistantMessage.ToolCalls, types.ToolCall{
					ID:         toolCall.ID,
					ToolName:   toolCall.Function.Name,
					Parameters: params,
				})
			}
		}

		// Add assistant message to context
		a.contextManager.AddMessage(session, assistantMessage)

		// If no tool calls, we're done
		if len(assistantMessage.ToolCalls) == 0 {
			// Display final response
			if choice.Message.Content != "" {
				ui.PrintOutput(choice.Message.Content)
			}
			break
		}

		// Execute tool calls
		for _, toolCall := range assistantMessage.ToolCalls {
			result, err := a.executeToolCall(ctx, session, toolCall)
			if err != nil {
				logging.Error("Tool execution failed",
					"tool", toolCall.ToolName,
					"error", err,
				)
				// Add error result to context
				result = &types.ToolResult{
					ToolCallID: toolCall.ID,
					Success:    false,
					Error:      err.Error(),
				}
			}

			// Add tool result to context
			a.contextManager.AddToolResult(session, *result)

			// Add tool result message
			toolResultMessage := types.Message{
				Role:      types.MessageRoleTool,
				Content:   result.Output,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"tool_call_id": result.ToolCallID,
				},
			}
			a.contextManager.AddMessage(session, toolResultMessage)

			// Display result
			formatter := ui.NewFormatter(types.OutputFormatHuman)
			ui.PrintInfo(formatter.FormatToolResult(result))
		}

		// Save session after each iteration
		if err := a.sessionManager.Save(session); err != nil {
			logging.Warn("Failed to save session", "error", err)
		}
	}

	// Save final session
	session.Status = types.SessionStatusCompleted
	if err := a.sessionManager.Save(session); err != nil {
		logging.Warn("Failed to save final session", "error", err)
	}

	// Display session info
	formatter := ui.NewFormatter(types.OutputFormatHuman)
	ui.PrintInfo(formatter.FormatSessionInfo(session.ID, len(session.Messages)))

	return nil
}

// executeToolCall executes a single tool call with approval
func (a *Agent) executeToolCall(ctx context.Context, session *types.Session, toolCall types.ToolCall) (*types.ToolResult, error) {
	logging.Debug("Executing tool call",
		"tool", toolCall.ToolName,
		"tool_call_id", toolCall.ID,
	)

	// Check approval
	approved, autoApproved, ruleDescription, err := a.approvalWorkflow.CheckApproval(toolCall.ToolName, toolCall.Parameters)
	if err != nil {
		return nil, fmt.Errorf("approval check failed: %w", err)
	}

	if !approved {
		return &types.ToolResult{
			ToolCallID:      toolCall.ID,
			Success:         false,
			Error:           "Operation rejected by user",
			ExecutionTimeMs: 0,
		}, nil
	}

	// Display auto-approval notification
	if autoApproved {
		formatter := ui.NewFormatter(types.OutputFormatHuman)
		ui.PrintInfo(formatter.FormatAutoApproval(toolCall.ToolName, ruleDescription))
	}

	// Execute tool
	result, err := a.toolRegistry.Execute(ctx, toolCall.ToolName, toolCall.Parameters, session.WorkingDir)
	if err != nil {
		return &types.ToolResult{
			ToolCallID:      toolCall.ID,
			Success:         false,
			Error:           err.Error(),
			ExecutionTimeMs: 0,
		}, err
	}

	// Set tool call ID
	result.ToolCallID = toolCall.ID

	return result, nil
}

// Model returns the model being used
func (a *Agent) Model() string {
	return a.llmClient.Model()
}

// BaseURL returns the LLM base URL
func (a *Agent) BaseURL() string {
	return a.baseURL
}
