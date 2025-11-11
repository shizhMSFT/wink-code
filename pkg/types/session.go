// Package types defines core types for session management
package types

import "time"

// SessionStatus represents the state of a session
type SessionStatus string

const (
	// SessionStatusActive - Session is currently running
	SessionStatusActive SessionStatus = "active"
	// SessionStatusPaused - Session paused, can be continued
	SessionStatusPaused SessionStatus = "paused"
	// SessionStatusCompleted - Session finished successfully
	SessionStatusCompleted SessionStatus = "completed"
	// SessionStatusErrored - Session ended with error
	SessionStatusErrored SessionStatus = "errored"
)

// MessageRole represents the role of a message sender
type MessageRole string

const (
	// MessageRoleUser - Message from user
	MessageRoleUser MessageRole = "user"
	// MessageRoleAssistant - Message from LLM assistant
	MessageRoleAssistant MessageRole = "assistant"
	// MessageRoleSystem - System message
	MessageRoleSystem MessageRole = "system"
	// MessageRoleTool - Tool execution result
	MessageRoleTool MessageRole = "tool"
)

// Session represents a conversation session
type Session struct {
	ID          string        `json:"id"`
	WorkingDir  string        `json:"working_dir"`
	Model       string        `json:"model"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Messages    []Message     `json:"messages"`
	ToolResults []ToolResult  `json:"tool_results"`
	Status      SessionStatus `json:"status"`
}

// Message represents a single message in the conversation
type Message struct {
	Role      MessageRole            `json:"role"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a request from the LLM to execute a tool
type ToolCall struct {
	ID         string                 `json:"id"`
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
}
