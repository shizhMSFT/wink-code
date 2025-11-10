# Implementation Plan: Wink CLI Coding Agent

**Branch**: `001-cli-agent` | **Date**: 2025-11-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-cli-agent/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a lightweight CLI coding agent named `wink` that connects to local Ollama server via OpenAI-compatible API for rapid script generation and coding assistance. The tool provides 10 built-in tools (file operations, search, terminal commands, web fetching) with user approval workflow and auto-approval configuration. Supports both interactive and non-interactive modes with session continuation. Primary use case: developers generating and modifying code through natural language prompts without leaving the terminal.

## Technical Context

**Language/Version**: Go 1.25 (latest stable for robust CLI tooling and cross-platform support)  
**Primary Dependencies**: 
- Ollama Go SDK or OpenAI Go SDK (for LLM API communication)
- cobra (CLI framework for command/flag management)
- viper (configuration management)
- go-homedir (cross-platform home directory detection)

**Storage**: 
- Local filesystem for generated files and tool operations
- JSON configuration file (~/.wink/config.json) for settings and auto-approval rules
- JSON session file (~/.wink/sessions/) for conversation history and state

**Testing**: 
- Go's built-in testing framework (testing package)
- testify/assert for assertions
- httptest for mocking LLM API calls
- Integration tests for tool execution workflows

**Target Platform**: Cross-platform CLI (Windows, macOS, Linux) - compiled binaries for each
**Project Type**: Single project (standalone CLI application)
**Performance Goals**: 
- CLI startup time ≤500ms (cold start)
- Tool execution overhead <100ms (excluding actual operation time)
- LLM API calls with 30s timeout
- Memory footprint ≤500MB during typical usage

**Constraints**: 
- All file operations restricted to current working directory
- Must work offline except for LLM API calls
- Cross-platform path handling and shell command execution
- Graceful degradation when Ollama server unreachable

**Scale/Scope**: 
- Single-user local tool (no multi-user/server components)
- Session history limited to last 100 interactions per session
- Configuration file <1MB
- Support for projects up to 100k files in workspace

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Code Quality First
- [x] Linting and formatting tools configured for project language (golangci-lint, gofmt)
- [x] Type hints/annotations strategy defined for public interfaces (Go's type system enforces this)
- [x] Complexity thresholds enforceable (cyclomatic complexity ≤10) - can use gocyclo
- [x] Documentation standards established for public API (Go doc comments required)
- [x] Code review process includes quality verification (GitHub PR workflow)

### II. Testing Standards
- [x] TDD workflow confirmed: tests before implementation
- [x] Unit test coverage target ≥90% achievable for this feature (go test -cover)
- [x] Integration test scope identified for contracts/APIs (tool execution, LLM API interaction, approval workflow)
- [x] Edge case scenarios documented (see spec.md edge cases section)
- [x] Test isolation and determinism strategy defined (mock LLM responses, temp directories for file ops)

### III. User Experience Consistency
- [x] CLI command naming follows project conventions (cobra standard: wink -p, --prompt, -m, --model, --continue)
- [x] Error message patterns defined (actionable, no raw traces - wrap errors with context)
- [x] Input/output protocols respect stdin/stdout/stderr separation (prompts to stderr, data to stdout)
- [x] Interactive response time ≤2s achievable (local operations <100ms, network bounded by 30s timeout)
- [x] User documentation includes examples for all commands (quickstart.md will include all examples)

### IV. Performance Requirements
- [x] LLM API timeout and retry strategy defined (30s timeout, exponential backoff with 3 retries)
- [x] File operation efficiency considered (streaming for large files, lazy loading for directory traversal)
- [x] Memory footprint estimated and within ≤500MB target (Go's efficient memory model, session history limited)
- [x] Performance-critical paths identified for benchmarking (CLI startup, tool execution overhead, LLM request/response)
- [x] Token usage optimization planned (structured tool outputs, context window management, session pruning)

**Gate Status**: ✅ PASSED - All constitution requirements satisfied

## Project Structure

### Documentation (this feature)

```text
specs/001-cli-agent/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── tools-api.md     # Tool interface definitions
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
wink-code/
├── cmd/
│   └── wink/
│       └── main.go           # CLI entry point
├── internal/
│   ├── agent/
│   │   ├── agent.go          # Core agent orchestration
│   │   ├── session.go        # Session management
│   │   └── context.go        # Conversation context handling
│   ├── llm/
│   │   ├── client.go         # LLM API client (OpenAI-compatible)
│   │   ├── ollama.go         # Ollama-specific connection
│   │   └── retry.go          # Retry logic with exponential backoff
│   ├── tools/
│   │   ├── registry.go       # Tool registration and dispatch
│   │   ├── approval.go       # Approval workflow and auto-approval
│   │   ├── file.go           # File operation tools (create, read, replace)
│   │   ├── search.go         # Search tools (file_search, grep_search)
│   │   ├── directory.go      # Directory tools (list_dir, create_directory)
│   │   ├── terminal.go       # Terminal tools (run_in_terminal, terminal_last_command)
│   │   └── web.go            # Web tools (fetch_webpage)
│   ├── config/
│   │   ├── config.go         # Configuration loading/saving
│   │   └── approval.go       # Auto-approval rule management
│   └── ui/
│       ├── prompt.go         # User prompts and input handling
│       └── output.go         # Output formatting (JSON/human-readable)
├── pkg/
│   └── types/
│       ├── tool.go           # Tool operation types
│       ├── session.go        # Session state types
│       └── approval.go       # Approval rule types
├── tests/
│   ├── integration/
│   │   ├── agent_test.go     # End-to-end agent tests
│   │   ├── tools_test.go     # Tool execution tests
│   │   └── session_test.go   # Session persistence tests
│   └── unit/
│       ├── llm_test.go       # LLM client unit tests
│       ├── tools_test.go     # Individual tool unit tests
│       ├── approval_test.go  # Approval logic tests
│       └── config_test.go    # Config management tests
├── go.mod
├── go.sum
├── .golangci.yml             # Linter configuration
├── Makefile                  # Build and test commands
└── README.md
```

**Structure Decision**: Single project layout selected. Go's standard project structure with `cmd/` for executables, `internal/` for private packages, and `pkg/` for potentially reusable types. Tests organized by integration vs unit testing strategy. This structure supports the CLI application pattern and enforces encapsulation through Go's internal package visibility rules.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations - all constitution checks passed. No complexity justification required.

---

## Planning Phase Completion

### ✅ Phase 0: Research (Complete)

**Artifacts Generated**:
- `research.md` - Technology decisions and best practices
  - Programming language: Go 1.25
  - CLI framework: Cobra + Viper
  - LLM integration: OpenAI-compatible SDK for Ollama
  - Session management: JSON file-based storage
  - Tool approval: Interactive prompts + regex auto-approval
  - Path security: Working directory jail with validation
  - Error handling: Wrapped errors with user-friendly messages
  - Debug logging: log/slog with --debug flag for verbose output
  - Testing strategy: Table-driven tests with mocks
  - Performance: Lazy loading + streaming
  - Cross-platform: Platform-specific shell detection

**Key Decisions**:
- All technical unknowns resolved
- Best practices identified for each component
- Implementation patterns documented

### ✅ Phase 1: Design & Contracts (Complete)

**Artifacts Generated**:
- `data-model.md` - Core entities and relationships
  - 7 entities: Session, Message, ToolCall, ToolResult, Config, ApprovalRule, Tool
  - State diagrams for tool execution and session flow
  - Storage schema for JSON persistence
  - Validation rules and constraints
  
- `contracts/tools-api.md` - Tool interface specifications
  - Universal tool interface definition
  - Complete specs for all 10 tools with parameters, validation, responses
  - Error handling patterns
  - LLM function calling format
  - Testing contracts

- `quickstart.md` - User documentation
  - Installation and setup instructions
  - Basic usage examples
  - Common workflows for all 6 user stories
  - Configuration guide
  - Safety and security guidelines
  - Troubleshooting and CLI reference

- `.github/copilot-instructions.md` - Updated with Go 1.25 context

**Constitution Re-Check**: ✅ ALL GATES PASSED
- Code quality tools identified (golangci-lint, gofmt, gocyclo)
- Testing framework established (Go testing, testify, httptest)
- UX patterns defined (cobra conventions, error formatting)
- Performance targets validated (startup <500ms, memory <500MB)

### 📋 Next Steps

**Phase 2: Task Breakdown** (Run `/speckit.tasks`)
- Generate detailed implementation tasks
- Organize by user story priority (P1-P6)
- Include TDD workflow (tests first)
- Specify file paths and dependencies

**Ready for Implementation**: All planning complete, technical decisions made, contracts defined.
