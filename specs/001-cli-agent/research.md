# Research Document: Wink CLI Coding Agent

**Feature**: 001-cli-agent  
**Created**: 2025-11-10  
**Purpose**: Technology decisions, best practices, and implementation patterns

## Overview

This document captures research findings and technical decisions for implementing the Wink CLI coding agent in Go, connecting to local Ollama server via OpenAI-compatible API.

## Technology Decisions

### 1. Programming Language: Go 1.22+

**Decision**: Use Go 1.22 or later for implementation

**Rationale**:
- **Cross-platform**: Single codebase compiles to native binaries for Windows, macOS, Linux
- **Performance**: Fast startup times (<500ms requirement), low memory footprint, efficient concurrency
- **Standard library**: Rich stdlib includes HTTP client, JSON, file I/O, command execution
- **CLI ecosystem**: Mature libraries (cobra, viper) with consistent patterns
- **Type safety**: Strong static typing prevents entire classes of bugs
- **Deployment**: Single binary distribution, no runtime dependencies

**Alternatives Considered**:
- Python: Slower startup, requires runtime, harder to distribute
- Rust: Steeper learning curve, longer compile times, less mature CLI ecosystem
- Node.js: Runtime dependency, larger memory footprint, inconsistent cross-platform behavior

### 2. CLI Framework: Cobra + Viper

**Decision**: Use cobra for command structure and viper for configuration

**Rationale**:
- **Industry standard**: Used by kubectl, Hugo, GitHub CLI, and thousands of tools
- **Flag management**: Built-in support for short/long flags (-p/--prompt, -m/--model)
- **Subcommands**: Easy to extend with future commands (wink init, wink config, etc.)
- **Help generation**: Automatic help text and usage documentation
- **Configuration**: Viper integrates seamlessly for config file + env vars + flags
- **POSIX compliance**: Follows standard CLI conventions

**Implementation Pattern**:
```go
rootCmd := &cobra.Command{
    Use:   "wink",
    Short: "AI coding agent for quick script generation",
}
rootCmd.PersistentFlags().StringP("prompt", "p", "", "Natural language prompt")
rootCmd.PersistentFlags().StringP("model", "m", "qwen3:8b", "LLM model to use")
rootCmd.PersistentFlags().Bool("continue", false, "Continue previous session")
```

### 3. LLM Integration: OpenAI-Compatible SDK

**Decision**: Use OpenAI Go SDK with custom base URL for Ollama

**Rationale**:
- **Standard protocol**: Ollama implements OpenAI-compatible API endpoints
- **Function calling**: OpenAI SDK supports tool/function calling (required for our 10 tools)
- **Maintained**: Official SDK with regular updates and bug fixes
- **Flexibility**: Easy to swap between Ollama, OpenAI, or other compatible providers
- **Streaming**: Built-in support for streaming responses (useful for interactive mode)

**Configuration**:
```go
// Point OpenAI SDK to local Ollama
client := openai.NewClient(
    openai.WithBaseURL("http://localhost:11434/v1"),
    openai.WithAPIKey("ollama"), // Ollama doesn't require real key
)
```

**Alternatives Considered**:
- Custom HTTP client: More control but reinventing the wheel, no streaming support
- Ollama Go library: Less flexible, ties us to Ollama-only deployment

### 4. Session Management: JSON File-Based Storage

**Decision**: Store sessions as JSON files in ~/.wink/sessions/

**Rationale**:
- **Simplicity**: No database setup, portable across systems
- **Human-readable**: Users can inspect/debug sessions manually
- **Version control friendly**: Can be committed to git if desired
- **Privacy**: Local storage, no cloud sync (aligns with security model)
- **Performance**: Fast for small session histories (target: 100 interactions max)

**Session Schema**:
```go
type Session struct {
    ID           string                 `json:"id"`
    WorkingDir   string                 `json:"working_dir"`
    Model        string                 `json:"model"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
    Messages     []Message              `json:"messages"`
    ToolResults  []ToolResult           `json:"tool_results"`
}
```

**Pruning Strategy**: Keep last 50 messages in context, archive older messages to separate file

### 5. Tool Approval System: Interactive Prompts + Regex Rules

**Decision**: Two-tier approval system with user prompts and persistent auto-approval rules

**Rationale**:
- **Safety first**: Explicit approval prevents unintended file modifications
- **Productivity**: Auto-approval eliminates repetitive prompts for trusted operations
- **Transparency**: Show exactly what will be executed before doing it
- **Flexibility**: Regex patterns allow fine-grained control

**Approval Flow**:
1. Agent proposes tool call with parameters
2. Check auto-approval rules (regex match on tool + params)
3. If no match, prompt user: "Execute [tool] [params]? (y/n/always)"
4. If "always", generate regex from params and save to config
5. Execute tool and return result to LLM

**Config Schema**:
```go
type ApprovalRule struct {
    ToolName       string `json:"tool_name"`
    ParamPattern   string `json:"param_pattern"` // regex
    Description    string `json:"description"`
    CreatedAt      time.Time `json:"created_at"`
}
```

**Example Rules**:
- Read all .txt files: `{"tool_name": "read_file", "param_pattern": ".*\\.txt$"}`
- List any directory: `{"tool_name": "list_dir", "param_pattern": ".*"}`

### 6. Path Security: Working Directory Jail

**Decision**: Validate all file paths against working directory before any operation

**Rationale**:
- **Security**: Prevent LLM from accessing/modifying files outside project
- **User trust**: Critical for adoption - users need confidence their system is safe
- **Explicit boundary**: Working directory is the clear, user-visible security boundary

**Implementation**:
```go
func ValidatePath(workingDir, requestedPath string) error {
    // Resolve to absolute path
    absRequested, err := filepath.Abs(requestedPath)
    if err != nil {
        return err
    }
    absWorking, _ := filepath.Abs(workingDir)
    
    // Check if requested path is within working dir
    relPath, err := filepath.Rel(absWorking, absRequested)
    if err != nil || strings.HasPrefix(relPath, "..") {
        return fmt.Errorf("path outside working directory")
    }
    return nil
}
```

**Edge Cases Handled**:
- Symlinks: Resolve before validation
- Relative paths with ..: Normalize before check
- Case sensitivity: Use case-preserving comparison on all platforms

### 7. Error Handling: Wrapped Errors with User-Friendly Messages

**Decision**: Use Go's error wrapping with custom user-facing messages

**Rationale**:
- **Constitution compliance**: No raw stack traces to users
- **Debuggability**: Full context preserved for developers
- **Actionable**: Every error includes suggested next steps

**Pattern**:
```go
// Internal error with full context
if err != nil {
    return fmt.Errorf("failed to read file %s: %w", path, err)
}

// User-facing error formatting
func FormatUserError(err error) string {
    // Extract user-friendly message and suggest action
    return fmt.Sprintf("Error: %s\nTry: %s", userMsg, suggestion)
}
```

### 8. Testing Strategy: Table-Driven Tests + Mocks

**Decision**: Use Go's table-driven test pattern with interface-based mocking

**Rationale**:
- **Coverage**: Easy to test many scenarios per function
- **Maintainability**: Test cases clearly separated from logic
- **Isolation**: Mock external dependencies (LLM API, filesystem, shell)
- **Speed**: Unit tests complete in milliseconds

**Example Structure**:
```go
func TestToolExecution(t *testing.T) {
    tests := []struct {
        name        string
        tool        string
        params      map[string]string
        mockResponse string
        want        ToolResult
        wantErr     bool
    }{
        {name: "create_file success", tool: "create_file", ...},
        {name: "create_file path escape", tool: "create_file", wantErr: true, ...},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### 9. Performance Optimization: Lazy Loading + Streaming

**Decision**: Implement lazy loading for large operations and streaming for LLM responses

**Rationale**:
- **Startup time**: Don't load unnecessary resources at CLI launch
- **Memory**: Stream large files instead of loading into memory
- **UX**: Show partial results as they arrive (streaming LLM responses)
- **Token limits**: Only send relevant file content to LLM

**Strategies**:
- **File reading**: Stream in chunks for files >10MB
- **Directory listing**: Paginate results for directories with >1000 files
- **LLM context**: Prune old messages, summarize when approaching token limit
- **Config loading**: Load on-demand, cache in memory

### 10. Cross-Platform Shell Execution

**Decision**: Use exec.Command with platform-specific shell detection

**Rationale**:
- **User expectations**: Commands should work as typed in user's shell
- **Platform differences**: cmd.exe vs bash vs PowerShell have different syntax
- **Environment**: Preserve user's environment variables and PATH

**Implementation**:
```go
func ExecuteCommand(cmdStr string) (string, error) {
    var cmd *exec.Cmd
    
    switch runtime.GOOS {
    case "windows":
        // Detect PowerShell vs cmd
        if hasPowerShell() {
            cmd = exec.Command("powershell", "-Command", cmdStr)
        } else {
            cmd = exec.Command("cmd", "/C", cmdStr)
        }
    default:
        // Unix-like systems
        shell := os.Getenv("SHELL")
        if shell == "" {
            shell = "/bin/sh"
        }
        cmd = exec.Command(shell, "-c", cmdStr)
    }
    
    // Set working directory, inherit environment
    cmd.Dir = workingDir
    cmd.Env = os.Environ()
    
    return cmd.CombinedOutput()
}
```

## Best Practices Applied

### Go Project Organization
- **internal/ packages**: Enforce encapsulation, prevent external imports
- **pkg/ types**: Shared types that could be reused
- **cmd/ entry points**: Clean separation of CLI from business logic
- **Flat structure**: Avoid deep nesting, keep related code together

### Configuration Management
- **Precedence**: CLI flags > Environment vars > Config file > Defaults
- **Validation**: Validate on load, fail fast with clear messages
- **Defaults**: Sensible defaults for all settings (qwen3:8b model)
- **Discovery**: Search standard locations (~/.wink/, .wink.json in cwd)

### CLI UX Patterns
- **Progressive disclosure**: Simple usage first, advanced features discoverable
- **Consistent flags**: -p/--prompt, -m/--model follow conventions
- **Helpful errors**: Include command to fix issue in error message
- **Exit codes**: 0 for success, 1 for user error, 2 for system error

### Security Considerations
- **Input validation**: Sanitize all user input before passing to shell
- **Path traversal**: Strict working directory enforcement
- **API keys**: Support environment variables, never log sensitive data
- **Temp files**: Clean up on exit, use secure temp directories

## Implementation Priorities

### Phase 1 (P1 - MVP): Quick Script Generation
1. Basic CLI setup with cobra
2. Ollama client connection
3. Simple tool execution (create_file only)
4. Interactive approval prompts
5. Basic error handling

### Phase 2 (P2 - Production Ready): Safety & Configuration
1. All 10 tools implemented
2. Auto-approval system with config file
3. Path validation and security
4. Comprehensive error messages
5. Session persistence basics

### Phase 3 (P3+ - Enhanced): Advanced Features
1. Session continuation (--continue flag)
2. Interactive mode with REPL
3. Streaming responses
4. Performance optimization
5. Web content fetching

## Open Questions & Decisions Needed

None - all technical decisions resolved during research phase.

## References

- [Cobra Documentation](https://cobra.dev/)
- [Viper Configuration](https://github.com/spf13/viper)
- [OpenAI Go SDK](https://github.com/sashabaranov/go-openai)
- [Ollama API Compatibility](https://github.com/ollama/ollama/blob/main/docs/openai.md)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
