# .servo File Specification

## Overview

The `.servo` file is a YAML-based package definition format that contains all the information needed to install, configure, and run an MCP server. This specification defines the structure, validation rules, and semantics for `.servo` files.

## File Structure

```yaml
servo_version: "1.0"                    # Required: Servo spec version
name: "package-name"                    # Required: Package name (lowercase, hyphens)
version: "1.0.0"                        # Optional: Semantic version
description: "Package description"      # Optional: Short description
author: "Author Name"                   # Optional: Author name and contact
license: "MIT"                          # Optional: License identifier (SPDX)
metadata: {}                            # Optional: Additional package metadata
requirements: {}                        # Optional: System and runtime requirements
install: {}                            # Required: Installation instructions
dependencies: {}                        # Optional: Service dependencies  
configuration_schema: {}               # Optional: Interactive configuration
server: {}                             # Required: Server execution config
services: {}                           # Optional: Service dependencies (alternative to dependencies)
clients: {}                            # Optional: Client compatibility info
documentation: {}                      # Optional: Documentation links
```

## Schema Definition

### Root Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `servo_version` | string | ✅ | Servo specification version (currently "1.0") |
| `name` | string | ✅ | Package name (lowercase, hyphens) |
| `version` | string | ❌ | Semantic version (e.g., "1.2.0") |
| `description` | string | ❌ | Short description |
| `author` | string | ❌ | Author name and contact |
| `license` | string | ❌ | License identifier (SPDX) |
| `metadata` | object | ❌ | Additional package metadata (homepage, repository, tags) |
| `requirements` | object | ❌ | System and runtime requirements |
| `install` | object | ✅ | Installation method and commands |
| `dependencies` | object | ❌ | Service dependencies (legacy field) |
| `configuration_schema` | object | ❌ | Interactive configuration schema |
| `server` | object | ✅ | Server execution configuration |
| `services` | object | ❌ | Service dependencies (preferred over dependencies) |
| `clients` | object | ❌ | Client compatibility information |
| `documentation` | object | ❌ | Documentation and examples |

### Metadata Schema

The metadata section now contains only optional fields for additional package information:

```yaml
metadata:
  homepage: string                      # Optional: Project homepage URL
  repository: string                    # Optional: Source repository URL
  tags: []string                        # Optional: Keywords for discovery
```

**Validation Rules:**

Top-level fields:
- `name`: Required, must match `^[a-z][a-z0-9-]*[a-z0-9]$` (lowercase, hyphens, no leading/trailing hyphens)
- `version`: Optional, must follow semantic versioning (e.g., "1.2.0", "2.0.0-beta.1") if provided
- `description`: Optional, maximum 200 characters if provided
- `license`: Optional, should be valid SPDX license identifier if provided
- `author`: Optional, author name and contact information if provided

Metadata fields:
- `homepage`, `repository`: Must be valid URLs if provided
- `tags`: Each tag must match `^[a-z][a-z0-9-]*$`

### Requirements Schema

```yaml
# Note: Both 'requirements' and 'runtime_requirements' are supported for compatibility
requirements:
  system:                              # Array of system requirements
    - name: string                     # System tool name
      description: string              # Human-readable description
      check_command: string            # Command to verify installation
      install_hint: string             # Generic installation hint
      platforms:                       # Platform-specific install commands (map)
        darwin: string                 # macOS installation command
        linux: string                  # Linux installation command
        windows: string                # Windows installation command
  runtimes:                            # Array of runtime requirements
    - name: string                     # Runtime name (python, node, go, uv, etc.)
      version: string                  # Version requirement (>=3.10, ^18.0.0)

# Alternative format (from ServoDefinition):
runtime_requirements:
  python:
    version: ">=3.10"                  # Version requirement
  uv: {}                               # Latest version
  node:
    version: "^18.0.0"                 # Node version constraint
```

**Validation Rules:**
- `system.name`: Must be a valid command name
- `system.check_command`: Must be a safe shell command
- `runtimes.version`: Must be valid version constraint

### Install Schema

```yaml
install:
  type: string                          # Required: git, local, file, remote
  method: string                        # Required: Explicit install method
  repository: string                    # For git: repository URL
  subdirectory: string                  # Optional: subdirectory path
  setup_commands: []string              # Required: Installation commands
  build_commands: []string              # Optional: Build commands
  test_commands: []string               # Optional: Test commands
```

**Installation Types:**

#### Git Installation
```yaml
install:
  type: "git"
  method: "git"
  repository: "https://github.com/user/repo.git"
  subdirectory: "mcp_server"           # Optional
  setup_commands:
    - "uv sync"
    - "uv run pip install -e ."
```

#### Local Installation
```yaml
install:
  type: "local"
  method: "local"
  subdirectory: "."                    # Relative to .servo file
  setup_commands:
    - "npm install"
    - "npm run build"
```

**Validation Rules:**
- `type`: Must be one of: "git", "local", "file", "remote"
- `method`: Must match the `type` value
- `repository`: Required for git type, must be valid git URL
- `setup_commands`: At least one command required
- All commands are validated for safety (no arbitrary code execution)

### Dependencies Schema

```yaml
dependencies:
  services:
    service_name:
      image: string                     # Required: Docker image
      ports: []string                   # Optional: Exposed ports
      environment: map[string]string    # Optional: Environment variables
      volumes: []string                 # Optional: Volume mounts
      command: []string                 # Optional: Override command
      healthcheck:                      # Optional: Health check config
        test: []string                  # Health check command
        interval: string                # Check interval (30s)
        timeout: string                 # Check timeout (10s)
        retries: int                    # Retry count (3)
      auto_generate_password: bool      # Optional: Generate secure password
      shared: bool                      # Optional: Share across scopes (default: false)
```

**Example:**
```yaml
dependencies:
  services:
    neo4j:
      image: "neo4j:5.13"
      ports: ["7687"]
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
      shared: false
```

**Template Variables:**
- `{{.Password}}`: Auto-generated secure password
- `{{.DataPath}}`: Service data directory path
- `{{.LogsPath}}`: Service logs directory path
- `{{.ConfigPath}}`: Service config directory path

**Validation Rules:**
- `image`: Must be valid Docker image reference
- `ports`: Each port must be valid port number (1-65535)
- `environment`: Values can contain template variables
- `healthcheck.interval/timeout`: Must be valid duration strings
- `auto_generate_password`: Only allowed with template variables in environment

### Configuration Schema

```yaml
configuration_schema:
  secrets:
    secret_name:
      description: string               # Required: Human-readable description
      type: string                      # Required: secret type
      required: bool                    # Required: whether secret is mandatory
      validation: string                # Optional: regex validation pattern
      prompt: string                    # Optional: custom prompt text
      env_var: string                   # Required: environment variable name
  config:
    config_name:
      description: string               # Required: Human-readable description
      type: string                      # Required: config type
      required: bool                    # Optional: default false
      default: any                      # Optional: default value
      options: []any                    # Optional: valid options (for select)
      validation: string                # Optional: regex validation
      env_var: string                   # Required: environment variable name
```

**Secret Types:**
- `api_key`: API key or token
- `password`: Password or secret string
- `certificate`: SSL certificate or key
- `url`: URL with embedded credentials

**Config Types:**
- `string`: Text input
- `integer`: Numeric input
- `boolean`: True/false checkbox
- `select`: Dropdown with predefined options
- `multiselect`: Multiple selection
- `file`: File path input
- `url`: URL input

**Example:**
```yaml
configuration_schema:
  secrets:
    openai_api_key:
      description: "OpenAI API key for embeddings"
      type: "api_key"
      required: true
      validation: "^sk-[a-zA-Z0-9]{20,}$"
      prompt: "Enter your OpenAI API key (starts with sk-)"
      env_var: "OPENAI_API_KEY"
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
```

### Server Schema

```yaml
server:
  transport: string                     # Required: stdio, sse, http
  command: string                       # Required: executable command
  args: []string                        # Required: command arguments
  environment: map[string]string        # Optional: environment variables
  working_directory: string             # Optional: working directory
  timeout: string                       # Optional: startup timeout (30s)
```

**Example:**
```yaml
server:
  transport: "stdio"
  command: "{{.RuntimePath.uv}}"
  args:
    - "run"
    - "--directory={{.InstallPath}}"
    - "graphiti_mcp_server.py"
    - "--transport={{.Transport}}"
  environment:
    # Auto-injected service connections
    NEO4J_URI: "{{.Services.neo4j.uri}}"
    NEO4J_USER: "{{.Services.neo4j.user}}"
    NEO4J_PASSWORD: "{{.Services.neo4j.password}}"
    # User-provided secrets
    OPENAI_API_KEY: "{{.Secrets.openai_api_key}}"
    # User configuration
    MODEL_NAME: "{{.Config.model_name}}"
  working_directory: "{{.InstallPath}}"
  timeout: "30s"
```

**Template Variables:**
- `{{.RuntimePath.name}}`: Path to runtime executable (python, node, etc.)
- `{{.InstallPath}}`: Server installation directory
- `{{.Transport}}`: Selected transport method
- `{{.Services.name.property}}`: Service connection details
- `{{.Secrets.name}}`: User-provided secrets
- `{{.Config.name}}`: User configuration values

**Validation Rules:**
- `transport`: Must be one of: "stdio", "sse", "http"
- `command`: Must be valid executable name or template variable
- `timeout`: Must be valid duration string

### Clients Schema

```yaml
clients:
  recommended: []string                 # Optional: recommended clients
  tested: []string                      # Optional: tested clients
  excluded: []string                    # Optional: incompatible clients
  requirements:                         # Optional: client-specific requirements
    client_name:
      minimum_version: string           # Minimum client version
      features: []string                # Required client features
```

**Example:**
```yaml
clients:
  recommended: ["vscode", "claude-code", "cursor"]
  tested: ["vscode", "claude-code", "cursor", "crewai"]
  excluded: ["legacy-client"]
  requirements:
    vscode:
      minimum_version: "1.85.0"
      features: ["mcp-support"]
```

### Documentation Schema

```yaml
documentation:
  readme: string                        # Optional: README file path
  changelog: string                     # Optional: CHANGELOG file path
  examples:                            # Optional: example configurations
    - name: string                      # Example name
      description: string               # Example description
      file: string                      # Example file path
```

## Template System

The `.servo` file supports a powerful template system for dynamic configuration:

### Available Template Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `{{.InstallPath}}` | Server installation directory | `.servo/sessions/default/servers/graphiti` |
| `{{.Transport}}` | Selected transport method | `stdio` |
| `{{.RuntimePath.python}}` | Python executable path | `/usr/bin/python3` |
| `{{.RuntimePath.uv}}` | UV executable path | `/usr/bin/uv` |
| `{{.RuntimePath.node}}` | Node.js executable path | `/usr/bin/node` |
| `{{.Services.name.uri}}` | Service connection URI | `bolt://localhost:7687` |
| `{{.Services.name.user}}` | Service username | `neo4j` |
| `{{.Services.name.password}}` | Service password | `auto-generated-password` |
| `{{.Secrets.name}}` | User-provided secret | `sk-1234567890abcdef` |
| `{{.Config.name}}` | User configuration value | `gpt-4o-mini` |
| `{{.Password}}` | Auto-generated password | `secure-random-password` |
| `{{.DataPath}}` | Service data directory | `.servo/sessions/default/volumes/graphiti/neo4j` |
| `{{.LogsPath}}` | Service logs directory | `.servo/sessions/default/volumes/graphiti/logs` |

### Template Functions

| Function | Description | Example |
|----------|-------------|---------|
| `{{.Env "VAR"}}` | Environment variable | `{{.Env "HOME"}}` |
| `{{.Port 7687}}` | Available port | `{{.Port 7687}}` returns available port |
| `{{.Random 32}}` | Random string | `{{.Random 32}}` generates 32-char string |

## Validation Rules

### Schema Validation
1. **Structure**: Must conform to YAML schema
2. **Required Fields**: All required fields must be present
3. **Type Checking**: All fields must match expected types
4. **Format Validation**: Strings must match format constraints

### Security Validation
1. **Command Safety**: No arbitrary code execution in commands
2. **Path Safety**: No path traversal in file paths
3. **URL Safety**: Only HTTPS URLs allowed (except localhost)
4. **Template Safety**: Only allowed template variables

### Compatibility Validation
1. **Platform Support**: Check platform-specific requirements
2. **Runtime Versions**: Validate version constraints
3. **Client Compatibility**: Check client requirements

## Examples

### Minimal .servo File
```yaml
servo_version: "1.0"
name: "simple-server"
version: "1.0.0"
description: "A simple MCP server"
author: "Example Author <author@example.com>"
license: "MIT"

install:
  type: "local"
  method: "local"
  setup_commands:
    - "echo 'No setup required'"

server:
  transport: "stdio"
  command: "python"
  args: ["server.py"]
```

### Complex .servo File
```yaml
servo_version: "1.0"
name: "knowledge-graph"
version: "2.1.0"
description: "Advanced knowledge graph with temporal relationships"
author: "Knowledge Systems Inc. <info@knowledge.com>"
license: "Apache-2.0"
metadata:
  homepage: "https://knowledge.com"
  repository: "https://github.com/knowledge/graph-server.git"
  tags: ["knowledge-graph", "temporal", "neo4j", "ai"]

# Using runtime_requirements format (matches ServoDefinition)
runtime_requirements:
  python:
    version: ">=3.10"
  uv: {}

# Using legacy requirements format (still supported)  
requirements:
  system:
    - name: "docker"
      description: "Docker for service management"
      check_command: "docker --version"
      install_hint: "Install Docker Desktop or Docker Engine"
      platforms:
        darwin: "brew install --cask docker"
        linux: "apt-get install docker.io"
        windows: "Download Docker Desktop"

install:
  type: "git"
  method: "git"
  repository: "https://github.com/knowledge/graph-server.git"
  subdirectory: "mcp_server"
  setup_commands:
    - "uv sync"
    - "uv run pip install -e ."
  test_commands:
    - "uv run pytest tests/"

dependencies:
  services:
    neo4j:
      image: "neo4j:5.13"
      ports: ["7687", "7474"]
      environment:
        NEO4J_AUTH: "neo4j/{{.Password}}"
        NEO4J_PLUGINS: '["apoc", "graph-data-science"]'
        NEO4J_dbms_security_procedures_unrestricted: "gds.*,apoc.*"
      volumes:
        - "{{.DataPath}}:/data"
        - "{{.LogsPath}}:/logs"
      healthcheck:
        test: ["CMD-SHELL", "cypher-shell 'RETURN 1'"]
        interval: 30s
        timeout: 10s
        retries: 5
      auto_generate_password: true
      shared: false
    redis:
      image: "redis:7-alpine"
      ports: ["6379"]
      command: ["redis-server", "--appendonly", "yes"]
      volumes:
        - "{{.DataPath}}:/data"
      healthcheck:
        test: ["CMD", "redis-cli", "ping"]
        interval: 15s
        timeout: 5s
        retries: 3
      shared: false

configuration_schema:
  secrets:
    openai_api_key:
      description: "OpenAI API key for embeddings and LLM operations"
      type: "api_key"
      required: true
      validation: "^sk-[a-zA-Z0-9]{20,}$"
      prompt: "Enter your OpenAI API key (starts with sk-)"
      env_var: "OPENAI_API_KEY"
    anthropic_api_key:
      description: "Anthropic API key for Claude operations (optional)"
      type: "api_key"
      required: false
      validation: "^sk-ant-[a-zA-Z0-9-]+$"
      prompt: "Enter your Anthropic API key (optional, starts with sk-ant-)"
      env_var: "ANTHROPIC_API_KEY"
  config:
    embedding_model:
      description: "Model for generating embeddings"
      type: "select"
      required: false
      default: "text-embedding-3-small"
      options: ["text-embedding-3-small", "text-embedding-3-large", "text-embedding-ada-002"]
      env_var: "EMBEDDING_MODEL"
    llm_model:
      description: "Large language model for reasoning"
      type: "select"
      required: false
      default: "gpt-4o-mini"
      options: ["gpt-4o-mini", "gpt-4o", "claude-3-sonnet", "claude-3-opus"]
      env_var: "LLM_MODEL"
    vector_dimensions:
      description: "Vector embedding dimensions"
      type: "integer"
      required: false
      default: 1536
      validation: "^(512|768|1536|3072)$"
      env_var: "VECTOR_DIMENSIONS"
    max_context_length:
      description: "Maximum context length for processing"
      type: "integer"
      required: false
      default: 8192
      env_var: "MAX_CONTEXT_LENGTH"
    enable_temporal_features:
      description: "Enable temporal relationship tracking"
      type: "boolean"
      required: false
      default: true
      env_var: "ENABLE_TEMPORAL_FEATURES"

server:
  transport: "stdio"
  command: "{{.RuntimePath.uv}}"
  args:
    - "run"
    - "--directory={{.InstallPath}}"
    - "knowledge_graph_server.py"
    - "--transport={{.Transport}}"
    - "--log-level=info"
  environment:
    # Service connections
    NEO4J_URI: "{{.Services.neo4j.uri}}"
    NEO4J_USER: "{{.Services.neo4j.user}}"
    NEO4J_PASSWORD: "{{.Services.neo4j.password}}"
    REDIS_URL: "{{.Services.redis.uri}}"
    # User secrets
    OPENAI_API_KEY: "{{.Secrets.openai_api_key}}"
    ANTHROPIC_API_KEY: "{{.Secrets.anthropic_api_key}}"
    # User configuration
    EMBEDDING_MODEL: "{{.Config.embedding_model}}"
    LLM_MODEL: "{{.Config.llm_model}}"
    VECTOR_DIMENSIONS: "{{.Config.vector_dimensions}}"
    MAX_CONTEXT_LENGTH: "{{.Config.max_context_length}}"
    ENABLE_TEMPORAL_FEATURES: "{{.Config.enable_temporal_features}}"
  working_directory: "{{.InstallPath}}"
  timeout: "45s"

clients:
  recommended: ["vscode", "claude-code", "cursor"]
  tested: ["vscode", "claude-code", "cursor", "crewai"]
  requirements:
    vscode:
      minimum_version: "1.85.0"
      features: ["mcp-support"]
    claude-code:
      minimum_version: "1.0.0"
      features: ["mcp-support"]

documentation:
  readme: "README.md"
  changelog: "CHANGELOG.md"
  examples:
    - name: "Basic Knowledge Graph"
      description: "Simple knowledge graph setup for personal use"
      file: "examples/basic.md"
    - name: "Research Assistant"
      description: "Academic research with paper indexing"
      file: "examples/research.md"
    - name: "Enterprise Setup"
      description: "Large-scale deployment configuration"
      file: "examples/enterprise.md"
```

## Migration and Versioning

### Version Compatibility
- Servo maintains backward compatibility within major versions
- Minor version updates may add new optional fields
- Breaking changes require major version increment

### Migration Path
When updating `.servo` files:
1. Check current `servo_version`
2. Review new fields and options
3. Test with `servo validate`
4. Update incrementally

### Deprecation Policy
- Features marked deprecated in documentation
- Minimum 6 months notice before removal
- Migration guides provided for breaking changes