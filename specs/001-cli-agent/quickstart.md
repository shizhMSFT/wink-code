# Quickstart Guide: Wink CLI Coding Agent

**Feature**: 001-cli-agent  
**Created**: 2025-11-10  
**Purpose**: Developer onboarding and common workflows

## Overview

Wink is a lightweight CLI coding agent that connects to local Ollama server for quick script generation and coding assistance. It provides 10 built-in tools with safety approval workflows.

## Prerequisites

- **Ollama**: Installed and running locally ([installation guide](https://ollama.ai))
- **Go 1.22+**: For building from source
- **Supported Models**: `qwen3-coder:30b` or `qwen3:8b` (default) pulled in Ollama

## Installation

### From Source
```bash
git clone https://github.com/shizhMSFT/wink-code.git
cd wink-code
make build
sudo make install  # or add ./bin to PATH
```

### Verify Installation
```bash
wink --version
# Output: wink v1.0.0
```

## Basic Usage

### Generate Your First Script

```bash
# Navigate to your project directory
cd ~/my-project

# Generate a script with a prompt
wink -p "create a Python script that backs up all .txt files to a backup folder"
```

**What happens**:
1. Wink connects to local Ollama
2. LLM analyzes your request and proposes tool calls
3. You're prompted to approve each operation
4. Script is created in current directory

**Example interaction**:
```
Tool: create_file
Parameters:
  path: backup_texts.py
  content: <245 bytes>
Files affected: backup_texts.py
Risk level: safe_write

Approve? (y/n/always): y

✓ Created file: backup_texts.py (245 bytes)
  Location: /home/user/my-project/backup_texts.py
```

### Using Different Models

```bash
# Use the more powerful coder model
wink -m qwen3-coder:30b -p "create a REST API server"

# Use custom model
wink --model codellama:13b -p "refactor this function"
```

### Non-Interactive Mode with Auto-Approval

```bash
# Set up auto-approval for read operations
wink -p "read all config files"
# When prompted: choose 'always' to create auto-approval rule

# Now subsequent reads won't prompt
wink -p "analyze config files for errors"
# Reads happen automatically, no prompts
```

## Common Workflows

### 1. Quick Script Generation (P1)

Generate standalone scripts for automation:

```bash
# Bash script
wink -p "create a bash script to find and delete log files older than 30 days"

# Python script with dependencies
wink -p "create a Python script that reads CSV and generates charts using matplotlib"

# PowerShell script (on Windows)
wink -p "create a PowerShell script to backup registry keys"
```

### 2. Read and Analyze Code (P4)

Understand an unfamiliar codebase:

```bash
# Find specific files
wink -p "find all Python files that import requests"

# Search for patterns
wink -p "search for all TODO comments and list them"

# Analyze directory structure
wink -p "show me the src directory structure and explain what each folder contains"

# Read specific files
wink -p "read the main.py file and explain what it does"
```

### 3. Modify Existing Code (P2)

Update files with safety approvals:

```bash
# Update a specific function
wink -p "in utils.py, add error handling to the parse_config function"

# Refactor code
wink -p "read models.py and extract the User class to a separate file"

# Add features
wink -p "add logging to all functions in app.py"
```

### 4. Automated Tasks (P5)

Execute commands and make decisions:

```bash
# Check status and generate reports
wink -p "check git status and create a commit message based on changes"

# System analysis
wink -p "check disk usage and create a cleanup script if usage > 80%"

# Test automation
wink -p "run pytest and if there are failures, analyze and suggest fixes"
```

### 5. Session Continuation (P3)

Continue conversations across invocations:

```bash
# Start a task
wink -p "create a Flask API with user authentication"

# Later, continue the session
wink --continue -p "now add password reset functionality"

# Keep going
wink --continue -p "add rate limiting to the login endpoint"
```

### 6. Web Research Integration (P6)

Reference online documentation:

```bash
# Fetch and use docs
wink -p "fetch the FastAPI documentation and create a hello world API"

# Integration guides
wink -p "get the Stripe API docs and create a payment processing script"
```

## Configuration

### Config File Location

Primary: `~/.wink/config.json`  
Override: `./.wink.json` (in current directory)

### Example Config

```json
{
  "default_model": "qwen3:8b",
  "ollama_base_url": "http://localhost:11434",
  "api_timeout_seconds": 30,
  "max_session_messages": 100,
  "output_format": "human",
  "auto_approval_rules": [
    {
      "id": "auto-read-txt",
      "tool_name": "read_file",
      "param_pattern": ".*\\.txt$",
      "description": "Auto-approve reading .txt files"
    },
    {
      "id": "auto-list-dirs",
      "tool_name": "list_dir",
      "param_pattern": ".*",
      "description": "Auto-approve listing any directory"
    }
  ]
}
```

### Environment Variables

```bash
# Override Ollama URL
export WINK_OLLAMA_URL=http://192.168.1.100:11434

# Set default model
export WINK_MODEL=qwen3-coder:30b

# Set API timeout
export WINK_TIMEOUT=60
```

## Safety & Security

### Working Directory Restriction

**All file operations are restricted to the current working directory and its subdirectories.**

```bash
cd ~/my-project
wink -p "read /etc/passwd"
# ❌ Error: Path outside working directory

wink -p "read config.txt"
# ✓ Allowed (within ~/my-project)
```

### Approval Workflow

Every tool operation requires approval by default:

```
Tool: run_in_terminal
Command: git status
Working directory: /home/user/my-project
Risk level: dangerous

Approve? (y/n/always):
  y      - Execute once
  n      - Cancel
  always - Auto-approve similar operations
```

**Important**: For `run_in_terminal`, approval is at the **command level**:
- Each unique command requires separate approval
- Approving `git status` does NOT auto-approve `git push`
- Auto-approval rules match the specific command string

### Auto-Approval Rules

Rules use regex patterns to match operations:

```bash
# Create rule interactively
wink -p "read all markdown files"
# Choose 'always' when prompted → creates rule

# View rules
cat ~/.wink/config.json

# Edit rules manually
vim ~/.wink/config.json
# Remove unwanted rules or tighten patterns
```

**Best Practices**:
- Start with manual approval, add auto-approval as you trust patterns
- Use specific patterns (e.g., `.*\\.txt$`) instead of wildcards (`.*`)
- Review rules periodically: `cat ~/.wink/config.json | jq .auto_approval_rules`
- **For `run_in_terminal`**: Each command creates a separate rule (e.g., "git status" vs "git push")

**Example Rules**:
```json
{
  "auto_approval_rules": [
    {
      "tool_name": "read_file",
      "param_pattern": ".*\\.txt$",
      "description": "Auto-approve reading .txt files"
    },
    {
      "tool_name": "run_in_terminal",
      "param_pattern": "^{\"command\":\"git status\".*}$",
      "description": "Auto-approve 'git status' command only"
    },
    {
      "tool_name": "run_in_terminal",
      "param_pattern": "^{\"command\":\"npm test\".*}$",
      "description": "Auto-approve 'npm test' command only"
    }
  ]
}
```

## Troubleshooting

### Ollama Connection Failed

```bash
# Check if Ollama is running
ollama list

# Start Ollama (if not running)
ollama serve

# Test connection
curl http://localhost:11434/api/version
```

### Model Not Found

```bash
# Pull required model
ollama pull qwen3:8b

# Or use the coder variant
ollama pull qwen3-coder:30b

# List available models
ollama list
```

### Slow Response Times

```bash
# Use smaller model
wink -m qwen3:8b -p "simple task"

# Check Ollama resource usage
ollama ps

# Increase timeout for complex tasks
export WINK_TIMEOUT=60
wink -p "complex analysis task"
```

### Session Not Continuing

```bash
# List sessions
ls ~/.wink/sessions/

# Check session file exists
cat ~/.wink/sessions/<session-id>.json

# Session may have expired - start fresh
wink -p "new task"
```

### Path Validation Errors

```bash
# Use relative paths only
wink -p "read ./src/main.py"  # ✓ Good
wink -p "read /home/user/src/main.py"  # ❌ Absolute path rejected

# Stay within project
cd ~/my-project
wink -p "read ../other-project/file.py"  # ❌ Outside working dir
```

## CLI Reference

### Flags

```
-p, --prompt string      Natural language prompt (required)
-m, --model string       LLM model name (default: "qwen3:8b")
    --continue           Continue previous session
    --json               Output in JSON format
-h, --help              Show help
-v, --version           Show version
```

### Examples

```bash
# Basic usage
wink -p "create hello world script"

# With specific model
wink -m qwen3-coder:30b -p "complex code generation"

# Continue session
wink --continue -p "add more features"

# JSON output
wink --json -p "read config.json"
```

## Development Workflow

### Typical Development Session

```bash
# 1. Start in project directory
cd ~/my-app

# 2. Initial analysis
wink -p "analyze the project structure and summarize what this app does"

# 3. Generate new feature
wink -p "create a new user registration endpoint"

# 4. Review and refine
wink -p "read the registration code and add input validation"

# 5. Add tests
wink -p "create unit tests for the registration endpoint"

# 6. Check and commit
wink -p "run tests, and if they pass, check git status and suggest commit message"
```

## Advanced Usage

### Custom Tool Patterns

Create reusable command patterns:

```bash
# Create alias in ~/.bashrc or ~/.zshrc
alias wink-analyze='wink -p "analyze all Python files for code quality issues"'
alias wink-test='wink -p "run all tests and summarize results"'
alias wink-docs='wink -p "update README.md to reflect current code"'

# Use aliases
wink-analyze
wink-test
```

### Batch Operations

```bash
# Process multiple files
wink -p "for each Python file in src/, add type hints and docstrings"

# Apply patterns across codebase
wink -p "find all database queries and add error handling"
```

### Integration with Git

```bash
# Smart commits
wink -p "review staged changes and write a conventional commit message"

# Release notes
wink -p "generate release notes from commits since last tag"

# Code review prep
wink -p "analyze my changes and create a PR description"
```

## Performance Tips

1. **Use appropriate model**: `qwen3:8b` for simple tasks, `qwen3-coder:30b` for complex code
2. **Auto-approve safe operations**: Reduce prompt overhead for trusted operations
3. **Keep sessions focused**: Start fresh session for unrelated tasks
4. **Limit context**: Be specific in prompts to avoid unnecessary file reads

## Next Steps

- **Read the spec**: `specs/001-cli-agent/spec.md` for detailed requirements
- **Review tools**: `specs/001-cli-agent/contracts/tools-api.md` for tool documentation
- **Contribute**: Check CONTRIBUTING.md for development guidelines
- **Get help**: Open an issue on GitHub or join community discussions

## Support

- **GitHub Issues**: https://github.com/shizhMSFT/wink-code/issues
- **Documentation**: https://github.com/shizhMSFT/wink-code/docs
- **Ollama Help**: https://ollama.ai/docs

---

**Version**: 1.0.0  
**Last Updated**: 2025-11-10
