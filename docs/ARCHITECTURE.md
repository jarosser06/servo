# Servo Architecture Documentation

## Overview

Servo is a project-based MCP server management system with containerized development environments:

1. **Project-First Design**: Complete project isolation with `.servo/` directories
2. **Session-Based Environments**: Multiple environments per project (dev, staging, prod)  
3. **Automated Configuration**: Zero-config devcontainer and client generation

## Core Components

### Client Plugin System (`clients/`)
Supports VS Code, Claude Code, and Cursor through a standardized registry interface.

#### Registry Interface
```go
type ClientRegistry interface {
    Register(client Client) error
    Get(name string) (Client, error)
    List() []Client
    IsRegistered(name string) bool
}

type Client interface {
    Name() string
    IsInstalled() bool
    GenerateConfig(manifests []ServoDefinition, secretsProvider func(string) (string, error)) error
}
```

#### Built-in Client Implementations
- **VS Code**: Updates `.vscode/mcp.json` with MCP server configurations
- **Claude Code**: Updates `.mcp.json` with server configurations
- **Cursor**: Updates `.cursor/settings.json` with MCP configurations

### 2. Project Session Management (`internal/project/`, `internal/session/`)

Servo uses project-based sessions for complete isolation and portability:

#### Project Structure
```
project/
├── .servo/
│   ├── project.yaml          # Project configuration and server declarations
│   ├── secrets.yaml          # Base64-encoded secrets (local only)
│   ├── .gitignore           # Excludes secrets and volumes
│   └── sessions/
│       ├── default/         # Default session
│       │   ├── manifests/   # Downloaded .servo file definitions
│       │   ├── config/      # Session-specific config overrides
│       │   └── volumes/     # Docker volumes (gitignored)
│       └── development/     # Development session
│           ├── manifests/   
│           ├── config/      
│           └── volumes/     
├── .devcontainer/           # Generated development container
│   ├── devcontainer.json   # Dev container configuration
│   └── docker-compose.yml  # Service dependencies
├── .vscode/
│   └── settings.json        # VS Code MCP configuration
└── .mcp.json               # Claude Code MCP configuration
```

#### Project Manager Interface
```go
type Manager struct{}

func (m *Manager) Init(sessionName string, clients []string) (*Project, error)
func (m *Manager) Get() (*Project, error)
func (m *Manager) AddMCPServerToSession(serverName, source string, clients []string, sessionName string, forceUpdate bool) error
func (m *Manager) AddRequiredSecret(name, description string) error
func (m *Manager) GetMissingSecrets() ([]RequiredSecret, error)
func (m *Manager) GetConfiguredSecrets() (map[string]bool, error)
```

#### Session Resolution Algorithm
1. **Project Detection**: Check for `.servo/project.yaml` in current directory
2. **Session Selection**: Use specified session, active session, or default session
3. **Path Resolution**: All operations within session-specific directory structure
4. **Client Generation**: Generate client configs from session manifests

### 3. Service Orchestration (`internal/config/`)

Servo automatically manages service dependencies through generated Docker Compose files and devcontainer configurations.

#### Service Definition in .servo Files
```yaml
# From ServoDefinition type
dependencies:
  services:
    neo4j:
      image: "neo4j:5.13"
      ports: ["7687:7687", "7474:7474"]
      environment:
        NEO4J_AUTH: "neo4j/${NEO4J_PASSWORD}"
      volumes:
        - "neo4j_data:/data"
      auto_generate_password: true
```

#### Configuration Generator Interface
```go
type ConfigGeneratorManager struct {
    servoDir string
}

func (c *ConfigGeneratorManager) GenerateDevcontainer() error
func (c *ConfigGeneratorManager) GenerateDockerCompose() error
```

#### Generated Configurations
- **Devcontainer**: `.devcontainer/devcontainer.json` with runtime features and service integrations
- **Docker Compose**: `.devcontainer/docker-compose.yml` with all service dependencies
- **MCP Configurations**: Client-specific configuration files in project root

#### Service Management Features
- **Auto-generation**: Create devcontainer and Docker Compose from manifest definitions
- **Runtime Detection**: Automatic feature mapping for programming language requirements
- **Data Persistence**: Isolated data volumes per session in `.servo/sessions/*/volumes/`
- **Environment Variables**: Template-based secret substitution
- **Lifecycle Control**: Services managed through devcontainer lifecycle

### 4. Manifest Management (`internal/manifest/`)

Handles .servo file parsing, storage, and retrieval for MCP server definitions.

#### Manifest Store Interface
```go
type Store struct {
    sessionDir string
    parser     *mcp.Parser
}

func NewStore(sessionDir string, parser *mcp.Parser) *Store
func (s *Store) StoreManifest(serverName, source string) error
func (s *Store) LoadManifest(serverName string) (*pkg.ServoDefinition, error)
func (s *Store) ListManifests() (map[string]*pkg.ServoDefinition, error)
func (s *Store) RemoveManifest(serverName string) error
```

#### Project Configuration Schema
```yaml
# .servo/project.yaml (from Project type)
clients: ["vscode", "claude-code", "cursor"]
default_session: default
active_session: development

mcp_servers:
  - name: "graphiti"
    source: "https://github.com/getzep/graphiti.git"
    clients: ["vscode", "claude-code"]
    sessions: ["default", "development"]

required_secrets:
  - name: "openai_api_key"
    description: "OpenAI API key for embeddings"
```

#### Servo File Definition Schema
```yaml
# From ServoDefinition type (pkg/types.go)
servo_version: "1.0"
name: "graphiti"
version: "1.2.0"
description: "Temporal knowledge graphs for dynamic data"
author: "Zep AI"
license: "MIT"

install:
  type: "git"
  setup_commands: ["uv sync"]
  build_commands: ["uv build"]

server:
  transport: "stdio"
  command: "uv"
  args: ["run", "python", "-m", "graphiti.mcp_server"]
  working_directory: "."

dependencies:
  services:
    neo4j:
      image: "neo4j:5.13"
      ports: ["7687:7687", "7474:7474"]
      environment:
        NEO4J_AUTH: "neo4j/${NEO4J_PASSWORD}"

configuration_schema:
  secrets:
    openai_api_key:
      description: "OpenAI API key for embeddings"
      required: true
      type: "string"
      env_var: "OPENAI_API_KEY"
```

### 5. Secrets Management (`internal/cli/commands/secrets.go`)

Simple secrets management using base64 encoding for obscurity.

#### Simple Storage Strategy
- **Encoding**: Base64 encoding for basic obscurity (not cryptographic security)
- **Storage**: Plain YAML in `.servo/secrets.yaml` (never synchronized)
- **Access**: Direct file read/write operations
- **Portability**: Simple YAML format for easy backup and restore

#### Secrets Command Interface
```go
type SecretsCommand struct {
    projectManager *project.Manager
}

func (c *SecretsCommand) setSecret(args []string) error
func (c *SecretsCommand) getSecret(args []string) error
func (c *SecretsCommand) listSecrets(args []string) error
func (c *SecretsCommand) deleteSecret(args []string) error
func (c *SecretsCommand) exportSecrets(args []string) error
func (c *SecretsCommand) importSecrets(args []string) error
```

#### SecretsData Structure
```yaml
# .servo/secrets.yaml
version: "1.0"
secrets:
  openai_api_key: "c2stMTIzNDU2Nzg5MGFiY2RlZg=="  # base64 encoded
  neo4j_password: "bXlfc2VjdXJlX3Bhc3N3b3Jk"      # base64 encoded
```

#### Storage Process
1. **Encoding**: Base64 encode secret values for basic obscurity
2. **Storage**: Write to `.servo/secrets.yaml` with 0600 permissions
3. **Git Exclusion**: Automatically excluded via `.servo/.gitignore`
4. **Team Sharing**: Secrets declared in project config but values stored locally

### 6. MCP Server Parsing (`internal/mcp/`)

Handles parsing and validation of .servo files from various sources.

#### Parser Interface
```go
type Parser struct{}

func (p *Parser) ParseFromFile(filePath string) (*pkg.ServoDefinition, error)
func (p *Parser) ParseFromURL(url string) (*pkg.ServoDefinition, error) 
func (p *Parser) ParseFromGitRepo(repoURL, subPath string) (*pkg.ServoDefinition, error)
func (p *Parser) ParseFromDirectory(dirPath string) (*pkg.ServoDefinition, error)
```

#### Validator Interface
```go
type Validator struct{}

func (v *Validator) ValidateServoDefinition(def *pkg.ServoDefinition) error
func (v *Validator) ValidateSource(source string) error
```

#### Source Types
- **Git Repositories**: Clone repos and find `.servo` files or `servo.yaml`
- **Local Files**: Parse `.servo` files directly from filesystem
- **Local Directories**: Scan directories for `.servo` files
- **Remote URLs**: Fetch and parse `.servo` files from HTTP/HTTPS

#### Validation Features
- **Schema validation**: Ensure `.servo` files match expected structure
- **Dependency checking**: Verify service definitions are valid
- **Requirements validation**: Check runtime requirements are supported

### 7. Command Line Interface (`internal/cli/`)

Comprehensive CLI application built using urfave/cli/v2 framework.

#### Application Structure
```go
type App struct {
    *cli.App
}

func NewApp(version string) (*App, error)
```

#### Core Commands
- **init**: Initialize new project with session and client configuration
- **install**: Install MCP servers from various sources with git authentication
- **status**: Show project status, servers, and missing secrets
- **work**: Generate devcontainer and client configurations
- **session**: Manage project sessions (create, list, activate, delete)
- **config**: Manage project configuration settings  
- **secrets**: Manage encrypted secrets with AES-256-GCM
- **clients**: List available MCP clients and their status
- **validate**: Validate .servo files and installation sources

#### Command Factory Pattern
Commands are instantiated with dependency injection:
```go
installCmd := commands.NewInstallCommand(parser, validator)
secretsCmd := commands.NewSecretsCommand(projectManager)
```

## Data Flow

### Installation Flow (Install Command)
1. **Project Detection**: Verify current directory contains `.servo/project.yaml`
2. **Session Resolution**: Use specified session, active session, or default session
3. **Source Parsing**: Parse `.servo` file from source (git, local file, directory, URL)
4. **Server Name Extraction**: Extract server name from manifest metadata
5. **Validation**: Check if server already exists in target session
6. **Manifest Storage**: Store parsed `.servo` file in session manifests directory
7. **Secret Extraction**: Add required secrets to project configuration
8. **Configuration Generation**: Generate devcontainer, docker-compose, and MCP client configs
9. **Project Update**: Update project.yaml with new server declaration

### Work Command Flow (Generate Development Environment)
1. **Project Detection**: Verify current directory contains `.servo/project.yaml`
2. **Session Resolution**: Use specified session or active session
3. **Manifest Collection**: Load all installed .servo files from session manifests
4. **Devcontainer Generation**: Generate `.devcontainer/devcontainer.json` with runtime features
5. **Docker Compose Generation**: Generate `.devcontainer/docker-compose.yml` with services
6. **MCP Configuration**: Generate client configs (VS Code settings, Claude Code .mcp.json, etc.)
7. **Template Resolution**: Resolve environment variables and secret placeholders

## Security Considerations

### Secret Protection
- **Base64 Encoding**: Basic obscurity for secret values (not cryptographic security)
- **Storage Isolation**: Secrets stored in project-local `.servo/secrets.yaml` files
- **Git Exclusion**: Secrets automatically excluded via `.servo/.gitignore`
- **Access Control**: 0600 file permissions on secrets files
- **Team Workflow**: Secrets declared in project config, values stored locally

### Project Isolation
- **Directory Isolation**: Each project completely isolated in its own `.servo/` directory
- **Session Separation**: Session-specific volumes and configurations in `.servo/sessions/`
- **No Global State**: No system-wide configuration or data sharing
- **Container Isolation**: Services run in project-specific Docker containers
- **Network Isolation**: Each project uses dedicated Docker networks

### Input Validation
- **Source Validation**: Validate all installation sources before parsing
- **Schema Validation**: Ensure `.servo` files match expected structure
- **Path Safety**: Prevent directory traversal in file operations
- **Command Sanitization**: Validate and sanitize all shell commands
- **Permission Control**: Appropriate file permissions on all created files

## Performance Considerations

### Efficient Operations
- **Project-Local Operations**: All operations within project directory, no global traversal
- **Session-Based Caching**: Manifests cached in session directories
- **Minimal Parsing**: Parse .servo files only when needed, cache results
- **Concurrent Configuration**: Generate multiple client configs in parallel

### Resource Management
- **Docker Integration**: Leverage Docker's resource management and caching
- **File System Efficiency**: Use hardlinks and symlinks where appropriate
- **Memory Management**: Minimize memory usage for large .servo files
- **Temporary File Cleanup**: Clean up temporary files after operations

## Extensibility Points

### Client Registry System
- **New Client Support**: Register new MCP clients by implementing the Client interface
- **Auto-Detection**: Add detection logic for new client installations
- **Configuration Templates**: Define client-specific configuration formats

### Parser Extensions
- **New Source Types**: Extend parser to support additional installation sources
- **Custom Validation**: Add validation rules for specific .servo file patterns
- **Protocol Support**: Add support for new git authentication methods

### Configuration Generators
- **Runtime Support**: Add new programming language runtime detection
- **Service Templates**: Support additional service types in devcontainer generation
- **Template Engine**: Extend template resolution for custom variable types

## Testing Strategy

### Unit Tests (`*_test.go`)
- **Project Manager**: Test project initialization, configuration management
- **Secrets Management**: Test encryption/decryption, key derivation
- **Parser**: Test .servo file parsing from various sources
- **Client Registry**: Test client detection and configuration generation

### Integration Tests (`test/`)
- **Complete Workflows**: Test init -> install -> work -> secrets workflows
- **Git Authentication**: Test with various git authentication methods
- **Docker Integration**: Test devcontainer and service generation
- **Client Integration**: Verify generated configurations work with actual clients

### Error Handling

All errors result in exit code 1 with descriptive messages:
- **User Errors**: Invalid arguments, missing files, permission issues
- **Network Errors**: Git clone failures, download failures  
- **Configuration Errors**: Invalid .servo files, missing requirements
- **System Errors**: Docker unavailable, filesystem issues

Error messages are designed to be actionable and help users resolve issues.

## Implementation Status

### Current Implementation
- ✅ **Project Management**: Complete project-local isolation
- ✅ **Session Support**: Multiple environments per project
- ✅ **MCP Client Integration**: VS Code, Claude Code, Cursor
- ✅ **Secrets Management**: AES-256-GCM encryption with PBKDF2
- ✅ **Git Authentication**: SSH keys, HTTP tokens, credentials
- ✅ **Devcontainer Generation**: Automatic runtime feature detection
- ✅ **CLI Framework**: Comprehensive command set with urfave/cli/v2

### Future Enhancements
- **Registry Support**: Package registry integration for .servo files  
- **Service Management**: Direct Docker service lifecycle control
- **Configuration Validation**: Advanced .servo file linting
- **Template Engine**: More sophisticated variable substitution
- **Plugin System**: Third-party client plugin support