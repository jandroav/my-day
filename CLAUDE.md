# Claude Context for my-day

## Project Overview
`my-day` is a colorful Golang CLI tool that helps DevOps team members track Jira tickets across multiple teams and generate daily standup reports with AI-powered summarization.

## Key Features
- Multi-team Jira ticket tracking (DevOps, Interop, Foundation, Enterprise, LBIO)
- OAuth 2.0 integration with Jira Cloud
- AI-powered summarization (embedded LLM or Ollama)
- Colorful console and markdown report generation
- Debug mode with quality indicators
- Cross-platform support

## Project Structure
```
my-day/
├── cmd/                    # Cobra CLI commands
│   ├── auth.go            # OAuth authentication
│   ├── config.go          # Configuration management
│   ├── init.go            # Initialize config
│   ├── llm.go             # LLM integration commands
│   ├── report.go          # Report generation
│   ├── root.go            # Root command
│   ├── sync.go            # Jira sync
│   └── version.go         # Version information
├── internal/
│   ├── config/            # Configuration handling
│   ├── jira/              # Jira API client
│   ├── llm/               # LLM processing (embedded & Ollama)
│   └── report/            # Report generation
├── docs/                  # Documentation
├── .github/workflows/     # GitHub Actions
└── main.go               # Entry point
```

## Technology Stack
- **Language**: Go 1.24.3
- **CLI Framework**: Cobra + Viper
- **Authentication**: OAuth 2.0
- **AI/LLM**: Embedded model + Ollama support
- **Output**: Colorful terminal + Markdown
- **Build**: Cross-platform (Linux, macOS, Windows × amd64/arm64)

## Configuration
- Default config: `~/.my-day/config.yaml`
- Environment variables with `MY_DAY_` prefix
- CLI flags override all other settings
- OAuth tokens stored in `~/.my-day/auth.json`

## Key Commands
- `my-day init` - Initialize configuration
- `my-day auth` - OAuth authentication
- `my-day sync` - Sync Jira tickets
- `my-day report` - Generate daily standup report
- `my-day config show` - Display configuration
- `my-day llm test` - Test LLM functionality

## Development Commands
```bash
# Build
go build -o my-day

# Run tests
go test ./...

# Run with debug
./my-day report --debug --show-quality --verbose
```

## Release Process
- **Automatic**: Push git tags (v1.0.0)
- **Manual**: GitHub Actions workflow_dispatch
- **Artifacts**: Cross-platform binaries with checksums
- **Repository**: https://github.com/jandroav/my-day

## LLM Integration
- **Embedded mode**: Default lightweight model
- **Ollama mode**: External Ollama server
- **Features**: Technical pattern matching, debug mode, quality scoring
- **Debug flags**: `--debug`, `--show-quality`, `--verbose`

## Common Tasks
- Adding new CLI commands: Add to `cmd/` directory
- Modifying LLM behavior: Edit `internal/llm/` files
- Updating configuration: Modify `internal/config/` files
- Adding tests: Use Go's testing framework
- Building releases: Use GitHub Actions workflow