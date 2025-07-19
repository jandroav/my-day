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

### Standard Setup (Recommended)

1. **Initialize Configuration**
   ```bash
   my-day init
   ```

2. **Edit Jira URL** in `~/.my-day/config.yaml`:
   ```yaml
   jira:
     base_url: "https://your-company.atlassian.net"
   ```

3. **Get API Token**: Visit https://id.atlassian.com/manage-profile/security/api-tokens

4. **Authenticate**:
   ```bash
   my-day auth --email your-email@example.com --token your-api-token
   ```

5. **Start Using**:
   ```bash
   my-day sync     # Get your tickets
   my-day report   # Generate report
   ```

### Guided Setup (For First-Time Users)

1. **Initialize with Guided Mode**
   ```bash
   my-day init --guided
   ```

2. **Follow the Simple Instructions** displayed after initialization

3. **Edit one line** in the config file to set your Jira URL

4. **Authenticate and you're ready**:
   ```bash
   my-day auth --email your-email@example.com --token your-api-token
   my-day sync && my-day report
   ```

### Alternative Authentication Methods

**Environment Variables** (CI/CD Friendly):
```bash
export MY_DAY_JIRA_EMAIL="your-email@example.com"
export MY_DAY_JIRA_TOKEN="your-api-token"
my-day auth
```

**Configuration File** (Less Secure):
```yaml
jira:
  base_url: "https://your-company.atlassian.net"
  email: "your-email@example.com"
  token: "your-api-token"
```

**🔒 Security Note:** Environment variables are recommended for better security, especially in shared environments.

## 📋 Complete Command Reference

### Root Command
```bash
my-day [global-flags] <command> [command-flags] [arguments]
```

### Global Flags
Available on all commands:

| Flag | Description | Default | Config |
|------|-------------|---------|--------|
| `--config` | Config file path | `$HOME/.my-day/config.yaml` | *file location* |
| `-v, --verbose` | Enable verbose output (config: `verbose`) | `false` | `verbose` |
| `-q, --quiet` | Enable quiet output (config: `quiet`) | `false` | `quiet` |
| `--jira-url` | Jira base URL (config: `jira.base_url`) | - | `jira.base_url` |
| `--jira-email` | Jira email for API token (config: `jira.email`) | - | `jira.email` |
| `--jira-token` | Jira API token (config: `jira.token`) | - | `jira.token` |
| `--projects` | Jira project keys, comma-separated (config: `jira.projects`) | - | `jira.projects` |
| `--llm-mode` | LLM mode: embedded\|ollama\|disabled (config: `llm.mode`) | `ollama` | `llm.mode` |
| `--llm-model` | LLM model name (config: `llm.model`) | `qwen2.5:3b` | `llm.model` |
| `--llm-enabled` | Enable LLM features (config: `llm.enabled`) | `true` | `llm.enabled` |
| `--llm-debug` | Enable LLM debug mode (config: `llm.debug`) | `false` | `llm.debug` |
| `--llm-style` | LLM summary style: technical\|business\|brief (config: `llm.summary_style`) | `technical` | `llm.summary_style` |
| `--llm-max-length` | Maximum LLM summary length, 0 for no limit (config: `llm.max_summary_length`) | `0` | `llm.max_summary_length` |
| `--llm-technical-details` | Include technical details in summaries (config: `llm.include_technical_details`) | `true` | `llm.include_technical_details` |
| `--llm-fallback` | LLM fallback strategy: graceful\|strict (config: `llm.fallback_strategy`) | `graceful` | `llm.fallback_strategy` |
| `--ollama-url` | Ollama base URL (config: `llm.ollama.base_url`) | `http://localhost:11434` | `llm.ollama.base_url` |
| `--ollama-model` | Ollama model name (config: `llm.ollama.model`) | `qwen2.5:3b` | `llm.ollama.model` |
| `--report-format` | Report format: console\|markdown (config: `report.format`) | `console` | `report.format` |
| `--include-yesterday` | Include yesterday's work (config: `report.include_yesterday`) | `true` | `report.include_yesterday` |
| `--include-today` | Include today's work (config: `report.include_today`) | `true` | `report.include_today` |
| `--include-in-progress` | Include in-progress tickets (config: `report.include_in_progress`) | `true` | `report.include_in_progress` |

### Commands

#### 1. `my-day init`
Initialize configuration file with production-ready settings

**Usage:**
```bash
my-day init [flags]
```

**Flags:**
- `--force` - Overwrite existing configuration file
- `--guided` - Interactive guided setup with simplified configuration

**Examples:**
```bash
# Standard setup with comprehensive configuration
my-day init

# Guided setup for first-time users
my-day init --guided

# Force overwrite existing config
my-day init --force
```

#### 2. `my-day auth`
Authenticate with Jira using API token

**Usage:**
```bash
my-day auth [flags]
```

**Flags:**
- `--email` - Email address for API token authentication (config: `jira.email`)
- `--token` - API token for authentication (config: `jira.token`)
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
- `--comments-since` - Look for your comments since this duration ago (default: 24h)

**Examples:**
```bash
my-day sync
my-day sync --max-results 200
my-day sync --force
my-day sync --since 48h
my-day sync --comments-since 12h
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
- `--no-llm` - Disable LLM summarization for this report
- `--detailed` - Include detailed ticket information
- `--debug` - Enable debug output for LLM processing (config: `llm.debug`)
- `--show-quality` - Show summary quality indicators
- `--verbose` - Show verbose LLM processing information (config: `verbose`)
- `--export` - Export report to markdown file (config: `report.export.enabled`)
- `--export-folder` - Folder path for exported reports (config: `report.export.folder_path`)
- `--export-tags` - Additional tags for exported report (config: `report.export.tags`)
- `--field` - Group report by specified Jira custom field (config: `jira.custom_fields`)

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
my-day report --field squad
my-day report --field team --detailed
my-day report --field customfield_12944
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

##### `my-day llm models`
List available LLM models for the current mode

**Usage:**
```bash
my-day llm models
```

##### `my-day llm switch`
Switch to a different LLM model

**Usage:**
```bash
my-day llm switch [model-name]
```

**Examples:**
```bash
my-day llm switch qwen2.5:7b
my-day llm switch llama3.1:8b
my-day llm switch enhanced-embedded
```

##### `my-day llm start`
Start Docker LLM container

**Usage:**
```bash
my-day llm start
```

##### `my-day llm stop`
Stop Docker LLM container

**Usage:**
```bash
my-day llm stop
```

#### 7. `my-day completion`
Generate shell autocompletion scripts

**Usage:**
```bash
my-day completion [bash|zsh|fish|powershell]
```

**Examples:**
```bash
# Bash completion
my-day completion bash > /etc/bash_completion.d/my-day

# Zsh completion (add to your .zshrc)
my-day completion zsh > ~/.my-day-completion.zsh
source ~/.my-day-completion.zsh

# Fish completion
my-day completion fish > ~/.config/fish/completions/my-day.fish
```

#### 8. `my-day version`
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
| `MY_DAY_LLM_MODE` | LLM mode | `ollama` |
| `MY_DAY_LLM_MODEL` | LLM model name | `qwen2.5:3b` |
| `MY_DAY_LLM_ENABLED` | Enable LLM features | `true` |
| `MY_DAY_LLM_DEBUG` | Enable LLM debug mode | `false` |
| `MY_DAY_LLM_SUMMARY_STYLE` | LLM summary style | `technical` |
| `MY_DAY_LLM_MAX_SUMMARY_LENGTH` | Maximum summary length | `0` |
| `MY_DAY_LLM_INCLUDE_TECHNICAL_DETAILS` | Include technical details | `true` |
| `MY_DAY_LLM_FALLBACK_STRATEGY` | LLM fallback strategy | `graceful` |
| `MY_DAY_LLM_OLLAMA_BASE_URL` | Ollama base URL | `http://localhost:11434` |
| `MY_DAY_LLM_OLLAMA_MODEL` | Ollama model name | `qwen2.5:3b` |
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
  base_url: "https://your-instance.atlassian.net"  # CLI: --jira-url
  email: "your-email@example.com"                   # CLI: --jira-email
  token: "your-api-token"                           # CLI: --jira-token
  projects:                                         # CLI: --projects
    - key: "DEVOPS"
      name: "DevOps Team"
    - key: "INTEROP"
      name: "Interop Team"
    # Add more projects...
  # Custom Fields Configuration (used with --field flag)
  custom_fields:
    squad:
      field_id: "customfield_12944"
      display_name: "Squad"
      field_type: "select"
    team:
      field_id: "customfield_12945"
      display_name: "Team"
      field_type: "select"
    component:
      field_id: "customfield_12946"
      display_name: "Component"
      field_type: "multi-select"

llm:
  enabled: true                             # CLI: --llm-enabled
  mode: "ollama"                           # CLI: --llm-mode (embedded, ollama, disabled)
  model: "qwen2.5:3b"                      # CLI: --llm-model
  debug: false                             # CLI: --llm-debug
  summary_style: "technical"               # CLI: --llm-style (technical, business, brief)
  max_summary_length: 0                    # CLI: --llm-max-length (0 for no limit)
  include_technical_details: true          # CLI: --llm-technical-details
  prioritize_recent_work: true             # Focus on recent activity
  fallback_strategy: "graceful"            # CLI: --llm-fallback (graceful, strict)
  ollama:
    base_url: "http://localhost:11434"     # CLI: --ollama-url
    model: "qwen2.5:3b"                    # CLI: --ollama-model

report:
  format: "console"                        # CLI: --report-format (console, markdown)
  include_yesterday: true                  # CLI: --include-yesterday
  include_today: true                      # CLI: --include-today
  include_in_progress: true                # CLI: --include-in-progress
  export:
    enabled: false                         # CLI: --export
    folder_path: "~/Documents/my-day-reports"  # CLI: --export-folder
    filename_date: "2006-01-02"           # Date format for filenames
    tags: ["report", "my-day"]             # CLI: --export-tags

# Global settings
verbose: false                             # CLI: -v, --verbose
quiet: false                               # CLI: -q, --quiet
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

## 🏷️ Custom Field Grouping

`my-day` supports grouping your daily standup reports by any Jira custom field. This is perfect for organizing reports by Squad, Team, Component, Epic, or any other custom field in your Jira instance.

### Quick Start

Use the `--field` flag to group your report by any field:

```bash
# Group by Squad
my-day report --field squad

# Group by Team
my-day report --field team

# Group by any custom field ID
my-day report --field customfield_12944
```

### Supported Field Types

The field grouping feature supports:

- **Standard Jira Fields**: `project`, `priority`, `status`, `issuetype`, `assignee`, `reporter`
- **Custom Fields**: Any custom field in your Jira instance (by field ID or configured name)
- **Common Fields**: Pre-configured mappings for `squad`, `team`, `component`, `epic`, `sprint`

### Configuration Setup

For the best experience, configure your custom fields in your config file:

```yaml
jira:
  base_url: "https://your-company.atlassian.net"
  email: "your-email@example.com"
  projects:
    - key: "DEVOPS"
      name: "DevOps Team"
    - key: "INTEROP" 
      name: "Interop Team"
  
  # Custom Fields Configuration
  custom_fields:
    squad:
      field_id: "customfield_12944"
      display_name: "DevOps Infrastructure Squad"
      field_type: "select"
    team:
      field_id: "customfield_12945"  
      display_name: "Team Assignment"
      field_type: "select"
    component:
      field_id: "customfield_12946"
      display_name: "System Component"
      field_type: "multi-select"
    epic:
      field_id: "customfield_10014"
      display_name: "Epic Link"
      field_type: "epic"
    sprint:
      field_id: "customfield_10007"
      display_name: "Sprint"
      field_type: "sprint"
```

### Finding Custom Field IDs

To find the field ID for any custom field in your Jira instance:

#### Method 1: Via Jira UI (Recommended)
1. Go to any Jira issue that has the custom field
2. Click "..." → "Configure" 
3. Click on the custom field
4. The field ID will be shown in the URL: `customfield_XXXXX`

#### Method 2: Via Browser Developer Tools
1. Open any Jira issue with the custom field
2. Right-click on the custom field → "Inspect Element"
3. Look for `data-field-id` or similar attributes
4. The field ID will be in format `customfield_XXXXX`

#### Method 3: Via Jira Admin (For Admins)
1. Go to Jira Administration → Issues → Custom Fields
2. Find your custom field in the list
3. The field ID is shown in the "ID" column

### Usage Examples

#### Basic Field Grouping
```bash
# Group by squad (uses configured field mapping)
my-day report --field squad

# Group by team with detailed info
my-day report --field team --detailed

# Group by priority (standard field)
my-day report --field priority
```

#### Direct Field ID Usage
```bash
# Use field ID directly if not configured
my-day report --field customfield_12944

# Group by Epic Link
my-day report --field customfield_10014

# Group by Sprint
my-day report --field customfield_10007
```

#### Combined with Other Features
```bash
# Grouped report with AI analysis
my-day report --field squad --debug --show-quality

# Export grouped report to Obsidian
my-day report --field team --export --export-tags squad,team-report

# Generate markdown grouped report
my-day report --field component --report-format markdown --output team-report.md
```

### Report Output Example

When using `my-day report --field squad`, your output will look like:

```
🚀 Daily Standup Report - July 19, 2025
==================================================
📝 Issues grouped by Squad

🤖 AI SUMMARY OF TODAY'S WORK
Completed infrastructure deployments across DevOps and Platform squads. Database migrations tested successfully and security configurations updated.

📊 SUMMARY
• Total issues: 8
• Groups by squad: 3
• Total comments added: 12
• Worklog entries: 5

🏷️  DEVOPS INFRASTRUCTURE SQUAD (5 issues)
------------------------------
🔄 Currently Working On:
  🔄 DEVOPS-123 [DEVOPS] AWS Lambda deployment automation
    💬 Today's work: Configured Terraform modules and tested in staging environment

✅ Recently Completed:
  ✅ DEVOPS-122 [DEVOPS] VPC security group updates
    💬 Today's work: Applied security policies and verified connectivity

🏷️  PLATFORM SQUAD (2 issues)  
------------------------------
🔄 Currently Working On:
  🔄 PLAT-456 [PLATFORM] Kubernetes ingress configuration

🏷️  UNASSIGNED (1 issues)
------------------------------
📋 To Do:
  📋 MISC-789 [SUPPORT] Documentation updates
```

### Best Practices

#### 1. Consistent Field Values
Ensure your custom field values are consistent across issues:
- Use standardized squad/team names
- Avoid typos and variations
- Consider using Jira's select lists for consistency

#### 2. Configuration Management
- Configure commonly used fields in your config file
- Use descriptive display names for better readability
- Document field IDs for your team

#### 3. Reporting Workflows
```bash
# Daily squad standup
my-day report --field squad --detailed

# Weekly team review  
my-day report --field team --date $(date -d "last monday" +%Y-%m-%d)

# Component-specific analysis
my-day report --field component --debug --show-quality
```

#### 4. Team Integration
```bash
# Generate squad-specific reports for different teams
my-day report --field squad --projects DEVOPS,PLATFORM --export
my-day report --field team --projects INTEROP,FOUNDATION --export
```

### Troubleshooting

#### Issue: Field grouping shows all tickets as "Unassigned"
**Solution**: Verify the field ID is correct and the field exists on your issues:
```bash
# Check your Jira configuration
my-day config show

# Test with a known standard field first
my-day report --field project
```

#### Issue: Custom field not found
**Solution**: Double-check the field ID format:
- Field IDs must start with `customfield_`
- Example: `customfield_12944` (not `12944`)
- Use browser dev tools to verify the exact field ID

#### Issue: Empty field values
**Solution**: Ensure issues have values for the custom field:
- Check that the field is populated on your issues
- Consider using a different field that has consistent values
- Issues without field values will appear under "Unassigned"

### Environment Variables

You can also configure custom field mappings via environment variables:

```bash
# Configure a squad field mapping
export MY_DAY_JIRA_CUSTOM_FIELDS_SQUAD_FIELD_ID="customfield_12944"
export MY_DAY_JIRA_CUSTOM_FIELDS_SQUAD_DISPLAY_NAME="Squad"

# Use in report
my-day report --field squad
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

## 🧠 LLM Integration & Model Management

my-day provides flexible LLM integration with easy model switching and configuration options.

### Quick Model Switching

```bash
# List available models for your current mode
my-day llm models

# Switch to a different model
my-day llm switch qwen2.5:7b

# Test the new model
my-day llm test

# Check current status
my-day llm status
```

### LLM Modes

#### 1. Ollama Mode (Default - Recommended)

High-quality AI summarization using Docker-based Ollama models.

**Setup:**
```bash
# Auto-start Docker container with model
my-day llm start

# Or use CLI flags
my-day report --llm-mode ollama --ollama-model qwen2.5:3b
```

**Available Models:**
- **qwen2.5:3b** (1.9GB) - Fast, current default
- **llama3.2:3b** (2.0GB) - Meta's efficient model
- **phi3:3.8b** (2.3GB) - Microsoft's technical content model
- **llama3.1:8b** (4.7GB) - Better understanding, slower
- **codellama:7b** (3.8GB) - Specialized for code/technical content
- **mistral:7b** (4.1GB) - General-purpose model

**Model Management:**
```bash
# List available models
my-day llm models

# Switch models easily
my-day llm switch llama3.1:8b
my-day llm switch codellama:7b

# Install new models via Ollama
ollama pull mistral:7b
```

#### 2. Embedded Mode

Lightweight built-in summarization for basic needs.

**Setup:**
```bash
# Switch to embedded mode
my-day report --llm-mode embedded --llm-model enhanced-embedded
```

**Available Models:**
- **enhanced-embedded** - Pattern matching with technical term recognition
- **basic-embedded** - Simple keyword extraction

**Features:**
- No external dependencies
- Fast processing
- Technical pattern matching
- DevOps terminology recognition

#### 3. Disabled Mode

Disable AI features entirely:

```bash
# Disable via CLI
my-day report --llm-enabled=false

# Or via config
llm:
  enabled: false
```

### Advanced LLM Configuration

#### CLI Flags for Fine-Tuning

```bash
# Customize LLM behavior on-the-fly
my-day report \
  --llm-mode ollama \
  --ollama-model qwen2.5:7b \
  --llm-style business \
  --llm-max-length 150 \
  --llm-debug \
  --llm-technical-details=false

# Quick model comparison
my-day report --ollama-model llama3.1:8b --debug
my-day report --ollama-model codellama:7b --debug
```

#### Configuration File Options

```yaml
llm:
  enabled: true
  mode: "ollama"
  model: "qwen2.5:3b"
  debug: false
  summary_style: "technical"      # technical, business, brief
  max_summary_length: 0          # 0 for no limit
  include_technical_details: true
  prioritize_recent_work: true
  fallback_strategy: "graceful"   # graceful, strict
  ollama:
    base_url: "http://localhost:11434"
    model: "qwen2.5:3b"
```

#### Environment Variables

```bash
# Switch models via environment
export MY_DAY_LLM_MODE="ollama"
export MY_DAY_LLM_OLLAMA_MODEL="mistral:7b"
export MY_DAY_LLM_SUMMARY_STYLE="business"
export MY_DAY_LLM_DEBUG="true"

my-day report
```

### Model Recommendations by Use Case

#### DevOps/Infrastructure Teams
```bash
# Best for technical summaries with infrastructure terms
my-day llm switch codellama:7b
my-day report --llm-style technical --llm-technical-details
```

#### Management/Business Users  
```bash
# Better for business-focused summaries
my-day llm switch qwen2.5:7b
my-day report --llm-style business --llm-max-length 100
```

#### Quick Daily Standups
```bash
# Fast model for quick morning reports
my-day llm switch qwen2.5:3b
my-day report --llm-style brief
```

#### Detailed Analysis
```bash
# High-quality model for comprehensive analysis
my-day llm switch llama3.1:8b
my-day report --llm-style technical --debug --show-quality
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
- **Switch to a better model**: `my-day llm switch codellama:7b` or `my-day llm switch llama3.1:8b`
- **Add detailed comments** to Jira tickets
- **Use technical terms** (AWS, Terraform, Kubernetes)
- **Include specific actions** (deployed, configured, tested)
- **Enable debug mode** to see quality metrics: `my-day report --debug --show-quality`
- **Try different summary styles**: `my-day report --llm-style technical` or `--llm-style business`

#### Q: How do I switch between LLM models?
A: Multiple ways to switch models:
```bash
# Discover available models
my-day llm models

# Switch via command
my-day llm switch qwen2.5:7b

# Test via CLI flag
my-day report --ollama-model mistral:7b

# Set via environment
export MY_DAY_LLM_OLLAMA_MODEL="codellama:7b"

# Update config file
# Edit ~/.my-day/config.yaml ollama.model setting
```

#### Q: Which LLM model should I use?
A: Depends on your needs:
- **qwen2.5:3b** - Good default, fast and balanced
- **codellama:7b** - Best for technical/DevOps content
- **llama3.1:8b** - Higher quality, slower
- **phi3:3.8b** - Good for Microsoft tech stacks
- **embedded** - No setup required, basic functionality

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