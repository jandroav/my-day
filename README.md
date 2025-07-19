# my-day

A colorful Golang CLI tool that helps DevOps team members track Jira tickets across multiple teams and generate daily standup reports with AI-powered summarization.

## 📑 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Complete Command Reference](#-complete-command-reference)
- [Configuration](#configuration)
- [Usage Examples & Workflows](#-usage-examples--workflows)
- [Jira API Token Setup](#jira-api-token-setup)
- [LLM Integration](#llm-integration)
- [Troubleshooting & FAQ](#-troubleshooting--faq)
- [Development](#️-development)
- [Documentation](#documentation)
- [Support](#support)

## ✨ Features

- 🎯 **Multi-team Support**: Track tickets across DevOps, Interop, Foundation, Enterprise, and LBIO teams
- 🔐 **Simple Authentication**: Secure API token authentication with Jira Cloud (recommended by Atlassian)
- 📊 **Daily Reports**: Generate colorful console or markdown reports for standups
- 📝 **Obsidian Export**: Export reports to Obsidian-compatible markdown with interconnected daily notes
- ⚙️ **Flexible Configuration**: YAML config, CLI flags, and environment variables
- 🚀 **Fast & Offline**: Local caching for quick report generation
- 🤖 **AI Summarization**: Optional embedded LLM or Ollama integration with enhanced features
- 🌈 **Colorful Output**: Beautiful terminal interface with status icons
- 🔍 **Debug Mode**: Detailed processing information and quality indicators
- 📈 **Quality Metrics**: Summary quality scoring and recommendations

## 📦 Installation

### Option 1: Download Binary (Recommended)

1. **Download the latest release** from [releases](https://github.com/jandroav/my-day/releases)
2. **Extract the binary** to your preferred location
3. **Make it executable** (Linux/macOS):
   ```bash
   chmod +x my-day
   ```
4. **Add to PATH** (optional but recommended):
   ```bash
   sudo mv my-day /usr/local/bin/
   ```

### Option 2: Build from Source

**Prerequisites:**
- Go 1.21 or higher
- Git

**Steps:**
```bash
# Clone the repository
git clone https://github.com/jandroav/my-day.git
cd my-day

# Build the binary
go build -o my-day

# Optional: Install to your PATH
sudo mv my-day /usr/local/bin/
```

### Option 3: Install with Go

```bash
go install github.com/jandroav/my-day@latest
```

## 🚀 Quick Start

### 1. Initialize Configuration

```bash
my-day init
```

This creates `~/.my-day/config.yaml` with default settings.

### 2. Configure Jira Base URL

Edit your configuration file to add your Jira URL:

```yaml
jira:
  base_url: "https://your-company.atlassian.net"
```

### 3. Create API Token

1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a name (e.g., "my-day CLI") and copy the token

### 4. Authenticate

You can authenticate in three ways:

#### Option A: CLI Flags (Quick Setup)
```bash
my-day auth --email your-email@example.com --token your-api-token
```

#### Option B: Configuration File (Recommended)
Add to your `~/.my-day/config.yaml`:
```yaml
jira:
  base_url: "https://your-company.atlassian.net"
  email: "your-email@example.com"
  token: "your-api-token"
```
Then run:
```bash
my-day auth
```

#### Option C: Environment Variables (CI/CD Friendly)
```bash
export MY_DAY_JIRA_EMAIL="your-email@example.com"
export MY_DAY_JIRA_TOKEN="your-api-token"
my-day auth
```

All methods save your credentials for future use.

**🔒 Security Note:** For better security, consider using environment variables instead of storing tokens in config files, especially in shared environments.

### 5. Sync Your Tickets

```bash
my-day sync
```

### 6. Generate Daily Report

```bash
my-day report
```

## 📋 Complete Command Reference

### Root Command
```bash
my-day [global-flags] <command> [command-flags] [arguments]
```

### Global Flags
Available on all commands:

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Config file path | `$HOME/.my-day/config.yaml` |
| `-v, --verbose` | Enable verbose output | `false` |
| `-q, --quiet` | Enable quiet output | `false` |
| `--jira-url` | Jira base URL | - |
| `--jira-email` | Jira email for API token | - |
| `--jira-token` | Jira API token | - |
| `--projects` | Jira project keys (comma-separated) | - |
| `--llm-mode` | LLM mode (embedded\|ollama\|disabled) | `embedded` |
| `--llm-model` | LLM model name | `tinyllama` |
| `--llm-enabled` | Enable LLM features | `true` |
| `--ollama-url` | Ollama base URL | `http://localhost:11434` |
| `--ollama-model` | Ollama model name | `llama3.1` |
| `--report-format` | Report format (console\|markdown) | `console` |
| `--include-yesterday` | Include yesterday's work | `true` |
| `--include-today` | Include today's work | `true` |
| `--include-in-progress` | Include in-progress tickets | `true` |

### Commands

#### 1. `my-day init`
Initialize configuration file

**Usage:**
```bash
my-day init [flags]
```

**Flags:**
- `--force` - Overwrite existing configuration file

**Examples:**
```bash
my-day init
my-day init --force
```

#### 2. `my-day auth`
Authenticate with Jira using API token

**Usage:**
```bash
my-day auth [flags]
```

**Flags:**
- `--email` - Email address for API token authentication (can be set in config)
- `--token` - API token for authentication (can be set in config)
- `--clear` - Clear existing authentication
- `--test` - Test existing authentication

**Examples:**
```bash
# Using CLI flags
my-day auth --email your-email@example.com --token your-api-token

# Using config file (email and token set in config.yaml)
my-day auth

# Using environment variables
MY_DAY_JIRA_EMAIL=your-email@example.com MY_DAY_JIRA_TOKEN=your-token my-day auth

# Clear authentication
my-day auth --clear

# Test authentication
my-day auth --test
```

#### 3. `my-day sync`
Sync tickets from Jira

**Usage:**
```bash
my-day sync [flags]
```

**Flags:**
- `--max-results` - Maximum tickets to fetch (default: 100)
- `--force` - Force sync even if recently synced
- `--worklog` - Include worklog entries (default: true)
- `--since` - Sync tickets updated since duration ago (default: 168h)

**Examples:**
```bash
my-day sync
my-day sync --max-results 200
my-day sync --force
my-day sync --since 48h
my-day sync --worklog=false
```

#### 4. `my-day report`
Generate daily standup report

**Usage:**
```bash
my-day report [flags]
```

**Flags:**
- `--date` - Generate report for specific date (YYYY-MM-DD)
- `--output` - Output file path (default: stdout)
- `--no-llm` - Disable LLM summarization
- `--detailed` - Include detailed ticket information
- `--debug` - Enable debug output for LLM processing
- `--show-quality` - Show summary quality indicators
- `--verbose` - Show verbose LLM processing information
- `--export` - Export report to Obsidian-compatible markdown file
- `--export-folder` - Folder path for exported reports (overrides config)
- `--export-tags` - Additional tags for exported report (overrides config)

**Examples:**
```bash
my-day report
my-day report --date 2024-07-15
my-day report --output report.md
my-day report --no-llm
my-day report --detailed
my-day report --debug --show-quality --verbose
my-day report --export
my-day report --export --export-folder ~/obsidian-vault/daily-reports
my-day report --export --export-tags work,standup,devops
```

#### 5. `my-day config`
Manage configuration settings

**Subcommands:**

##### `my-day config show`
Show current configuration

**Usage:**
```bash
my-day config show [flags]
```

**Flags:**
- `--json` - Output configuration as JSON
- `--sources` - Show configuration sources

**Examples:**
```bash
my-day config show
my-day config show --json
my-day config show --sources
```

##### `my-day config edit`
Edit configuration file

**Usage:**
```bash
my-day config edit
```

##### `my-day config path`
Show configuration file path

**Usage:**
```bash
my-day config path
```

#### 6. `my-day llm`
Manage LLM integration

**Subcommands:**

##### `my-day llm test`
Test LLM connectivity and functionality

**Usage:**
```bash
my-day llm test
```

##### `my-day llm status`
Show LLM configuration and status

**Usage:**
```bash
my-day llm status
```

#### 7. `my-day version`
Show version information

**Usage:**
```bash
my-day version
```

### Environment Variables

All configuration can be overridden using environment variables with the `MY_DAY_` prefix:

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `MY_DAY_JIRA_BASE_URL` | Jira base URL | - |
| `MY_DAY_JIRA_EMAIL` | Jira email for API token | - |
| `MY_DAY_JIRA_TOKEN` | Jira API token | - |
| `MY_DAY_JIRA_PROJECTS` | Comma-separated project keys | - |
| `MY_DAY_LLM_MODE` | LLM mode | `embedded` |
| `MY_DAY_LLM_MODEL` | LLM model name | `tinyllama` |
| `MY_DAY_LLM_ENABLED` | Enable LLM features | `true` |
| `MY_DAY_LLM_OLLAMA_BASE_URL` | Ollama base URL | `http://localhost:11434` |
| `MY_DAY_LLM_OLLAMA_MODEL` | Ollama model name | `llama3.1` |
| `MY_DAY_REPORT_FORMAT` | Report format | `console` |
| `MY_DAY_REPORT_INCLUDE_YESTERDAY` | Include yesterday's work | `true` |
| `MY_DAY_REPORT_INCLUDE_TODAY` | Include today's work | `true` |
| `MY_DAY_REPORT_INCLUDE_IN_PROGRESS` | Include in-progress tickets | `true` |
| `MY_DAY_REPORT_EXPORT_ENABLED` | Enable export to markdown | `false` |
| `MY_DAY_REPORT_EXPORT_FOLDER_PATH` | Export folder path | `~/Documents/my-day-reports` |
| `MY_DAY_REPORT_EXPORT_FILENAME_DATE` | Date format for filenames | `2006-01-02` |
| `MY_DAY_REPORT_EXPORT_TAGS` | Comma-separated export tags | `report,my-day` |
| `MY_DAY_VERBOSE` | Enable verbose output | `false` |
| `MY_DAY_QUIET` | Enable quiet output | `false` |

### Configuration Priority

Configuration values are applied in the following order (highest to lowest priority):

1. **Command line flags** (highest priority)
2. **Environment variables**
3. **Configuration file**
4. **Default values** (lowest priority)

## Configuration

### Configuration File

Default location: `~/.my-day/config.yaml`

```yaml
jira:
  base_url: "https://your-instance.atlassian.net"
  email: "your-email@example.com"      # Optional: API token email
  token: "your-api-token"              # Optional: API token (consider using env vars for security)
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
  export:
    enabled: false                           # Enable export to markdown
    folder_path: "~/Documents/my-day-reports"  # Export folder path
    filename_date: "2006-01-02"             # Date format for filenames
    tags: ["report", "my-day"]              # Tags for Obsidian
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
export MY_DAY_LLM_MODE="disabled"
```

## 📚 Usage Examples & Workflows

### Basic Daily Report

```bash
my-day report
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

### Daily Workflow Examples

#### Morning Standup Preparation
```bash
# Get today's work with enhanced analysis
my-day report --debug --show-quality

# Generate markdown report for sharing
my-day report --report-format markdown --output standup-$(date +%Y-%m-%d).md

# Export to Obsidian for daily notes
my-day report --export --export-folder ~/obsidian-vault/daily-reports

# Quick sync before standup
my-day sync --since 24h && my-day report --detailed
```

#### Weekly Review
```bash
# Generate reports for the past week
for day in {1..7}; do
  date=$(date -d "-$day days" +%Y-%m-%d)
  my-day report --date $date --output weekly-review-$date.md
done

# Export past week to Obsidian (creates interconnected daily notes)
for day in {1..7}; do
  date=$(date -d "-$day days" +%Y-%m-%d)
  my-day report --date $date --export --export-tags weekly-review,standup
done

# Sync with extended timeframe
my-day sync --since 168h --max-results 200
```

#### Custom Team Reports
```bash
# DevOps team only
my-day report --projects DEVOPS --detailed

# Multiple teams with enhanced LLM
my-day report --projects DEVOPS,INTEROP,FOUND --debug --show-quality
```

### Advanced Usage Examples

#### Enhanced LLM Analysis
```bash
# Full diagnostic mode
my-day report --debug --show-quality --verbose

# Test LLM functionality
my-day llm test

# Check LLM status and configuration
my-day llm status
```

#### Custom Configurations
```bash
# Override LLM settings
my-day report --llm-mode ollama --ollama-model llama3.1

# Disable LLM for quick report
my-day report --no-llm --detailed

# Custom date range with specific format
my-day report --date 2024-07-15 --report-format markdown --output report.md
```

#### Troubleshooting & Debug
```bash
# Debug configuration issues
my-day config show --sources

# Test authentication
my-day auth --test

# Clear and re-authenticate
my-day auth --clear && my-day auth

# Force sync with debug output
my-day sync --force --verbose
```

### Automation & Scripting

#### Daily Automation Script
```bash
#!/bin/bash
# daily-standup.sh

# Sync latest data
my-day sync --since 24h

# Generate enhanced report
my-day report --debug --show-quality --output "standup-$(date +%Y-%m-%d).md"

# Send to team channel (example with Slack)
# curl -X POST -H 'Content-type: application/json' \
#   --data '{"text":"Daily standup report ready"}' \
#   YOUR_SLACK_WEBHOOK_URL
```

#### CI/CD Integration
```bash
# In your CI/CD pipeline
export MY_DAY_JIRA_BASE_URL="https://company.atlassian.net"

# Note: For CI/CD, you'll need to authenticate once locally and copy the auth.json file
# or use a dedicated service account with API token

# Generate deployment report
my-day sync --projects DEVOPS --since 24h
my-day report --projects DEVOPS --detailed --output deployment-report.md
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

## Jira API Token Setup

**Quick Setup:**
1. Go to [Atlassian Account Security](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Click "Create API token"
3. Give it a name (e.g., "my-day CLI")
4. Copy the generated token
5. Run `my-day auth --email your-email@example.com --token your-api-token`

**Why API tokens?** Atlassian recommends API tokens for CLI tools as they're more secure and simpler to set up than OAuth applications.

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

## 📝 Obsidian Export Integration

Export your daily standup reports to Obsidian-compatible markdown files with full interconnection for the graph view.

### Features

- **YAML Frontmatter**: Proper Obsidian 2025 properties format
- **Date-based Filenames**: Configurable date format for consistent naming
- **Wiki-style Links**: Automatic navigation links to previous/next day reports
- **Tag Integration**: Configurable tags plus automatic date tags
- **Graph View Connectivity**: Perfect interconnection in Obsidian's graph view
- **Folder Organization**: Configurable export folder path

### Quick Setup

1. **Enable in Configuration**:
```yaml
report:
  export:
    enabled: true
    folder_path: "~/obsidian-vault/daily-reports"
    filename_date: "2006-01-02"
    tags: ["report", "my-day", "standup"]
```

2. **Or Use CLI Flags**:
```bash
my-day report --export --export-folder ~/obsidian-vault/daily-reports
```

### Generated File Structure

Each exported report creates a markdown file with:

```markdown
---
date: 2025-07-19
title: Daily Standup Report - July 19, 2025
type: daily-report
tags:
  - report
  - my-day
  - standup
  - 2025-07-19
created: 2025-07-19T08:30:00-07:00
---

## Navigation

← [[2025-07-18]] | [[2025-07-20]] →

# Daily Standup Report - July 19, 2025

[Your report content here...]

---

## Tags

#report #my-day #standup #2025-07-19 

## Related Notes

*This section will be automatically populated by Obsidian's backlinks*
```

### Usage Examples

#### Basic Export
```bash
# Export today's report
my-day report --export

# Export specific date
my-day report --date 2025-07-18 --export
```

#### Custom Configuration
```bash
# Custom folder and tags
my-day report --export \
  --export-folder ~/documents/obsidian-vault/work/daily-reports \
  --export-tags work,devops,standup,team-update

# For team leads - export with team-specific tags
my-day report --export --export-tags devops-team,leadership,status-update
```

#### Automation for Daily Notes
```bash
#!/bin/bash
# daily-obsidian-export.sh

# Sync latest data and export to Obsidian
my-day sync --since 24h
my-day report --export \
  --export-folder ~/obsidian-vault/daily-notes \
  --export-tags daily,work,standup,$(date +%A)

echo "Daily report exported to Obsidian!"
```

#### Backfill Historical Reports
```bash
# Export past month of reports for Obsidian
for i in {1..30}; do
  date=$(date -d "-$i days" +%Y-%m-%d)
  my-day report --date $date --export \
    --export-tags historical,standup,backfill
  echo "Exported report for $date"
done
```

### Configuration Options

| Setting | Description | Default | Example |
|---------|-------------|---------|---------|
| `enabled` | Enable export functionality | `false` | `true` |
| `folder_path` | Export destination folder | `~/Documents/my-day-reports` | `~/obsidian-vault/daily-reports` |
| `filename_date` | Date format for filenames | `2006-01-02` | `2006-01-02` (YYYY-MM-DD) |
| `tags` | Default tags for exported files | `["report", "my-day"]` | `["work", "standup", "devops"]` |

### Obsidian Integration Tips

1. **Graph View**: Daily reports will automatically link to adjacent days
2. **Tags Panel**: Use consistent tags for easy filtering and organization
3. **Search**: All reports are searchable through Obsidian's search
4. **Backlinks**: Related tickets and team members will show automatic backlinks
5. **Templates**: Reports follow consistent structure for template-based workflows

### Environment Variables

```bash
# Enable export via environment variables
export MY_DAY_REPORT_EXPORT_ENABLED=true
export MY_DAY_REPORT_EXPORT_FOLDER_PATH="~/obsidian-vault/daily-reports"
export MY_DAY_REPORT_EXPORT_TAGS="report,my-day,devops"

# Then run normally
my-day report
```

### Troubleshooting Export

**Problem**: Export folder not created
- **Solution**: Ensure parent directories exist or use absolute paths

**Problem**: Permission denied when writing files
- **Solution**: Check folder permissions: `chmod 755 ~/obsidian-vault/daily-reports`

**Problem**: Obsidian not recognizing files
- **Solution**: Ensure files are saved with `.md` extension and valid YAML frontmatter

**Problem**: Tags not appearing in Obsidian
- **Solution**: Use the 2025 format with tags as lists in frontmatter (automatic in my-day)

## ❓ Troubleshooting & FAQ

### Common Issues

#### Installation Issues

**Problem**: Binary not found after installation
```bash
# Check if binary is in PATH
which my-day

# If not found, add to PATH
export PATH=$PATH:/path/to/my-day

# Or move to standard location
sudo mv my-day /usr/local/bin/
```

**Problem**: Permission denied when running
```bash
# Make executable
chmod +x my-day
```

#### Configuration Issues

**Problem**: Config file not found
```bash
# Check config location
my-day config path

# Initialize if missing
my-day init
```

**Problem**: Invalid configuration format
```bash
# Show current config with sources
my-day config show --sources

# Edit config file
my-day config edit
```

#### Authentication Issues

**Problem**: API token authentication fails
```bash
# Test current authentication
my-day auth --test

# Clear and re-authenticate
my-day auth --clear
my-day auth --email your-email@example.com --token your-api-token
```

**Problem**: Token expired or invalid
```bash
# Re-authenticate with new token
my-day auth --email your-email@example.com --token your-new-api-token
```

#### Sync Issues

**Problem**: No tickets found
```bash
# Check projects configuration
my-day config show

# Sync with verbose output
my-day sync --verbose

# Force sync with extended timeframe
my-day sync --force --since 168h --max-results 200
```

**Problem**: Sync takes too long
```bash
# Reduce sync scope
my-day sync --since 24h --max-results 50

# Sync specific projects only
my-day sync --projects DEVOPS,INTEROP
```

#### LLM Issues

**Problem**: LLM not working
```bash
# Check LLM status
my-day llm status

# Test LLM functionality
my-day llm test

# Use debug mode
my-day report --debug --show-quality
```

**Problem**: Poor summary quality
```bash
# Enable quality indicators
my-day report --show-quality

# Use verbose mode for details
my-day report --debug --verbose
```

**Problem**: Ollama connection issues
```bash
# Test Ollama connectivity
curl http://localhost:11434/api/tags

# Check configuration
my-day config show | grep ollama
```

### FAQ

#### Q: How do I set up Jira authentication?
A: Follow the [Jira API Token Setup](#jira-api-token-setup) section. You'll need to create an API token in your Atlassian account.

#### Q: Can I use this with Jira Server (on-premises)?
A: Currently, my-day only supports Jira Cloud with API token authentication. Jira Server support is planned for future releases.

#### Q: How do I add custom projects?
A: Edit your config file or use environment variables:
```bash
# Config file
my-day config edit

# Environment variable
export MY_DAY_JIRA_PROJECTS="DEVOPS,INTEROP,CUSTOM"
```

#### Q: Can I customize the report format?
A: Yes! Use `--report-format markdown` or modify the configuration. You can also use `--output` to save to a file.

#### Q: How do I improve LLM summary quality?
A: 
- Add detailed comments to Jira tickets
- Use technical terms (AWS, Terraform, Kubernetes)
- Include specific actions (deployed, configured, tested)
- Enable debug mode to see quality metrics

#### Q: Can I run this in CI/CD?
A: Yes! You can copy the auth.json file or use a dedicated service account with API token. See [Automation & Scripting](#automation--scripting) examples.

#### Q: How do I use the Obsidian export feature?
A: Enable export in config or use `--export` flag. Reports are exported as interconnected markdown files perfect for Obsidian's graph view. See [Obsidian Export Integration](#-obsidian-export-integration) for details.

#### Q: Can I customize the export format and location?
A: Yes! Configure `report.export.folder_path`, `report.export.tags`, and `report.export.filename_date` in your config file, or use CLI flags like `--export-folder` and `--export-tags`.

#### Q: What are the different ways to set my Jira email and API token?
A: You can set them via: 1) CLI flags (`--jira-email`, `--jira-token`), 2) Config file (`jira.email`, `jira.token`), or 3) Environment variables (`MY_DAY_JIRA_EMAIL`, `MY_DAY_JIRA_TOKEN`). Environment variables are recommended for better security.

#### Q: How do I backup my configuration?
A: Copy the config file:
```bash
cp ~/.my-day/config.yaml ~/.my-day/config.yaml.backup
```

### Performance Tips

- Use `--since 24h` for daily reports instead of syncing all history
- Limit `--max-results` for faster syncing
- Use `--no-llm` for quick reports when LLM analysis isn't needed
- Cache is stored in `~/.my-day/cache.json` - delete if you have issues

### Security Notes

- API tokens are stored locally in `~/.my-day/auth.json`
- Configuration files no longer contain sensitive information
- API tokens can be easily revoked from your Atlassian account
- Regularly rotate API tokens for security

## 🛠️ Development

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

- 📋 **[Jira Setup Guide](docs/jira-setup.md)** - Complete API token setup walkthrough
- 🚀 **[Installation Guide](docs/installation.md)** - Install and initial setup
- 📖 **[Usage Guide](docs/usage.md)** - Commands, workflows, and examples
- ⚙️ **[Configuration Reference](docs/configuration.md)** - All configuration options

## Support

- 🐛 [Issues](https://github.com/jandroav/my-day/issues)
- 💬 [Discussions](https://github.com/jandroav/my-day/discussions)