# my-day

A colorful Golang CLI tool that helps DevOps team members track Jira tickets across multiple teams and generate daily standup reports.

## Features

- 🎯 **Multi-team Support**: Track tickets across DevOps, Interop, Foundation, Enterprise, and LBIO teams
- 🔐 **OAuth 2.0 Integration**: Secure authentication with Jira Cloud
- 📊 **Daily Reports**: Generate colorful console or markdown reports for standups
- ⚙️ **Flexible Configuration**: YAML config, CLI flags, and environment variables
- 🚀 **Fast & Offline**: Local caching for quick report generation
- 🤖 **AI Summarization**: Optional embedded LLM or Ollama integration
- 🌈 **Colorful Output**: Beautiful terminal interface with status icons

## Quick Start

### 1. Install

Download the latest binary from [releases](https://github.com/your-org/my-day/releases) or build from source:

```bash
git clone https://github.com/your-org/my-day.git
cd my-day
go build -o my-day
```

### 2. Initialize Configuration

```bash
./my-day init
```

This creates `~/.my-day/config.yaml` with default settings.

### 3. Configure Jira OAuth

Edit your configuration file to add your Jira details:

```yaml
jira:
  base_url: "https://your-company.atlassian.net"
  oauth:
    client_id: "your-oauth-client-id"
    client_secret: "your-oauth-client-secret"
```

### 4. Authenticate

```bash
./my-day auth
```

This opens your browser to complete OAuth authentication.

### 5. Sync Your Tickets

```bash
./my-day sync
```

### 6. Generate Daily Report

```bash
./my-day report
```

## Commands

| Command | Description |
|---------|-------------|
| `init` | Initialize configuration file |
| `auth` | Authenticate with Jira OAuth |
| `sync` | Pull latest tickets from Jira |
| `report` | Generate daily standup report |
| `report --debug` | Generate report with debug information |
| `report --show-quality` | Generate report with quality indicators |
| `report --verbose` | Generate report with verbose LLM processing |
| `llm status` | Check LLM configuration and connectivity |
| `llm test` | Test LLM summarization functionality |
| `config` | Manage configuration settings |
| `version` | Show version information |

## Configuration

### Configuration File

Default location: `~/.my-day/config.yaml`

```yaml
jira:
  base_url: "https://your-instance.atlassian.net"
  oauth:
    client_id: "your-oauth-client-id"
    client_secret: "your-oauth-client-secret"
    redirect_uri: "http://localhost:8080/callback"
  projects:
    - key: "DEVOPS"
      name: "DevOps Team"
    - key: "INTEROP"
      name: "Interop Team"
    # Add more projects...

llm:
  enabled: true
  mode: "embedded"  # embedded, ollama, disabled
  model: "tinyllama"
  # Enhanced LLM Configuration
  config:
    debug: false                     # Enable debug logging
    summary_style: "technical"       # technical, business, brief
    max_summary_length: 200         # Maximum summary length
    include_technical_details: true  # Include technical terms
    prioritize_recent_work: true     # Focus on recent activity
    fallback_strategy: "graceful"    # Error handling strategy
  ollama:
    base_url: "http://localhost:11434"
    model: "llama3.1"

report:
  format: "console"  # console, markdown
  include_yesterday: true
  include_today: true
  include_in_progress: true
```

### CLI Flags

All configuration options can be overridden with CLI flags:

```bash
./my-day report --report-format markdown --no-llm --detailed
./my-day sync --max-results 50 --projects DEVOPS,INTEROP
```

### Environment Variables

All settings support environment variables with `MY_DAY_` prefix:

```bash
export MY_DAY_JIRA_BASE_URL="https://company.atlassian.net"
export MY_DAY_JIRA_CLIENT_ID="abc123"
export MY_DAY_LLM_MODE="disabled"
```

## Usage Examples

### Basic Daily Report

```bash
./my-day report
```

Output:
```
🚀 Daily Standup Report - January 15, 2025
==================================================

📊 SUMMARY
• Issues: 5
• Worklog entries: 3

🔄 CURRENTLY WORKING ON
  🔄 DEV-123 [DEVOPS] Fix CI/CD pipeline timeout
  🔄 INT-456 [INTEROP] API integration issues

✅ RECENTLY COMPLETED
  ✅ DEV-122 [DEVOPS] Update deployment scripts
  ✅ FOUND-789 [FOUND] Database migration
```

### Markdown Report

```bash
./my-day report --report-format markdown --output daily-report.md
```

### Specific Date Report

```bash
./my-day report --date 2025-01-14 --detailed
```

### Override Configuration

```bash
./my-day report --projects DEVOPS,INTEROP --no-llm
```

### Enhanced Report Features

**Debug Mode**: Get detailed processing information
```bash
./my-day report --debug
```

**Quality Indicators**: View summary quality metrics
```bash
./my-day report --show-quality
```

**Verbose Mode**: See detailed LLM processing steps
```bash
./my-day report --verbose
```

**Combined Enhanced Mode**: Full diagnostic output
```bash
./my-day report --debug --show-quality --verbose
```

Example enhanced output:
```
🚀 Daily Standup Report - January 15, 2025
==================================================

🤖 AI SUMMARY OF TODAY'S WORK (Enhanced)
Completed AWS Lambda deployment using Terraform configuration. VPC security groups updated and Kubernetes ingress configured. Database migration scripts tested successfully.

🔑 Key Activities:
  • Terraform infrastructure deployment
  • AWS security configuration
  • Kubernetes service setup
  • Database migration testing

📊 SUMMARY QUALITY INDICATORS
------------------------------
Overall Quality Score: 85/100

Quality Factors:
  ✓ Appropriate length
  ✓ Contains meaningful content
  ✓ Contains 4 technical terms
  ✓ Complete data available

🔍 LLM DEBUG INFORMATION
==================================================
Configuration:
  • LLM Mode: embedded
  • Model: enhanced-embedded
  • Debug Mode: true
  • Show Quality: true

LLM Processing Report:
  • Session ID: abc123
  • Processing Steps: 8
  • Success Rate: 100%
  • Quality Score: 85/100
```

## Jira OAuth Setup

**Quick Setup:**
1. Go to [Atlassian Developer Console](https://developer.atlassian.com/console/myapps/)
2. Create OAuth 2.0 app with redirect URI: `http://localhost:8080/callback`
3. Add scopes: `read:jira-user`, `read:jira-work`
4. Copy Client ID and Secret to your config
5. Run `my-day auth` to authenticate

**📋 For detailed setup instructions, see [Jira Setup Guide](docs/jira-setup.md)**

## LLM Integration

### Embedded Mode (Default)

Uses a lightweight embedded model for basic ticket summarization. No additional setup required.

**Enhanced Features:**
- **Technical Pattern Matching**: Recognizes DevOps terminology (AWS, Terraform, Kubernetes, etc.)
- **Intelligent Summarization**: Context-aware comment analysis
- **Debug Mode**: Detailed processing information with `--debug` flag
- **Quality Indicators**: Summary quality scoring with `--show-quality` flag

### Ollama Mode

For more advanced summarization:

1. Install [Ollama](https://ollama.ai/)
2. Pull a model: `ollama pull llama3.1`
3. Configure my-day:

```yaml
llm:
  mode: "ollama"
  ollama:
    base_url: "http://localhost:11434"
    model: "llama3.1"
```

### Disabled Mode

```yaml
llm:
  enabled: false
```

### Enhanced LLM Features

Generate reports with enhanced analysis:

```bash
# Debug mode with detailed processing info
./my-day report --debug

# Show summary quality indicators
./my-day report --show-quality

# Verbose processing with enhanced context
./my-day report --verbose --debug --show-quality
```

### LLM Troubleshooting

**Problem**: LLM summarization not working
- **Solution**: Check LLM configuration and run `./my-day llm status`

**Problem**: Poor summary quality
- **Solution**: Enable debug mode to see processing details: `./my-day report --debug --show-quality`

**Problem**: Ollama connection issues
- **Solution**: Verify Ollama is running: `curl http://localhost:11434/api/tags`

**Problem**: Summaries too generic
- **Solution**: Add more detailed comments to Jira tickets with technical terms

**Problem**: Missing technical context
- **Solution**: Include keywords like "terraform", "aws", "kubernetes" in comments

**Examples of Good vs Poor LLM Summaries**:

**Good Summary** (detailed, technical):
```
Completed AWS Lambda deployment using Terraform. VPC security groups updated and tested in staging environment. Database migration scripts ready for production.
```

**Poor Summary** (generic):
```
Recent activity: 3 issues, 5 comments
```

**Tips for Better Summaries**:
- Use specific technical terms in Jira comments
- Include deployment status and environments
- Mention specific actions taken (deployed, configured, tested)
- Reference infrastructure components and tools used

## Development

### Building

```bash
go build -o my-day
```

### Running Tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Documentation

- 📋 **[Jira Setup Guide](docs/jira-setup.md)** - Complete OAuth configuration walkthrough
- 🚀 **[Installation Guide](docs/installation.md)** - Install and initial setup
- 📖 **[Usage Guide](docs/usage.md)** - Commands, workflows, and examples
- ⚙️ **[Configuration Reference](docs/configuration.md)** - All configuration options

## Support

- 🐛 [Issues](https://github.com/your-org/my-day/issues)
- 💬 [Discussions](https://github.com/your-org/my-day/discussions)