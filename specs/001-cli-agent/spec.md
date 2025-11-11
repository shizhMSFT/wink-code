# Feature Specification: Wink CLI Coding Agent

**Feature Branch**: `001-cli-agent`  
**Created**: 2025-11-10  
**Status**: Draft  
**Input**: User description: "Build a lightweight CLI coding agent that connects local and remote LLMs via OpenAI-compatible API for quick script generation and coding assistance. The CLI tool is named `wink`. The CLI shall have the following built-in tools / capabilities: create_file, read_file, replace_string_in_file, list_dir, file_search, grep_search, run_in_terminal, terminal_last_command, create_directory, fetch_webpage. If a path is used in above tools, the path must be in the current working folder when the CLI starts. When a tool is invoked, ask for approval first. The tools can also be auto approved by using a config where parameter should be a regex match. When the user is asked for approval, the user can choose to always approve (i.e. auto approve) and the CLI should record that in the config."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Quick Script Generation (Priority: P1)

A developer needs to generate a simple script or code snippet quickly without opening an IDE or switching context. They invoke `wink` with a natural language prompt describing what code they need, and the CLI generates the appropriate script file immediately.

**Why this priority**: This is the core value proposition - enabling rapid code generation through natural language. It's the minimum viable product that delivers immediate value and can be independently tested without any other features.

**Independent Test**: Can be fully tested by running `wink -p "create a Python script that reads a CSV file and prints row count"` and verifying that a valid Python script is generated and saved to the current directory.

**Acceptance Scenarios**:

1. **Given** a developer is in a project directory, **When** they run `wink -p "create a bash script to backup logs"` or `wink --prompt "create a bash script to backup logs"`, **Then** a functional bash script is generated and saved in the current directory
2. **Given** a developer provides an ambiguous request, **When** `wink` processes the request, **Then** the LLM generates reasonable code based on common patterns and notifies the user of assumptions made
3. **Given** a developer wants to modify generated code, **When** they run `wink -p "update the script to add error handling"` referencing the previous output, **Then** the tool updates the generated file accordingly

---

### User Story 2 - File Operations with Approval (Priority: P2)

A developer wants to use `wink` to read, create, or modify files in their workspace, but needs control over which operations execute. The CLI prompts for approval before each file operation, giving the developer visibility and control.

**Why this priority**: Safety and trust are critical for adoption. This story establishes the approval workflow that prevents unintended file modifications, making the tool production-ready.

**Independent Test**: Can be tested by running `wink -p "read myfile.txt and create a summary in summary.txt"` and verifying that the tool prompts for approval before reading myfile.txt and again before creating summary.txt.

**Acceptance Scenarios**:

1. **Given** `wink` needs to execute a file operation, **When** the operation is about to run, **Then** the user receives a clear prompt showing the tool name, operation details, and file path affected
2. **Given** a user is prompted for approval, **When** they respond with "yes" or "y", **Then** the operation executes and the result is displayed
3. **Given** a user is prompted for approval, **When** they respond with "no" or "n", **Then** the operation is cancelled and the user is informed without error
4. **Given** a user wants to skip repetitive approvals, **When** they respond with "always" or "a" at an approval prompt, **Then** the tool records their preference and auto-approves similar operations in the future

---

### User Story 3 - Auto-Approval Configuration (Priority: P3)

A developer working in a trusted environment wants to streamline their workflow by pre-approving certain operations. They configure regex patterns for automatic approval, eliminating repetitive prompts for safe, expected operations.

**Why this priority**: This enhances power-user productivity but isn't essential for initial adoption. It builds on the approval system established in P2.

**Independent Test**: Can be tested by configuring auto-approval for read operations on `*.txt` files, then running `wink -p "read all txt files"` and verifying no prompts appear for matching operations.

**Acceptance Scenarios**:

1. **Given** a user selects "always approve" during an operation, **When** the approval is recorded, **Then** a configuration file is created/updated with a regex pattern matching that operation
2. **Given** auto-approval rules exist in the config, **When** `wink` encounters a matching operation, **Then** the operation executes without prompting and the user sees a notification that it was auto-approved
3. **Given** a user wants to review or remove auto-approvals, **When** they inspect the configuration file, **Then** they can see and manually edit or delete approval patterns
4. **Given** conflicting auto-approval patterns exist, **When** evaluating an operation, **Then** the most specific pattern takes precedence

---

### User Story 4 - Workspace Search and Analysis (Priority: P4)

A developer needs to understand or navigate an unfamiliar codebase. They use `wink` to search for files, grep for patterns, list directory contents, and get explanations of what code does, all through natural language queries.

**Why this priority**: This extends `wink` beyond code generation into code comprehension, but requires the foundational file operations from P2.

**Independent Test**: Can be tested by running `wink -p "find all Python files that import requests"` in a project directory and verifying correct search results are returned.

**Acceptance Scenarios**:

1. **Given** a user needs to find files, **When** they request `wink -p "find all config files"`, **Then** `wink` uses file_search with appropriate glob patterns and returns matching file paths
2. **Given** a user needs to search code content, **When** they request `wink -p "search for all TODO comments"`, **Then** `wink` uses grep_search and presents results with file paths and line numbers
3. **Given** a user wants directory information, **When** they request `wink -p "show me what's in the src folder"`, **Then** `wink` uses list_dir and presents a readable directory tree or list

---

### User Story 5 - Command Execution and Automation (Priority: P5)

A developer wants to automate repetitive tasks that involve both code generation and command execution. They use `wink` to execute shell commands, check command outputs, and make decisions based on results.

**Why this priority**: This enables advanced automation workflows but requires robust error handling and safety measures. It's valuable but not essential for initial release.

**Independent Test**: Can be tested by running `wink -p "check git status and create a commit script"` and verifying the tool executes git status, analyzes output, and generates an appropriate script.

**Acceptance Scenarios**:

1. **Given** `wink` needs to execute a shell command, **When** the command is requested, **Then** approval is required (following P2 approval workflow) before execution
2. **Given** a command is approved and executes, **When** the command completes, **Then** the output is captured and made available for the LLM to analyze
3. **Given** a command fails, **When** the error occurs, **Then** the error message is returned to the user with context and suggested next steps
4. **Given** a user wants to reference previous commands, **When** they use terminal_last_command, **Then** the most recent command and its output are retrieved

---

### User Story 6 - Web Content Integration (Priority: P6)

A developer needs to reference online documentation or API specifications while generating code. They use `wink` to fetch web content and incorporate it into code generation context.

**Why this priority**: This is a productivity enhancement that makes `wink` more capable, but it's not essential for core code generation functionality.

**Independent Test**: Can be tested by running `wink -p "fetch the OpenAI API docs and generate a client"` and verifying the tool retrieves web content and uses it to inform code generation.

**Acceptance Scenarios**:

1. **Given** a user references a URL in their request, **When** `wink` processes the request, **Then** fetch_webpage is used to retrieve content from the URL
2. **Given** web content is fetched, **When** generating code, **Then** the LLM incorporates relevant information from the fetched content
3. **Given** a webpage fails to load, **When** the error occurs, **Then** the user is notified and code generation continues with available information

---

### Edge Cases

- What happens when the LLM API is unreachable or times out?
- How does the system handle LLM requests that exceed the configured timeout?
- What happens when a user configures an unreasonably short timeout (e.g., 1 second)?
- How does the progress indicator behave when LLM responds very quickly (under 2 seconds)?
- How does the system handle file operations when the target path doesn't exist or isn't writable?
- What happens when a user's natural language request is too ambiguous to execute?
- How does the tool behave when the current working directory changes during execution?
- What happens when auto-approval patterns in config conflict or create security risks?
- How does the system handle very large files or directory trees that exceed token limits?
- What happens when a shell command hangs or runs indefinitely?
- How does the tool handle special characters or spaces in file paths across different operating systems?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST connect to LLM providers through OpenAI-compatible API endpoints
- **FR-002**: System MUST restrict all file operations to the current working directory and its subdirectories
- **FR-003**: System MUST prompt for user approval before executing any tool operation (create_file, read_file, replace_string_in_file, run_in_terminal, etc.)
- **FR-004**: System MUST display clear information about each operation before requesting approval (tool name, operation type, target path/command)
- **FR-005**: System MUST accept "yes/y", "no/n", and "always/a" as approval responses
- **FR-006**: System MUST persist auto-approval preferences in a configuration file when user selects "always"
- **FR-007**: System MUST support regex pattern matching for auto-approval configuration
- **FR-008**: System MUST validate file paths to ensure they are within the current working directory before any operation
- **FR-009**: System MUST support the following tools: create_file, read_file, replace_string_in_file, list_dir, file_search, grep_search, run_in_terminal, terminal_last_command, create_directory, fetch_webpage
- **FR-010**: System MUST handle LLM API errors gracefully with clear user-facing error messages
- **FR-011**: System MUST support reading file contents with optional line range specification (for read_file tool)
- **FR-012**: System MUST support glob patterns for file search operations (file_search tool)
- **FR-013**: System MUST support both literal text and regex patterns for grep_search operations
- **FR-014**: System MUST capture and return stdout and stderr from shell commands executed via run_in_terminal
- **FR-015**: System MUST parse user natural language input and determine which tools to invoke
- **FR-016**: System MUST maintain conversation context across multiple tool invocations within a single session
- **FR-017**: System MUST support cross-platform operation (Windows, macOS, Linux) with appropriate shell command handling
- **FR-018**: System MUST validate that the current working directory exists and is accessible on startup
- **FR-019**: System MUST allow users to configure LLM API endpoint, API key, and model parameters through environment variables or config file
- **FR-020**: System MUST timeout LLM API requests after a configurable duration (default 30 seconds) per constitution performance requirements
- **FR-021**: System MUST accept prompts via `-p` or `--prompt` command-line flag
- **FR-022**: System MUST support a `-d` or `--debug` flag that enables verbose logging including LLM API requests/responses, tool execution details, and internal state for troubleshooting
- **FR-023**: System MUST support a `--timeout` flag that allows users to configure LLM API request timeout duration in seconds
- **FR-024**: System MUST support `WINK_TIMEOUT` environment variable as an alternative to `--timeout` flag for configuring timeout duration
- **FR-025**: System MUST display a progress indicator when waiting for LLM API responses that take longer than 2 seconds
- **FR-026**: Progress indicator MUST show elapsed time and indicate ongoing activity (e.g., spinner, animated dots, or progress bar)
- **FR-027**: Progress indicator MUST update at least once per second to provide responsive feedback to the user

### Key Entities

- **Tool Operation**: Represents a single tool invocation with parameters, approval status, and execution result
- **Approval Rule**: Configuration entry containing a regex pattern, tool type, and auto-approval decision
- **Session Context**: Maintains conversation history, working directory, and executed operations for a single wink invocation
- **LLM Connection**: Configuration and state for communicating with the LLM API endpoint

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can generate a working script from a natural language prompt in under 30 seconds (end-to-end, including LLM response time)
- **SC-002**: 95% of file operations correctly prompt for approval before execution
- **SC-003**: Users can complete a typical coding task (generate script, review file, modify) with no more than 3 approval prompts when using auto-approval
- **SC-004**: The tool successfully prevents file operations outside the working directory 100% of the time
- **SC-005**: 90% of users successfully complete their first code generation task without consulting documentation
- **SC-006**: Average tool response time (excluding LLM API latency) is under 500ms

### Quality & Performance Criteria *(per constitution)*

- **QC-001**: Code coverage ≥90% with all edge cases tested (Constitution: Testing Standards)
- **QC-002**: All public APIs documented with examples (Constitution: Code Quality First)
- **QC-003**: Cyclomatic complexity ≤10 per function (Constitution: Code Quality First)
- **QC-004**: Interactive response time ≤2 seconds for approval prompts and tool execution feedback (Constitution: UX Consistency)
- **QC-005**: Error messages are actionable and user-friendly, never exposing raw stack traces (Constitution: UX Consistency)
- **QC-006**: Memory footprint ≤500MB during typical usage (Constitution: Performance Requirements)
- **QC-007**: CLI startup time ≤500ms (Constitution: Performance Requirements)
- **QC-008**: LLM API calls implement 30-second timeout with exponential backoff retry logic (Constitution: Performance Requirements)
- **QC-009**: Tool maintains consistent command patterns and argument conventions across all features (Constitution: UX Consistency)

## Assumptions

1. Users have network access to reach LLM API endpoints
2. Users have appropriate permissions to read/write files in their working directory
3. LLM API endpoints follow OpenAI's function calling / tool use protocol
4. Configuration file will be stored in a standard location (e.g., `~/.wink/config.json` or `.wink.json` in working directory)
5. Shell commands will be executed in the user's default shell (bash, zsh, PowerShell, etc.)
6. Users understand basic CLI interaction and file system concepts
7. File operations are synchronous (blocking) rather than async/concurrent
8. The tool operates as a single-shot command (not a persistent daemon or REPL) unless the user's prompt requires multiple iterations
9. Standard web scraping ethics apply to fetch_webpage (respects robots.txt, rate limiting)
