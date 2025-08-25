# Custom Configuration Guide

## Overview

Servo allows you to customize the generated `docker-compose.yml` and `devcontainer.json` files by placing override configuration files in your project. This enables you to:

- Add custom services to your development environment
- Modify existing service configurations
- Add development tools and features
- Customize VS Code settings and extensions
- Override ports, volumes, and environment variables

## Configuration Override Locations

### Project-Level Overrides
Apply to **all sessions** in the project:
```
.servo/config/
├── docker-compose.yml    # Docker Compose overrides
└── devcontainer.json     # Devcontainer overrides
```

### Session-Level Overrides  
Apply to **specific sessions** (higher precedence):
```
.servo/sessions/<session-name>/config/
├── docker-compose.yml    # Session-specific Docker Compose overrides  
└── devcontainer.json     # Session-specific devcontainer overrides
```

## Override Precedence

Configuration merging follows this precedence (highest wins):

1. **Session-level overrides** (`.servo/sessions/<session>/config/`)
2. **Project-level overrides** (`.servo/config/`)  
3. **Generated base configuration** (from .servo manifests)

## Docker Compose Customization

### Adding Custom Services

Create `.servo/config/docker-compose.yml` or `.servo/sessions/<session>/config/docker-compose.yml`:

```yaml
services:
  # Add a custom database for development
  postgres:
    image: postgres:15
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: myapp_dev
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: devpass
    volumes:
      - postgres_data:/var/lib/postgresql/data

  # Add Redis for caching
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

### Modifying Existing Services

Override specific properties of generated services:

```yaml
services:
  # Modify an existing service from a .servo manifest
  neo4j:
    environment:
      # Override the auto-generated password
      NEO4J_AUTH: "neo4j/my_custom_password"
      # Add additional environment variables
      NEO4J_PLUGINS: '["apoc"]'
    ports:
      # Change port mapping
      - "7687:7687"
      - "7474:7474" 
      - "7473:7473"  # Additional port
```

### Adding Networks

```yaml
networks:
  custom-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16

services:
  my-service:
    networks:
      - custom-network
```

## Devcontainer Customization

### Adding VS Code Extensions and Settings

Create `.servo/config/devcontainer.json`:

```json
{
  "features": {
    "ghcr.io/devcontainers/features/docker-outside-of-docker:1": {
      "version": "latest"
    },
    "ghcr.io/devcontainers/features/node:1": {
      "nodeGypDependencies": true,
      "version": "18"
    }
  },
  "customizations": {
    "vscode": {
      "settings": {
        "python.defaultInterpreterPath": "/usr/local/bin/python",
        "python.linting.enabled": true,
        "python.linting.pylintEnabled": true,
        "terminal.integrated.shell.linux": "/bin/bash"
      },
      "extensions": [
        "ms-python.python",
        "ms-python.pylint",
        "ms-vscode.vscode-json",
        "GitHub.copilot"
      ]
    }
  },
  "forwardPorts": [8000, 9000],
  "postCreateCommand": "pip install -r requirements-dev.txt"
}
```

### Adding Development Tools

```json
{
  "features": {
    "ghcr.io/devcontainers/features/git:1": {},
    "ghcr.io/devcontainers/features/github-cli:1": {},
    "ghcr.io/devcontainers/features/docker-outside-of-docker:1": {
      "version": "latest",
      "enableNonRootDocker": "true"
    }
  },
  "mounts": [
    "source=${localWorkspaceFolder}/.git,target=/workspace/.git,type=bind,consistency=cached"
  ]
}
```

## Session-Specific Customization

Use session-level overrides for environment-specific configurations:

### Development Session Override
`.servo/sessions/development/config/docker-compose.yml`:
```yaml
services:
  postgres:
    environment:
      POSTGRES_DB: myapp_development
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: devpass
    ports:
      - "5432:5432"
```

### Staging Session Override  
`.servo/sessions/staging/config/docker-compose.yml`:
```yaml
services:
  postgres:
    environment:
      POSTGRES_DB: myapp_staging
      POSTGRES_USER: staging_user
      POSTGRES_PASSWORD: staging_pass
    ports:
      - "5433:5432"  # Different port to avoid conflicts
```

## Complete Example

Here's a complete example showing project-level and session-specific customizations:

### Project-Level Override (`.servo/config/docker-compose.yml`)
```yaml
services:
  # Add Redis to all sessions
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  # Modify neo4j for all sessions
  neo4j:
    environment:
      NEO4J_PLUGINS: '["apoc", "graph-data-science"]'

volumes:
  redis_data:
```

### Session-Level Override (`.servo/sessions/development/config/docker-compose.yml`) 
```yaml
services:
  # Development-specific database
  postgres:
    image: postgres:15
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: myapp_dev
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: devpass
    volumes:
      - dev_postgres_data:/var/lib/postgresql/data

  # Override Redis port for development
  redis:
    ports:
      - "6380:6379"  # Different port for development

volumes:
  dev_postgres_data:
```

### Project-Level Devcontainer Override (`.servo/config/devcontainer.json`)
```json
{
  "features": {
    "ghcr.io/devcontainers/features/git:1": {},
    "ghcr.io/devcontainers/features/github-cli:1": {}
  },
  "customizations": {
    "vscode": {
      "settings": {
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
          "source.organizeImports": true
        }
      },
      "extensions": [
        "ms-python.python",
        "ms-vscode.vscode-json"
      ]
    }
  }
}
```

## Best Practices

### File Organization
- Use **project-level overrides** for configurations common to all environments
- Use **session-level overrides** for environment-specific settings
- Keep override files minimal - only specify what you need to change

### Service Naming
- Avoid conflicts with generated service names from .servo manifests
- Use descriptive names for custom services
- Prefix custom services with your project name if needed

### Port Management  
- Use different port ranges for different sessions to avoid conflicts
- Document port usage in your project README
- Consider using port ranges (e.g., 8000-8099 for development, 8100-8199 for staging)

### Volume Management
- Use named volumes for persistent data
- Session-specific volumes are stored in `.servo/sessions/<session>/volumes/`
- Include volume names in `.gitignore` to avoid committing data

## Generated Configuration

After creating override files, run:

```bash
servo work <session-name>
```

Servo will:
1. Load base configuration from .servo manifests
2. Apply project-level overrides from `.servo/config/`
3. Apply session-level overrides from `.servo/sessions/<session>/config/`
4. Generate final configurations in `.devcontainer/`

The final generated files will contain your customizations merged with the base infrastructure requirements.
