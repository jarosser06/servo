# Generate Servo File from MCP GitHub Repository

You are an expert at analyzing Model Context Protocol (MCP) server repositories and generating `.servo` configuration files. Your task is to examine a GitHub repository containing an MCP server and create a complete, working `.servo` file that follows the specification.

## Analysis Process

### 1. Repository Structure Analysis
Examine the repository and identify:
- **Programming language** (Python, Node.js, Go, etc.)
- **Package manager** (pip, uv, npm, yarn, go mod, etc.)
- **Entry points** (main files, executables, scripts)
- **Dependencies** (requirements.txt, package.json, go.mod, pyproject.toml)
- **Configuration files** (environment variables, config examples)
- **Documentation** (README, setup instructions, examples)

### 2. MCP Server Detection
Look for MCP server implementation patterns:
- Transport method (stdio, SSE, HTTP)
- Server initialization code
- Tool/resource definitions
- Command-line arguments or configuration options

### 3. Dependencies and Services Analysis
Identify external service dependencies:
- Databases (PostgreSQL, MySQL, MongoDB, Neo4j, Redis)
- APIs (OpenAI, Anthropic, external services)
- File systems or storage requirements
- Authentication requirements

## Servo File Generation Rules

### Required Fields
- `servo_version: "1.0"`
- `name`: Derive from repository name (lowercase, hyphens only)
- `install`: Complete installation configuration
- `server`: Server execution configuration

### Installation Configuration
Based on the detected setup:

**For Python projects:**
```yaml
install:
  type: "git"
  method: "git"
  repository: "[GITHUB_URL]"
  setup_commands:
    - "uv sync"  # if pyproject.toml with uv
    - "pip install -e ."  # if setup.py or simple requirements.txt
```

**For Node.js projects:**
```yaml
install:
  type: "git"
  method: "git"
  repository: "[GITHUB_URL]"
  setup_commands:
    - "npm install"  # or "yarn install"
    - "npm run build"  # if build step exists
```

### Server Configuration
Determine the correct server execution:

**Python with uv:**
```yaml
server:
  transport: "stdio"
  command: "{{.RuntimePath.uv}}"
  args:
    - "run"
    - "--directory={{.InstallPath}}"
    - "[MAIN_MODULE_OR_SCRIPT]"
  working_directory: "{{.InstallPath}}"
```

**Python with pip:**
```yaml
server:
  transport: "stdio"
  command: "{{.RuntimePath.python}}"
  args:
    - "-m"
    - "[MODULE_NAME]"
  working_directory: "{{.InstallPath}}"
```

**Node.js:**
```yaml
server:
  transport: "stdio"
  command: "{{.RuntimePath.node}}"
  args:
    - "[MAIN_FILE]"
  working_directory: "{{.InstallPath}}"
```

### Runtime Requirements
Include appropriate runtime requirements:

```yaml
runtime_requirements:
  python:
    version: ">=3.10"  # Adjust based on project requirements
  uv: {}  # If using uv
  # OR
  node:
    version: "^18.0.0"  # Adjust based on package.json engines
```

### Configuration Schema
Analyze environment variables and create configuration schema:

```yaml
configuration_schema:
  secrets:
    api_key_name:
      description: "[Description from code/docs]"
      type: "api_key"
      required: true
      env_var: "[ENV_VAR_NAME]"
  config:
    setting_name:
      description: "[Description]"
      type: "string"  # or integer, boolean, select
      required: false
      default: "[DEFAULT_VALUE]"
      env_var: "[ENV_VAR_NAME]"
```

### Service Dependencies
For detected external services:

```yaml
dependencies:
  services:
    postgres:
      image: "postgres:15"
      ports: ["5432"]
      environment:
        POSTGRES_DB: "{{.Config.database_name}}"
        POSTGRES_USER: "{{.Config.database_user}}"
        POSTGRES_PASSWORD: "{{.Password}}"
      volumes:
        - "{{.DataPath}}:/var/lib/postgresql/data"
      healthcheck:
        test: ["CMD-SHELL", "pg_isready -U {{.Config.database_user}}"]
        interval: 30s
        timeout: 10s
        retries: 3
      auto_generate_password: true
```

## Output Requirements

Generate a complete `.servo` file that includes:

1. **Metadata section** with repository information
2. **Runtime requirements** based on detected language/tools
3. **Installation configuration** with proper setup commands
4. **Server configuration** with correct execution parameters
5. **Configuration schema** for all detected environment variables
6. **Service dependencies** for external services
7. **Client compatibility** information

## Analysis Instructions

Please analyze the provided GitHub repository URL and generate a complete `.servo` file. Follow this process:

1. **Examine the repository structure** and identify the programming language, package manager, and entry points
2. **Find the MCP server implementation** and determine transport method and execution parameters
3. **Identify dependencies** including external services, APIs, and configuration requirements
4. **Extract environment variables** from code, documentation, and example configurations
5. **Generate the complete `.servo` file** with all necessary sections properly configured

Make sure the generated `.servo` file:
- ✅ Follows the exact YAML schema specified
- ✅ Uses correct template variables ({{.InstallPath}}, {{.RuntimePath.python}}, etc.)
- ✅ Includes all detected environment variables in configuration_schema
- ✅ Properly configures service dependencies with health checks
- ✅ Uses appropriate runtime requirements for the detected language
- ✅ Has working installation and server execution commands

Provide the complete `.servo` file as your output, along with a brief explanation of your analysis and any assumptions made.