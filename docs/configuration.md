# Configuration Reference

Complete reference for configuring my-day CLI tool.

## Configuration Sources

my-day uses a layered configuration system with the following priority order:

1. **Command line flags** (highest priority)
2. **Environment variables** (`MY_DAY_*` prefix)
3. **Configuration file** (YAML)
4. **Default values** (lowest priority)

## Configuration File

### Location

- Default: `~/.my-day/config.yaml`
- Custom: Use `--config /path/to/config.yaml` flag

### Structure

```yaml
# Complete configuration example
jira:
  base_url: "https://your-company.atlassian.net"
  oauth:
    client_id: "your-oauth-client-id"
    client_secret: "your-oauth-client-secret"
    redirect_uri: "http://localhost:8080/callback"
  projects:
    - key: "DEVOPS"
      name: "DevOps Team"
    - key: "INTEROP"
      name: "Interop Team"
    - key: "FOUNDATION"
      name: "Foundation Team"
    - key: "ENTERPRISE"
      name: "Enterprise Team"
    - key: "LBIO"
      name: "LBIO Team"

llm:
  enabled: true
  mode: "embedded"  # embedded, ollama, disabled
  model: "tinyllama"
  ollama:
    base_url: "http://localhost:11434"
    model: "llama3.1"

report:
  format: "console"  # console, markdown
  include_yesterday: true
  include_today: true
  include_in_progress: true
```

## Jira Configuration

### Base URL
The URL of your Jira instance.

```yaml
jira:
  base_url: "https://your-company.atlassian.net"
```

**CLI Flag:** `--jira-url`  
**Environment:** `MY_DAY_JIRA_BASE_URL`

### OAuth Settings

OAuth 2.0 credentials for Jira authentication.

```yaml
jira:
  oauth:
    client_id: "your-oauth-client-id"
    client_secret: "your-oauth-client-secret"
    redirect_uri: "http://localhost:8080/callback"
```

**CLI Flags:**
- `--jira-client-id`
- `--jira-client-secret`
- `--jira-redirect-uri`

**Environment Variables:**
- `MY_DAY_JIRA_CLIENT_ID`
- `MY_DAY_JIRA_CLIENT_SECRET`
- `MY_DAY_JIRA_REDIRECT_URI`

### Projects

List of Jira projects to track.

```yaml
jira:
  projects:
    - key: "DEVOPS"
      name: "DevOps Team"
    - key: "SEC"
      name: "Security Team"
```

**CLI Flag:** `--projects DEVOPS,SEC,OPS`  
**Environment:** `MY_DAY_JIRA_PROJECTS=DEVOPS,SEC,OPS`

## LLM Configuration

### Basic Settings

```yaml
llm:
  enabled: true
  mode: "embedded"  # embedded, ollama, disabled
  model: "tinyllama"
```

**CLI Flags:**
- `--llm-enabled`
- `--llm-mode`
- `--llm-model`

**Environment Variables:**
- `MY_DAY_LLM_ENABLED`
- `MY_DAY_LLM_MODE`
- `MY_DAY_LLM_MODEL`

### LLM Modes

#### Embedded Mode (Default)
Uses a lightweight embedded model for basic summarization.

```yaml
llm:
  mode: "embedded"
  model: "tinyllama"  # Model to embed
```

**Pros:**
- Zero configuration
- Works offline
- No external dependencies

**Cons:**
- Larger binary size
- Limited model capabilities

#### Ollama Mode
Connects to a local Ollama instance for advanced summarization.

```yaml
llm:
  mode: "ollama"
  ollama:
    base_url: "http://localhost:11434"
    model: "llama3.1"
```

**CLI Flags:**
- `--ollama-url`
- `--ollama-model`

**Environment Variables:**
- `MY_DAY_OLLAMA_BASE_URL`
- `MY_DAY_OLLAMA_MODEL`

**Requirements:**
- [Ollama](https://ollama.ai/) installed and running
- Model pulled: `ollama pull llama3.1`

#### Disabled Mode
Disables all LLM functionality.

```yaml
llm:
  enabled: false
  mode: "disabled"
```

## Report Configuration

### Format

Output format for reports.

```yaml
report:
  format: "console"  # console, markdown
```

**CLI Flag:** `--report-format`  
**Environment:** `MY_DAY_REPORT_FORMAT`

**Options:**
- `console` - Colorful terminal output
- `markdown` - Markdown format for documentation

### Date Ranges

Control which tickets are included based on update dates.

```yaml
report:
  include_yesterday: true
  include_today: true
  include_in_progress: true
```

**CLI Flags:**
- `--include-yesterday`
- `--include-today`
- `--include-in-progress`

**Environment Variables:**
- `MY_DAY_INCLUDE_YESTERDAY`
- `MY_DAY_INCLUDE_TODAY`
- `MY_DAY_INCLUDE_IN_PROGRESS`

## Environment Variables

All configuration options can be set via environment variables with the `MY_DAY_` prefix:

### Jira Variables
```bash
MY_DAY_JIRA_BASE_URL="https://company.atlassian.net"
MY_DAY_JIRA_CLIENT_ID="your-client-id"
MY_DAY_JIRA_CLIENT_SECRET="your-secret"
MY_DAY_JIRA_REDIRECT_URI="http://localhost:8080/callback"
MY_DAY_JIRA_PROJECTS="DEVOPS,INTEROP,FOUNDATION"
```

### LLM Variables
```bash
MY_DAY_LLM_ENABLED="true"
MY_DAY_LLM_MODE="embedded"
MY_DAY_LLM_MODEL="tinyllama"
MY_DAY_OLLAMA_BASE_URL="http://localhost:11434"
MY_DAY_OLLAMA_MODEL="llama3.1"
```

### Report Variables
```bash
MY_DAY_REPORT_FORMAT="markdown"
MY_DAY_INCLUDE_YESTERDAY="true"
MY_DAY_INCLUDE_TODAY="true"
MY_DAY_INCLUDE_IN_PROGRESS="true"
```

### Application Variables
```bash
MY_DAY_VERBOSE="true"
MY_DAY_QUIET="false"
```

## CLI Flags Reference

### Global Flags

Available on all commands:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | string | `~/.my-day/config.yaml` | Config file path |
| `--verbose, -v` | bool | false | Verbose output |
| `--quiet, -q` | bool | false | Quiet output |

### Jira Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--jira-url` | string | - | Jira base URL |
| `--jira-client-id` | string | - | OAuth client ID |
| `--jira-client-secret` | string | - | OAuth client secret |
| `--jira-redirect-uri` | string | `http://localhost:8080/callback` | OAuth redirect URI |
| `--projects` | []string | - | Project keys to track |

### LLM Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--llm-enabled` | bool | true | Enable LLM features |
| `--llm-mode` | string | `embedded` | LLM mode |
| `--llm-model` | string | `tinyllama` | Model name |
| `--ollama-url` | string | `http://localhost:11434` | Ollama base URL |
| `--ollama-model` | string | `llama3.1` | Ollama model |

### Report Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--report-format` | string | `console` | Output format |
| `--include-yesterday` | bool | true | Include yesterday's work |
| `--include-today` | bool | true | Include today's work |
| `--include-in-progress` | bool | true | Include in-progress tickets |

## Command-Specific Flags

### `my-day init`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Overwrite existing config |

### `my-day auth`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--clear` | bool | false | Clear stored authentication |
| `--test` | bool | false | Test existing authentication |
| `--no-browser` | bool | false | Don't auto-open browser |

### `my-day sync`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--max-results` | int | 100 | Maximum tickets to fetch |
| `--force` | bool | false | Force sync even if recent |
| `--worklog` | bool | true | Include worklog entries |
| `--since` | duration | 168h | Sync tickets updated since |

### `my-day report`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--date` | string | today | Report date (YYYY-MM-DD) |
| `--output` | string | stdout | Output file path |
| `--detailed` | bool | false | Include detailed info |
| `--no-llm` | bool | false | Disable LLM for this report |

### `my-day config show`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | false | Output as JSON |
| `--sources` | bool | false | Show config sources |

## Configuration Examples

### Minimal Configuration

```yaml
jira:
  base_url: "https://company.atlassian.net"
  oauth:
    client_id: "abc123"
    client_secret: "secret456"
  projects:
    - key: "DEV"
      name: "Development"
```

### Team-Specific Configuration

```yaml
# DevOps team configuration
jira:
  base_url: "https://company.atlassian.net"
  oauth:
    client_id: "devops-client-id"
    client_secret: "devops-secret"
  projects:
    - key: "DEVOPS"
      name: "DevOps Team"
    - key: "INFRA"
      name: "Infrastructure"
    - key: "SEC"
      name: "Security"

llm:
  enabled: true
  mode: "ollama"  # Using Ollama for better summaries
  ollama:
    model: "llama3.1"

report:
  format: "markdown"  # For documentation
  include_in_progress: true
```

### CI/CD Configuration

```yaml
# Minimal config for automated environments
jira:
  base_url: "https://company.atlassian.net"
  # OAuth credentials via environment variables
  projects:
    - key: "AUTO"
      name: "Automation"

llm:
  enabled: false  # Disable for CI/CD

report:
  format: "markdown"
  include_yesterday: false  # Only current work
```

### Security Considerations

- **Never commit secrets**: Use environment variables for sensitive data
- **File permissions**: Config files should be readable only by user (644)
- **OAuth tokens**: Stored in `~/.my-day/auth.json` with 600 permissions
- **Local server**: OAuth callback server only binds to localhost

### Troubleshooting Configuration

**View current configuration:**
```bash
my-day config show
```

**Check configuration sources:**
```bash
my-day config show --sources
```

**Override for testing:**
```bash
my-day report --jira-url https://test.atlassian.net --projects TEST
```

**Reset configuration:**
```bash
my-day init --force
```