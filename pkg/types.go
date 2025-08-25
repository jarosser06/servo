package pkg

import (
	"errors"
	"time"

	"gopkg.in/yaml.v3"
)

// Common errors
var (
	ErrInvalidScope   = errors.New("invalid scope")
	ErrConfigNotFound = errors.New("configuration not found")
)

// ServoDefinition represents a complete .servo file
type ServoDefinition struct {
	ServoVersion        string                        `yaml:"servo_version" json:"servo_version"`
	Name                string                        `yaml:"name" json:"name"`
	Version             string                        `yaml:"version,omitempty" json:"version,omitempty"`
	Description         string                        `yaml:"description,omitempty" json:"description,omitempty"`
	Author              string                        `yaml:"author,omitempty" json:"author,omitempty"`
	License             string                        `yaml:"license,omitempty" json:"license,omitempty"`
	Metadata            *Metadata                     `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	Requirements        *Requirements                 `yaml:"requirements,omitempty" json:"requirements,omitempty"`
	Install             Install                       `yaml:"install" json:"install"`
	Dependencies        *Dependencies                 `yaml:"dependencies,omitempty" json:"dependencies,omitempty"`
	ConfigurationSchema *ConfigurationSchema          `yaml:"configuration_schema,omitempty" json:"configuration_schema,omitempty"`
	Server              Server                        `yaml:"server" json:"server"`
	Services            map[string]*ServiceDependency `yaml:"services,omitempty" json:"services,omitempty"`
	Clients             *ClientInfo                   `yaml:"clients,omitempty" json:"clients,omitempty"`
	Documentation       *Documentation                `yaml:"documentation,omitempty" json:"documentation,omitempty"`
}

// Metadata contains optional package metadata
type Metadata struct {
	Homepage   string   `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	Repository string   `yaml:"repository,omitempty" json:"repository,omitempty"`
	Tags       []string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// Requirements defines system and runtime requirements
type Requirements struct {
	System   []SystemRequirement  `yaml:"system,omitempty" json:"system,omitempty"`
	Runtimes []RuntimeRequirement `yaml:"runtimes,omitempty" json:"runtimes,omitempty"`
}

// SystemRequirement defines a system-level dependency
type SystemRequirement struct {
	Name         string            `yaml:"name" json:"name"`
	Description  string            `yaml:"description" json:"description"`
	CheckCommand string            `yaml:"check_command" json:"check_command"`
	InstallHint  string            `yaml:"install_hint" json:"install_hint"`
	Platforms    map[string]string `yaml:"platforms,omitempty" json:"platforms,omitempty"`
}

// RuntimeRequirement defines a runtime dependency
type RuntimeRequirement struct {
	Name    string `yaml:"name" json:"name"`
	Version string `yaml:"version" json:"version"`
}

// Install defines installation method and commands
type Install struct {
	Type          string   `yaml:"type" json:"type"`
	Method        string   `yaml:"method" json:"method"`
	Repository    string   `yaml:"repository,omitempty" json:"repository,omitempty"`
	Subdirectory  string   `yaml:"subdirectory,omitempty" json:"subdirectory,omitempty"`
	SetupCommands []string `yaml:"setup_commands" json:"setup_commands"`
	BuildCommands []string `yaml:"build_commands,omitempty" json:"build_commands,omitempty"`
	TestCommands  []string `yaml:"test_commands,omitempty" json:"test_commands,omitempty"`
}

// Dependencies defines service dependencies
type Dependencies struct {
	Services map[string]ServiceDependency `yaml:"services,omitempty" json:"services,omitempty"`
}

// ServiceDependency defines a Docker service dependency
type ServiceDependency struct {
	Image                string            `yaml:"image" json:"image"`
	Ports                []string          `yaml:"ports,omitempty" json:"ports,omitempty"`
	Environment          map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	Volumes              []string          `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Command              []string          `yaml:"command,omitempty" json:"command,omitempty"`
	HealthCheck          *HealthCheck      `yaml:"healthcheck,omitempty" json:"healthcheck,omitempty"`
	AutoGeneratePassword bool              `yaml:"auto_generate_password,omitempty" json:"auto_generate_password,omitempty"`
	Shared               bool              `yaml:"shared,omitempty" json:"shared,omitempty"`
}

// HealthCheck defines service health check configuration
type HealthCheck struct {
	Test     []string `yaml:"test" json:"test"`
	Interval string   `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout  string   `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retries  int      `yaml:"retries,omitempty" json:"retries,omitempty"`
}

// ConfigurationSchema defines interactive configuration options
type ConfigurationSchema struct {
	Secrets map[string]SecretSchema `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	Config  map[string]ConfigSchema `yaml:"config,omitempty" json:"config,omitempty"`
}

// SecretSchema defines a secret configuration field
type SecretSchema struct {
	Description string `yaml:"description" json:"description"`
	Type        string `yaml:"type" json:"type"`
	Required    bool   `yaml:"required" json:"required"`
	Validation  string `yaml:"validation,omitempty" json:"validation,omitempty"`
	Prompt      string `yaml:"prompt,omitempty" json:"prompt,omitempty"`
	EnvVar      string `yaml:"env_var" json:"env_var"`
}

// ConfigSchema defines a configuration field
type ConfigSchema struct {
	Description string        `yaml:"description" json:"description"`
	Type        string        `yaml:"type" json:"type"`
	Required    bool          `yaml:"required,omitempty" json:"required,omitempty"`
	Default     interface{}   `yaml:"default,omitempty" json:"default,omitempty"`
	Options     []interface{} `yaml:"options,omitempty" json:"options,omitempty"`
	Validation  string        `yaml:"validation,omitempty" json:"validation,omitempty"`
	EnvVar      string        `yaml:"env_var" json:"env_var"`
}

// Server defines server execution configuration
type Server struct {
	Transport        string            `yaml:"transport" json:"transport"`
	Command          string            `yaml:"command" json:"command"`
	Args             []string          `yaml:"args" json:"args"`
	Environment      map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`
	WorkingDirectory string            `yaml:"working_directory,omitempty" json:"working_directory,omitempty"`
	Timeout          string            `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// ClientInfo contains client compatibility information
type ClientInfo struct {
	Recommended  []string                     `yaml:"recommended,omitempty" json:"recommended,omitempty"`
	Tested       []string                     `yaml:"tested,omitempty" json:"tested,omitempty"`
	Excluded     []string                     `yaml:"excluded,omitempty" json:"excluded,omitempty"`
	Requirements map[string]ClientRequirement `yaml:"requirements,omitempty" json:"requirements,omitempty"`
}

// ClientRequirement defines client-specific requirements
type ClientRequirement struct {
	MinimumVersion string   `yaml:"minimum_version,omitempty" json:"minimum_version,omitempty"`
	Features       []string `yaml:"features,omitempty" json:"features,omitempty"`
}

// Documentation contains documentation links
type Documentation struct {
	Readme    string    `yaml:"readme,omitempty" json:"readme,omitempty"`
	Changelog string    `yaml:"changelog,omitempty" json:"changelog,omitempty"`
	Examples  []Example `yaml:"examples,omitempty" json:"examples,omitempty"`
}

// Example defines a documentation example
type Example struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	File        string `yaml:"file" json:"file"`
}

// Configuration represents Servo's configuration
type Configuration struct {
	Version     string        `yaml:"version" json:"version"`
	Scope       string        `yaml:"scope" json:"scope"`
	SessionName string        `yaml:"session_name,omitempty" json:"session_name,omitempty"`
	Settings    Settings      `yaml:"settings" json:"settings"`
	Sync        *SyncConfig   `yaml:"sync,omitempty" json:"sync,omitempty"`
	Clients     ClientConfigs `yaml:"clients" json:"clients"`
	Servers     ServerConfigs `yaml:"servers" json:"servers"`
}

// Settings contains global settings
type Settings struct {
	AutoStartServices   bool   `yaml:"auto_start_services" json:"auto_start_services"`
	InheritGlobal       bool   `yaml:"inherit_global,omitempty" json:"inherit_global,omitempty"`
	LogRetentionDays    int    `yaml:"log_retention_days" json:"log_retention_days"`
	UpdateCheckInterval string `yaml:"update_check_interval" json:"update_check_interval"`
	SecretsEncryption   bool   `yaml:"secrets_encryption" json:"secrets_encryption"`
}

// SyncConfig contains synchronization settings
type SyncConfig struct {
	Enabled        bool   `yaml:"enabled" json:"enabled"`
	Repository     string `yaml:"repository" json:"repository"`
	AutoPush       bool   `yaml:"auto_push" json:"auto_push"`
	ExcludeSecrets bool   `yaml:"exclude_secrets" json:"exclude_secrets"`
}

// ClientConfigs maps client names to their configurations
type ClientConfigs map[string]ClientConfig

// ClientConfig contains client-specific configuration
type ClientConfig struct {
	Enabled       bool     `yaml:"enabled" json:"enabled"`
	AutoConfigure bool     `yaml:"auto_configure" json:"auto_configure"`
	ConfigPath    string   `yaml:"config_path,omitempty" json:"config_path,omitempty"`
	PluginPath    string   `yaml:"plugin_path,omitempty" json:"plugin_path,omitempty"`
	Extensions    []string `yaml:"extensions,omitempty" json:"extensions,omitempty"`
}

// ServerConfigs maps server names to their configurations
type ServerConfigs map[string]ServerConfig

// ServerConfig contains server installation configuration
type ServerConfig struct {
	Version         string                 `yaml:"version" json:"version"`
	Clients         []string               `yaml:"clients" json:"clients"`
	Config          map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty"`
	SecretsRequired []string               `yaml:"secrets_required,omitempty" json:"secrets_required,omitempty"`
}

// MCPConfig represents the configuration for an MCP client
type MCPConfig struct {
	Servers map[string]MCPServerConfig `json:"mcpServers,omitempty" yaml:"mcpServers,omitempty"`
}

// MCPServerConfig represents an individual MCP server configuration
type MCPServerConfig struct {
	Command          string            `json:"command" yaml:"command"`
	Args             []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Environment      map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	WorkingDirectory string            `json:"cwd,omitempty" yaml:"cwd,omitempty"`
}

// ClientScope represents a configuration scope for a client
type ClientScope string

const (
	LocalScope ClientScope = "local"
)

// ServiceConnection contains connection details for a service
type ServiceConnection struct {
	URI      string `json:"uri" yaml:"uri"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
}

// InstallationStatus represents the status of a server installation
type InstallationStatus struct {
	Name        string    `json:"name" yaml:"name"`
	Version     string    `json:"version" yaml:"version"`
	Scope       string    `json:"scope" yaml:"scope"`
	Status      string    `json:"status" yaml:"status"`
	Clients     []string  `json:"clients" yaml:"clients"`
	Services    []string  `json:"services" yaml:"services"`
	InstallPath string    `json:"install_path" yaml:"install_path"`
	InstallTime time.Time `json:"install_time" yaml:"install_time"`
	LastUsed    time.Time `json:"last_used" yaml:"last_used"`
}

// ServiceStatus represents the status of a service
type ServiceStatus struct {
	Name        string            `json:"name" yaml:"name"`
	Server      string            `json:"server" yaml:"server"`
	Scope       string            `json:"scope" yaml:"scope"`
	Status      string            `json:"status" yaml:"status"`
	Image       string            `json:"image" yaml:"image"`
	Ports       map[string]string `json:"ports" yaml:"ports"`
	Health      string            `json:"health" yaml:"health"`
	LastStarted time.Time         `json:"last_started" yaml:"last_started"`
}

// SyncStatus represents synchronization status
type SyncStatus struct {
	Repository   string    `json:"repository" yaml:"repository"`
	LastSync     time.Time `json:"last_sync" yaml:"last_sync"`
	Status       string    `json:"status" yaml:"status"`
	LocalChanges bool      `json:"local_changes" yaml:"local_changes"`
	RemoteCommit string    `json:"remote_commit" yaml:"remote_commit"`
	LocalCommit  string    `json:"local_commit" yaml:"local_commit"`
}

// Additional types for backward compatibility and testing
type ServiceConfig = ServiceDependency
type SecretConfig = SecretSchema
type ConfigParam = ConfigSchema

// TemplateContext contains variables for template rendering
type TemplateContext struct {
	InstallPath string                       `json:"install_path"`
	Transport   string                       `json:"transport"`
	RuntimePath map[string]string            `json:"runtime_path"`
	Services    map[string]ServiceConnection `json:"services"`
	Secrets     map[string]string            `json:"secrets"`
	Config      map[string]interface{}       `json:"config"`
	Env         map[string]string            `json:"env"`
	DataPath    string                       `json:"data_path"`
	LogsPath    string                       `json:"logs_path"`
	ConfigPath  string                       `json:"config_path"`
}

// ToYAML converts a ServoDefinition to YAML format
func (s *ServoDefinition) ToYAML() (string, error) {
	data, err := yaml.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
