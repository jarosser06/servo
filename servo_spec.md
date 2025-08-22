# Servo - MCP Server Package Manager Specification

## **Overview**
Servo is a unified tool for managing Model Context Protocol (MCP) servers across development environments with flexible client support and extensible architecture. Think "homebrew for MCP servers" with pluggable client integrations and configurable scope management.

## **User Stories**

### **Flexible Installation Scope**
```
As a developer, I want to choose installation scope based on my needs
So I can have different MCP setups for different projects or a shared global setup

GIVEN I'm working on a specific project with unique requirements
WHEN I run `mcp install graphiti --scope project`
THEN Graphiti is installed only for the current project directory
WHEN I run `mcp install memento --scope global`
THEN Memento is available across all my projects and tools
```

```
As a developer, I want to specify which clients should use each MCP server
So I can have different configurations for different tools

GIVEN I install Graphiti globally
WHEN I run `servo install graphiti --clients claude-desktop,vscode`
THEN only Claude Desktop and VS Code get configured to use Graphiti
AND Claude Code and Cursor remain unchanged
```

### **Extensible Client Support**
```
As a developer, I want to easily add support for new MCP-compatible tools
So the tool stays useful as the MCP ecosystem grows

GIVEN a new MCP client tool becomes available
WHEN I create a client plugin configuration
THEN the tool can automatically configure that client
AND existing installations can be updated to support the new client
```

```
As a developer, I want to see which clients are supported and configured
So I can understand my current setup and available options

GIVEN I have multiple MCP-compatible tools installed
WHEN I run `servo clients list`
THEN I see all detected and supported clients with their status
WHEN I run `servo clients configure vscode --enable`
THEN VS Code is configured to use my installed MCP servers
```

### **Project-Specific vs Global Management**
```
As a developer, I want project-specific MCP configurations
So different projects can have different memory/knowledge requirements

GIVEN I'm working on a research project needing academic paper indexing
WHEN I run `servo install research-papers --scope local`
THEN the MCP server is only available in this project directory
AND my global setup remains unchanged
```

```
As a developer, I want to inherit global configs in projects
So I get baseline functionality everywhere with project-specific additions

GIVEN I have global MCP servers installed
WHEN I install a local server `servo install tool --scope local`
THEN supported clients get both global + local servers
AND clients that don't support local scope continue using global only
```

### **Core Installation & Management**
```
As a developer, I want to install MCP servers with a single command
So I don't have to manually clone repos, configure dependencies, and wire up configs

GIVEN I want to use Graphiti for knowledge graphs
WHEN I run `servo install graphiti`
THEN it automatically installs Neo4j, clones Graphiti, configures everything, and updates my Claude configs
```

```
As a developer, I want to see what MCP servers are available and installed
So I can discover new servers and manage my current setup

GIVEN I'm exploring MCP options
WHEN I run `servo list --available`
THEN I see a catalog of available servers with descriptions and features
WHEN I run `servo list --installed` 
THEN I see currently installed servers with status (running/stopped)
```

### **Service Management**
```
As a developer, I want MCP servers and their dependencies to start automatically
So I don't have to remember to manually start Neo4j, Redis, etc.

GIVEN I have Graphiti installed with Neo4j dependency
WHEN my system starts up
THEN Neo4j automatically starts in the background
AND Graphiti connects successfully when Claude needs it
```

```
As a developer, I want to control services when needed
So I can troubleshoot issues or save resources

GIVEN I have multiple MCP servers with database dependencies
WHEN I run `mcp stop graphiti` 
THEN it stops Graphiti and optionally its Neo4j dependency
WHEN I run `mcp restart --all`
THEN all services restart in dependency order
```

### **Cross-Machine Synchronization**
```
As a developer, I want my MCP setup to sync across machines
So I have the same capabilities on my laptop, desktop, and work machine

GIVEN I've configured MCP servers on my main laptop
WHEN I run `servo sync push`
THEN my configuration is saved to a git repository
WHEN I run `servo sync pull` on a new machine
THEN all my MCP servers are installed and configured identically
```

### **Interactive Configuration & Secrets Management**
```
As a developer, I want to easily configure secrets and settings for MCP servers
So I don't have to manually edit config files or remember where to put API keys

GIVEN I install Graphiti which requires an OpenAI API key
WHEN I run `servo configure graphiti`
THEN I'm prompted interactively for required configuration
AND secrets are stored separately from the main config
AND the system automatically wires everything together
```

```
As a developer, I want secrets to be stored securely and separately from config
So I can check in my configuration without exposing sensitive data

GIVEN I have configured multiple MCP servers with API keys
WHEN I run `servo sync push`
THEN configuration is synced but secrets remain local
AND on a new machine, I'm prompted only for the secrets I need
```

```
As a developer, I want dependency-managed secrets to be handled automatically
So I don't have to manually configure database passwords and connection strings

GIVEN I install a server that uses Neo4j
WHEN the system sets up the Neo4j container
THEN it automatically generates secure passwords
AND configures the MCP server with the correct connection details
AND I never have to manually handle database credentials
```
```
As a developer, I want to easily debug MCP issues
So I can quickly identify and fix problems

GIVEN an MCP server isn't working properly
WHEN I run `servo doctor graphiti`
THEN it checks dependencies, connectivity, configs, and reports specific issues
WHEN I run `servo logs graphiti --tail`
THEN I see real-time logs from the server and its dependencies
```

## **Architecture Requirements**

### **Pluggable Client System**
```
clients/
├── core/
│   ├── interface.go          # Client plugin interface
│   └── registry.go           # Client registry management
├── claude_desktop/
│   ├── client.go            # Claude Desktop implementation
│   ├── config.go            # Config file management
│   └── detector.go          # Installation detection
├── claude_code/
│   ├── client.go            # Claude Code CLI integration
│   └── detector.go
├── vscode/
│   ├── client.go            # VS Code extension integration
│   ├── copilot.go           # GitHub Copilot MCP support
│   └── detector.go
├── cursor/
│   ├── client.go
│   └── detector.go
└── template/
    └── example_client.go     # Template for new clients
```

### **Client Plugin Interface**
```go
type Client interface {
    // Metadata
    Name() string
    Description() string
    SupportedPlatforms() []string
    
    // Detection
    IsInstalled() bool
    GetVersion() (string, error)
    
    // Scope Management
    GetSupportedScopes() []ClientScope
    ValidateScope(scope string) error
    
    // Configuration
    GetCurrentConfig(scope string) (*MCPConfig, error)
    WriteConfig(scope string, servers []ServerConfig) error
    ValidateConfig(scope string) error
    
    // Lifecycle
    RequiresRestart() bool
    TriggerReload() error
}

type ClientScope struct {
    Name        string   // "global" or "local"
    ConfigPath  string   // Where this scope's config lives  
    Description string   // Human-readable description
}
```

### **Scope Management Architecture**
```
~/.servo/
├── global/
│   ├── config.yaml           # Global configuration
│   ├── secrets.enc           # Encrypted secrets file (not synced)
│   ├── servers/              # Global server installations
│   └── services/             # Global background services
├── projects/
│   └── project-hash-123/
│       ├── config.yaml       # Project-specific config
│       ├── secrets.enc       # Project-specific secrets (not synced)
│       ├── servers/          # Project-specific servers
│       └── .project-root     # Link to actual project directory
├── services/                 # Docker service management
│   ├── global/
│   │   ├── graphiti/
│   │   │   ├── docker-compose.yml    # Generated by Servo
│   │   │   ├── .env                  # Generated secrets
│   │   │   ├── neo4j/                # Data volume
│   │   │   └── redis/                # Data volume
│   │   └── memento/
│   │       ├── docker-compose.yml
│   │       └── postgres/
│   └── local/
│       └── project-hash-abc123/
│           └── research-tool/
│               ├── docker-compose.yml
│               └── elasticsearch/
├── cache/
│   ├── client-detections.json
│   └── server-catalog.json
└── sync/
    ├── .git/                 # Git repo for config sync (excludes secrets)
    └── .gitignore            # Excludes *.enc files

# Client-specific config files (managed by client plugins)
~/.config/claude_desktop/claude_desktop_config.json  # Claude Desktop (global only)
~/.mcp.json                                          # VS Code (global)
.vscode/mcp.json                                     # VS Code (local/project)
~/.cursor/mcp.json                                   # Cursor (global) 
.cursor/mcp.json                                     # Cursor (local/project)
```

### **Project Structure**
```
mcp-manager/
├── cmd/
│   └── mcp/
│       └── main.go           # CLI entry point
├── internal/
│   ├── config/              # Config management
│   ├── server/              # MCP server definitions
│   ├── service/             # Docker/process management
│   ├── client/              # Client plugin system
│   ├── scope/               # Global vs project scope
│   └── sync/                # Git sync functionality
├── clients/                 # Client plugin implementations
│   ├── claude_desktop/
│   ├── claude_code/
│   ├── vscode/
│   └── cursor/
├── server-definitions/       # Legacy: Built-in server configs (deprecated in favor of .servo files)
│   ├── graphiti.servo
│   ├── memento.servo
│   └── basic-memory.servo
├── scripts/
│   ├── bootstrap.sh
│   └── install.sh
└── docs/
    ├── plugin-development.md
    └── client-integration.md
```

## **Integration Points**

### **MCP Client Integrations**
- **Claude Desktop**: Automatically update `~/.config/claude_desktop/claude_desktop_config.json`
- **Claude Code**: Interface with `claude mcp` commands
- **Cursor**: Update `~/.cursor/mcp.json` and `.cursor/mcp.json`
- **VS Code**: Support `.vscode/settings.json` and extension configurations
- **GitHub Copilot**: MCP integration through VS Code extension

### **Dependency Management**
- **Docker**: For containerized service dependencies (Neo4j, Redis, PostgreSQL) via generated docker-compose files
- **Python/uv**: For Python-based MCP servers
- **Node.js/npm**: For TypeScript/JavaScript MCP servers  
- **Binary downloads**: For compiled servers
- **Git repositories**: For source-based installations

### **Service Management**
- **System startup integration**: launchd (macOS), systemd (Linux), Task Scheduler (Windows)
- **Process monitoring**: Health checks and automatic restarts via docker-compose
- **Port management**: Automatic port assignment to avoid conflicts
- **Data isolation**: Per-server, per-scope data volumes in ~/.servo/services/
- **Docker orchestration**: Generated docker-compose files for service management

### **Configuration Sync**
- **Git backend**: Store configs in private git repositories (secrets excluded)
- **Secret management**: Separate encrypted files that never sync
- **Environment handling**: Support for dev/staging/prod environments
- **Conflict resolution**: Handle divergent configs across machines
- **Interactive setup**: Prompt for missing secrets when syncing to new machines

## **Baseline Dependencies**

### **System Requirements**
- **Operating Systems**: macOS 10.15+, Linux (Ubuntu 20.04+), Windows 10+
- **Architecture**: x86_64, ARM64 (Apple Silicon)
- **Memory**: 4GB RAM minimum (for typical MCP server + database combinations)
- **Storage**: 2GB free space for tool and common dependencies

### **Required Tools**
- **Docker**: For containerized service dependencies via generated docker-compose files
- **Git**: For repository cloning and config synchronization
- **curl/wget**: For downloading binaries and installation scripts

### **Language Runtimes** (installed on-demand)
- **Python 3.10+**: Via uv package manager
- **Node.js 18+**: For npm-based MCP servers
- **Go 1.20+**: For compiled servers requiring source builds

### **Optional Dependencies**
- **Hammerspoon** (macOS): For advanced automation integration
- **SSH key**: For private git repository access
- **Cloud CLI tools**: For cloud-hosted databases (AWS CLI, etc.)

## **Configuration Schema**

### **Main Configuration with Scope Support**
```yaml
# ~/.servo/config.yaml (global) or ./.servo/config.yaml (local)
version: "1.0"
scope: "global"  # or "local"

settings:
  auto_start_services: true
  inherit_global: true  # (local scope only)
  log_retention_days: 7
  update_check_interval: "24h"
  secrets_encryption: true
  
sync:
  enabled: true
  repository: "git@github.com:username/servo-config.git"
  auto_push: false
  exclude_secrets: true  # Always true - secrets never sync
  
# Client-specific configurations
clients:
  claude_desktop:
    enabled: true
    auto_configure: true
    config_path: "~/.config/claude_desktop/claude_desktop_config.json"
  claude_code:
    enabled: true
    auto_configure: true
  vscode:
    enabled: false
    extensions: ["github.copilot"]  # Which VS Code extensions support MCP
    config_path: "~/.vscode/settings.json"
  cursor:
    enabled: false
    auto_configure: false
  custom_client:
    enabled: false
    plugin_path: "./plugins/custom_client.so"

# Server installations with client targeting
servers:
  graphiti:
    version: "1.0.0"
    clients: ["claude_desktop", "vscode"]  # Only configure these clients
    # User configuration (not secrets)
    config:
      model_name: "gpt-4o-mini"
      vector_dimensions: 1536
    # Reference to secrets (stored in secrets.enc)
    secrets_required:
      - openai_api_key
  memento:
    version: "latest"
    clients: ["*"]  # Configure all enabled clients
    secrets_required:
      - openai_api_key
    config:
      confidence_decay_days: 30
```

## **Installation Method Interface**

Servo supports pluggable installation methods through a simple interface. The `.servo` file specifies which method to use, and Servo handles the rest.

```go
type InstallMethod interface {
    Name() string
    CanInstall(source string) bool
    Install(servoFile ServoDefinition, installPath string) error
    Validate(servoFile ServoDefinition) error
    GetRequiredTools() []string
}

// Core implementations:
type GitInstaller struct{}     // Clone git repositories  
type LocalInstaller struct{}   // Install from local directory
type FileInstaller struct{}    // Install from standalone .servo file
type RemoteInstaller struct{}  // Download remote .servo files
```

### **Docker Service Management**

Servo generates and manages docker-compose files behind the scenes to orchestrate service dependencies. Users define simple service requirements in `.servo` files, and Servo handles the complexity of port assignment, data persistence, and container orchestration.

#### **User's .servo Definition:**
```yaml
dependencies:
  services:
    neo4j:
      image: "neo4j:5.13"
      ports: ["7687"]  # Servo auto-assigns external port
      environment:
        NEO4J_AUTH: "neo4j/{{.Password}}"
      auto_generate_password: true
```

#### **Generated docker-compose.yml:**
```yaml
# ~/.servo/services/global/graphiti/docker-compose.yml
version: '3.8'
services:
  neo4j-graphiti:
    image: neo4j:5.13
    container_name: servo-neo4j-graphiti-global
    ports:
      - "7687:7687"  # Or auto-assigned port to avoid conflicts
    environment:
      NEO4J_AUTH: "neo4j/auto-generated-password-xyz"
    volumes:
      - ~/.servo/services/global/graphiti/neo4j:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "cypher-shell 'RETURN 1'"]
      interval: 30s

volumes:
  neo4j-data:
    driver: local
```

#### **Service Management Commands:**
```bash
# Servo manages docker-compose behind the scenes
servo start graphiti
# Runs: docker-compose -f ~/.servo/services/global/graphiti/docker-compose.yml up -d

servo services logs neo4j-graphiti
# Runs: docker-compose logs in the appropriate service directory

servo services stop neo4j-graphiti  
# Stops specific service container

servo stop graphiti --services
# Stops all services for the graphiti server
```

## **MCP Server Package Definition**

Servo uses a unified `.servo` file format for all MCP server definitions. This file can be:
- **Embedded in server repositories** (as `.servo` in the repo root)
- **Standalone files** that can be shared, stored, or distributed independently
- **Remote files** accessible via URLs

### **Package Definition File: `.servo`**

The `.servo` file is the complete package definition that contains all installation instructions, dependencies, and configuration schema needed to install and run an MCP server.

```yaml
# .servo - MCP Server Package Definition
servo_version: "1.0"
metadata:
  name: "graphiti"
  version: "1.2.0"
  description: "Temporal knowledge graphs with weighted relationships"
  author: "Zep <team@getzep.com>"
  license: "MIT"
  homepage: "https://github.com/getzep/graphiti"
  repository: "https://github.com/getzep/graphiti.git"
  tags: ["knowledge-graph", "temporal", "neo4j", "memory"]

# System requirements that Servo cannot install automatically  
requirements:
  system:
    - name: "ffmpeg"
      description: "Required for audio/video processing"
      check_command: "ffmpeg -version"
      install_hint: "brew install ffmpeg"
      platforms:
        darwin: "brew install ffmpeg"
        linux: "apt-get install ffmpeg"
        windows: "choco install ffmpeg"
  runtimes:
    - name: "python"
      version: ">=3.10"
    - name: "uv"
      version: "latest"
  
# Installation instructions
install:
  type: "git"  # git or local
  method: "git"  # Explicitly specify install method interface
  subdirectory: "mcp_server"  # Optional: if server code is in subdirectory
  setup_commands:
    - "uv sync"
    - "uv run pip install -e ."
  
# Dependencies (auto-managed by Servo via generated docker-compose files)
dependencies:
  services:
    neo4j:
      image: "neo4j:5.13"
      ports: ["7687"]  # Servo auto-assigns external port to avoid conflicts
      environment:
        NEO4J_AUTH: "neo4j/{{.Password}}"
        NEO4J_PLUGINS: '["apoc"]'
      volumes:
        - "{{.DataPath}}:/data"
        - "{{.LogsPath}}:/logs"
      healthcheck:
        test: ["CMD-SHELL", "cypher-shell 'RETURN 1'"]
        interval: 30s
        timeout: 10s
        retries: 3
      auto_generate_password: true
      shared: false  # Isolated per scope by default
    redis:
      image: "redis:7"
      ports: ["6379"]
      command: ["redis-server", "--appendonly", "yes"]
      volumes:
        - "{{.DataPath}}:/data"
      shared: false

# Configuration schema (for interactive setup)
configuration_schema:
  secrets:
    openai_api_key:
      description: "OpenAI API key for embeddings"
      type: "api_key"
      required: true
      validation: "^sk-[a-zA-Z0-9]{20,}$"
      prompt: "Enter your OpenAI API key (starts with sk-)"
      env_var: "OPENAI_API_KEY"
    anthropic_api_key:
      description: "Anthropic API key (optional)"
      type: "api_key"
      required: false
      validation: "^sk-ant-[a-zA-Z0-9-]+$"
      prompt: "Enter your Anthropic API key (optional, starts with sk-ant-)"
      env_var: "ANTHROPIC_API_KEY"
  config:
    model_name:
      description: "OpenAI model for embeddings"
      type: "select"
      required: false
      default: "gpt-4o-mini"
      options: ["gpt-4o-mini", "gpt-4o", "text-embedding-3-small"]
      env_var: "MODEL_NAME"
    vector_dimensions:
      description: "Vector embedding dimensions"
      type: "integer"
      required: false
      default: 1536
      validation: "^(384|768|1536|3072)$"
      env_var: "NEO4J_VECTOR_DIMENSIONS"

# Server execution configuration
server:
  transport: "stdio"  # stdio, sse, http
  command: "{{.RuntimePath.uv}}"
  args:
    - "run"
    - "--directory={{.InstallPath}}"
    - "graphiti_mcp_server.py"
    - "--transport={{.Transport}}"
  environment:
    # Auto-injected by Servo
    NEO4J_URI: "{{.Services.neo4j.uri}}"
    NEO4J_USER: "{{.Services.neo4j.user}}"  
    NEO4J_PASSWORD: "{{.Services.neo4j.password}}"
    # User-provided secrets
    OPENAI_API_KEY: "{{.Secrets.openai_api_key}}"
    ANTHROPIC_API_KEY: "{{.Secrets.anthropic_api_key}}"
    # User configuration
    MODEL_NAME: "{{.Config.model_name}}"
    NEO4J_VECTOR_DIMENSIONS: "{{.Config.vector_dimensions}}"

# Optional: Client-specific configurations
clients:
  recommended: ["claude-desktop", "vscode", "cursor"]
  tested: ["claude-desktop", "vscode", "cursor", "claude-code"]
  
# Optional: Documentation and examples
documentation:
  readme: "README.md"
  examples:
    - name: "Basic Usage"
      description: "Simple knowledge graph setup"
      file: "examples/basic.md"
    - name: "Research Assistant"
      description: "Academic paper indexing"
      file: "examples/research.md"
```

### **Installation from .servo Files**

```bash
# Install from git repo with .servo file in root
servo install https://github.com/getzep/graphiti.git

# Install specific version/branch
servo install https://github.com/getzep/graphiti.git@v1.2.0
servo install https://github.com/user/custom-server.git@main

# Install from local directory with .servo file
servo install ./my-custom-server/

# Install from standalone .servo file
servo install ./custom-graphiti.servo
servo install https://example.com/configs/my-server.servo

# Install with scope and client targeting (works with any source)
servo install ./research-tool.servo --scope local --clients vscode
```

### **Package Inspection and Validation**

```bash
# Inspect any .servo source before installing
servo inspect https://github.com/getzep/graphiti.git
servo inspect ./my-custom-server/
servo inspect ./custom-config.servo
servo inspect https://example.com/configs/server.servo

# All show the same output format:
# Package: graphiti v1.2.0
# Description: Temporal knowledge graphs with weighted relationships
# Author: Zep <team@getzep.com>
# 
# Dependencies:
#   Services: neo4j (required), redis (optional)
#   Runtimes: python >=3.10, uv latest
# 
# Configuration Required:
#   Secrets: openai_api_key (required), anthropic_api_key (optional)
#   Config: model_name, vector_dimensions
# 
# Tested Clients: claude-desktop, vscode, cursor, claude-code

# Validate any .servo file
servo validate ./.servo
servo validate ./my-config.servo
servo validate https://github.com/user/server.git
```
### **Built-in Server Definition Format (Legacy Support)**

For backward compatibility, Servo continues to support built-in server definitions, but these are being migrated to standalone `.servo` files.

```yaml
# servers/graphiti.servo (same format as standalone .servo files)
servo_version: "1.0"
metadata:
  name: graphiti
  description: "Temporal knowledge graphs with weighted relationships"
  version: "1.0.0"
  tags: ["knowledge-graph", "temporal", "neo4j"]
  
install:
  type: git
  repository: "https://github.com/getzep/graphiti.git"
  subdirectory: "mcp_server"
  setup_commands:
    - "uv sync"
    
# ... rest of .servo format
```

### **Secrets Management Schema**
```yaml
# secrets.enc (encrypted file, not synced)
version: "1.0"
encryption: "aes-256-gcm"
encrypted_data: "base64-encoded-encrypted-json"

# When decrypted, contains:
secrets:
  openai_api_key: "sk-1234567890abcdef"
  anthropic_api_key: "sk-ant-api03-xyz"
  custom_service_token: "abc123"
  
# Auto-generated by dependency services (not user-editable)
service_secrets:
  neo4j:
    password: "auto-generated-secure-password"
    uri: "bolt://localhost:7687"
    user: "neo4j"
  redis:
    password: "auto-generated-password"
    uri: "redis://localhost:6379"
```

### **Client Plugin Configuration**
```yaml
# clients/vscode/config.yaml
metadata:
  name: "vscode"
  description: "Visual Studio Code with MCP support"
  supported_platforms: ["darwin", "linux", "windows"]
  
scopes:
  global:
    config_path: "~/.mcp.json"
    description: "System-wide VS Code MCP settings"
  local:
    config_path: ".vscode/mcp.json"
    description: "Project-specific VS Code MCP settings"
    
configuration:
  format: "vscode_mcp"  # Uses VS Code's native MCP format
  restart_required: false
  supports_hot_reload: true
  
integration:
  supports_stdio: true
  supports_sse: true
  supports_websocket: false
```

### **Client Implementation Examples**

#### **VS Code Client Plugin**
```go
type VSCodeManager struct{}

func (v *VSCodeManager) GetSupportedScopes() []ClientScope {
    return []ClientScope{
        {
            Name: "global",
            ConfigPath: "~/.mcp.json", 
            Description: "System-wide VS Code MCP settings",
        },
        {
            Name: "local",
            ConfigPath: ".vscode/mcp.json",
            Description: "Project-specific VS Code MCP settings", 
        },
    }
}

func (v *VSCodeManager) WriteConfig(scope string, servers []ServerConfig) error {
    var configPath string
    switch scope {
    case "global":
        configPath = expandPath("~/.mcp.json")
    case "local":
        configPath = ".vscode/mcp.json"
        // Ensure .vscode directory exists
        if err := os.MkdirAll(".vscode", 0755); err != nil {
            return err
        }
    default:
        return fmt.Errorf("unsupported scope: %s", scope)
    }
    
    config := VSCodeMCPConfig{Servers: make(map[string]VSCodeServer)}
    for _, server := range servers {
        config.Servers[server.Name] = VSCodeServer{
            Command: server.Command,
            Args:    server.Args,
            Env:     server.Environment,
        }
    }
    
    return writeJSONFile(configPath, config)
}
```

#### **Claude Desktop Client Plugin**
```go
type ClaudeDesktopManager struct{}

func (c *ClaudeDesktopManager) GetSupportedScopes() []ClientScope {
    return []ClientScope{
        {
            Name: "global",
            ConfigPath: "~/.config/claude_desktop/claude_desktop_config.json",
            Description: "Claude Desktop configuration (global only)",
        },
        // No local scope - Claude Desktop doesn't support project-specific configs
    }
}

func (c *ClaudeDesktopManager) ValidateScope(scope string) error {
    if scope != "global" {
        return fmt.Errorf("claude-desktop only supports global scope")
    }
    return nil
}
```

## **Command Line Interface**

### **Scope Management Commands**
```bash
# Scope operations
servo scopes list [--client <client>]  # Show available scopes, optionally filtered by client
servo scopes status [--client <client>]  # Show current global + local server status
servo scope init                       # Initialize local scope in current directory (create .vscode/, .cursor/ dirs)
```

### **Client Management Commands**
```bash
# Client operations  
servo clients list                # Show all configured clients and their supported scopes
servo clients add claude-desktop  # Explicitly add client (no auto-detection)
servo clients add vscode --config-path ~/.vscode/settings.json  # Override config path
servo clients remove cursor      # Remove client configuration
servo clients configure claude-desktop --dry-run  # Show what would be configured
```

### **Enhanced Installation Commands**
```bash
# Installation with scope and client targeting
servo install graphiti.servo --scope global --clients claude-desktop,vscode
servo install https://github.com/user/memento.git --scope local --clients vscode,cursor
servo install ./research-papers.servo --scope local --clients cursor
servo uninstall graphiti --scope global   # Remove from global scope only

# Unified installation from any .servo source
servo install https://github.com/getzep/graphiti.git      # .servo in repo root
servo install ./my-custom-server/                         # .servo in directory
servo install ./custom-graphiti.servo                     # Standalone .servo file
servo install https://configs.example.com/server.servo    # Remote .servo file
servo install https://github.com/user/server.git@v1.2.0  # Specific version/branch

# Default scope (global) and client management
servo install <servo-source>          # Install from any .servo source globally
servo install <servo-source> --dry-run  # Show what would be installed
servo install <servo-source> --client vscode  # Install for specific client
servo uninstall <server>              # Remove server and cleanup
servo update <server>                 # Update to latest version
servo search <query>                  # Search available servers

# Client-specific management
servo disable <server> --client <client>  # Disable server for specific client
servo enable <server> --client <client>   # Enable server for specific client
```

### **Management Commands**
```bash
servo list                        # Show installed servers
servo list --available            # Show all available servers  
servo status                      # Show status of all services and servers
servo start [server]              # Start server and its services
servo stop [server]               # Stop server and its services
servo restart [server]            # Restart server and its services

# Service-specific management (generated docker-compose)
servo services list [--server <name>] [--scope <scope>]  # List running services
servo services stop <service-name>     # Stop specific service container
servo services restart <service-name>  # Restart specific service container
servo services logs <service-name>     # View service logs
servo start <server> --services-only   # Start only services, not the MCP server
servo stop <server> --services         # Stop all services for a server
```

### **Configuration Commands**
```bash
servo config set <key> <value>    # Set configuration value
servo config get <key>            # Get configuration value
servo config edit                 # Edit config in $EDITOR
servo config set default_client <client>  # Set default client for commands

# Secret management
servo secrets set <key> <value>   # Set secret (encrypted storage)
servo secrets get <key>           # Get secret value (if allowed)
servo secrets list                # List all secret keys (not values)
servo secrets delete <key>        # Remove secret
servo secrets status              # Show which secrets are needed/configured

# Interactive configuration
servo configure <server>          # Interactive setup for server secrets/config
servo configure <server> --reconfigure  # Reconfigure existing server

# Requirements checking
servo doctor --check-requirements  # Check system requirements
servo requirements install         # Attempt to install missing system requirements
```

### **Synchronization Commands**
```bash
servo sync init <repo-url>        # Initialize sync with git repo
servo sync push                   # Push config to remote
servo sync pull                   # Pull and apply remote config
servo sync status                 # Show sync status
```

### **Plugin Management Commands**
```bash
# Plugin operations
servo plugins list                # Show available and installed client plugins
servo plugins install my-client  # Install custom client plugin
servo plugins develop ./my-plugin  # Load plugin from local development path
```

### **Package Management Commands**
```bash
# Package operations (works with any .servo source)
servo install <servo-source>          # Install from git repo, local dir, or .servo file
servo install <directory>             # Install from local directory with .servo file  
servo install <servo-file>            # Install from standalone .servo file
servo inspect <servo-source>          # Show package info before installing
servo validate <servo-source>         # Validate .servo file or source

# Package development
servo init package                    # Create template .servo in current directory
servo package validate                # Validate current directory's .servo
servo package build                   # Test installation process locally
```
### **Debugging Commands**
```bash
servo doctor [server]             # Run health checks
servo logs <server>               # Show logs
servo logs <server> --follow      # Tail logs
servo debug <server>              # Start in debug mode
```

## **Client Integration Specifications**

### **Claude Desktop Integration**
- **Config Path**: `~/.config/claude_desktop/claude_desktop_config.json`
- **Method**: Direct JSON manipulation
- **Restart Required**: Yes
- **Supported Scopes**: `global` only
- **Config Format**: `mcpServers` object with nested server configurations

### **Claude Code Integration**  
- **Method**: `claude mcp` CLI commands
- **Restart Required**: No
- **Supported Scopes**: `global`, `local` (checks current directory)
- **Config Format**: CLI-managed, no direct file manipulation

### **VS Code Integration**
- **Global Config**: `~/.mcp.json`
- **Local Config**: `.vscode/mcp.json` (workspace-specific)
- **Method**: Direct JSON manipulation of MCP config files
- **Restart Required**: No (immediate reload)
- **Supported Scopes**: `global`, `local`
- **Config Format**: `servers` object with command/args/env structure

### **Cursor Integration**
- **Global Config**: `~/.cursor/mcp.json`
- **Local Config**: `.cursor/mcp.json` (project-specific)
- **Method**: Direct JSON manipulation
- **Restart Required**: Yes
- **Supported Scopes**: `global`, `local`
- **Config Format**: Server name keys with command/args properties

### **Custom Client Plugin Template**
```go
package main

import "github.com/servo/core"

type MyClient struct{}

func (c *MyClient) Name() string { return "my-client" }

func (c *MyClient) IsInstalled() bool {
    // Detection logic
    return true
}

func (c *MyClient) WriteConfig(scope string, servers []core.ServerConfig) error {
    // Configuration update logic
    return nil
}

// Export plugin
func NewClient() core.Client {
    return &MyClient{}
}
```

## **Success Criteria**

1. **Flexible scope management**: Both global and project-specific configurations work seamlessly
2. **Extensible client support**: New MCP clients can be added via plugins without core changes
3. **Intelligent client detection**: Automatically discovers installed MCP-compatible tools
4. **Granular client targeting**: Can configure specific servers for specific clients
5. **Cross-platform plugin system**: Client plugins work on macOS, Linux, and Windows
6. **Zero-config installation**: `servo install graphiti` should work without any manual setup
7. **Cross-platform compatibility**: Works on macOS, Linux, and Windows
8. **Automatic service management**: Dependencies start/stop automatically with system
9. **Seamless client integration**: No manual JSON editing required
10. **Reliable synchronization**: Config sync works across multiple machines
11. **Comprehensive troubleshooting**: Clear error messages and debugging tools
12. **Performance**: Commands complete in <5 seconds for typical operations
13. **Backwards compatibility**: Existing MCP configurations continue to work
14. **Developer experience**: Clear plugin development documentation and templates

## **Implementation Phases**

### **Phase 1: Core + Basic Clients**
- Core architecture with plugin system
- Claude Desktop, Claude Code, Cursor clients
- Global scope support
- Basic server management

### **Phase 2: Advanced Scope & VS Code**
- Project scope implementation
- VS Code integration with extension support
- Config inheritance system
- Plugin development tooling

### **Phase 3: Extensibility & Polish**
- Plugin SDK and documentation
- Custom client plugin support
- Advanced debugging and troubleshooting
- Performance optimizations

## **Example Usage Workflows**

### **Interactive Server Configuration**
```bash
# Install server - prompts for required secrets/config
servo install graphiti
# Interactive prompts:
# "Enter OpenAI API key (starts with sk-): "
# "Select embedding model [gpt-4o-mini]: "
# "Vector dimensions [1536]: "

# Reconfigure existing server
servo configure graphiti --reconfigure
# Shows current config, prompts for changes

# Set secrets manually
servo secrets set openai_api_key sk-1234567890abcdef

# Set configuration manually  
servo config set graphiti.model_name gpt-4o
servo config set graphiti.vector_dimensions 768
```

### **Dependency-Managed vs User-Managed Secrets**
```bash
# Install server with database dependency
servo install memento --with redis

# System automatically:
# 1. Starts Redis container with generated password
# 2. Configures Memento with Redis connection details
# 3. Only prompts for user secrets (OpenAI API key)

# Check what secrets are managed vs required
servo secrets list
# Output:
# User Secrets (required):
#   openai_api_key ✓ configured
#   anthropic_api_key ✗ not set
# 
# Service Secrets (auto-managed):
#   redis.password ✓ auto-generated
#   redis.uri ✓ auto-configured
```

### **Scope-Aware Installation**
```bash
# Initialize local scope in current directory
cd ~/my-project/
servo scope init
# Creates .vscode/ and .cursor/ directories for local configs

# Install servers with explicit scope targeting
servo install graphiti --scope global --clients claude-desktop,vscode
# Result:
# - Claude Desktop: Updates ~/.config/claude_desktop/claude_desktop_config.json
# - VS Code: Updates ~/.mcp.json

servo install research-tool --scope local --clients vscode,cursor  
# Result:
# - VS Code: Updates .vscode/mcp.json
# - Cursor: Updates .cursor/mcp.json
# - Claude Desktop: Skipped (doesn't support local scope)

# Error handling for unsupported combinations
servo install tool --scope local --clients claude-desktop
# Error: Client 'claude-desktop' does not support local scope
#        Supported scopes for claude-desktop: global
```

### **Scope Status and Management**
```bash
# Show current scope status across all clients
servo scopes status
# Output:
# Current directory: ~/my-project/
# 
# Global servers: graphiti, basic-memory
# Local servers: research-tool, academic-papers
#
# Client configurations:
#   claude-desktop (global only): 
#     ~/.config/claude_desktop/claude_desktop_config.json
#     Servers: graphiti, basic-memory, research-tool, academic-papers (all merged)
#   
#   vscode (global + local):
#     Global: ~/.mcp.json → graphiti, basic-memory  
#     Local: .vscode/mcp.json → research-tool, academic-papers
#   
#   cursor (global + local):
#     Global: ~/.cursor/mcp.json → graphiti, basic-memory
#     Local: .cursor/mcp.json → research-tool, academic-papers

# List available scopes per client
servo scopes list
# Output:
# Available scopes by client:
#   claude-desktop:
#     global: ~/.config/claude_desktop/claude_desktop_config.json
#   
#   vscode:
#     global: ~/.mcp.json (system-wide)
#     local: .vscode/mcp.json (project-specific)
#   
#   cursor:
#     global: ~/.cursor/mcp.json (system-wide) 
#     local: .cursor/mcp.json (project-specific)
```
### **Setting up Graphiti for Research Project**
```bash
# Initialize project-specific MCP setup
cd ~/projects/research-assistant
servo scope init

# Install Graphiti globally for all clients
servo install graphiti --scope global

# Install research-specific tool locally for supported clients
servo install research-papers --scope local --clients vscode,cursor

# Start services
servo start graphiti

# Verify setup
servo status
servo doctor graphiti
```

### **Global Setup with Multiple Clients**
```bash
# Set up global configuration
servo clients add claude-desktop
servo clients add vscode
servo clients add cursor

# Install servers globally (prompts for secrets interactively)
servo install memento --scope global --clients all
# Prompts: "Enter OpenAI API key: "

servo install basic-memory --scope global --clients claude-code
# No secrets required for basic-memory

# Sync config to other machines (secrets stay local)
servo sync init git@github.com:username/servo-config.git
servo sync push

# On new machine
servo sync pull
# Prompts: "Server 'memento' requires OpenAI API key: "
# User enters key, system configures everything else
```

### **Secret Management Workflows**
```bash
# Check what secrets are needed across all servers
servo secrets status
# Output:
# ✓ openai_api_key (used by: graphiti, memento)
# ✗ anthropic_api_key (used by: claude-server) - NOT SET
# ✓ neo4j.password (auto-managed by: graphiti)

# Set missing secrets
servo secrets set anthropic_api_key sk-ant-api03-xyz

# Rotate a secret across all servers
servo secrets set openai_api_key sk-new-key-here
servo restart --affected  # Restart only servers using this secret

# View secret metadata (not values)
servo secrets list --verbose
# Output:
# openai_api_key
#   Used by: graphiti, memento  
#   Last updated: 2025-01-15
#   Type: api_key
```

### **Adding Custom Client Support**
```bash
# Add custom client plugin
servo plugins install my-custom-client

# Add the client explicitly
servo clients add my-custom-client

# Install server for the new client
servo install graphiti --scope global --clients my-custom-client
```