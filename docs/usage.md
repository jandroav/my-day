# Usage Guide

Complete guide to using my-day CLI for daily standup reporting.

## Quick Reference

```bash
# Setup
my-day init                    # Initialize configuration
my-day auth                    # Authenticate with Jira
my-day sync                    # Pull latest tickets

# Daily workflow
my-day report                  # Generate today's report
my-day report --detailed       # Detailed report with descriptions
my-day report --date 2025-01-14  # Report for specific date

# Configuration
my-day config show            # View current settings
my-day config edit            # Edit config file
```

## Commands

### `my-day init`

Initialize configuration file with defaults.

```bash
my-day init [flags]
```

**Flags:**
- `--force` - Overwrite existing configuration

**Examples:**
```bash
my-day init                    # Create default config
my-day init --force           # Recreate config file
```

### `my-day auth`

Manage Jira authentication.

```bash
my-day auth [flags]
```

**Flags:**
- `--clear` - Clear stored authentication
- `--test` - Test existing authentication
- `--no-browser` - Don't auto-open browser

**Examples:**
```bash
my-day auth                    # Authenticate with Jira
my-day auth --test            # Test current authentication
my-day auth --clear           # Clear and re-authenticate
my-day auth --no-browser      # Manual browser opening
```

### `my-day sync`

Sync tickets from Jira to local cache.

```bash
my-day sync [flags]
```

**Flags:**
- `--max-results int` - Maximum tickets to fetch (default: 100)
- `--force` - Force sync even if recently synced
- `--worklog` - Include worklog entries (default: true)
- `--since duration` - Sync tickets updated since duration ago (default: 168h)

**Examples:**
```bash
my-day sync                           # Standard sync
my-day sync --max-results 200        # Fetch more tickets
my-day sync --force                   # Force immediate sync
my-day sync --since 48h              # Only last 48 hours
my-day sync --worklog=false          # Skip worklog entries
```

### `my-day report`

Generate daily standup report.

```bash
my-day report [flags]
```

**Flags:**
- `--date string` - Report date (YYYY-MM-DD format)
- `--output string` - Output file path (default: stdout)
- `--detailed` - Include detailed ticket information
- `--no-llm` - Disable LLM summarization

**Examples:**
```bash
my-day report                         # Today's report
my-day report --date 2025-01-14     # Specific date
my-day report --detailed             # With descriptions
my-day report --output report.md     # Save to file
my-day report --no-llm              # Without AI summaries
```

### `my-day config`

Manage configuration settings.

#### `my-day config show`

Display current configuration.

```bash
my-day config show [flags]
```

**Flags:**
- `--json` - Output as JSON
- `--sources` - Show configuration sources

**Examples:**
```bash
my-day config show                   # Human-readable format
my-day config show --json           # JSON format
my-day config show --sources        # Show config sources
```

#### `my-day config edit`

Open configuration file in editor.

```bash
my-day config edit
```

Uses `$EDITOR` environment variable or defaults to `vi`.

#### `my-day config path`

Show configuration file path.

```bash
my-day config path
```

### `my-day version`

Show version information.

```bash
my-day version
```

## Global Flags

Available on all commands:

### Configuration
- `--config string` - Config file path (default: `~/.my-day/config.yaml`)

### Jira Settings
- `--jira-url string` - Jira base URL
- `--jira-client-id string` - OAuth client ID
- `--jira-client-secret string` - OAuth client secret
- `--projects strings` - Project keys to track

### LLM Settings
- `--llm-mode string` - LLM mode: embedded, ollama, disabled
- `--llm-enabled` - Enable LLM features (default: true)
- `--llm-model string` - Model name

### Report Settings
- `--report-format string` - Output format: console, markdown
- `--include-yesterday` - Include yesterday's work (default: true)
- `--include-today` - Include today's work (default: true)
- `--include-in-progress` - Include in-progress tickets (default: true)

### Output
- `--verbose, -v` - Verbose output
- `--quiet, -q` - Quiet output

## Common Workflows

### Daily Standup Preparation

```bash
# Morning routine - sync and generate report
my-day sync && my-day report --detailed

# Save report for sharing
my-day report --output standup-$(date +%Y-%m-%d).md --report-format markdown
```

### Team-Specific Reports

```bash
# DevOps team only
my-day report --projects DEVOPS

# Multiple teams
my-day report --projects DEVOPS,INTEROP,FOUNDATION
```

### Different Report Formats

```bash
# Console output (default)
my-day report

# Markdown for documentation
my-day report --report-format markdown

# Detailed with descriptions
my-day report --detailed

# Minimal without LLM
my-day report --no-llm --report-format console
```

### Working with Dates

```bash
# Yesterday's work
my-day report --date $(date -d "yesterday" +%Y-%m-%d)

# Last Friday (macOS)
my-day report --date $(date -v-fri +%Y-%m-%d)

# Specific date
my-day report --date 2025-01-10
```

### Configuration Override Examples

```bash
# Use different Jira instance
my-day sync --jira-url https://other-company.atlassian.net

# Disable LLM for this report
my-day report --llm-mode disabled

# Different project set
my-day sync --projects DEV,QA,OPS --max-results 50
```

## Environment Variables

Override any setting with environment variables:

```bash
# Jira settings
export MY_DAY_JIRA_BASE_URL="https://company.atlassian.net"
export MY_DAY_JIRA_CLIENT_ID="your-client-id"
export MY_DAY_JIRA_CLIENT_SECRET="your-secret"

# Project settings
export MY_DAY_JIRA_PROJECTS="DEVOPS,INTEROP,FOUNDATION"

# LLM settings
export MY_DAY_LLM_MODE="ollama"
export MY_DAY_LLM_ENABLED="true"

# Report settings
export MY_DAY_REPORT_FORMAT="markdown"
export MY_DAY_INCLUDE_YESTERDAY="false"

# Then run without flags
my-day report  # Uses environment variables
```

## Output Examples

### Console Report
```
🚀 Daily Standup Report - January 15, 2025
==================================================

📊 SUMMARY
• Issues: 5
• Worklog entries: 3

🔄 CURRENTLY WORKING ON
  🔄 DEV-123 [DEVOPS] Fix CI/CD pipeline timeout issues
  🔄 INT-456 [INTEROP] API integration for new service

✅ RECENTLY COMPLETED
  ✅ DEV-122 [DEVOPS] Update deployment scripts for k8s
  ✅ FOUND-789 [FOUND] Database migration for user table

📋 TO DO
  📋 DEV-124 [DEVOPS] Review security scan results

⏰ WORK LOG
  ⏱️  [DEV-123] Jan 15, 09:30
    Investigated timeout issue in build pipeline
  
  ⏱️  [INT-456] Jan 15, 14:15
    Fixed API endpoint response format

---
Generated by my-day CLI 🤖
```

### Markdown Report
```markdown
# Daily Standup Report - January 15, 2025

## Summary
- **Issues**: 5
- **Worklog entries**: 3

## 🔄 Currently Working On
- 🔄 **[DEV-123]** Fix CI/CD pipeline timeout issues
- 🔄 **[INT-456]** API integration for new service

## ✅ Recently Completed
- ✅ **[DEV-122]** Update deployment scripts for k8s
- ✅ **[FOUND-789]** Database migration for user table

## ⏰ Work Log
- ⏱️ **[DEV-123]** Jan 15, 09:30
  - Investigated timeout issue in build pipeline

---
*Generated by my-day CLI*
```

## Tips and Best Practices

### Automation

Create shell aliases for common tasks:

```bash
# Add to ~/.bashrc or ~/.zshrc
alias standup='my-day sync && my-day report --detailed'
alias yesterday='my-day report --date $(date -d "yesterday" +%Y-%m-%d)'
alias saveup='my-day report --output ~/standups/$(date +%Y-%m-%d).md --report-format markdown'
```

### Scheduling

Set up daily sync with cron:

```bash
# Add to crontab: sync every morning at 8 AM
0 8 * * 1-5 /usr/local/bin/my-day sync >/dev/null 2>&1
```

### Team Integration

Share reports via Slack/Teams:

```bash
#!/bin/bash
# daily-standup.sh
my-day sync
REPORT=$(my-day report --report-format markdown)
curl -X POST -H 'Content-type: application/json' \
  --data "{\"text\":\"$REPORT\"}" \
  YOUR_SLACK_WEBHOOK_URL
```

### Troubleshooting Commands

```bash
# Debug authentication
my-day auth --test --verbose

# Check configuration
my-day config show --sources

# Force fresh data
my-day sync --force --verbose

# Test with minimal settings
my-day report --no-llm --projects DEVOPS
```