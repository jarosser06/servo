package pkg

import (
	"context"
	"io"
)

// Client represents an MCP (Model Context Protocol) client plugin that integrates
// with various code editors and development environments.
//
// Clients handle MCP server configuration, lifecycle management, and platform-specific
// integration details. Each client implementation should support local configuration
// scoping and provide devcontainer integration capabilities.
type Client interface {
	// Name returns the unique identifier for this client (e.g., "vscode", "cursor")
	Name() string

	// Description returns a human-readable description of the client
	Description() string

	// SupportedPlatforms returns the list of operating systems this client supports
	// Platform names should follow Go's GOOS format: "darwin", "linux", "windows"
	SupportedPlatforms() []string

	// IsPlatformSupported returns true if the client is supported on the current platform
	IsPlatformSupported() bool

	// IsInstalled checks if the client application is installed and accessible
	IsInstalled() bool

	// GetVersion retrieves the installed version of the client application
	// Returns an error if the client is not installed or version cannot be determined
	GetVersion() (string, error)

	// GetSupportedScopes returns the configuration scopes this client supports
	// Currently only LocalScope is supported across all clients
	GetSupportedScopes() []ClientScope

	// ValidateScope ensures the provided scope is supported by this client
	// Returns an error if the scope is not supported
	ValidateScope(scope string) error

	// GetCurrentConfig retrieves the current MCP configuration for the specified scope
	// Returns a default empty configuration if no config file exists
	GetCurrentConfig(scope string) (*MCPConfig, error)

	// ValidateConfig validates the syntax and structure of the configuration file
	// Returns an error if the configuration is invalid or cannot be read
	ValidateConfig(scope string) error

	// RemoveServer removes a named MCP server from the client configuration
	RemoveServer(scope string, serverName string) error

	// ListServers returns the names of all configured MCP servers for the given scope
	ListServers(scope string) ([]string, error)

	// GenerateConfig creates complete MCP configuration from Servo manifests
	// The secretsProvider function is called to resolve secret placeholders in manifests
	// Configuration is written directly to the client's config files
	GenerateConfig(manifests []ServoDefinition, secretsProvider func(string) (string, error)) error

	// RequiresRestart returns true if the client must be restarted to apply config changes
	RequiresRestart() bool

	// TriggerReload attempts to hot-reload the client configuration without restart
	// Returns an error if hot-reload is not supported or fails
	TriggerReload() error

	// GetLaunchCommand returns the shell command to launch this client with a project
	// The projectPath parameter should be properly quoted for shell execution
	GetLaunchCommand(projectPath string) string

	// SupportsDevcontainers returns true if this client has devcontainer support
	SupportsDevcontainers() bool
}

// ClientRegistry manages available client plugins and provides discovery capabilities.
//
// The registry maintains a collection of registered MCP clients and supports
// dynamic client detection based on installed applications.
type ClientRegistry interface {
	// Register adds a new client to the registry
	// Returns an error if a client with the same name is already registered
	Register(client Client) error

	// Unregister removes a client from the registry by name
	// Returns an error if the client is not found
	Unregister(name string) error

	// Get retrieves a registered client by name
	// Returns an error if the client is not found
	Get(name string) (Client, error)

	// List returns all registered clients
	List() []Client

	// Detect returns a list of clients that are both registered and installed
	// This performs installation detection on all registered clients
	Detect() ([]Client, error)
}

// InstallMethod defines how to install MCP servers from different source types.
//
// Each method handles a specific source format (e.g., git repositories, local files,
// npm packages) and manages the installation process including dependency resolution.
type InstallMethod interface {
	// Name returns the unique identifier for this installation method
	Name() string

	// CanInstall determines if this method can handle the given source
	// Sources can be URLs, file paths, package names, etc.
	CanInstall(source string) bool

	// Install performs the actual installation of the MCP server
	// The servoFile contains manifest details, installPath is the target directory
	Install(ctx context.Context, servoFile ServoDefinition, installPath string) error

	// Validate checks if the servo file is compatible with this installation method
	// Should be called before attempting installation
	Validate(servoFile ServoDefinition) error

	// GetRequiredTools returns a list of external tools needed for this install method
	// Used for pre-installation validation and user guidance
	GetRequiredTools() []string
}

// InstallMethodRegistry manages and routes to appropriate installation methods.
//
// The registry automatically selects the correct installation method based on
// source format and maintains a collection of available methods.
type InstallMethodRegistry interface {
	// Register adds a new installation method to the registry
	Register(method InstallMethod) error

	// GetMethod finds the appropriate installation method for a source
	// Returns an error if no compatible method is found
	GetMethod(source string) (InstallMethod, error)

	// List returns all registered installation methods
	List() []InstallMethod
}

// ConfigManager handles configuration persistence and retrieval
type ConfigManager interface {
	Load(scope string) (*Configuration, error)
	Save(scope string, config *Configuration) error
	Get(scope string, key string) (interface{}, error)
	Set(scope string, key string, value interface{}) error
	Delete(scope string, key string) error
	Merge(global *Configuration, local *Configuration) (*Configuration, error)
}

// SecretsManager handles encrypted secret storage
type SecretsManager interface {
	Set(scope string, key string, value string) error
	Get(scope string, key string) (string, error)
	Delete(scope string, key string) error
	List(scope string) ([]string, error)
	RequiredSecrets(serverName string) ([]string, error)
	HasRequired(serverName string) (bool, error)
	Encrypt(scope string) error
	Decrypt(scope string) error
}

// ServiceManager handles Docker service orchestration
type ServiceManager interface {
	CreateService(ctx context.Context, scope string, serverName string, service ServiceDependency) error
	StartService(ctx context.Context, scope string, serviceName string) error
	StopService(ctx context.Context, scope string, serviceName string) error
	RestartService(ctx context.Context, scope string, serviceName string) error
	GetServiceStatus(ctx context.Context, scope string, serviceName string) (*ServiceStatus, error)
	ListServices(ctx context.Context, scope string) ([]ServiceStatus, error)
	GetServiceLogs(ctx context.Context, scope string, serviceName string, follow bool) (io.ReadCloser, error)
	GetServiceConnection(scope string, serviceName string) (*ServiceConnection, error)
}

// ServerManager handles MCP server lifecycle
type ServerManager interface {
	Install(ctx context.Context, source string, options InstallOptions) error
	Uninstall(ctx context.Context, serverName string, scope string) error
	Start(ctx context.Context, serverName string, scope string) error
	Stop(ctx context.Context, serverName string, scope string) error
	Restart(ctx context.Context, serverName string, scope string) error
	List(scope string) ([]InstallationStatus, error)
	GetStatus(serverName string, scope string) (*InstallationStatus, error)
	Update(ctx context.Context, serverName string, scope string) error
	Configure(ctx context.Context, serverName string, scope string) error
}

// SyncManager handles configuration synchronization
type SyncManager interface {
	Init(repository string) error
	Push(ctx context.Context) error
	Pull(ctx context.Context) error
	Status() (*SyncStatus, error)
	Configure(settings SyncConfig) error
}

// TemplateRenderer handles template variable substitution
type TemplateRenderer interface {
	Render(template string, context TemplateContext) (string, error)
	RenderMap(templates map[string]string, context TemplateContext) (map[string]string, error)
	RenderSlice(templates []string, context TemplateContext) ([]string, error)
}

// Validator validates .servo files and configurations
type Validator interface {
	ValidateServoFile(servoFile ServoDefinition) error
	ValidateConfiguration(config Configuration) error
	ValidateInstallation(serverName string, scope string) error
	ValidateSecrets(serverName string, scope string) error
	ValidateServices(scope string, serverName string) error
}

// PackageManager coordinates all package operations
type PackageManager interface {
	Install(ctx context.Context, source string, options InstallOptions) error
	Uninstall(ctx context.Context, serverName string, options UninstallOptions) error
	Update(ctx context.Context, serverName string, options UpdateOptions) error
	List(options ListOptions) ([]InstallationStatus, error)
	Inspect(source string) (*ServoDefinition, error)
	Validate(source string) error
	Configure(ctx context.Context, serverName string, options ConfigureOptions) error
}

// Logger provides structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
}

// InstallOptions contains options for installation
type InstallOptions struct {
	Scope       string   `json:"scope"`
	Clients     []string `json:"clients"`
	DryRun      bool     `json:"dry_run"`
	Force       bool     `json:"force"`
	SkipSecrets bool     `json:"skip_secrets"`
	SkipStart   bool     `json:"skip_start"`
}

// UninstallOptions contains options for uninstallation
type UninstallOptions struct {
	Scope        string `json:"scope"`
	KeepServices bool   `json:"keep_services"`
	KeepSecrets  bool   `json:"keep_secrets"`
	Force        bool   `json:"force"`
}

// UpdateOptions contains options for updates
type UpdateOptions struct {
	Scope   string `json:"scope"`
	Version string `json:"version"`
	Force   bool   `json:"force"`
}

// ListOptions contains options for listing
type ListOptions struct {
	Scope     string `json:"scope"`
	Available bool   `json:"available"`
	Status    string `json:"status"`
}

// ConfigureOptions contains options for configuration
type ConfigureOptions struct {
	Scope       string `json:"scope"`
	Reconfigure bool   `json:"reconfigure"`
	Interactive bool   `json:"interactive"`
}

// RuntimeRequirements represents checked runtime requirements
type RuntimeRequirements struct {
	Docker RuntimeInfo `json:"docker"`
	Git    RuntimeInfo `json:"git"`
	Python RuntimeInfo `json:"python"`
	Node   RuntimeInfo `json:"node"`
	UV     RuntimeInfo `json:"uv"`
	Go     RuntimeInfo `json:"go"`
}

// RuntimeInfo contains information about a runtime
type RuntimeInfo struct {
	Available bool   `json:"available"`
	Version   string `json:"version"`
	Path      string `json:"path"`
	Error     string `json:"error,omitempty"`
}

// RequirementsChecker validates system requirements
type RequirementsChecker interface {
	CheckAll() (*RuntimeRequirements, error)
	CheckRuntime(name string) (*RuntimeInfo, error)
	CheckSystemRequirement(req SystemRequirement) error
	InstallMissing(ctx context.Context) error
}

// PortManager handles port allocation for services
type PortManager interface {
	AllocatePort(preferredPort int) (int, error)
	ReleasePort(port int) error
	IsPortAvailable(port int) bool
	GetAllocatedPorts() []int
}

// DockerComposeGenerator generates docker-compose files
type DockerComposeGenerator interface {
	Generate(scope string, serverName string, services map[string]ServiceDependency) ([]byte, error)
	GenerateFile(scope string, serverName string, services map[string]ServiceDependency, outputPath string) error
}

// CLI represents the command line interface
type CLI interface {
	Execute(args []string) error
	RegisterCommand(cmd Command) error
	GetCommand(name string) (Command, error)
}

// Command represents a CLI command
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(ctx context.Context, args []string) error
	Flags() []Flag
}

// Flag represents a command line flag
type Flag struct {
	Name        string
	Short       string
	Description string
	Required    bool
	Default     interface{}
	Type        string
}

// InteractivePrompter handles user interaction
type InteractivePrompter interface {
	PromptString(message string, defaultValue string) (string, error)
	PromptPassword(message string) (string, error)
	PromptBool(message string, defaultValue bool) (bool, error)
	PromptSelect(message string, options []string, defaultValue string) (string, error)
	PromptMultiSelect(message string, options []string) ([]string, error)
	PromptInt(message string, defaultValue int) (int, error)
	Confirm(message string) (bool, error)
}
