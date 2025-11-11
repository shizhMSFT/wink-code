// Package types defines approval-related types
package types

import "time"

// ApprovalStatus represents the state of a tool call approval
type ApprovalStatus string

const (
	// ApprovalStatusPending - Awaiting approval
	ApprovalStatusPending ApprovalStatus = "pending"
	// ApprovalStatusApproved - Approved for execution
	ApprovalStatusApproved ApprovalStatus = "approved"
	// ApprovalStatusRejected - Rejected by user
	ApprovalStatusRejected ApprovalStatus = "rejected"
	// ApprovalStatusExecuted - Successfully executed
	ApprovalStatusExecuted ApprovalStatus = "executed"
	// ApprovalStatusFailed - Execution failed
	ApprovalStatusFailed ApprovalStatus = "failed"
)

// ApprovalMethod indicates how approval was granted
type ApprovalMethod string

const (
	// ApprovalMethodManual - User manually approved
	ApprovalMethodManual ApprovalMethod = "manual"
	// ApprovalMethodAuto - Auto-approved by rule
	ApprovalMethodAuto ApprovalMethod = "auto"
	// ApprovalMethodConfigRule - Matched config rule
	ApprovalMethodConfigRule ApprovalMethod = "config_rule"
)

// ApprovalRule represents an auto-approval rule
type ApprovalRule struct {
	ID           string    `json:"id"`
	ToolName     string    `json:"tool_name"`
	ParamPattern string    `json:"param_pattern"` // regex
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at,omitempty"`
	UseCount     int       `json:"use_count"`
}

// ToolCallWithApproval extends ToolCall with approval tracking
type ToolCallWithApproval struct {
	ToolCall
	Status         ApprovalStatus `json:"status"`
	ApprovalMethod ApprovalMethod `json:"approval_method,omitempty"`
	ApprovalRuleID string         `json:"approval_rule_id,omitempty"`
	RequestedAt    time.Time      `json:"requested_at"`
	ExecutedAt     time.Time      `json:"executed_at,omitempty"`
}
