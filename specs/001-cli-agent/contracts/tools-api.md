# Tool API Contracts

**Feature**: 001-cli-agent  
**Created**: 2025-11-10  
**Purpose**: Define interfaces and contracts for all tool implementations

## Overview

This document specifies the interface contracts for all 10 built-in tools. Each tool follows a consistent pattern for registration, validation, execution, and error handling.

## Universal Tool Interface

All tools MUST implement this interface:

```go
type Tool interface {
    // Name returns the unique identifier for this tool
    Name() string
    
    // Description returns what this tool does (for LLM and users)
    Description() string
    
    // ParametersSchema returns JSON Schema defining parameters
    ParametersSchema() map[string]interface{}
    
    // Validate checks if parameters are valid before execution
    Validate(params map[string]interface{}, workingDir string) error
    
    // Execute runs the tool and returns result
    Execute(params map[string]interface{}, workingDir string) (ToolResult, error)
    
    // RequiresApproval returns true if this tool needs user approval
    RequiresApproval() bool
    
    // RiskLevel returns the risk category of this tool
    RiskLevel() RiskLevel
}
```

## Tool Catalog

### 1. create_file

**Purpose**: Create a new file with specified content

**Parameters**:
```json
{
  "path": {
    "type": "string",
    "description": "Relative path to the file to create",
    "required": true
  },
  "content": {
    "type": "string",
    "description": "Content to write to the file",
    "required": true
  }
}
```

**Validation Rules**:
- `path` MUST be within working directory
- `path` MUST NOT already exist (fail if file exists)
- `path` directory structure will be created if needed
- `content` size limited to 10MB

**Success Response**:
```json
{
  "success": true,
  "output": "Created file: hello.py (125 bytes)",
  "files_affected": ["hello.py"],
  "execution_time_ms": 5
}
```

**Error Cases**:
- Path outside working directory: "Error: Path '../etc/passwd' is outside working directory"
- File already exists: "Error: File 'test.txt' already exists. Use replace_string_in_file to modify"
- Permission denied: "Error: Cannot create file 'protected.txt' - permission denied"

**Risk Level**: safe_write

---

### 2. read_file

**Purpose**: Read the contents of a file, optionally specifying line range

**Parameters**:
```json
{
  "path": {
    "type": "string",
    "description": "Relative path to the file to read",
    "required": true
  },
  "start_line": {
    "type": "integer",
    "description": "Starting line number (1-indexed, optional)",
    "required": false
  },
  "end_line": {
    "type": "integer",
    "description": "Ending line number (1-indexed, inclusive, optional)",
    "required": false
  }
}
```

**Validation Rules**:
- `path` MUST be within working directory
- `path` MUST exist and be a regular file
- If `start_line` specified, it MUST be positive
- If `end_line` specified, it MUST be >= `start_line`
- File size limited to 10MB (larger files truncated with warning)

**Success Response**:
```json
{
  "success": true,
  "output": "Contents of hello.py (lines 1-10):\n...",
  "files_affected": [],
  "execution_time_ms": 2,
  "metadata": {
    "total_lines": 50,
    "lines_returned": 10,
    "file_size_bytes": 1250
  }
}
```

**Error Cases**:
- File not found: "Error: File 'missing.txt' not found"
- Line range invalid: "Error: Line range invalid - file has 20 lines, requested 30-40"
- Binary file: "Warning: File appears to be binary, displaying first 1KB as hex"

**Risk Level**: read_only

---

### 3. replace_string_in_file

**Purpose**: Replace a specific string in a file with new content

**Parameters**:
```json
{
  "path": {
    "type": "string",
    "description": "Relative path to the file to modify",
    "required": true
  },
  "old_string": {
    "type": "string",
    "description": "Exact string to find and replace",
    "required": true
  },
  "new_string": {
    "type": "string",
    "description": "String to replace with",
    "required": true
  }
}
```

**Validation Rules**:
- `path` MUST be within working directory
- `path` MUST exist
- `old_string` MUST NOT be empty
- `old_string` MUST exist in file (fail if not found)
- If multiple matches, only replace first occurrence

**Success Response**:
```json
{
  "success": true,
  "output": "Replaced 1 occurrence in hello.py",
  "files_affected": ["hello.py"],
  "execution_time_ms": 3,
  "metadata": {
    "occurrences_found": 1,
    "occurrences_replaced": 1,
    "lines_changed": [5]
  }
}
```

**Error Cases**:
- String not found: "Error: String 'old_text' not found in file"
- Multiple matches: "Warning: Found 3 occurrences, replaced only the first at line 5"

**Risk Level**: dangerous

---

### 4. list_dir

**Purpose**: List contents of a directory

**Parameters**:
```json
{
  "path": {
    "type": "string",
    "description": "Relative path to directory (default: current directory)",
    "required": false,
    "default": "."
  }
}
```

**Validation Rules**:
- `path` MUST be within working directory
- `path` MUST exist and be a directory
- Results limited to 1000 entries (paginated if more)

**Success Response**:
```json
{
  "success": true,
  "output": "Contents of ./src:\n  app.py\n  utils.py\n  models/\n",
  "files_affected": [],
  "execution_time_ms": 2,
  "metadata": {
    "total_entries": 3,
    "files": 2,
    "directories": 1
  }
}
```

**Error Cases**:
- Not a directory: "Error: Path 'file.txt' is not a directory"
- Permission denied: "Error: Cannot list directory - permission denied"

**Risk Level**: read_only

---

### 5. file_search

**Purpose**: Search for files matching a glob pattern

**Parameters**:
```json
{
  "pattern": {
    "type": "string",
    "description": "Glob pattern (e.g., '**/*.py', 'src/**/*.go')",
    "required": true
  },
  "base_path": {
    "type": "string",
    "description": "Base directory to search from (default: current directory)",
    "required": false,
    "default": "."
  }
}
```

**Validation Rules**:
- `base_path` MUST be within working directory
- `pattern` MUST be valid glob syntax
- Results limited to 1000 files
- Search depth limited to 20 levels

**Success Response**:
```json
{
  "success": true,
  "output": "Found 5 files matching '**/*.go':\n  cmd/main.go\n  internal/agent.go\n  ...",
  "files_affected": [],
  "execution_time_ms": 15,
  "metadata": {
    "matches": 5,
    "pattern": "**/*.go"
  }
}
```

**Error Cases**:
- Invalid pattern: "Error: Invalid glob pattern '**/*[.go'"
- Too many results: "Warning: Found 2000 matches, showing first 1000"

**Risk Level**: read_only

---

### 6. grep_search

**Purpose**: Search file contents for text or regex pattern

**Parameters**:
```json
{
  "pattern": {
    "type": "string",
    "description": "Text or regex pattern to search for",
    "required": true
  },
  "is_regex": {
    "type": "boolean",
    "description": "Whether pattern is regex (default: false)",
    "required": false,
    "default": false
  },
  "file_pattern": {
    "type": "string",
    "description": "Glob pattern to limit files searched (default: all files)",
    "required": false
  },
  "max_results": {
    "type": "integer",
    "description": "Maximum number of results to return (default: 100)",
    "required": false,
    "default": 100
  }
}
```

**Validation Rules**:
- `pattern` MUST NOT be empty
- If `is_regex` true, pattern MUST be valid regex
- `max_results` MUST be between 1 and 1000
- Search limited to text files (skip binary files)

**Success Response**:
```json
{
  "success": true,
  "output": "Found 3 matches for 'TODO':\n  app.py:15: # TODO: implement\n  ...",
  "files_affected": [],
  "execution_time_ms": 25,
  "metadata": {
    "total_matches": 3,
    "files_searched": 10,
    "pattern": "TODO"
  }
}
```

**Error Cases**:
- Invalid regex: "Error: Invalid regex pattern: unclosed group"
- Timeout: "Warning: Search timed out after 30s, returning partial results"

**Risk Level**: read_only

---

### 7. run_in_terminal

**Purpose**: Execute a shell command

**Parameters**:
```json
{
  "command": {
    "type": "string",
    "description": "Shell command to execute",
    "required": true
  },
  "timeout_seconds": {
    "type": "integer",
    "description": "Timeout in seconds (default: 30)",
    "required": false,
    "default": 30
  }
}
```

**Validation Rules**:
- `command` MUST NOT be empty
- `timeout_seconds` MUST be between 1 and 300
- Command executed in user's default shell
- Working directory set to session working directory

**Success Response**:
```json
{
  "success": true,
  "output": "Command output:\n<stdout content>\n",
  "files_affected": [],
  "execution_time_ms": 150,
  "metadata": {
    "exit_code": 0,
    "stdout_lines": 5,
    "stderr_lines": 0
  }
}
```

**Error Cases**:
- Command not found: "Error: Command 'invalid_cmd' not found"
- Timeout: "Error: Command timed out after 30 seconds"
- Non-zero exit: "Error: Command failed with exit code 1:\n<stderr>"

**Risk Level**: dangerous

---

### 8. terminal_last_command

**Purpose**: Retrieve the last shell command that was executed

**Parameters**: None

**Validation Rules**: None

**Success Response**:
```json
{
  "success": true,
  "output": "Last command: git status\nOutput: ...",
  "files_affected": [],
  "execution_time_ms": 1,
  "metadata": {
    "command": "git status",
    "exit_code": 0,
    "executed_at": "2025-11-10T10:05:00Z"
  }
}
```

**Error Cases**:
- No command history: "Error: No previous command in this session"

**Risk Level**: read_only

---

### 9. create_directory

**Purpose**: Create a directory structure (equivalent to mkdir -p)

**Parameters**:
```json
{
  "path": {
    "type": "string",
    "description": "Relative path to directory to create",
    "required": true
  }
}
```

**Validation Rules**:
- `path` MUST be within working directory
- Creates parent directories as needed
- No error if directory already exists

**Success Response**:
```json
{
  "success": true,
  "output": "Created directory: src/models",
  "files_affected": ["src/models"],
  "execution_time_ms": 3,
  "metadata": {
    "created_parents": ["src"]
  }
}
```

**Error Cases**:
- Path is a file: "Error: Path 'test.txt' exists as a file, cannot create directory"
- Permission denied: "Error: Cannot create directory - permission denied"

**Risk Level**: safe_write

---

### 10. fetch_webpage

**Purpose**: Fetch content from a web page

**Parameters**:
```json
{
  "url": {
    "type": "string",
    "description": "URL to fetch (must be http or https)",
    "required": true
  },
  "timeout_seconds": {
    "type": "integer",
    "description": "Request timeout (default: 10)",
    "required": false,
    "default": 10
  }
}
```

**Validation Rules**:
- `url` MUST start with http:// or https://
- `url` MUST be valid URL format
- `timeout_seconds` MUST be between 1 and 60
- Response size limited to 1MB

**Success Response**:
```json
{
  "success": true,
  "output": "Fetched content from https://example.com (5.2 KB)",
  "files_affected": [],
  "execution_time_ms": 250,
  "metadata": {
    "url": "https://example.com",
    "status_code": 200,
    "content_length": 5324,
    "content_type": "text/html"
  }
}
```

**Error Cases**:
- Invalid URL: "Error: Invalid URL format"
- Network error: "Error: Failed to fetch - connection refused"
- Timeout: "Error: Request timed out after 10 seconds"
- Too large: "Error: Response too large (>1MB), truncating"

**Risk Level**: dangerous

---

## Common Patterns

### Error Handling

All tools MUST return errors in this format:

```go
type ToolError struct {
    Code        string  // machine-readable code (e.g., "path_outside_working_dir")
    Message     string  // user-friendly message
    Suggestion  string  // what user should do to fix
}
```

Example:
```go
return ToolError{
    Code:       "file_not_found",
    Message:    "File 'missing.txt' not found in working directory",
    Suggestion: "Check the file name and try again, or use file_search to find it",
}
```

### Approval Display

When prompting for approval, show:
```
Tool: create_file
Parameters:
  path: hello.py
  content: <20 bytes>
Files affected: hello.py
Risk level: safe_write

Approve? (y/n/always):
```

### Auto-Approval Matching

For a tool call to match an approval rule:
1. Tool name must exactly match
2. Serialized parameters (JSON) must match regex pattern
3. Rule must not be expired or disabled

Example match:
```
Rule pattern: ^{"path":".*\\.txt"}$
Tool call:    {"path":"test.txt"}
Result:       ✅ Match - auto-approve
```

### Output Formatting

Human-readable format (default):
```
✓ Created file: hello.py (125 bytes)
  Location: /home/user/project/hello.py
  Time: 5ms
```

JSON format (with --json flag):
```json
{
  "tool": "create_file",
  "success": true,
  "output": "Created file: hello.py (125 bytes)",
  "files_affected": ["hello.py"],
  "execution_time_ms": 5
}
```

## LLM Function Calling Format

Tools are exposed to the LLM via OpenAI function calling format:

```json
{
  "name": "create_file",
  "description": "Create a new file with specified content",
  "parameters": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "Relative path to the file to create"
      },
      "content": {
        "type": "string",
        "description": "Content to write to the file"
      }
    },
    "required": ["path", "content"]
  }
}
```

## Testing Contracts

Each tool MUST have test coverage for:
- ✅ Successful execution with valid parameters
- ✅ Path validation (outside working directory)
- ✅ Parameter validation (missing, invalid types)
- ✅ Error conditions specific to that tool
- ✅ Edge cases (empty files, large files, special characters)
- ✅ Cross-platform behavior (Windows vs Unix paths)

Minimum test scenarios per tool: 10
Target code coverage per tool: ≥95%
