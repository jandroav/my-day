# Jira Integration Setup Guide

Complete guide to setting up Jira OAuth integration for my-day CLI.

## Overview

my-day uses OAuth 2.0 to securely connect to your Jira Cloud instance. This guide walks you through creating the OAuth application and configuring authentication.

## Prerequisites

- **Jira Cloud instance** with admin access
- **Atlassian account** with developer permissions
- **my-day CLI** installed and initialized

## Step 1: Create Atlassian OAuth App

### 1.1 Access Developer Console

1. **Navigate to Developer Console**
   - Open https://developer.atlassian.com/console/myapps/
   - Sign in with your Atlassian account (same account that has access to your Jira)

2. **Create New App**
   - Click the blue **"Create"** button
   - Select **"OAuth 2.0 (3LO)"** from the dropdown
   - This opens the app creation wizard

### 1.2 App Information Page

**Fill out the required fields:**

- **App name**: `my-day CLI` (or your preferred name like "Daily Standup Tool")
- **Privacy policy URL**: Leave blank or enter `http://localhost` (not required for private apps)
- **User license agreement URL**: Leave blank (not required)

**Click "Next"** to continue.

### 1.3 Distribution Settings

- **Select "Private"** for distribution
  - This means only your organization can use this app
  - Perfect for internal team tools
- **Click "Create"** to create the app

### 1.4 Configure OAuth Settings

After creating the app, you'll be taken to the app management page:

1. **Click on "Authorization" in the left sidebar**
   - You should see "OAuth 2.0 (3LO)" section
   - This is where you'll configure all OAuth settings

2. **Add Callback URL:**
   - In the "OAuth 2.0 (3LO)" section, find "Callback URLs"
   - Click **"Add"** next to "Callback URLs"
   - Enter exactly: `http://localhost:8080/callback`
   - **Critical**: This must match exactly what's in your my-day config
   - **No HTTPS**: Use `http://` (not `https://`) for localhost
   - Click **"Save"** or **"Save changes"**

3. **Configure Permissions:**
   - Still in the Authorization section
   - Click **"Add"** next to "Scopes"
   - Add these required scopes:

   **Required Scopes:**
   - `read:jira-user` - Read user account information
   - `read:jira-work` - Read issues, projects, and worklogs

   **How to add scopes:**
   - Type the scope name in the search box
   - Click on the scope when it appears
   - Click **"Add"** to add it to your app

   **Optional Scopes (for future features):**
   - `write:jira-work` - Create worklogs (if you want to add time logging)
   - `read:jira-project` - Additional project information

4. **Add Your Jira Site:**
   - **This is a critical step often missed!**
   - In the Authorization section, find **"Configure"** or **"Jira platform REST API"**
   - Click **"Configure"** next to "Jira platform REST API"
   - Click **"Add"** to add a Jira site
   - Enter your Jira URL: `https://datical.atlassian.net` (your specific instance)
   - Click **"Save"**
   - **Verification**: You should see your Jira site listed under "Jira platform REST API"

5. **Save Configuration:**
   - Click **"Save changes"** at the bottom of the page

### 1.5 Get Your Credentials

**Important: Do this step after configuring scopes and callback URL**

1. **Find Your Client ID:**
   - In the "OAuth 2.0 (3LO)" section
   - Copy the **"Client ID"** 
   - It looks like: `wnHViZ5PJspKhilGfAwhyPTHIZazhzse`
   - This is safe to share and goes in your config file

2. **Generate Client Secret:**
   - Click **"New secret"** button
   - Copy the **"Client Secret"** immediately
   - It looks like: `{{ATLASSIAN_API_TOKEN}}`
   - **Important**: You can only see this once! Store it securely.
   - **Never commit this to version control**

### 1.6 Verify Your App Configuration

Before proceeding, double-check these settings in your OAuth app:

- ✅ **Callback URL**: `http://localhost:8080/callback`
- ✅ **Scopes**: `read:jira-user`, `read:jira-work`
- ✅ **Jira Site Added**: Your Jira instance (`https://datical.atlassian.net`) is listed under Jira platform REST API
- ✅ **Client ID**: Copied and ready
- ✅ **Client Secret**: Copied and stored securely
- ✅ **App Status**: Should show as "Active" or "Enabled"

**Common Issues at This Step:**
- **Missing Jira Site**: Forgot to add your specific Jira instance (most common cause of 401 errors)
- **Callback URL typo**: Must be exactly `http://localhost:8080/callback`
- **Missing scopes**: Both `read:jira-user` and `read:jira-work` are required
- **App not saved**: Make sure you clicked "Save changes"

## Step 2: Configure my-day

### 2.1 Initialize Configuration

If you haven't already:

```bash
my-day init
```

This creates `~/.my-day/config.yaml`

### 2.2 Edit Configuration

Open the config file:

```bash
my-day config edit
```

Or manually edit `~/.my-day/config.yaml`:

```yaml
jira:
  base_url: "https://your-company.atlassian.net"  # Replace with your Jira URL
  oauth:
    client_id: "ari:cloud:ecosystem::app/your-app-id"     # From Step 1.4
    client_secret: "your-oauth-client-secret"             # From Step 1.4
    redirect_uri: "http://localhost:8080/callback"        # Must match OAuth app
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
```

**Important Configuration Notes:**
- **base_url**: Your Jira Cloud URL (format: `https://yourcompany.atlassian.net`)
- **client_id**: Exact value from Atlassian Developer Console (from Step 1.5)
- **client_secret**: Keep this secure, never commit to version control (from Step 1.5)
- **redirect_uri**: Must exactly match what's configured in OAuth app (`http://localhost:8080/callback`)
- **projects**: Add your actual project keys

**⚠️ Critical**: Based on your config, you have real credentials now. Make sure:
- Never commit this config file to git
- Keep your client_secret secure
- The base_url `https://datical.atlassian.net` is correct for your organization

### 2.3 Find Your Project Keys

To find your Jira project keys:

1. Go to your Jira instance
2. Navigate to **Projects** → **View all projects**
3. The project key is shown next to each project name (e.g., "DEV", "INFRA")
4. Add relevant project keys to your config

## Step 3: Authenticate

### 3.1 Test Configuration

First, verify your configuration:

```bash
my-day config show
```

Check that all Jira settings are correct.

### 3.2 Authenticate with Jira

Run the authentication command:

```bash
my-day auth
```

This will:
1. Start a local server on port 8080
2. Open your browser to Jira's OAuth authorization page
3. Ask you to approve permissions for my-day
4. Save the authentication token securely

**What you'll see:**
1. **Browser opens** to Atlassian OAuth page
2. **Review permissions** that my-day is requesting
3. **Click "Accept"** to approve
4. **Success page** confirms authentication
5. **Return to terminal** - authentication complete

### 3.3 Verify Authentication

Test the connection:

```bash
my-day auth --test
```

Should show:
```
✓ Authentication is valid
✓ Connection to Jira verified
```

## Step 4: First Sync and Report

### 4.1 Sync Your Data

Pull your recent tickets:

```bash
my-day sync
```

This fetches:
- Issues assigned to you
- Issues you created
- Recent worklogs
- Stores data locally for fast reporting

### 4.2 Generate Your First Report

```bash
my-day report
```

You should see a colorful daily standup report with your Jira activity!

## Configuration Options

### Environment Variables

For CI/CD or shared environments, use environment variables instead of config file:

```bash
export MY_DAY_JIRA_BASE_URL="https://company.atlassian.net"
export MY_DAY_JIRA_CLIENT_ID="ari:cloud:ecosystem::app/..."
export MY_DAY_JIRA_CLIENT_SECRET="your-secret"
export MY_DAY_JIRA_PROJECTS="DEVOPS,INTEROP,FOUNDATION"
```

### CLI Flags

Override config for specific commands:

```bash
my-day sync --jira-url https://test.atlassian.net --projects DEV,QA
my-day report --projects DEVOPS --max-results 50
```

### Multiple Jira Instances

Use different config files for different environments:

```bash
my-day --config ~/work-config.yaml sync
my-day --config ~/personal-config.yaml report
```

## Troubleshooting

### Common Issues

#### "OAuth client not found" or "failed to retrieve client"
- **Cause**: Incorrect Client ID, app not properly saved, or invalid OAuth app
- **Solution**: 
  1. Double-check Client ID in Atlassian Developer Console matches your config exactly
  2. Verify OAuth app is saved and shows as "Active" status
  3. Ensure you completed all configuration steps (callback URL, scopes)
  4. Try generating a new Client Secret if issue persists
- **Check**: `my-day config show` displays correct client_id
- **Debug**: Look for "invalid_request" or "client not found" in error messages

#### "Invalid redirect URI"
- **Cause**: Mismatch between config and OAuth app settings
- **Solution**: 
  1. In Atlassian Console: verify callback URL is exactly `http://localhost:8080/callback`
  2. In my-day config: verify redirect_uri matches exactly (no trailing slashes)
  3. Case sensitive - must match exactly
- **Default**: `http://localhost:8080/callback`

#### "Authentication successful but connection test failed: 401"
- **Cause**: Missing Jira site configuration in OAuth app (most common)
- **Solution**:
  1. Go back to Atlassian Developer Console
  2. Open your OAuth app → Authorization tab
  3. Find "Jira platform REST API" section
  4. Click "Configure" and add your Jira site: `https://datical.atlassian.net`
  5. Save changes and try authentication again
- **This is the most commonly missed step!**

#### "Invalid client" or "Client authentication failed"
- **Cause**: Incorrect Client Secret or expired credentials
- **Solution**:
  1. Generate a new Client Secret in Atlassian Console (click "New secret")
  2. Update your my-day config with the new secret immediately
  3. Ensure you copied the entire secret string without extra spaces
- **Note**: Client secrets can only be viewed once when generated

#### "Access denied" during OAuth
- **Cause**: User doesn't have permission to authorize, or app needs approval
- **Solution**: 
  1. Contact Jira admin to approve OAuth app
  2. Verify user has access to the Jira projects in config
  3. Check if app requires admin approval in your organization

#### "No issues found" after sync
- **Cause**: Project keys don't match or user has no access
- **Solution**: 
  1. Verify project keys: `my-day config show`
  2. Check Jira permissions for those projects
  3. Try broader sync: `my-day sync --projects ""`

#### Browser doesn't open automatically
- **Solution**: Copy the displayed URL and open manually
- **Alternative**: `my-day auth --no-browser`

### Debug Commands

**Test authentication:**
```bash
my-day auth --test --verbose
```

**Check configuration:**
```bash
my-day config show --sources
```

**Verbose sync:**
```bash
my-day sync --verbose --force
```

**Clear and re-authenticate:**
```bash
my-day auth --clear
my-day auth
```

### Network Issues

#### Corporate Firewall
If behind a corporate firewall:

1. **Proxy Configuration**: Set standard proxy environment variables
   ```bash
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   ```

2. **Port Issues**: If port 8080 is blocked, change redirect URI:
   ```yaml
   oauth:
     redirect_uri: "http://localhost:3000/callback"
   ```
   (Update both config and OAuth app settings)

3. **SSL Issues**: Contact IT about Atlassian certificate trust

## Security Best Practices

### Protecting Credentials

**Never commit secrets:**
```bash
# Add to .gitignore
~/.my-day/config.yaml
~/.my-day/auth.json
*.yaml
```

**Use environment variables in CI/CD:**
```bash
# GitHub Actions example
MY_DAY_JIRA_CLIENT_SECRET: ${{ secrets.JIRA_CLIENT_SECRET }}
```

**File permissions:**
```bash
chmod 600 ~/.my-day/config.yaml
chmod 600 ~/.my-day/auth.json
```

### Token Management

- **Tokens auto-refresh**: my-day handles token renewal automatically
- **Token storage**: Stored securely in `~/.my-day/auth.json`
- **Token expiry**: OAuth tokens typically last 1 hour, refresh tokens last 90 days
- **Revoke access**: Delete app from Atlassian Developer Console to revoke all tokens

### Audit and Monitoring

**Check OAuth app usage:**
1. Go to Atlassian Developer Console
2. Select your app
3. Check **"Authorization"** tab for active tokens

**Monitor API usage:**
- Jira Cloud has rate limits (typically 300 requests/minute)
- my-day respects rate limits and caches data locally
- Check Jira admin logs for API usage

## Advanced Configuration

### Custom Scopes

For advanced features, add additional scopes to your OAuth app:

```
read:jira-project     # Extended project information
write:jira-work      # Create worklogs from CLI
read:issue-details   # Additional issue metadata
```

### Multiple Teams Setup

For organizations with multiple teams:

1. **Create team-specific OAuth apps** with appropriate project access
2. **Use different config files** per team
3. **Set up shared environments** with environment variables

```bash
# DevOps team config
my-day --config ~/.my-day/devops-config.yaml sync

# Security team config  
my-day --config ~/.my-day/security-config.yaml sync
```

### API Rate Limiting

Configure sync behavior for high-volume environments:

```yaml
jira:
  # Limit requests per sync
  max_results_per_sync: 50
  
  # Sync frequency
  min_sync_interval: "10m"
  
  # Request timeout
  request_timeout: "30s"
```

## Next Steps

Once Jira integration is working:

1. **Automate daily sync**: Set up cron job or scheduled task
2. **Team sharing**: Share config templates with your team
3. **Report automation**: Integrate with Slack/Teams for automated standup reports
4. **Custom project filtering**: Fine-tune project lists for your workflow

For more configuration options, see [Configuration Reference](configuration.md).

For usage examples and workflows, see [Usage Guide](usage.md).