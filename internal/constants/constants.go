package constants

// File Extensions
const (
	ExtServo      = ".servo"
	ExtYAML       = ".yaml" 
	ExtYML        = ".yml"
	ExtJSON       = ".json"
	ExtMarkdown   = ".md"
)

// Directory Names
const (
	// Project structure
	DirServo         = ".servo"
	DirSessions      = "sessions"
	DirManifests     = "manifests"
	DirConfig        = "config"
	DirVolumes       = "volumes"
	DirLogs          = "logs"
	
	// Client directories
	DirVSCode        = ".vscode"
	DirCursor        = ".cursor"
	DirDevcontainer  = ".devcontainer"
)

// File Names
const (
	// Project files
	FileProjectConfig     = "project.yaml"
	FileSessionConfig     = "session.yaml"
	FileActiveSession     = "active_session"
	FileGitignore         = ".gitignore"
	FileSecretsConfig     = "secrets.yaml"
	FileEnvironmentConfig = "environment.yaml"
	
	// Configuration files
	FileMCPConfig         = "mcp.json"
	FileDevcontainer      = "devcontainer.json"
	FileDockerCompose     = "docker-compose.yml"
	FileDockerComposeYAML = "docker-compose.yaml"
)

// Client Names
const (
	ClientVSCode     = "vscode"
	ClientClaudeCode = "claude-code"
	ClientCursor     = "cursor"
	ClientClaudeApp  = "claude-desktop"
)

// Supported client list
var SupportedClients = []string{
	ClientVSCode,
	ClientClaudeCode,
	ClientCursor,
	ClientClaudeApp,
}

// Transport Types
const (
	TransportStdio = "stdio"
	TransportHTTP  = "http"
	TransportSSE   = "sse"
)

// Install Methods
const (
	InstallMethodGit    = "git"
	InstallMethodLocal  = "local"
	InstallMethodDocker = "docker"
	InstallMethodURL    = "url"
)

// Install Types
const (
	InstallTypeGit    = "git"
	InstallTypeLocal  = "local"
	InstallTypeDocker = "docker"
)

// Environment Names
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTesting     = "testing"
)

// Default Values
const (
	DefaultSessionName    = "default"
	DefaultServoVersion   = "1.0"
	DefaultTransport      = TransportStdio
	DefaultInstallMethod  = InstallMethodGit
)

// YAML Field Names
const (
	FieldName              = "name"
	FieldVersion           = "version"
	FieldDescription       = "description"
	FieldAuthor            = "author"
	FieldLicense           = "license"
	FieldServoVersion      = "servo_version"
	FieldInstall           = "install"
	FieldServer            = "server"
	FieldClients           = "clients"
	FieldRequirements      = "requirements"
	FieldRuntimes          = "runtimes"
	FieldServices          = "services"
	FieldDependencies      = "dependencies"
	FieldConfiguration     = "configuration_schema"
	FieldSecrets           = "secrets"
	FieldEnvironment       = "environment"
	FieldTransport         = "transport"
	FieldCommand           = "command"
	FieldArgs              = "args"
	FieldType              = "type"
	FieldMethod            = "method"
	FieldRepository        = "repository"
	FieldSetupCommands     = "setup_commands"
	FieldRecommended       = "recommended"
	FieldTested            = "tested"
	FieldActiveSession     = "active_session"
	FieldDefaultSession    = "default_session"
	FieldImage             = "image"
	FieldPorts             = "ports"
	FieldVolumes           = "volumes"
	FieldHealthcheck       = "healthcheck"
	FieldDependsOn         = "depends_on"
)

// Configuration Keys
const (
	ConfigMCPServers = "mcpServers"
	ConfigServers    = "servers"
)

// Secret-related constants
const (
	SecretPrefix    = "SERVO_SECRET:"
	SecretPlaceholder = "{{SERVO_SECRET:%s}}"
)

// Docker-related constants
const (
	DockerComposeVersion = "3.8"
	DockerNetwork        = "servo-network"
)

// Command-related constants
const (
	CommandInit      = "init"
	CommandInstall   = "install"
	CommandConfigure = "configure"
	CommandWork      = "work"
	CommandStatus    = "status"
	CommandSession   = "session"
	CommandSecrets   = "secrets"
	CommandValidate  = "validate"
	CommandEnv       = "env"
)

// Validation constants
const (
	MinNameLength = 1
	MaxNameLength = 255
	MinVersionLength = 1
	MaxDescriptionLength = 1000
)

// Regex patterns as constants
const (
	PatternSessionName = `^[a-zA-Z0-9_-]+$`
	PatternVersion     = `^[0-9]+\.[0-9]+(\.[0-9]+)?(-[a-zA-Z0-9.-]+)?(\+[a-zA-Z0-9.-]+)?$`
	PatternAPIKey      = `^sk-[a-zA-Z0-9]{20,}$`
)

// Error messages
const (
	ErrProjectNotFound      = "not in a servo project directory"
	ErrSessionNotFound      = "session not found"
	ErrSessionExists        = "session already exists"
	ErrServerExists         = "server already exists"
	ErrInvalidName          = "invalid name"
	ErrRequiredField        = "required field missing"
	ErrPermissionDenied     = "permission denied"
	ErrFileNotFound         = "file not found"
	ErrInvalidConfiguration = "invalid configuration"
)

// Platform constants
const (
	PlatformDarwin  = "darwin"
	PlatformLinux   = "linux"
	PlatformWindows = "windows"
)

// Common file permissions
const (
	FilePermReadWrite = 0644
	FilePermExecute   = 0755
	DirPermDefault    = 0755
)

// Environment variable names
const (
	EnvServoDir              = "SERVO_DIR"
	EnvServoMasterPassword   = "SERVO_MASTER_PASSWORD"
	EnvServoNonInteractive   = "SERVO_NON_INTERACTIVE"
	EnvNodeEnv               = "NODE_ENV"
	EnvAPIKey                = "API_KEY"
	EnvDatabaseURL           = "DATABASE_URL"
	EnvLogLevel              = "LOG_LEVEL"
)

// URL and repository patterns
const (
	GitHubDomain   = "github.com"
	GitHubRawDomain = "raw.githubusercontent.com"
	HTTPSPrefix    = "https://"
	HTTPPrefix     = "http://"
	GitPrefix      = "git://"
	SSHPrefix      = "git@"
)

// Timeouts and limits
const (
	DefaultTimeout    = 30 // seconds
	MaxRetries       = 3
	RetryDelay       = 1 // seconds
	MaxFileSize      = 10 * 1024 * 1024 // 10MB
	MaxManifestSize  = 1024 * 1024      // 1MB
)

// Session states
const (
	SessionStateActive   = "active"
	SessionStateInactive = "inactive"
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// Common status messages
const (
	StatusReady       = "ready"
	StatusStarting    = "starting"
	StatusRunning     = "running"
	StatusStopped     = "stopped"
	StatusError       = "error"
)

// Help text constants
const (
	HelpUsagePrefix = "Usage: "
	HelpExamplePrefix = "Example: "
	HelpOptionsHeader = "Options:"
	HelpCommandsHeader = "Commands:"
)

// Build information (can be set via ldflags)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)