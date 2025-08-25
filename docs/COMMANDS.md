# Servo Commands Reference

This document provides a comprehensive reference for all Servo CLI commands, their options, and usage patterns.

## Table of Contents

- [Global Options](#global-options)
- [Project Management](#project-management)
- [Session Management](#session-management)
- [Configuration Management](#configuration-management)
- [Environment Variables Management](#environment-variables-management)
- [Secrets Management](#secrets-management)
- [Client Management](#client-management)
- [Validation](#validation)
- [Global Environment Variables](#global-environment-variables)
- [Exit Codes](#exit-codes)

## Global Options

### `--no-interactive`, `-n`
Disable interactive prompts (useful for CI/CD). Environment variable: `SERVO_NON_INTERACTIVE`

### `--help`, `-h`
Show help information for the command.

### `--version`, `-v`
Display the Servo version.

## Project Management

### `servo init`

Initialize a new Servo project in the current directory.

**Syntax:**
```bash
servo init [PROJECT_NAME] [OPTIONS]
```

**Arguments:**
- `PROJECT_NAME` (optional) - Name of the project. Defaults to current directory name.

**Options:**
- `--session, -s <name>` - Default session name (default: "default")
- `--clients, -c <list>` - Comma-separated list of MCP clients to configure
- `--description <text>` - Project description

**Supported Clients:**
- `vscode` - Visual Studio Code
- `claude-code` - Claude Code
- `cursor` - Cursor Editor

**Examples:**
```bash
# Basic initialization
servo init

# With project name and description
servo init my-app --description "My awesome application"

# With custom session and clients
servo init web-project --session development --clients vscode,claude-code

# All options
servo init data-analysis \
  --description "Data analysis project" \
  --session research \
  --clients vscode,claude-code,cursor
```

**Created Structure:**
```
.servo/
├── project.yaml           # Project configuration
├── .gitignore            # Git ignore rules
└── sessions/
    └── default/          # Default session (or custom name)
        ├── session.yaml  # Session metadata
        ├── manifests/    # Server definitions
        └── volumes/      # Docker volumes (gitignored)
```

**Exit Codes:**
- `0` - Success  
- `1` - Any error (project already exists, invalid options, etc.)
- `3` - File system error

---

### `servo install`

Install an MCP server from various sources.

```bash
servo install <SOURCE> [OPTIONS]
```

**Sources:** Git repos, local directories, .servo files, or remote URLs

**Options:**
- `--session, -s <name>` - Target session
- `--clients, -c <list>` - Target clients
- `--update, -u` - Update if exists

**Git Authentication:** Use `--ssh-key`, `--http-token`, or environment variables like `GIT_TOKEN`

**Examples:**
```bash
servo install https://github.com/getzep/graphiti.git
servo install ./local-server --session development
servo install server.servo --update
```

---

### `servo status`

Show project status, servers, and configuration state.

```bash
servo status
```

Shows project info, active session, installed servers, missing secrets, and client configurations.

---

### `servo work`

Generate development environment and client configurations.

```bash
servo work [--client <name>]
```

**Generates:**
- `.devcontainer/devcontainer.json` - Dev container with runtime features
- `.devcontainer/docker-compose.yml` - Service dependencies  
- `.vscode/settings.json` - VS Code MCP configuration
- `.mcp.json` - Claude Code MCP configuration

## Session Management

### `servo session create <name> [--description <text>]`
Create a new project session.

### `servo session list`
List all project sessions.

### `servo session activate <name>`
Activate a specific session.

### `servo session delete <name>`
Delete a session and all its data permanently.

### `servo session rename <old-name> <new-name>`
Rename an existing session, updating all references.

## Configuration Management

### `servo configure [--client <name>]`
Generate MCP client configurations (VS Code, Claude Code, Cursor).

## Environment Variables Management

Manage non-sensitive environment variables stored in `.servo/env.yaml`.

### `servo env list`
Display all project environment variables.

### `servo env get <key>`
Get a specific environment variable value.

### `servo env set <key> <value>`
Set or update an environment variable.

### `servo env delete <key>`
Remove an environment variable.

### `servo env export <file>` / `servo env import <file>`
Export/import environment variables for backup.

## Configuration Management

Generate MCP client configurations and manage project-level environment variables.

### `servo configure`

Generate MCP client configurations independently of the install/work workflow.

**Syntax:**
```bash
servo configure [OPTIONS]
```

**Options:**
- `--client, -c <name>` - Target specific client for optimized configuration

**Supported Clients:**
- `vscode` - Visual Studio Code
- `claude-code` - Claude Code  
- `cursor` - Cursor

**Examples:**
```bash
# Generate configurations for all detected clients
servo configure

# Generate only VS Code configuration
servo configure --client vscode

# Generate only Claude Code configuration  
servo configure --client claude-code
```

**Generated Files:**
- **`.vscode/settings.json`** - VS Code MCP configuration
- **`.mcp.json`** - Claude Code MCP configuration  
- **`.cursor/settings.json`** - Cursor MCP configuration (if applicable)

**Process:**
1. Load project configuration
2. Determine active session
3. Collect installed servers
4. Generate client-specific MCP configurations
5. Validate generated configurations

**Exit Codes:**
- `0` - Success
- `1` - Any error (not in project directory, no servers installed, etc.)
- `3` - Missing required secrets

---

## Environment Variables Management

Manage non-sensitive environment variables that are automatically injected into MCP services during docker-compose generation.

**Storage:** Environment variables are stored in `.servo/env.yaml` and are safe to commit to version control.

**Integration:** Variables are automatically injected into all MCP server services in docker-compose.yml generation, with service-specific overrides supported.

### `servo env list`

Display all project environment variables.

**Syntax:**
```bash
servo env list
```

**No options or arguments.**

**Output Format:**
```
Project Environment Variables:
  API_TIMEOUT: "30s"
  DEBUG_MODE: "true"
  BASE_URL: "https://api.example.com"
  MAX_RETRIES: "3"
```

**Exit Codes:**
- `0` - Success  
- `1` - Any error (not in project directory, etc.)

---

### `servo env get`

Retrieve a specific environment variable value.

**Syntax:**
```bash
servo env get <KEY>
```

**Arguments:**
- `KEY` - Environment variable name (required)

**Examples:**
```bash
servo env get API_TIMEOUT
servo env get BASE_URL

# Use in scripts
API_URL=$(servo env get BASE_URL)
```

**Exit Codes:**
- `0` - Success
- `1` - Any error (key not found, not in project directory, etc.)

---

### `servo env set`

Set or update an environment variable.

**Syntax:**
```bash
servo env set <KEY> <VALUE>
```

**Arguments:**
- `KEY` - Environment variable name (required)
- `VALUE` - Environment variable value (required)

**Examples:**
```bash
servo env set API_TIMEOUT "60s"
servo env set BASE_URL "https://staging-api.example.com"
servo env set DEBUG_MODE "false"
servo env set MAX_CONNECTIONS "10"
```

**Usage Notes:**
- Values are stored in plaintext in `.servo/env.yaml`
- Use for non-sensitive configuration only (URLs, timeouts, feature flags)
- For sensitive data, use `servo secrets` instead
- Variables are automatically injected into MCP services via docker-compose

**Exit Codes:**
- `0` - Success
- `1` - Any error (invalid key, not in project directory, etc.)

---

### `servo env delete`

Remove an environment variable.

**Syntax:**
```bash
servo env delete <KEY>
```

**Arguments:**
- `KEY` - Environment variable name to remove (required)

**Examples:**
```bash
servo env delete DEBUG_MODE
servo env delete API_TIMEOUT
```

**Exit Codes:**
- `0` - Success
- `1` - Any error (key not found, not in project directory, etc.)

---

### `servo env export`

Export environment variables to a file for backup or sharing.

**Syntax:**
```bash
servo env export <OUTPUT_FILE>
```

**Arguments:**
- `OUTPUT_FILE` - Path to export file (required)

**Examples:**
```bash
servo env export ./project-env.yaml
servo env export /backup/project-environment.yaml
```

**Output Format:**
The exported file contains YAML-formatted environment variables that can be imported into other projects.

**Exit Codes:**
- `0` - Success
- `1` - Any error (write failed, not in project directory, etc.)

---

### `servo env import`

Import environment variables from a file.

**Syntax:**
```bash
servo env import <INPUT_FILE>
```

**Arguments:**
- `INPUT_FILE` - Path to import file (required)

**Examples:**
```bash
servo env import ./project-env.yaml
servo env import /backup/project-environment.yaml
```

**Behavior:**
- Merges imported variables with existing ones
- Overwrites existing variables with matching keys
- Preserves existing variables not present in import file

**Exit Codes:**
- `0` - Success
- `1` - Any error (file not found, invalid format, not in project directory, etc.)

## Secrets Management

Simple secrets management using base64 encoding for basic obscurity.

### Security Model
- **Encoding:** Base64 encoding for basic obscurity (not cryptographic security)
- **Storage:** Local `.servo/secrets.yaml` file (never committed)
- **Access:** Direct file access with 0600 permissions
- **Team Workflow:** Secrets declared in project config, values stored locally

### `servo secrets list`

List all configured secret names (values are not displayed).

**Syntax:**
```bash
servo secrets list
```

**No options or arguments.**

**Example Output:**
```
Project secrets:
  • openai_api_key
  • database_url
  • redis_password
  • webhook_secret
```

**Exit Codes:**
- `0` - Success  
- `1` - Any error (not in project directory, etc.)
- `2` - Secrets file corrupted

---

### `servo secrets set`

Set or update a secret value.

**Syntax:**
```bash
servo secrets set <KEY> <VALUE>
```

**Arguments:**
- `KEY` - Secret name (required)
- `VALUE` - Secret value (required)

**Examples:**
```bash
servo secrets set openai_api_key sk-1234567890abcdef
servo secrets set database_url "postgres://user:pass@localhost/db"
servo secrets set api_token "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9..."
```

**Security Notes:**
- Values are encrypted before storage
- Shell history may contain the value - consider using files or prompts for highly sensitive data
- Master password is required and prompted if not in environment

**Exit Codes:**
- `0` - Success
- `1` - Any error (encryption failed, wrong password, not in project directory, etc.)

---

### `servo secrets get`

Retrieve and display a secret value.

**Syntax:**
```bash
servo secrets get <KEY>
```

**Arguments:**
- `KEY` - Secret name (required)

**Examples:**
```bash
servo secrets get openai_api_key
servo secrets get database_url

# Use in scripts
DB_URL=$(servo secrets get database_url)
```

**Security Notes:**
- Value is displayed in plaintext
- Master password required for decryption
- Output goes to stdout and may be logged

**Exit Codes:**
- `0` - Success
- `1` - Any error (secret not found, decryption failed, etc.)
- `3` - Not in a project directory

---

### `servo secrets delete`

Remove a secret permanently.

**Syntax:**
```bash
servo secrets delete <KEY>
```

**Arguments:**
- `KEY` - Secret name to delete (required)

**Examples:**
```bash
servo secrets delete old_api_key
servo secrets delete temporary_token
```

**Warning:** This permanently removes the secret. Deleted secrets cannot be recovered.

**Exit Codes:**
- `0` - Success (even if secret didn't exist)
- `1` - Any error (encryption failed during save, not in project directory, etc.)

---

### `servo secrets export`

Export encrypted secrets to a backup file.

**Syntax:**
```bash
servo secrets export <OUTPUT_FILE>
```

**Arguments:**
- `OUTPUT_FILE` - Path to export file (required)

**Examples:**
```bash
servo secrets export secrets-backup.enc
servo secrets export /secure/backup/$(date +%Y%m%d)-secrets.yaml
```

**File Format:**
```json
{
  "version": "1.0",
  "encryption": "aes-256-gcm",
  "salt": "base64-encoded-salt",
  "encrypted_data": "base64-encoded-encrypted-secrets"
}
```

**Notes:**
- Export maintains encryption - safe to store
- Exported file can be imported on any system with the same master password
- Export file should still be treated as sensitive

**Exit Codes:**
- `0` - Success
- `1` - Any error (export failed, not in project directory, etc.)

---

### `servo secrets import`

Import encrypted secrets from a backup file.

**Syntax:**
```bash
servo secrets import <INPUT_FILE>
```

**Arguments:**
- `INPUT_FILE` - Path to import file (required)

**Examples:**
```bash
servo secrets import secrets-backup.enc
servo secrets import /backup/production-secrets.yaml
```

**Process:**
1. Validate import file format
2. Test decryption with current master password
3. Backup existing secrets file
4. Import new secrets
5. Verify import integrity

**Warnings:**
- Existing secrets will be backed up automatically
- Master password must match the one used for export
- Import completely replaces current secrets

**Exit Codes:**
- `0` - Success
- `1` - Any error (import file invalid, decryption failed, wrong password, etc.)
- `3` - Not in a project directory

## Client Management

### `servo client list`
List available MCP clients and their configuration status.

**No options or arguments.**

**Output Format:**
```
Available Clients:
NAME            INSTALLED  DESCRIPTION
----            ---------  -----------
vscode          Yes        Visual Studio Code MCP integration
claude-code     Yes        Claude Code desktop application
cursor          No         Cursor AI code editor
```

