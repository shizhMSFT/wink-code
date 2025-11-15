# Wink CLI Coding Agent

> [!CAUTION]
> **Experimental AI-Generated Project**: This tool executes commands and modifies files on your system. The codebase itself was AI-generated and is experimental, created for learning and exploration purposes. Always review actions before approval and use in a safe environment (e.g., version-controlled projects). No warranty or support is provided.

A lightweight CLI coding agent that connects to local LLMs (via Ollama) for rapid script generation and coding assistance.

## Features

- 🚀 **Quick Script Generation**: Generate code from natural language prompts
- 🔒 **Safe File Operations**: Approval workflow with auto-approval configuration
- 🔍 **Workspace Search**: Search files and code content through natural language
- ⚡ **Command Execution**: Run shell commands with safety checks
- 🌐 **Web Integration**: Fetch online documentation for context
- 💾 **Session Persistence**: Continue previous conversations with `--continue`

## Installation

### Prerequisites

- Go 1.25 or later
- [Ollama](https://ollama.ai/) installed and running locally

### From Source

```bash
# Clone the repository
git clone https://github.com/shizhMSFT/wink-code.git
cd wink-code

# Build and install
make install

# Or build manually
go install ./cmd/wink
```

### Binary Download

Download pre-built binaries from the [releases page](https://github.com/shizhMSFT/wink-code/releases).

## Quick Start

```bash
# Start Ollama (if not already running)
ollama serve

# Pull a model (recommended: qwen3:8b or qwen3-coder:30b)
ollama pull qwen3:8b

# Generate a script
wink -p "create a Python script that reads a CSV file and prints row count"

# With a specific model
wink -m qwen3-coder:30b -p "create a bash script to backup logs"

# Continue previous session
wink --continue

# Enable debug logging
wink -d -p "your prompt here"
```

## Usage

```
Usage: wink [flags]

Flags:
  -p, --prompt string    Natural language prompt (required)
  -m, --model string     LLM model to use (default "qwen3:8b")
      --continue         Continue previous session
  -d, --debug            Enable verbose debug logging
  -h, --help             Help for wink
```

### Examples

**Create a file:**
```bash
wink -p "create a hello world Python script"
```

**Read and modify files:**
```bash
wink -p "read config.json and update the port to 8080"
```

**Search code:**
```bash
wink -p "find all Python files that import requests"
```

**Run commands:**
```bash
wink -p "check git status and create a commit script"
```

## Configuration

Configuration is stored in `~/.wink/config.json`:

```json
{
  "default_model": "qwen3:8b",
  "ollama_base_url": "http://localhost:11434",
  "api_timeout_seconds": 30,
  "auto_approval_rules": []
}
```

### Auto-Approval

When prompted for approval, you can:
- Type `y` or `yes` to approve once
- Type `n` or `no` to reject
- Type `a` or `always` to auto-approve similar operations in the future

Auto-approval rules are saved to your config file and use regex patterns for matching.

## Safety & Security

- **Working Directory Jail**: All file operations are restricted to the current directory and subdirectories
- **Approval Workflow**: Every tool operation requires approval (unless auto-approved)
- **Command-Level Approval**: Shell commands require approval per unique command
- **Transparent Operations**: Clear display of what each operation will do before execution

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Check coverage (target: ≥90%)
make coverage

# Build for all platforms
make build-all

# Show all targets
make help
```

### Project Structure

```
wink-code/
├── cmd/wink/              # CLI entry point
├── internal/
│   ├── agent/             # Core agent orchestration
│   ├── llm/               # LLM API client
│   ├── tools/             # Tool implementations
│   ├── config/            # Configuration management
│   └── ui/                # User interface (prompts, output)
├── pkg/types/             # Shared types
├── tests/
│   ├── unit/              # Unit tests
│   └── integration/       # Integration tests
└── specs/                 # Feature specifications
```

## Architecture

Wink uses:
- **Go 1.25** for cross-platform CLI performance
- **Cobra** for command-line interface
- **Viper** for configuration management
- **OpenAI SDK** for LLM communication (Ollama-compatible)
- **log/slog** for structured logging

See [specs/001-cli-agent/](specs/001-cli-agent/) for detailed documentation:
- [Feature Specification](specs/001-cli-agent/spec.md)
- [Implementation Plan](specs/001-cli-agent/plan.md)
- [Quick Start Guide](specs/001-cli-agent/quickstart.md)

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

See [LICENSE](LICENSE) for details.

## Acknowledgments

- [Ollama](https://ollama.ai/) for local LLM serving
- [Cobra](https://cobra.dev/) for CLI framework
- [OpenAI](https://openai.com/) for API specification
