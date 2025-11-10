# Data Model: Wink CLI Coding Agent

**Feature**: 001-cli-agent  
**Created**: 2025-11-10  
**Purpose**: Define core entities, relationships, and state management

## Overview

This document defines the data structures and entities used throughout the Wink CLI agent. These are implementation-agnostic representations that will be realized as Go structs.

## Core Entities

### 1. Session

Represents a conversation session between the user and the LLM agent.

**Purpose**: Maintain conversation context, working directory, and execution history across multiple interactions

**Attributes**:
- `id`: Unique identifier (UUID)
- `working_dir`: Absolute path to the working directory (security boundary)
- `model`: LLM model name (e.g., "qwen3-coder:30b")
- `created_at`: Timestamp of session creation
- `updated_at`: Timestamp of last modification
- `messages`: Ordered list of conversation messages
- `tool_results`: History of tool executions
- `status`: Enum (active, paused, completed, errored)

**Validation Rules**:
- `working_dir` MUST be an absolute path
- `working_dir` MUST exist and be accessible
- `model` MUST NOT be empty
- `messages` array maintains chronological order
- Maximum 100 messages retained in active session (older messages archived)

**State Transitions**:
- Created → Active (on first message)
- Active → Paused (user exits with --continue available)
- Active → Completed (user explicitly ends session)
- Active → Errored (unrecoverable error occurs)

**Relationships**:
- Contains many Messages (1:N)
- Contains many ToolResults (1:N)

### 2. Message

Represents a single message in the conversation (from user or assistant).

**Purpose**: Capture conversation flow for LLM context and session replay

**Attributes**:
- `role`: Enum (user, assistant, system, tool)
- `content`: Message text content
- `timestamp`: When message was created
- `tool_calls`: Array of tool invocations (for assistant messages)
- `metadata`: Additional context (token count, model used, etc.)

**Validation Rules**:
- `role` MUST be one of: user, assistant, system, tool
- `content` MUST NOT be empty (except when tool_calls present)
- User messages MUST have only content (no tool_calls)
- Assistant messages MAY have content and/or tool_calls
- Tool messages MUST reference a tool_call_id

**Relationships**:
- Belongs to Session (N:1)
- May reference ToolCalls (1:N)

### 3. ToolCall

Represents a request from the LLM to execute a tool.

**Purpose**: Bridge between LLM function calling and actual tool execution

**Attributes**:
- `id`: Unique identifier for this tool call
- `tool_name`: Name of the tool to execute (e.g., "create_file")
- `parameters`: Key-value map of tool parameters
- `status`: Enum (pending, approved, rejected, executed, failed)
- `approval_method`: Enum (manual, auto, config_rule)
- `approval_rule_id`: Reference to auto-approval rule (if applicable)
- `requested_at`: Timestamp of LLM request
- `executed_at`: Timestamp of actual execution

**Validation Rules**:
- `tool_name` MUST match registered tool
- `parameters` MUST satisfy tool's parameter schema
- File path parameters MUST pass security validation
- Status transitions MUST follow: pending → (approved|rejected) → (executed|failed)

**State Transitions**:
- Pending → Approved (user says yes OR auto-approval matches)
- Pending → Rejected (user says no)
- Approved → Executed (tool completes successfully)
- Approved → Failed (tool errors)

**Relationships**:
- Belongs to Message (N:1)
- Produces one ToolResult (1:1)
- May reference ApprovalRule (N:1)

### 4. ToolResult

Represents the output from executing a tool.

**Purpose**: Capture tool execution results for LLM context and debugging

**Attributes**:
- `tool_call_id`: Reference to the ToolCall that produced this result
- `success`: Boolean indicating success/failure
- `output`: String output from tool execution
- `error`: Error message (if failed)
- `execution_time_ms`: Duration of execution
- `files_affected`: List of file paths modified/created
- `metadata`: Additional context (bytes written, lines changed, etc.)

**Validation Rules**:
- If `success` is true, `error` MUST be empty
- If `success` is false, `error` MUST NOT be empty
- `output` limited to 10KB (larger outputs truncated with summary)
- `files_affected` paths MUST be within working directory

**Relationships**:
- Belongs to ToolCall (1:1)
- Belongs to Session (N:1)

### 5. Config

Represents user configuration and preferences.

**Purpose**: Persist settings across sessions

**Attributes**:
- `default_model`: Default LLM model to use
- `ollama_base_url`: Base URL for Ollama API (default: http://localhost:11434)
- `api_timeout_seconds`: Timeout for LLM API calls (default: 30)
- `max_session_messages`: Maximum messages to keep in session (default: 100)
- `auto_approval_rules`: Array of approval rules
- `output_format`: Enum (human, json) for default output format
- `config_version`: Schema version for migrations

**Validation Rules**:
- `default_model` MUST NOT be empty
- `ollama_base_url` MUST be valid URL
- `api_timeout_seconds` MUST be between 5 and 300
- `max_session_messages` MUST be between 10 and 1000

**Storage Location**:
- Primary: `~/.wink/config.json`
- Override: `./.wink.json` in working directory (takes precedence)

### 6. ApprovalRule

Represents an auto-approval rule for tool operations.

**Purpose**: Enable workflow automation while maintaining safety

**Attributes**:
- `id`: Unique identifier (UUID)
- `tool_name`: Tool this rule applies to
- `param_pattern`: Regex pattern to match against serialized parameters
- `description`: Human-readable description of what this rule allows
- `created_at`: When rule was created
- `last_used_at`: Last time rule matched and auto-approved
- `use_count`: Number of times rule has been applied

**Validation Rules**:
- `tool_name` MUST be a registered tool
- `param_pattern` MUST be valid regex
- Pattern MUST NOT match all inputs (prevent blanket approvals)
- Description MUST NOT be empty

**Example Rules**:
```json
{
  "id": "uuid-123",
  "tool_name": "read_file",
  "param_pattern": "^{\"path\":\".*\\.txt\"}$",
  "description": "Auto-approve reading any .txt file",
  "use_count": 42
}
```

**Relationships**:
- Referenced by ToolCalls (1:N)
- Belongs to Config (N:1)

### 7. Tool

Represents a tool/capability available to the agent.

**Purpose**: Define tool interface and contract

**Attributes**:
- `name`: Unique tool identifier (e.g., "create_file")
- `description`: What the tool does (for LLM and users)
- `parameters_schema`: JSON Schema defining required/optional parameters
- `requires_approval`: Boolean (all tools require approval by default)
- `risk_level`: Enum (read_only, safe_write, dangerous)
- `examples`: Array of example invocations

**Validation Rules**:
- `name` MUST be unique across all tools
- `parameters_schema` MUST be valid JSON Schema
- All registered tools MUST implement the Tool interface

**Tool Catalog** (from spec):
1. `create_file`: Create new file with content
2. `read_file`: Read file contents with optional line range
3. `replace_string_in_file`: Edit file by replacing text
4. `list_dir`: List directory contents
5. `file_search`: Search files by glob pattern
6. `grep_search`: Search content with text/regex
7. `run_in_terminal`: Execute shell command
8. `terminal_last_command`: Get last command run
9. `create_directory`: Create directory structure
10. `fetch_webpage`: Fetch content from URL

**Risk Levels**:
- Read-only: read_file, list_dir, file_search, grep_search, terminal_last_command
- Safe write: create_file, create_directory
- Dangerous: replace_string_in_file, run_in_terminal, fetch_webpage

## Entity Relationships

```
Session (1) ─── (N) Message
                     │
                     └── (N) ToolCall (1) ─── (1) ToolResult
                                  │
                                  └── (1) ApprovalRule

Config (1) ─── (N) ApprovalRule

Tool (registry) ─── (N) ToolCall (references)
```

## State Diagrams

### Tool Execution Lifecycle

```
[LLM proposes ToolCall] 
         ↓
    [Status: Pending]
         ↓
    ┌────┴────┐
    ↓         ↓
[Check Auto-Approval Rules]
    ↓         ↓
 [Match]  [No Match]
    ↓         ↓
[Auto-Approve] [Prompt User]
    ↓         ↓
    └────┬────┘
         ↓
    ┌────┴────┐
    ↓         ↓
[Approved] [Rejected] → [Record rejection]
    ↓                        ↓
[Execute Tool]          [End: Not executed]
    ↓
┌───┴───┐
↓       ↓
[Success] [Error]
↓       ↓
[Create ToolResult with output]
↓       ↓
[Return to LLM for next step]
```

### Session State Flow

```
[User invokes wink -p "..."]
         ↓
    [Create Session]
         ↓
    [Status: Active]
         ↓
    [LLM processes prompt]
         ↓
    [Tool calls if needed]
         ↓
    [User continues or exits]
         ↓
    ┌────┴────┐
    ↓         ↓
[--continue] [Exit]
    ↓         ↓
[Status: Paused] [Status: Completed]
    ↓
[Save session to disk]
```

## Storage Schema

### Session File Format
Location: `~/.wink/sessions/{session-id}.json`

```json
{
  "id": "uuid",
  "working_dir": "/absolute/path",
  "model": "qwen3:8b",
  "created_at": "2025-11-10T10:00:00Z",
  "updated_at": "2025-11-10T10:05:00Z",
  "status": "active",
  "messages": [
    {
      "role": "user",
      "content": "create a hello world script",
      "timestamp": "2025-11-10T10:00:00Z"
    },
    {
      "role": "assistant",
      "content": "I'll create that for you.",
      "timestamp": "2025-11-10T10:00:01Z",
      "tool_calls": [
        {
          "id": "call_123",
          "tool_name": "create_file",
          "parameters": {"path": "hello.py", "content": "print('Hello')"}
        }
      ]
    },
    {
      "role": "tool",
      "content": "File created successfully",
      "timestamp": "2025-11-10T10:00:02Z",
      "tool_call_id": "call_123"
    }
  ],
  "tool_results": [
    {
      "tool_call_id": "call_123",
      "success": true,
      "output": "Created file: hello.py (20 bytes)",
      "execution_time_ms": 5,
      "files_affected": ["hello.py"]
    }
  ]
}
```

### Config File Format
Location: `~/.wink/config.json`

```json
{
  "config_version": "1.0",
  "default_model": "qwen3:8b",
  "ollama_base_url": "http://localhost:11434",
  "api_timeout_seconds": 30,
  "max_session_messages": 100,
  "output_format": "human",
  "auto_approval_rules": [
    {
      "id": "rule-1",
      "tool_name": "read_file",
      "param_pattern": ".*\\.txt$",
      "description": "Auto-approve reading .txt files",
      "created_at": "2025-11-10T09:00:00Z",
      "last_used_at": "2025-11-10T10:00:00Z",
      "use_count": 5
    }
  ]
}
```

## Data Validation & Constraints

### Path Validation
All file path parameters MUST pass these checks:
1. Resolve to absolute path
2. Must be within or under the working directory
3. No symlinks that escape working directory
4. Handle platform-specific path separators correctly

### Input Sanitization
- Shell commands: Escape special characters based on platform
- File content: Validate UTF-8 encoding for text files
- Regex patterns: Validate before compilation, catch invalid patterns

### Size Limits
- Tool output: 10KB maximum (truncate with summary)
- File content in context: 100KB maximum
- Session messages: 100 messages maximum (prune oldest)
- Config file: 1MB maximum

## Performance Considerations

### Caching Strategy
- Config: Load once on startup, cache in memory
- Sessions: Load on demand, keep active session in memory
- Tool registry: Initialize once, immutable during execution

### Indexing
- Sessions indexed by ID for O(1) lookup
- Approval rules indexed by tool_name for fast matching
- Session list sorted by updated_at for recency

### Memory Management
- Limit active session message count
- Archive old tool results to separate files
- Stream large file operations instead of loading into memory

## Migration Strategy

When data model changes:
1. Increment `config_version` in Config
2. Write migration function to transform old → new format
3. Auto-migrate on load with backup of original
4. Log migration actions for user visibility
