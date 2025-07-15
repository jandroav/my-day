# Installation Guide

This guide walks you through installing and setting up my-day CLI tool.

## System Requirements

- **Operating System**: macOS, Linux, Windows
- **Jira**: Jira Cloud instance with admin access for OAuth setup
- **Go** (for building from source): 1.21 or later

## Installation Methods

### Method 1: Download Binary (Recommended)

1. Go to the [Releases page](https://github.com/your-org/my-day/releases)
2. Download the appropriate binary for your platform:
   - `my-day-darwin-amd64` (macOS Intel)
   - `my-day-darwin-arm64` (macOS Apple Silicon)
   - `my-day-linux-amd64` (Linux 64-bit)
   - `my-day-windows-amd64.exe` (Windows 64-bit)

3. Make it executable and move to your PATH:

**macOS/Linux:**
```bash
chmod +x my-day-*
sudo mv my-day-* /usr/local/bin/my-day
```

**Windows:**
Move the `.exe` file to a directory in your PATH or add its location to PATH.

### Method 2: Build from Source

```bash
git clone https://github.com/your-org/my-day.git
cd my-day
go build -o my-day
sudo mv my-day /usr/local/bin/  # Optional: install globally
```

### Method 3: Go Install

```bash
go install github.com/your-org/my-day@latest
```

## Verification

Verify the installation:

```bash
my-day version
```

You should see output similar to:
```
my-day - DevOps Daily Standup Report Generator

Version: v1.0.0
Commit:  abc123
Date:    2025-01-15
```

## Initial Setup

### 1. Initialize Configuration

```bash
my-day init
```

This creates `~/.my-day/config.yaml` with default settings.

### 2. Jira OAuth Application Setup

Before you can authenticate, you need to create a Jira OAuth application.

**📋 For complete OAuth setup instructions, see [Jira Setup Guide](jira-setup.md)**

**Quick Steps:**

1. **Go to Atlassian Developer Console**
   - Visit: https://developer.atlassian.com/console/myapps/
   - Sign in with your Atlassian account

2. **Create OAuth 2.0 Integration**
   - Click "Create" → "OAuth 2.0 integration"
   - Name: "my-day CLI"
   - App URL: `http://localhost:8080` (or your preference)

3. **Configure Permissions**
   - Add these scopes:
     - `read:jira-user` - Read user information
     - `read:jira-work` - Read work items (issues, worklogs)

4. **Set Redirect URI**
   - Authorization callback URL: `http://localhost:8080/callback`
   - This must match exactly what's in your config

5. **Get Credentials**
   - Copy the **Client ID** and **Client Secret**
   - You'll need these for configuration

### 3. Configure my-day

Edit your configuration file:

```bash
my-day config edit
```

Or manually edit `~/.my-day/config.yaml`:

```yaml
jira:
  base_url: "https://your-company.atlassian.net"  # Replace with your Jira URL
  oauth:
    client_id: "your-client-id-here"              # From step 2
    client_secret: "your-client-secret-here"      # From step 2
    redirect_uri: "http://localhost:8080/callback"
  projects:
    - key: "DEVOPS"     # Replace with your project keys
      name: "DevOps Team"
    - key: "INTEROP"
      name: "Interop Team"
    # Add more projects as needed
```

### 4. Authenticate

```bash
my-day auth
```

This will:
1. Start a local server on port 8080
2. Open your browser to Jira's OAuth page
3. After you approve, save the authentication token
4. Test the connection

### 5. Initial Sync

```bash
my-day sync
```

This pulls your recent tickets and caches them locally.

### 6. Generate First Report

```bash
my-day report
```

You should see a colorful daily standup report!

## Troubleshooting

### Authentication Issues

**"OAuth client not found" error:**
- Verify your Client ID is correct
- Ensure the OAuth app is enabled in Atlassian Developer Console

**"Invalid redirect URI" error:**
- Check that redirect URI in config matches exactly what's set in OAuth app
- Default is `http://localhost:8080/callback`

**Browser doesn't open automatically:**
```bash
my-day auth --no-browser
```
Then manually open the displayed URL.

### Connection Issues

**"Connection refused" or timeout errors:**
- Verify your Jira base URL is correct
- Check if you're behind a corporate firewall
- Try testing connection: `my-day auth --test`

**SSL certificate errors:**
- Your Jira instance might use self-signed certificates
- Contact your Jira administrator

### Configuration Issues

**"Config file not found":**
```bash
my-day init --force  # Recreate config file
```

**"No project keys configured":**
Edit your config file and add your Jira project keys.

**Permission denied errors:**
```bash
# Fix config directory permissions
chmod 755 ~/.my-day
chmod 644 ~/.my-day/config.yaml
```

### Getting Help

**Show current configuration:**
```bash
my-day config show
```

**Show configuration file location:**
```bash
my-day config path
```

**Verbose output for debugging:**
```bash
my-day sync --verbose
```

**Clear authentication and start over:**
```bash
my-day auth --clear
my-day auth
```

## Next Steps

- **[Jira Setup Guide](jira-setup.md)** - Complete OAuth configuration walkthrough
- **[Usage Guide](usage.md)** - Learn all commands and features  
- **[Configuration Reference](configuration.md)** - Detailed configuration options
- [Contributing](../README.md#contributing) - Help improve my-day

## Security Notes

- OAuth tokens are stored in `~/.my-day/auth.json` with 600 permissions
- Never commit your config file with secrets to version control
- Use environment variables for CI/CD: `MY_DAY_JIRA_CLIENT_SECRET`
- The local OAuth server only runs during authentication and binds to localhost