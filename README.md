# Servo - MCP Server Project Manager

> "Project-focused containerized development environments for Model Context Protocol (MCP) servers"

Servo provides **project-based MCP server management** with isolated, containerized development environments. Each project maintains its own MCP servers, dependencies, and configuration while supporting team collaboration through git-friendly configs.

## What is Servo?

Servo is a command-line tool that simplifies managing Model Context Protocol (MCP) servers in development projects by:

### 🏗️ **Project Isolation** - Complete Development Environment Management
- **Directory-based projects**: Each project runs in its own `.servo/` directory with independent servers and dependencies
- **Containerized by default**: Projects run in isolated Docker environments with automatic devcontainer generation
- **Session support**: Multiple isolated environments per project (web_dev, dba, cloud_infra)
- **Team collaboration**: Project configs are git-friendly with declarative secrets (values stored locally)
- **Client integration**: Automatically configures for VS Code, Claude Code, and Cursor

### 🔧 **Automated Configuration** - Zero-Config Development Workflows
- **Devcontainer generation**: Automatically generates `.devcontainer/devcontainer.json` with runtime features
- **Docker Compose orchestration**: Manages service dependencies from .servo file definitions
- **MCP client configuration**: Updates client-specific settings for seamless MCP server access
- **Secret management**: base64-encoded secrets with team-friendly workflows

## Key Concepts

### Project-First Design

All Servo commands operate on **projects as the primary workflow**:

- **Automatic detection**: Servo detects if you're in a project directory (contains `.servo/`)
- **Project isolation**: Each project has completely independent servers, dependencies, and configuration
- **Session support**: Multiple environments per project for different development workflows
- **Team collaboration**: Configurations are git-friendly, secrets are declared but stored locally

### Servo Files (.servo)

MCP servers are packaged using `.servo` files that define:
- **Server configuration**: Transport, command, arguments
- **Dependencies**: Required services (databases, caches, etc.)
- **Runtime requirements**: Programming language versions and tools
- **Configuration schema**: Required secrets and settings

## Quick Start

### 1. Initialize a Project
```bash
# Initialize in current directory
cd /path/to/my-project
servo init my-app --description "My awesome application" --clients vscode,claude-code

# This creates:
# .servo/project.yaml       - Project configuration
# .servo/sessions/default/   - Default session directory
```

### 2. Install MCP Servers
```bash
# Install from various sources
servo install https://github.com/getzep/graphiti.git
servo install ./my-custom-server.servo
servo install ./local-mcp-server/
```

### 3. Configure Environment Variables (Optional)
```bash
# Set non-sensitive configuration values
servo env set API_TIMEOUT "30s"
servo env set DEBUG_MODE "false"
servo env set BASE_URL "https://api.example.com"
```

### 4. Configure Required Secrets
```bash
# Set secrets declared in .servo files
servo secrets set openai_api_key sk-your-key-here
servo secrets set database_url postgres://user:pass@localhost/db
```

### 5. Start Development Environment
```bash
servo work
# Generates:
# - .devcontainer/devcontainer.json (with runtime features)
# - .devcontainer/docker-compose.yml (with service dependencies and env vars)  
# - .vscode/settings.json (MCP server configuration)
# - .mcp.json (Claude Code configuration)

# Or generate just client configurations:
servo configure
```

## Installation

### Prerequisites
- **Docker** (for containerized environments and service dependencies)
- **Go 1.20+** (for building from source)

### From Source
```bash
git clone https://github.com/jarosser06/servo.git
cd servo
make build
sudo make install
```

### Verify Installation
```bash
servo --version
servo --help
```

## Commands Reference

### Global Flags
```bash
--no-interactive, -n    Disable interactive prompts (env: SERVO_NON_INTERACTIVE)
--help, -h              Show help
--version, -v           Show version
```

### Project Management

#### `servo init`
Initialize a new Servo project in the current directory.

```bash
servo init [PROJECT_NAME] [OPTIONS]
```

**Arguments:**
- `PROJECT_NAME` - Name of the project (optional, defaults to current directory name)

**Flags:**
- `--session, -s` - Default session name (default: "default")  
- `--clients, -c` - Comma-separated list of MCP clients (vscode,claude-code,cursor)
- `--description` - Project description

**Examples:**
```bash
servo init my-app --description "My application"
servo init web-project --session development --clients vscode,claude-code
```

#### `servo install`
Install MCP server from various sources.

```bash  
servo install <SOURCE> [OPTIONS]
```

**Arguments:**
- `SOURCE` - Installation source:
  - Git repository: `https://github.com/user/repo.git`
  - Local directory: `./path/to/server/`
  - .servo file: `./config.servo`

**Flags:**
- `--session, -s` - Install to specific session
- `--clients, -c` - Target MCP clients for this server
- `--update, -u` - Update server if it already exists

**Git Authentication Flags:**
- `--ssh-key` - SSH private key path (env: GIT_SSH_KEY)
- `--ssh-password` - SSH key passphrase (env: GIT_SSH_PASSWORD) 
- `--http-token` - HTTP token for git (env: GIT_TOKEN, GITHUB_TOKEN)
- `--http-username` - HTTP username (env: GIT_USERNAME)
- `--http-password` - HTTP password (env: GIT_PASSWORD)

**Examples:**
```bash
servo install https://github.com/getzep/graphiti.git
servo install ./my-server --clients vscode,claude-code
servo install server.servo --session production --update
```

#### `servo status`
Show status of current project including servers and configurations.

```bash
servo status
```

**Output:**
- Project information
- Active session
- Installed servers
- Missing secrets
- Client configurations

#### `servo configure`
Generate MCP client configurations independently of install/work workflows.

```bash
servo configure [OPTIONS]
```

**Flags:**
- `--client, -c` - Target specific client for optimized configuration

**Examples:**
```bash
servo configure                    # Generate configs for all clients
servo configure --client vscode   # Generate only VS Code configuration
```

**Generated Files:**
- `.vscode/settings.json` - VS Code MCP configuration
- `.mcp.json` - Claude Code MCP configuration
- `.cursor/settings.json` - Cursor MCP configuration (if applicable)

#### `servo work`
Generate development environment and client configurations.

```bash
servo work [OPTIONS]
```

**Flags:**
- `--client, -c` - Target specific client for development

**Examples:**
```bash
servo work                    # Generate configs for all clients
servo work --client vscode    # Focus on VS Code configuration
```

**Generated Files:**
- `.devcontainer/devcontainer.json` - Development container configuration
- `.devcontainer/docker-compose.yml` - Service dependencies (with env var injection)
- `.vscode/settings.json` - VS Code MCP configuration
- `.mcp.json` - Claude Code MCP configuration

### Session Management

#### `servo session create`
Create a new project session.

```bash
servo session create <SESSION_NAME> [OPTIONS]
```

**Arguments:**
- `SESSION_NAME` - Name of the new session

**Flags:**
- `--description, -d` - Session description

**Examples:**
```bash
servo session create development --description "Development environment"
servo session create staging
```

#### `servo session list`
List all project sessions.

```bash
servo session list
```

**Output:**
- Session names
- Descriptions  
- Active status

#### `servo session activate`
Activate a specific session.

```bash
servo session activate <SESSION_NAME>
```

**Arguments:**
- `SESSION_NAME` - Name of session to activate

#### `servo session delete`
Delete a session and all its data.

```bash
servo session delete <SESSION_NAME>
```

**Arguments:**
- `SESSION_NAME` - Name of session to delete

### Configuration Management

#### `servo configure`
Generate MCP client configurations independently of install/work workflows.

```bash
servo configure [OPTIONS]
```

**Flags:**
- `--client, -c` - Target specific client for optimized configuration

**Examples:**
```bash
servo configure                    # Generate configs for all clients
servo configure --client vscode   # Generate only VS Code configuration
```

### Environment Variables Management
Manage non-sensitive project-level environment variables that are automatically injected into MCP services.

Environment variables are stored in `.servo/env.yaml` and are safe to commit to version control.

#### `servo env list`
Display all project environment variables.

```bash
servo env list
```

#### `servo env get`
Get a specific environment variable value.

```bash
servo env get <KEY>
```

**Arguments:**
- `KEY` - Environment variable name

**Examples:**
```bash
servo env get API_TIMEOUT
servo env get BASE_URL
```

#### `servo env set`
Set or update an environment variable.

```bash
servo env set <KEY> <VALUE>
```

**Arguments:**
- `KEY` - Environment variable name
- `VALUE` - Environment variable value

**Examples:**
```bash
servo env set API_TIMEOUT "60s"
servo env set BASE_URL "https://api.example.com"
servo env set DEBUG_MODE "false"
```

#### `servo env delete`
Remove an environment variable.

```bash
servo env delete <KEY>
```

#### `servo env export`
Export environment variables to a file for backup or sharing.

```bash
servo env export <OUTPUT_FILE>
```

#### `servo env import`
Import environment variables from a file.

```bash
servo env import <INPUT_FILE>
```

### Secrets Management
All secrets are base64-encoded (not encrypted) and stored locally in `.servo/secrets.yaml`.

#### `servo secrets list`
List all configured secrets (names only, not values).

```bash
servo secrets list
```

#### `servo secrets set`
Set a secret value.

```bash
servo secrets set <KEY> <VALUE>
```

**Arguments:**
- `KEY` - Secret name
- `VALUE` - Secret value

**Examples:**
```bash
servo secrets set openai_api_key sk-1234567890abcdef
servo secrets set database_url "postgres://user:pass@localhost/db"
```

#### `servo secrets get`
Retrieve a secret value.

```bash
servo secrets get <KEY>
```

**Arguments:**
- `KEY` - Secret name

#### `servo secrets delete`
Delete a secret.

```bash
servo secrets delete <KEY>
```

**Arguments:**
- `KEY` - Secret name


#### `servo secrets export`
Export base64-encoded secrets to a file for backup.

```bash
servo secrets export <OUTPUT_FILE>
```

**Arguments:**
- `OUTPUT_FILE` - Path to export file

#### `servo secrets import`
Import base64-encoded secrets from a backup file.

```bash
servo secrets import <INPUT_FILE>
```

**Arguments:**
- `INPUT_FILE` - Path to import file

### Client Management

#### `servo client list`
List all available MCP clients and their installation status.

```bash
servo client list
```

#### `servo client enable`
Enable support for one or more MCP clients in the current project.

```bash
servo client enable <CLIENT> [<CLIENT> ...]
```

**Arguments:**
- `CLIENT` - Name of MCP client to enable (vscode, claude-code, cursor)

#### `servo client disable`
Disable support for one or more MCP clients in the current project.

```bash
servo client disable <CLIENT> [<CLIENT> ...]
```

**Arguments:**
- `CLIENT` - Name of MCP client to disable (vscode, claude-code, cursor)

**Output:**
- Client names
- Installation status
- Descriptions

### Validation

#### `servo validate`
Validate a .servo file or source.

```bash
servo validate <SOURCE>
```

**Arguments:**
- `SOURCE` - Path to .servo file or installation source

**Examples:**
```bash
servo validate ./server.servo
servo validate https://github.com/user/repo.git
```

## System Environment Variables

System-level environment variables that control Servo's behavior across all projects.

### Global Settings
- `SERVO_NON_INTERACTIVE` - Disable interactive prompts (for CI/scripts)

### Git Authentication  
- `GIT_SSH_KEY` - Path to SSH private key
- `GIT_SSH_PASSWORD` - SSH key passphrase
- `GIT_TOKEN` / `GITHUB_TOKEN` - HTTP token for git authentication
- `GIT_USERNAME` - HTTP username for git
- `GIT_PASSWORD` - HTTP password for git

## Project Structure

```
my-project/                           # Your project root
├── .servo/                          # Servo project directory
│   ├── project.yaml                 # Project configuration
│   ├── env.yaml                     # Environment variables (committed)
│   ├── secrets.yaml                 # Encoded secrets (local only)
│   ├── .gitignore                   # Ignores secrets and volumes
│   └── sessions/                    # Session directories
│       ├── default/                 # Default session
│       │   ├── session.yaml         # Session metadata
│       │   ├── manifests/           # Installed .servo files
│       │   │   └── server.servo     # Server definitions
│       │   └── volumes/             # Docker volumes (ignored)
│       └── development/             # Development session
│           ├── session.yaml
│           ├── manifests/
│           └── volumes/
├── .devcontainer/                   # Generated development container
│   ├── devcontainer.json           # Dev container configuration
│   └── docker-compose.yml          # Service dependencies (with env vars)
├── .vscode/                         # Generated VS Code settings
│   └── settings.json               # MCP server configuration
└── .mcp.json                       # Generated Claude Code configuration
```

## The .servo File Format

MCP servers are packaged using `.servo` files (YAML format):

```yaml
servo_version: "1.0"
name: "graphiti"
version: "1.2.0"
description: "Temporal knowledge graphs for dynamic data"
author: "Zep AI"
metadata:
  license: "MIT"

# Server runtime requirements  
runtime_requirements:
  python:
    version: ">=3.10"
  uv: {}

# Server configuration
server:
  transport: "stdio"
  command: "uv"
  args: ["run", "python", "-m", "graphiti.mcp_server"]
  working_directory: "."

# Service dependencies
dependencies:
  services:
    neo4j:
      image: "neo4j:5.13"
      ports: ["7687:7687", "7474:7474"] 
      environment:
        NEO4J_AUTH: "neo4j/${NEO4J_PASSWORD}"
      volumes:
        - "neo4j_data:/data"

# Configuration schema
configuration_schema:
  secrets:
    openai_api_key:
      description: "OpenAI API key for embeddings"
      required: true
      type: "string"
      env_var: "OPENAI_API_KEY"
    neo4j_password:
      description: "Neo4j database password"
      required: true
      type: "string"
      env_var: "NEO4J_PASSWORD"
```

### Key Sections:

- **metadata**: Basic package information
- **runtime_requirements**: Programming language and tool versions
- **server**: MCP server execution configuration  
- **dependencies**: Required services (databases, caches, etc.)
- **configuration_schema**: Secrets and environment variables

## Workflow Examples

### Team Development Workflow

**Team Lead Setup:**
```bash
cd awesome-project
servo init awesome-app --clients vscode,claude-code
servo install https://github.com/getzep/graphiti.git
servo install ./custom-analytics-server/

# Configure project environment variables
servo env set API_TIMEOUT "30s"
servo env set BASE_URL "https://api.example.com"
servo env set DEBUG_MODE "false"

# Configure project secrets
servo secrets set openai_api_key sk-proj-...
servo work

# Commit project configuration (env vars are included)
git add .servo/project.yaml .servo/env.yaml .devcontainer/ .vscode/ .mcp.json
git commit -m "Add servo project configuration"
git push
```

**Team Member Onboarding:**
```bash
git clone https://github.com/team/awesome-project.git
cd awesome-project

# Check what's needed
servo status
# Output: "Missing required secrets: openai_api_key"

# Configure local secrets (env vars already available from repo)
servo secrets set openai_api_key sk-proj-...

# Start development (env vars automatically injected into docker-compose)
servo work
# Opens VS Code with MCP servers automatically configured
```

### Multi-Environment Development

**Development Environment:**
```bash
servo session create development --description "Local development"
servo session activate development
servo install ./dev-tools-server.servo --session development

# Set environment variables for development
servo env set DEBUG_MODE "true"
servo env set LOG_LEVEL "debug"
servo env set API_TIMEOUT "60s"

# Set secrets
servo secrets set dev_api_key sk-dev-...
servo work
```

**Staging Environment:**
```bash
servo session create staging --description "Staging environment"
servo session activate staging  
servo install ./analytics-server.servo --session staging

# Set environment variables for staging
servo env set DEBUG_MODE "false"
servo env set LOG_LEVEL "info"
servo env set API_URL "https://staging-api.company.com"

# Set secrets
servo secrets set staging_api_key sk-staging-...
servo work --client vscode
```

### CI/CD Integration

```bash
# In CI pipeline
export SERVO_NON_INTERACTIVE=1

servo status                          # Validate project
servo secrets set api_key "${API_KEY}" # Configure secrets from CI
servo validate ./new-server.servo     # Validate new servers
```

## Supported MCP Clients

| Client | Auto-Detection | Configuration | Project Support |
|--------|----------------|---------------|-----------------|
| VS Code | ✅ | `.vscode/mcp.json` | ✅ |
| Claude Code | ✅ | `.mcp.json` | ✅ |  
| Cursor | ✅ | `.cursor/mcp.json` | ✅ |

## Contributing

See [Contributing Guidelines](CONTRIBUTING.md).

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.