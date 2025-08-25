# Servo Commands Reference

This document provides a comprehensive reference for all Servo CLI commands, their options, and usage patterns.

## Table of Contents

- [Global Options](#global-options)
- [Project Management](#project-management)
- [Session Management](#session-management)
- [Configuration Management](#configuration-management)
- [Secrets Management](#secrets-management)
- [Client Management](#client-management)
- [Validation](#validation)
- [Environment Variables](#environment-variables)
- [Exit Codes](#exit-codes)

## Global Options

These options are available for all Servo commands:

### `--no-interactive`, `-n`
Disable interactive prompts and use environment variables for required input.

**Environment Variable:** `SERVO_NON_INTERACTIVE`

**Usage:**
```bash
servo --no-interactive <command>
servo -n <command>
```

**Use Cases:**
- CI/CD pipelines
- Automated scripts
- Batch processing

### `--help`, `-h`
Show help information for the command.

```bash
servo --help              # Global help
servo <command> --help    # Command-specific help
```

### `--version`, `-v`
Display the Servo version.

```bash
servo --version
servo -v
```

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
- `crewai` - CrewAI Framework

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

**Syntax:**
```bash
servo install <SOURCE> [OPTIONS]
```

**Arguments:**
- `SOURCE` - Installation source (required)

**Source Types:**
- **Git Repository:** `https://github.com/user/repo.git`
- **Git with subdirectory:** `https://github.com/user/repo.git#subdirectory`
- **Local Directory:** `./path/to/server/`
- **Servo File:** `./config.servo`
- **Remote Servo File:** `https://example.com/server.servo`

**Options:**
- `--session, -s <name>` - Target session (default: active session)
- `--clients, -c <list>` - Comma-separated list of target clients
- `--update, -u` - Update server if it already exists

**Git Authentication Options:**
- `--ssh-key <path>` - SSH private key file path
- `--ssh-password <password>` - SSH key passphrase  
- `--http-token <token>` - HTTP token (GitHub personal access token)
- `--http-username <username>` - HTTP username
- `--http-password <password>` - HTTP password

**Environment Variables for Git Auth:**
- `GIT_SSH_KEY` - SSH private key path
- `GIT_SSH_PASSWORD` - SSH key passphrase
- `GIT_TOKEN`, `GITHUB_TOKEN` - HTTP token
- `GIT_USERNAME` - HTTP username
- `GIT_PASSWORD` - HTTP password

**Examples:**
```bash
# Install from public Git repository
servo install https://github.com/getzep/graphiti.git

# Install with specific clients
servo install https://github.com/user/repo.git --clients vscode,claude-code

# Install to specific session
servo install ./local-server --session development

# Install from subdirectory in Git repo
servo install https://github.com/user/mono-repo.git#mcp-server

# Install with update flag
servo install server.servo --update

# Install with Git authentication
servo install git@github.com:private/repo.git \
  --ssh-key ~/.ssh/id_rsa \
  --ssh-password mypassphrase

# Install with HTTP token
servo install https://github.com/private/repo.git \
  --http-token ghp_1234567890abcdef
```

**Process:**
1. Parse and validate source
2. Determine installation method
3. Clone/copy source to session directory
4. Parse .servo file
5. Validate requirements
6. Install dependencies (if needed)
7. Update client configurations
8. Generate development environment configs

**Exit Codes:**
- `0` - Success
- `1` - Any error (source not found, authentication failed, etc.)
- `3` - Validation failed
- `4` - Installation failed
- `5` - Not in a project directory

---

### `servo status`

Display comprehensive project status information.

**Syntax:**
```bash
servo status
```

**No options or arguments.**

**Output Includes:**
- **Project Information:**
  - Project name and description
  - Creation date
  - Directory location
  
- **Session Information:**
  - Active session
  - Default session
  - Available sessions
  
- **Installed Servers:**
  - Server names and versions
  - Installation sources
  - Target clients
  - Status (running/stopped/error)
  
- **Configuration Status:**
  - Required secrets (configured/missing)
  - Client configurations (valid/invalid)
  - Generated configs (devcontainer, docker-compose)
  
- **Client Detection:**
  - Detected MCP clients
  - Configuration status

**Example Output:**
```
📁 Project: awesome-app
   Description: My awesome application
   Directory: /home/user/projects/awesome-app/.servo
   Created: 2025-01-15T10:30:00Z

🔧 Sessions:
   • default (active) - Default session
   • development - Development environment
   • production - Production environment

📦 Installed Servers (default session):
   ✅ graphiti (v1.2.0) - from https://github.com/getzep/graphiti.git
      Clients: vscode, claude-code
   ✅ typescript-analyzer (v0.1.0) - from ./local-server
      Clients: vscode

🔐 Required Secrets:
   ✅ openai_api_key - configured
   ❌ database_url - missing
   
💻 MCP Clients:
   ✅ VS Code - configured (.vscode/settings.json)
   ✅ Claude Code - configured (.mcp.json)
   ❌ Cursor - not detected

🐳 Development Environment:
   ✅ .devcontainer/devcontainer.json - generated
   ✅ .devcontainer/docker-compose.yml - generated
```

**Exit Codes:**
- `0` - Success (project in good state)
- `1` - Any error (not in project directory, missing secrets, invalid configs, etc.)

---

### `servo work`

Generate development environment and client configurations.

**Syntax:**
```bash
servo work [OPTIONS]
```

**Options:**
- `--client, -c <name>` - Target specific client for optimized configuration

**Supported Clients:**
- `vscode` - Visual Studio Code
- `claude-code` - Claude Code
- `cursor` - Cursor
- `crewai` - CrewAI

**Examples:**
```bash
# Generate configurations for all detected clients
servo work

# Optimize for VS Code
servo work --client vscode

# Optimize for Claude Code
servo work --client claude-code
```

**Generated Files:**
- **`.devcontainer/devcontainer.json`** - Development container configuration
  - Runtime features (Python, Node.js, etc.)
  - Port forwarding
  - VS Code extensions
  - Post-creation commands
  
- **`.devcontainer/docker-compose.yml`** - Service dependencies
  - Database services
  - Cache services
  - Custom services from .servo files
  - Volume mounts
  - Network configuration
  
- **`.vscode/settings.json`** - VS Code MCP configuration
  - MCP server definitions
  - Server arguments and environment
  
- **`.mcp.json`** - Claude Code MCP configuration
  - Global MCP server settings
  
- **`.cursor/settings.json`** - Cursor MCP configuration (if applicable)

**Process:**
1. Load project configuration
2. Determine active session
3. Collect installed servers
4. Analyze runtime requirements
5. Generate devcontainer features
6. Create docker-compose services
7. Configure MCP clients
8. Validate generated configurations

**Exit Codes:**
- `0` - Success
- `1` - Any error (not in project directory, no servers installed, etc.)
- `3` - Missing required secrets
- `4` - Configuration generation failed

## Session Management

Sessions provide isolated environments within projects for different development workflows.

### `servo session create`

Create a new project session.

**Syntax:**
```bash
servo session create <SESSION_NAME> [OPTIONS]
```

**Arguments:**
- `SESSION_NAME` - Name of the new session (required)

**Options:**
- `--description, -d <text>` - Session description

**Examples:**
```bash
# Basic session creation
servo session create development

# With description
servo session create staging --description "Staging environment"

# Production session
servo session create production --description "Production deployment"
```

**Created Structure:**
```
.servo/sessions/<SESSION_NAME>/
├── session.yaml       # Session metadata
├── manifests/         # Server installations
└── volumes/           # Docker volumes (gitignored)
```

**Exit Codes:**
- `0` - Success  
- `1` - Any error (session already exists, invalid session name, etc.)
- `3` - Not in a project directory

---

### `servo session list`

List all project sessions with their status.

**Syntax:**
```bash
servo session list
```

**No options or arguments.**

**Output Format:**
```
Sessions:
  • default (active) - Default session
  • development - Development environment  
  • staging - Staging environment
  • production - Production deployment
```

**Exit Codes:**
- `0` - Success  
- `1` - Any error (not in project directory, etc.)

---

### `servo session activate`

Activate a specific session for the current project.

**Syntax:**
```bash
servo session activate <SESSION_NAME>
```

**Arguments:**
- `SESSION_NAME` - Name of session to activate (required)

**Examples:**
```bash
servo session activate development
servo session activate production
```

**Effects:**
- Updates project configuration to set active session
- Subsequent commands operate on the activated session
- Client configurations will be generated from active session servers

**Exit Codes:**
- `0` - Success
- `1` - Any error (session not found, not in project directory, etc.)

---

### `servo session delete`

Delete a session and all its data permanently.

**Syntax:**
```bash
servo session delete <SESSION_NAME>
```

**Arguments:**
- `SESSION_NAME` - Name of session to delete (required)

**Examples:**
```bash
servo session delete staging
servo session delete old-development
```

**Warning:** This operation permanently deletes:
- All installed servers in the session
- Session volumes and data
- Session-specific configurations

**Protection:** Cannot delete the active session. Must activate a different session first.

**Exit Codes:**
- `0` - Success
- `1` - Any error (session not found, cannot delete active session, etc.)
- `3` - Not in a project directory

### `servo session rename`

Rename an existing session. This operation updates all references to the session, including active session configuration and project default session references.

**Syntax:**
```bash
servo session rename <OLD_NAME> <NEW_NAME>
```

**Arguments:**
- `OLD_NAME` - Current name of the session to rename (required)
- `NEW_NAME` - New name for the session (required)

**Examples:**
```bash
servo session rename development dev
servo session rename staging prod-staging
servo session rename default main
```

**Behavior:**
- Renames the session directory and all internal references
- Updates active session reference if the renamed session is currently active
- Updates project configuration if the session is referenced as the default session
- Preserves all session data, manifests, and configuration
- Operation is atomic - either fully succeeds or rolls back on failure

**Exit Codes:**
- `0` - Success
- `1` - Any error (session not found, target name already exists, invalid names, etc.)
- `3` - Not in a project directory

## Configuration Management

Project-level configuration management for settings, preferences, and metadata.

### `servo config list`

Display all project configuration values.

**Syntax:**
```bash
servo config list
```

**No options or arguments.**

**Output Format:**
```
Project Configuration:
  name: awesome-app
  description: My awesome application
  created_at: 2025-01-15T10:30:00Z
  default_session: default
  active_session: development
  clients: [vscode, claude-code]
```

**Exit Codes:**
- `0` - Success  
- `1` - Any error (not in project directory, etc.)

---

### `servo config get`

Retrieve a specific configuration value.

**Syntax:**
```bash
servo config get <KEY>
```

**Arguments:**
- `KEY` - Configuration key to retrieve (required)

**Available Keys:**
- `name` - Project name
- `description` - Project description
- `created_at` - Project creation timestamp
- `default_session` - Default session name
- `active_session` - Currently active session
- `clients` - Configured MCP clients array

**Examples:**
```bash
servo config get name
servo config get active_session
servo config get clients
```

**Exit Codes:**
- `0` - Success
- `1` - Any error (key not found, not in project directory, etc.)

---

### `servo config set`

Set a project configuration value.

**Syntax:**
```bash
servo config set <KEY> <VALUE>
```

**Arguments:**
- `KEY` - Configuration key (required)
- `VALUE` - Configuration value (required)

**Settable Keys:**
- `name` - Project name
- `description` - Project description
- `default_session` - Default session name
- `active_session` - Currently active session

**Note:** Some keys like `created_at` are read-only.

**Examples:**
```bash
servo config set name "My New Project Name"
servo config set description "Updated description"
servo config set default_session production
servo config set active_session development
```

**Exit Codes:**
- `0` - Success
- `1` - Any error (invalid key, invalid value, etc.)
- `3` - Not in a project directory

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

Manage MCP client support for the current project.

### `servo client list`

List all available MCP clients with their detection and configuration status.

**Syntax:**
```bash
servo client list
```

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

**Detection Methods:**
- **VS Code:** Check for `code` command and `.vscode` directory support
- **Claude Code:** Check for Claude Code installation and config support
- **Cursor:** Check for `cursor` command availability

**Exit Codes:**
- `0` - Success
- `1` - Any error (client detection failed, etc.)

### `servo client enable`

Enable support for one or more MCP clients in the current project.

**Syntax:**
```bash
servo client enable <CLIENT> [<CLIENT> ...]
```

**Arguments:**
- `CLIENT` - Name of MCP client to enable (required, multiple allowed)

**Supported Clients:**
- `vscode` - Visual Studio Code
- `claude-code` - Claude Code
- `cursor` - Cursor Editor

**Examples:**
```bash
# Enable single client
servo client enable vscode

# Enable multiple clients
servo client enable vscode claude-code cursor

# Enable all clients
servo client enable vscode claude-code cursor
```

**Output Examples:**
```bash
✅ Enabled client(s): [vscode claude-code]
⚠️  Already enabled: [cursor]
❌ Failed to enable: [invalid-client (client 'invalid-client' not supported)]
```

**Project Impact:**
- Updates `project.yaml` with enabled clients
- Clients will be configured when running `servo work`
- Client-specific configuration files will be generated

**Exit Codes:**
- `0` - Success (at least one client enabled or already enabled)
- `1` - All clients failed to enable
- `3` - Not in a project directory

### `servo client disable`

Disable support for one or more MCP clients in the current project.

**Syntax:**
```bash
servo client disable <CLIENT> [<CLIENT> ...]
```

**Arguments:**
- `CLIENT` - Name of MCP client to disable (required, multiple allowed)

**Supported Clients:**
- `vscode` - Visual Studio Code
- `claude-code` - Claude Code
- `cursor` - Cursor Editor

**Examples:**
```bash
# Disable single client
servo client disable cursor

# Disable multiple clients
servo client disable vscode claude-code

# Disable all clients (not recommended)
servo client disable vscode claude-code cursor
```

**Output Examples:**
```bash
✅ Disabled client(s): [cursor]
⚠️  Not enabled: [vscode]
❌ Failed to disable: [invalid-client (client 'invalid-client' not found in project)]
```

**Project Impact:**
- Updates `project.yaml` removing disabled clients
- Client will no longer be configured when running `servo work`
- Existing client configuration files remain unchanged

**Exit Codes:**
- `0` - Success (at least one client disabled or already disabled)
- `1` - All clients failed to disable
- `3` - Not in a project directory

## Validation

Validate .servo files and installation sources.

### `servo validate`

Validate a .servo file or installation source for correctness and completeness.

**Syntax:**
```bash
servo validate <SOURCE>
```

**Arguments:**
- `SOURCE` - Path to .servo file or installation source (required)

**Source Types:**
- **Local File:** `./server.servo`
- **Git Repository:** `https://github.com/user/repo.git`
- **Remote File:** `https://example.com/server.servo`
- **Directory:** `./local-server-directory/`

**Validation Checks:**
- **Schema Validation:** YAML syntax and structure
- **Required Fields:** Presence of mandatory sections
- **Version Compatibility:** Servo version compatibility
- **Runtime Requirements:** Available tools and versions
- **Service Dependencies:** Docker image availability
- **Secret References:** Proper secret declarations
- **Security:** No obvious security issues

**Examples:**
```bash
# Validate local .servo file
servo validate ./graphiti.servo

# Validate git repository
servo validate https://github.com/getzep/graphiti.git

# Validate directory with .servo file
servo validate ./my-server-directory/

# Validate remote file
servo validate https://raw.githubusercontent.com/user/repo/main/server.servo
```

**Output Examples:**

**Valid File:**
```
✅ Validation successful: ./graphiti.servo
   Schema: Valid
   Runtime Requirements: python >=3.10, uv (available)
   Services: neo4j:5.13 (available)
   Secrets: 2 declared (openai_api_key, neo4j_password)
   Security: No issues found
```

**Invalid File:**
```
❌ Validation failed: ./broken.servo
   Schema: Invalid YAML syntax at line 15
   Runtime Requirements: python >=3.11 (not available - found 3.9)
   Services: postgres:15 (image not found)
   Secrets: Missing required description for 'api_key'
```

**Exit Codes:**
- `0` - Validation successful
- `1` - Any error (validation failed, source not found or inaccessible, etc.)
- `3` - Invalid source format

## Environment Variables

### Global Configuration

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `SERVO_NON_INTERACTIVE` | Disable interactive prompts (for CI/scripts) | `false` | `true` |
| `SERVO_DIR` | Custom servo directory location | `.servo` | `/home/user/.servo` |

### Git Authentication

| Variable | Description | Example |
|----------|-------------|---------|
| `GIT_SSH_KEY` | Path to SSH private key | `~/.ssh/id_rsa` |
| `GIT_SSH_PASSWORD` | SSH key passphrase | `my-key-passphrase` |
| `GIT_TOKEN` | Git HTTP token | `ghp_1234567890abcdef` |
| `GITHUB_TOKEN` | GitHub personal access token | `ghp_1234567890abcdef` |
| `GIT_USERNAME` | Git HTTP username | `myusername` |
| `GIT_PASSWORD` | Git HTTP password | `mypassword` |

### Usage in Scripts

**Non-interactive mode:**
```bash
#!/bin/bash
export SERVO_NON_INTERACTIVE=1

servo status
servo secrets set api_key "$API_KEY_VALUE"
servo work
```

**Git authentication:**
```bash
#!/bin/bash
export GITHUB_TOKEN="$MY_GITHUB_TOKEN"
servo install https://github.com/private/repo.git
```

**Custom servo directory:**
```bash
export SERVO_DIR="$HOME/.config/servo"
servo init my-project
```

## Exit Codes

Servo uses standard Unix exit codes:

| Code | Description |
|------|-------------|
| `0` | Success - Command completed successfully |
| `1` | Any error - Detailed error message printed to stderr |

All error conditions result in exit code 1 with a descriptive error message, including:
- Invalid command line arguments or usage
- File system issues, permissions, resources not found  
- Network connectivity or authentication failures
- Invalid configuration or missing requirements
- Validation failures for .servo files
- Runtime issues, missing tools, or service failures

### Using Exit Codes in Scripts

```bash
#!/bin/bash

servo status
if [ $? -ne 0 ]; then
    echo "Error: Not in a servo project or project has issues"
    exit 1
fi

servo secrets set api_key "$API_KEY"
if [ $? -ne 0 ]; then
    echo "Error: Failed to set secret"
    exit 1
fi

servo work
if [ $? -eq 0 ]; then
    echo "Development environment ready!"
else
    echo "Error: Failed to generate development environment"
    exit 1
fi
```

## Best Practices

### Command Line Usage

1. **Always check status first:**
   ```bash
   servo status  # Check project state before other operations
   ```

2. **Use non-interactive mode in scripts:**
   ```bash
   export SERVO_NON_INTERACTIVE=1
   ```

3. **Set secrets securely:**
   ```bash
   # Better than putting secrets in command line
   read -s -p "API Key: " API_KEY
   servo secrets set openai_api_key "$API_KEY"
   ```

4. **Validate before installing:**
   ```bash
   servo validate ./server.servo && servo install ./server.servo
   ```

### Project Organization

1. **Use descriptive session names:**
   ```bash
   servo session create local-development
   servo session create staging-test
   servo session create production-mirror
   ```

2. **Document your setup:**
   ```bash
   servo config set description "AI-powered web application with vector search"
   ```

3. **Keep secrets organized:**
   ```bash
   servo secrets set openai_api_key_dev "sk-dev-..."
   servo secrets set openai_api_key_prod "sk-prod-..."
   ```

### Team Collaboration

1. **Commit configurations, not secrets:**
   ```bash
   git add .servo/project.yaml .devcontainer/ .vscode/
   git commit -m "Add servo project configuration"
   # secrets.yaml is automatically gitignored
   ```

2. **Document required secrets:**
   ```bash
   # In your project README
   # Required secrets: openai_api_key, database_url, stripe_key
   ```

3. **Use consistent session names:**
   ```bash
   # Team standard
   servo session create development  # Local development
   servo session create staging      # Staging environment
   servo session create production   # Production mirror
   ```