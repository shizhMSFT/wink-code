// Package agent handles conversation context management
package agent

import (
	"github.com/shizhMSFT/wink-code/pkg/types"
)

// ContextManager manages conversation context and message pruning
type ContextManager struct {
	maxMessages int
}

// NewContextManager creates a new context manager
func NewContextManager(maxMessages int) *ContextManager {
	return &ContextManager{
		maxMessages: maxMessages,
	}
}

// AddMessage adds a message to the session and prunes if needed
func (cm *ContextManager) AddMessage(session *types.Session, message types.Message) {
	session.Messages = append(session.Messages, message)

	// Prune if exceeds max
	if len(session.Messages) > cm.maxMessages {
		cm.PruneMessages(session)
	}
}

// PruneMessages keeps only the most recent messages
func (cm *ContextManager) PruneMessages(session *types.Session) {
	if len(session.Messages) <= cm.maxMessages {
		return
	}

	// Keep the last N messages
	startIndex := len(session.Messages) - cm.maxMessages
	session.Messages = session.Messages[startIndex:]
}

// GetMessages returns messages suitable for LLM context
func (cm *ContextManager) GetMessages(session *types.Session) []types.Message {
	return session.Messages
}

// AddToolResult adds a tool execution result to the session
func (cm *ContextManager) AddToolResult(session *types.Session, result types.ToolResult) {
	session.ToolResults = append(session.ToolResults, result)

	// Optionally prune old tool results (keep last 50)
	if len(session.ToolResults) > 50 {
		session.ToolResults = session.ToolResults[len(session.ToolResults)-50:]
	}
}

// GetContext returns a summary of current context state
func (cm *ContextManager) GetContext(session *types.Session) map[string]interface{} {
	return map[string]interface{}{
		"session_id":    session.ID,
		"message_count": len(session.Messages),
		"tool_results":  len(session.ToolResults),
		"working_dir":   session.WorkingDir,
		"model":         session.Model,
		"status":        session.Status,
	}
}
