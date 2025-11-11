# Tasks: Wink CLI Coding Agent

**Branch**: `001-cli-agent` | **Date**: 2025-11-11  
**Input**: Design documents from `/specs/001-cli-agent/`  
**Prerequisites**: plan.md (complete), spec.md (complete), research.md (complete), data-model.md (complete), contracts/ (complete)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Implementation Strategy

**MVP-First Approach**: User Story 1 (P1) represents the minimum viable product - quick script generation with basic file creation. Subsequent stories build on this foundation incrementally.

**Independent Stories**: Stories P1-P6 are designed to be independently testable. Each story delivers working functionality that can be demonstrated and validated.

**Parallel Opportunities**: Tasks marked [P] can be executed in parallel when within the same phase, provided they operate on different files and have no dependencies.

---

## Phase 1: Setup (Project Initialization)

**Purpose**: Project initialization and basic structure

- [ ] T001 Initialize Go module with `go mod init github.com/shizhMSFT/wink-code` in repository root
- [ ] T002 [P] Create directory structure: cmd/wink/, internal/agent/, internal/llm/, internal/tools/, internal/config/, internal/ui/, pkg/types/, tests/unit/, tests/integration/
- [ ] T003 [P] Configure golangci-lint with .golangci.yml (cyclomatic complexity ≤10, gofmt, goimports)
- [ ] T004 [P] Create Makefile with build, test, lint, install targets
- [ ] T005 [P] Setup GitHub Actions CI workflow in .github/workflows/ci.yml (lint, test, build for Linux/Windows/macOS)
- [ ] T006 [P] Create README.md with installation and quick start instructions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T007 Install core dependencies: cobra, viper, OpenAI Go SDK in go.mod
- [ ] T008 [P] Create pkg/types/tool.go with Tool interface, ToolResult, RiskLevel enums
- [ ] T009 [P] Create pkg/types/session.go with Session, Message structs matching data-model.md
- [ ] T010 [P] Create pkg/types/approval.go with ApprovalRule, ToolCall structs
- [ ] T011 [P] Create pkg/types/config.go with Config struct
- [ ] T012 Implement internal/config/config.go for loading/saving config from ~/.wink/config.json with viper
- [ ] T013 [P] Implement internal/config/approval.go for auto-approval rule management (add, match, persist)
- [ ] T014 Implement internal/ui/prompt.go for user approval prompts (yes/no/always) with stdin/stdout
- [ ] T015 [P] Implement internal/ui/output.go for formatted output (human-readable and JSON modes)
- [ ] T016 Implement debug logging initialization in internal/logging/logger.go using log/slog with -d/--debug flag support
- [ ] T017 Implement path validation in internal/tools/security.go for working directory jail (ValidatePath function)
- [ ] T018 Implement internal/llm/client.go with OpenAI SDK pointing to Ollama base URL http://localhost:11434/v1
- [ ] T019 [P] Implement internal/llm/retry.go with exponential backoff (3 retries, 30s timeout per constitution)
- [ ] T020 Create internal/tools/registry.go with tool registration and dispatch logic
- [ ] T021 [P] Implement internal/agent/session.go for session persistence to ~/.wink/sessions/{id}.json
- [ ] T022 [P] Implement internal/agent/context.go for conversation context management (100 message limit)
- [ ] T023 Implement cmd/wink/main.go with cobra root command, -p/--prompt, -m/--model, --continue, -d/--debug flags

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Quick Script Generation (Priority: P1) 🎯 MVP

**Goal**: Enable developers to generate simple scripts/code through natural language prompts. This is the core value proposition and minimum viable product.

**Independent Test**: Run `wink -p "create a Python script that reads a CSV file and prints row count"` and verify a valid Python script is generated and saved.

### Implementation for User Story 1

- [ ] T024 [US1] Implement create_file tool in internal/tools/file.go with path validation, content writing, error handling per contracts/tools-api.md
- [ ] T025 [US1] Register create_file tool in internal/tools/registry.go with name, description, parameters schema
- [ ] T026 [US1] Implement approval workflow in internal/tools/approval.go: check auto-approval rules, prompt user, handle yes/no/always responses
- [ ] T027 [US1] Implement internal/agent/agent.go core orchestration: accept prompt, call LLM with tools, handle tool calls, return results
- [ ] T028 [US1] Wire up cobra command in cmd/wink/main.go to call agent with prompt from -p flag
- [ ] T029 [US1] Add integration test in tests/integration/quick_script_test.go for end-to-end script generation workflow
- [ ] T030 [US1] Add error handling for LLM unreachable, invalid prompts, file creation failures with user-friendly messages (Constitution: UX Consistency)
- [ ] T031 [US1] Validate startup time ≤500ms and tool execution overhead <100ms with benchmarks in tests/integration/benchmark_test.go (Constitution: Performance)

**MVP Delivery**: After T031, wink can generate files from natural language prompts with approval workflow. This is independently deployable.

---

## Phase 4: User Story 2 - File Operations with Approval (Priority: P2)

**Goal**: Establish comprehensive file operation capabilities with safety through approval workflow. Makes tool production-ready.

**Independent Test**: Run `wink -p "read myfile.txt and create a summary in summary.txt"` and verify approval prompts appear before each operation.

### Implementation for User Story 2

- [ ] T032 [P] [US2] Implement read_file tool in internal/tools/file.go with line range support, size limits (10MB), error handling per contracts/tools-api.md
- [ ] T033 [P] [US2] Implement replace_string_in_file tool in internal/tools/file.go with exact string matching, validation per contracts/tools-api.md
- [ ] T034 [P] [US2] Implement create_directory tool in internal/tools/directory.go with recursive creation, path validation per contracts/tools-api.md
- [ ] T035 [P] [US2] Implement list_dir tool in internal/tools/directory.go with formatting, pagination for >1000 files per contracts/tools-api.md
- [ ] T036 [US2] Register all file/directory tools (read_file, replace_string_in_file, create_directory, list_dir) in internal/tools/registry.go
- [ ] T037 [US2] Update approval workflow in internal/tools/approval.go to show clear operation details (tool name, path, action) per FR-004
- [ ] T038 [US2] Add unit tests for each file tool in tests/unit/file_tools_test.go with table-driven tests covering success, path escape, permission errors (≥90% coverage)
- [ ] T039 [US2] Add integration test in tests/integration/file_operations_test.go for multi-operation workflows with approval prompts
- [ ] T040 [US2] Implement -d/--debug flag logging for file operations showing paths, sizes, approval status in internal/tools/file.go

**Delivery**: After T040, wink supports comprehensive file operations with transparent approval workflow.

---

## Phase 5: User Story 3 - Auto-Approval Configuration (Priority: P3)

**Goal**: Enable power users to streamline workflows by pre-configuring trusted operations, eliminating repetitive prompts.

**Independent Test**: Configure auto-approval for read operations on `*.txt` files, then run `wink -p "read all txt files"` and verify no prompts appear.

### Implementation for User Story 3

- [ ] T041 [US3] Implement "always" response handler in internal/tools/approval.go that generates regex pattern from tool + parameters
- [ ] T042 [US3] Implement auto-approval rule persistence in internal/config/approval.go: save rule to config file, reload on next run
- [ ] T043 [US3] Implement regex matching in internal/config/approval.go: match incoming tool call against stored patterns, handle specificity precedence
- [ ] T044 [US3] Add auto-approval notification in internal/ui/output.go: "Auto-approved by rule: [description]"
- [ ] T045 [US3] Add unit tests for auto-approval logic in tests/unit/approval_test.go: pattern generation, matching, precedence, edge cases (≥90% coverage)
- [ ] T046 [US3] Add integration test in tests/integration/auto_approval_test.go: record rule via "always", verify auto-execution on next run
- [ ] T047 [US3] Document config file format and manual editing in specs/001-cli-agent/quickstart.md Auto-Approval Configuration section

**Delivery**: After T047, users can configure auto-approval for trusted operations, improving productivity.

---

## Phase 6: User Story 4 - Workspace Search and Analysis (Priority: P4)

**Goal**: Extend wink beyond code generation into code comprehension through search and analysis tools.

**Independent Test**: Run `wink -p "find all Python files that import requests"` and verify correct search results.

### Implementation for User Story 4

- [ ] T048 [P] [US4] Implement file_search tool in internal/tools/search.go with glob pattern matching, path filtering per contracts/tools-api.md
- [ ] T049 [P] [US4] Implement grep_search tool in internal/tools/search.go with regex support, line number reporting, result limiting per contracts/tools-api.md
- [ ] T050 [US4] Register search tools (file_search, grep_search) in internal/tools/registry.go with appropriate risk levels (read_only)
- [ ] T051 [US4] Update read_file tool to handle large file streaming for files >10MB in internal/tools/file.go (performance optimization)
- [ ] T052 [US4] Add unit tests for search tools in tests/unit/search_tools_test.go: glob patterns, regex, result limits, edge cases (≥90% coverage)
- [ ] T053 [US4] Add integration test in tests/integration/search_test.go: multi-file search, grep with context, combination workflows
- [ ] T054 [US4] Add -d/--debug logging for search operations showing patterns, match counts, performance metrics

**Delivery**: After T054, wink can search and analyze codebases through natural language queries.

---

## Phase 7: User Story 5 - Command Execution and Automation (Priority: P5)

**Goal**: Enable automation workflows that combine code generation with command execution. Requires robust safety measures.

**Independent Test**: Run `wink -p "check git status and create a commit script"` and verify command execution with approval and output capture.

### Implementation for User Story 5

- [ ] T055 [US5] Implement run_in_terminal tool in internal/tools/terminal.go with cross-platform shell detection (cmd/PowerShell/bash) per contracts/tools-api.md
- [ ] T056 [US5] Implement command-level approval for run_in_terminal: regex matches specific command string, not just tool name, in internal/tools/approval.go
- [ ] T057 [US5] Implement stdout/stderr capture in internal/tools/terminal.go with output size limits, timeout handling (30s default)
- [ ] T058 [US5] Implement terminal_last_command tool in internal/tools/terminal.go with command history tracking per contracts/tools-api.md
- [ ] T059 [US5] Register terminal tools (run_in_terminal, terminal_last_command) in internal/tools/registry.go with dangerous risk level
- [ ] T060 [US5] Add input sanitization for shell commands in internal/tools/terminal.go: escape special characters per platform
- [ ] T061 [US5] Add unit tests for terminal tools in tests/unit/terminal_tools_test.go: platform detection, command escaping, output capture (≥90% coverage)
- [ ] T062 [US5] Add integration test in tests/integration/terminal_test.go: execute safe commands, verify approval required, test command-level auto-approval
- [ ] T063 [US5] Add -d/--debug logging for command execution showing shell used, command sanitized, exit code, execution time
- [ ] T064 [US5] Document command-level approval security model in specs/001-cli-agent/quickstart.md Safety Guidelines section

**Delivery**: After T064, wink supports safe command execution with granular approval control.

---

## Phase 8: User Story 6 - Web Content Integration (Priority: P6)

**Goal**: Enable referencing online documentation and API specs during code generation for improved context.

**Independent Test**: Run `wink -p "fetch the OpenAI API docs and generate a client"` and verify web content retrieval and usage.

### Implementation for User Story 6

- [ ] T065 [US6] Implement fetch_webpage tool in internal/tools/web.go with HTTP client, timeout (30s), content extraction per contracts/tools-api.md
- [ ] T066 [US6] Add robots.txt checking in internal/tools/web.go to respect web scraping ethics
- [ ] T067 [US6] Add content size limits (100KB) and truncation in internal/tools/web.go to prevent token overflow
- [ ] T068 [US6] Register fetch_webpage tool in internal/tools/registry.go with dangerous risk level (external network access)
- [ ] T069 [US6] Add unit tests for web tool in tests/unit/web_tools_test.go: URL validation, timeout, size limits, error handling (≥90% coverage)
- [ ] T070 [US6] Add integration test in tests/integration/web_test.go using httptest mock server: fetch content, handle errors, verify approval workflow
- [ ] T071 [US6] Add -d/--debug logging for web requests showing URL, response status, content size, fetch duration

**Delivery**: After T071, wink can incorporate online documentation into code generation context.

---

## Phase 9: Session Continuation & Polish (Cross-Cutting Concerns)

**Goal**: Enable session persistence with --continue flag and final polish for production readiness.

### Implementation

- [ ] T072 [P] Implement session loading in internal/agent/session.go: read from ~/.wink/sessions/{id}.json, restore context
- [ ] T073 [P] Implement --continue flag handling in cmd/wink/main.go: find latest session, load, resume conversation
- [ ] T074 [P] Implement session pruning in internal/agent/context.go: keep last 100 messages, archive older messages
- [ ] T075 [P] Add session ID display and continuation instructions in internal/ui/output.go
- [ ] T076 [P] Add environment variable support in internal/config/config.go: WINK_MODEL, WINK_OLLAMA_URL, WINK_DEBUG
- [ ] T077 [P] Implement token usage tracking and reporting in internal/llm/client.go
- [ ] T078 [P] Add memory footprint monitoring in internal/agent/agent.go to validate ≤500MB target (Constitution: Performance)
- [ ] T079 Add comprehensive integration test in tests/integration/session_continuation_test.go: create session, exit, continue, verify context preserved
- [ ] T080 Add cross-platform testing in CI for Windows/macOS/Linux builds with platform-specific shell commands
- [ ] T081 [P] Update specs/001-cli-agent/quickstart.md with complete examples for all 6 user stories
- [ ] T082 [P] Create build scripts in scripts/ for cross-platform binary compilation (Linux/Windows/macOS)
- [ ] T083 Run full test suite and validate all constitution requirements: ≥90% coverage, cyclomatic complexity ≤10, performance targets met
- [ ] T084 Create release artifacts: binaries, README, LICENSE, installation instructions

**Final Delivery**: Production-ready wink CLI with all 6 user stories implemented, tested, and documented.

---

## Dependencies & Execution Order

### Story Dependencies

```
Phase 1 (Setup) → Phase 2 (Foundation) → [Independent User Stories] → Phase 9 (Polish)
                                          ↓
                              ┌───────────┼───────────┬───────────┐
                              ↓           ↓           ↓           ↓
                          US1 (P1)    US2 (P2)    US3 (P3)    US4 (P4)
                           MVP          ↓           ↓           ↓
                                       US5 (P5) ← depends on US2 tools
                                        ↓
                                       US6 (P6)
```

### Critical Path

1. **T001-T006**: Setup (can parallelize T003-T006)
2. **T007-T023**: Foundation (can parallelize most, but T007 blocks others)
3. **T024-T031**: US1 MVP (sequential within story)
4. **T032-T040**: US2 (can parallelize T032-T035, rest sequential)
5. **T041-T047**: US3 (builds on US2)
6. **T048-T054**: US4 (independent, can start after Phase 2)
7. **T055-T064**: US5 (requires US2 tools)
8. **T065-T071**: US6 (independent, can start after Phase 2)
9. **T072-T084**: Polish (can parallelize many tasks)

### Parallel Execution Examples

**After Phase 2 Complete**:
- Work on US1 (T024-T031) AND US4 (T048-T054) simultaneously
- Work on US2 (T032-T040) AND US6 (T065-T071) simultaneously

**Within Each Story**:
- US2: T032, T033, T034, T035 can all be done in parallel (different files)
- US4: T048 and T049 can be done in parallel (different tools)
- Polish: T072-T078, T081-T082 can be parallelized

### MVP Scope

**Minimum for First Release**: Phase 1 + Phase 2 + US1 (T001-T031)

This delivers:
- Working CLI with Ollama connection
- File creation from natural language prompts
- Approval workflow
- Basic error handling
- Performance targets met

**Recommended Initial Release**: Add US2 (T032-T040) for production-ready file operations

---

## Task Statistics

- **Total Tasks**: 84
- **Setup & Foundation**: 23 tasks (T001-T023)
- **User Story 1 (P1)**: 8 tasks (T024-T031) - MVP
- **User Story 2 (P2)**: 9 tasks (T032-T040)
- **User Story 3 (P3)**: 7 tasks (T041-T047)
- **User Story 4 (P4)**: 7 tasks (T048-T054)
- **User Story 5 (P5)**: 10 tasks (T055-T064)
- **User Story 6 (P6)**: 7 tasks (T065-T071)
- **Polish & Continuation**: 13 tasks (T072-T084)
- **Parallelizable Tasks**: 41 marked with [P]

**Estimated Effort**: 
- MVP (Phase 1-2 + US1): ~40-60 hours
- Production Ready (+US2-US3): ~80-100 hours
- Full Feature Set: ~120-160 hours

**Test Coverage Target**: ≥90% per constitution (tests included in each phase)
