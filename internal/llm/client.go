// Package llm handles LLM API communication
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/shizhMSFT/wink-code/internal/logging"
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// Client wraps the OpenAI client for Ollama
type Client struct {
	client  *openai.Client
	model   string
	timeout time.Duration
}

// NewClient creates a new LLM client pointing to Ollama
func NewClient(baseURL, model string, timeoutSeconds int) *Client {
	config := openai.DefaultConfig("ollama") // Ollama doesn't require real API key
	config.BaseURL = baseURL + "/v1"

	return &Client{
		client:  openai.NewClientWithConfig(config),
		model:   model,
		timeout: time.Duration(timeoutSeconds) * time.Second,
	}
}

// ChatCompletion sends a chat completion request with tool support
func (c *Client) ChatCompletion(ctx context.Context, messages []types.Message, tools []types.Tool) (*openai.ChatCompletionResponse, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Convert messages to OpenAI format
	openaiMessages := make([]openai.ChatCompletionMessage, 0, len(messages))
	for _, msg := range messages {
		openaiMsg := openai.ChatCompletionMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}

		// Add tool calls if present
		if len(msg.ToolCalls) > 0 {
			openaiMsg.ToolCalls = make([]openai.ToolCall, 0, len(msg.ToolCalls))
			for _, tc := range msg.ToolCalls {
				openaiMsg.ToolCalls = append(openaiMsg.ToolCalls, openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.ToolName,
						Arguments: mustMarshalJSON(tc.Parameters),
					},
				})
			}
		}

		openaiMessages = append(openaiMessages, openaiMsg)
	}

	// Convert tools to OpenAI format
	openaiTools := make([]openai.Tool, 0, len(tools))
	for _, tool := range tools {
		openaiTools = append(openaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.ParametersSchema(),
			},
		})
	}

	// Log request
	logging.Debug("LLM API request",
		"model", c.model,
		"message_count", len(openaiMessages),
		"tool_count", len(openaiTools),
	)

	// Create request
	req := openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	// Send request
	startTime := time.Now()
	resp, err := c.client.CreateChatCompletion(ctx, req)
	duration := time.Since(startTime)

	if err != nil {
		logging.Error("LLM API error",
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}

	// Log response
	logging.Debug("LLM API response",
		"duration_ms", duration.Milliseconds(),
		"completion_tokens", resp.Usage.CompletionTokens,
		"prompt_tokens", resp.Usage.PromptTokens,
		"total_tokens", resp.Usage.TotalTokens,
	)

	return &resp, nil
}

// Model returns the model name being used
func (c *Client) Model() string {
	return c.model
}

// mustMarshalJSON marshals to JSON string, panics on error (should never happen with valid data)
func mustMarshalJSON(v interface{}) string {
	// OpenAI SDK expects JSON string for function arguments
	data, err := json.Marshal(v)
	if err != nil {
		// This should never happen with valid tool parameters
		logging.Error("Failed to marshal tool parameters to JSON", "error", err)
		return "{}"
	}
	return string(data)
}
