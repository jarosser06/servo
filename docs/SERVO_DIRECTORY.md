# Servo Project Directory Structure

This document provides a comprehensive overview of the `.servo` project directory structure and explains what each component does in the Servo MCP Server Package Manager.

## Overview

Each Servo project contains a `.servo` directory that manages all project-specific configurations, sessions, and MCP server installations. This project-based approach ensures complete isolation and portability of development environments.

## Directory Structure

```
my-project/
├── .servo/                    # Project Servo directory  
│   ├── project.yaml          # Project configuration and server declarations
│   ├── secrets.yaml          # Base64-encoded secrets (local only)
│   ├── .gitignore           # Excludes secrets and volumes
│   ├── config/               # Project-level configuration overrides
│   └── sessions/             # Project sessions
│       ├── default/          # Default session (created automatically)
│       │   ├── manifests/    # Downloaded .servo file definitions
│       │   │   ├── graphiti.servo      # Graphiti server definition
│       │   │   └── playwright.servo    # Playwright server definition
│       │   ├── config/       # Session-specific configuration overrides
│       │   └── volumes/      # Session-specific Docker volumes (gitignored)
│       │       ├── graphiti/ # Graphiti service volumes
│       │       └── playwright/ # Playwright service volumes
│       ├── development/      # Development session
│       │   ├── manifests/    # Development server definitions
│       │   ├── config/       # Development configurations
│       │   └── volumes/      # Development Docker volumes
│       └── production/       # Production session
│           ├── manifests/    # Production server definitions
│           ├── config/       # Production configurations
│           └── volumes/      # Production Docker volumes
├── .vscode/                  # VS Code configuration
│   └── mcp.json             # VS Code MCP server configuration
├── .devcontainer/           # Generated development container
│   ├── devcontainer.json    # Dev container configuration with runtime features
│   └── docker-compose.yml   # Service dependencies
└── .mcp.json                # Claude Code MCP configuration
```

## Component Details

### Project Configuration (`.servo/project.yaml`)

The main project configuration file containing project metadata and MCP server declarations:

```yaml
# From the Project type in internal/project/manager.go
clients: ["vscode", "claude-code", "cursor"]
default_session: default
active_session: development
mcp_servers:
  - name: "graphiti"
    source: "https://github.com/getzep/graphiti.git"
    clients: ["vscode", "claude-code"]
    sessions: ["default", "development"]
  - name: "playwright"
    source: "./playwright-server.servo"
    clients: ["vscode", "claude-code"]
    sessions: ["default"]
required_secrets:
  - name: "openai_api_key"
    description: "OpenAI API key for embeddings"
  - name: "neo4j_password"
    description: "Neo4j database password"
```

### Base64-Encoded Secrets (`.servo/secrets.yaml`)

Simple base64-encoded secrets file (never synchronized):

```yaml
# Base64-encoded secrets for obscurity (not security)
version: "1.0"
secrets:
  openai_api_key: "c2stMTIzNDU2Nzg5MGFiY2RlZg=="  # base64 encoded
  neo4j_password: "bXlfc2VjdXJlX3Bhc3N3b3Jk"      # base64 encoded
```

### Project Sessions (`.servo/sessions/`)

Each session provides isolated environments for different workflows:

#### Session Directory Structure
```
sessions/
├── default/                  # Default session (created automatically)
│   ├── manifests/           # Downloaded .servo file definitions
│   │   ├── graphiti.servo   # Complete Graphiti server definition
│   │   └── playwright.servo # Complete Playwright server definition
│   ├── config/              # Session-specific configuration overrides
│   └── volumes/             # Session-specific Docker volumes (gitignored)
│       ├── graphiti/        # Graphiti service volumes
│       │   ├── neo4j/       # Neo4j data
│       │   └── logs/        # Service logs
│       └── playwright/      # Playwright service volumes
├── development/             # Development session
│   ├── manifests/           # Development server definitions
│   ├── config/              # Development configurations
│   └── volumes/             # Development Docker volumes
└── production/              # Production session
    ├── manifests/           # Production server definitions
    ├── config/              # Production configurations
    └── volumes/             # Production Docker volumes
```

### Client Configurations

MCP client configurations are automatically generated and stored in the project root:

#### VS Code Configuration (`.vscode/mcp.json`)
```json
{
  "mcpServers": {
    "graphiti": {
      "command": "uv",
      "args": ["run", "python", "-m", "graphiti.mcp_server"],
      "cwd": "/workspace",
      "env": {
        "NEO4J_URI": "bolt://localhost:7687",
        "NEO4J_USER": "neo4j",
        "NEO4J_PASSWORD": "auto-generated-password",
        "OPENAI_API_KEY": "user-secret-placeholder"
      }
    }
  }
}
```

#### Claude Code Configuration (`.mcp.json`)
```json
{
  "mcpServers": {
    "graphiti": {
      "command": "uv",
      "args": ["run", "python", "-m", "graphiti.mcp_server"],
      "cwd": "/workspace",
      "env": {
        "NEO4J_URI": "bolt://localhost:7687",
        "NEO4J_USER": "neo4j", 
        "NEO4J_PASSWORD": "auto-generated-password",
        "OPENAI_API_KEY": "user-secret-placeholder"
      }
    }
  }
}
```

### Development Container (`.devcontainer/`)

Automatically generated devcontainer configuration based on MCP server runtime requirements:

#### devcontainer.json
```json
{
  "name": "my-project-devcontainer",
  "dockerComposeFile": "docker-compose.yml", 
  "service": "devcontainer",
  "workspaceFolder": "/workspace",
  "features": {
    "ghcr.io/devcontainers/features/common-utils:2": {},
    "ghcr.io/devcontainers/features/python:1": {
      "version": "3.11",
      "installTools": true
    },
    "ghcr.io/devcontainers/features/docker-in-docker:2": {}
  },
  "customizations": {
    "vscode": {
      "extensions": ["ms-python.python"]
    }
  }
}
```

#### docker-compose.yml
```yaml
version: '3.8'
services:
  devcontainer:
    build:
      context: .
      dockerfile: Dockerfile
    volumes:
      - ../..:/workspaces:cached
    environment:
      - NEO4J_URI=bolt://neo4j:7687
      - NEO4J_USER=neo4j
      - NEO4J_PASSWORD=auto-generated-password
    depends_on:
      - neo4j
      
  neo4j:
    image: neo4j:5.13
    environment:
      NEO4J_AUTH: neo4j/auto-generated-password
      NEO4J_PLUGINS: '["apoc"]'
    ports:
      - "7687:7687"
      - "7474:7474"
    volumes:
      - neo4j_data:/data
      
volumes:
  neo4j_data:
```

## Session Management

### Default Session

Every project automatically gets a "default" session when initialized:
```bash
servo init my-project  # Creates default session
```

### Custom Sessions

Create projects with custom default sessions:
```bash
servo init my-project --session development
```

### Session-Specific Installation

Install MCP servers to specific sessions:
```bash
servo install graphiti.servo --session production
servo install dev-tools.servo --session development
```

### Session Isolation

Each session maintains complete isolation:
- **Separate server manifests** in `sessions/<name>/manifests/` - each session has its own .servo files
- **Isolated Docker volumes** in `sessions/<name>/volumes/` - session-specific service data
- **Independent configurations** in `sessions/<name>/config/` - session-specific overrides
- **Session-aware client configs** - generated configurations reference active session

## Project Portability

The project-based architecture ensures complete portability:
- **Self-contained**: All configurations and servers within project
- **Version controlled**: `.servo/project.yaml` tracks MCP server declarations
- **Reproducible**: Same setup across different environments
- **Isolated**: No global state or dependencies

## File Patterns

### Generated Files (do not edit directly)
- `.vscode/mcp.json` - Generated VS Code MCP configuration
- `.mcp.json` - Generated Claude Code configuration  
- `.devcontainer/devcontainer.json` - Generated devcontainer config with runtime features
- `.devcontainer/docker-compose.yml` - Generated Docker Compose with services

### Version Controlled Files
- `.servo/project.yaml` - Project metadata and server declarations (commit this)
- `.servo/sessions/<name>/manifests/` - Downloaded .servo files (commit these)
- `.servo/.gitignore` - Excludes secrets and volumes (commit this)

### Local Only Files (never commit)
- `.servo/secrets.yaml` - Base64-encoded secrets (automatically gitignored)
- `.servo/sessions/<name>/volumes/` - Docker volumes and service data (gitignored)

### Session Data Structure

Each session contains:
```
sessions/<session-name>/
├── manifests/           # Server definitions (.servo files)
│   ├── server1.servo   # Complete server definition
│   └── server2.servo   # Complete server definition
├── config/             # Session-specific config overrides
│   └── custom.yaml     # Optional custom configurations
└── volumes/            # Docker service volumes (gitignored)
    ├── server1/        # Server 1 volumes
    │   ├── data/       # Service data
    │   └── logs/       # Service logs
    └── server2/        # Server 2 volumes
```

## GitIgnore Strategy

The `.servo/.gitignore` file automatically excludes:
```gitignore
# Servo project files  
secrets.yaml
*.log

# Legacy structure (for backwards compatibility)
volumes/

# Session-specific files
sessions/*/config/
sessions/*/logs/

# OS files
.DS_Store
Thumbs.db
```

This ensures:
- ✅ **Project configuration is synced** (`.servo/project.yaml`)  
- ✅ **Server definitions are synced** (`.servo/sessions/*/manifests/*.servo`)
- ❌ **Secrets are never synced** (`.servo/secrets.yaml`)
- ❌ **Docker volumes are never synced** (`.servo/sessions/*/volumes/`)