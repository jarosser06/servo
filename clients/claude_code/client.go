package claude_code

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/servo/servo/internal/client"
	"github.com/servo/servo/internal/utils"
	"github.com/servo/servo/pkg"
)

// CommandExecutor abstracts command execution to enable testing with mocks.
//
// This interface allows the Claude Code client to execute CLI commands while
// providing a clean abstraction for unit testing scenarios.
type CommandExecutor interface {
	// Execute runs a command and returns its output
	Execute(name string, args ...string) ([]byte, error)

	// Run executes a command without capturing output
	Run(name string, args ...string) error
}

// DefaultCommandExecutor implements CommandExecutor using the system's exec functionality.
type DefaultCommandExecutor struct{}

func (e *DefaultCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	return cmd.Output()
}

func (e *DefaultCommandExecutor) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

// Client implements MCP (Model Context Protocol) integration for Claude Code CLI.
//
// This client manages MCP server configurations through the local .mcp.json file
// and integrates with Claude Code's command-line interface for server management.
// It supports hot-reloading and provides a programmatic interface for configuring
// MCP servers in Claude Code environments.
//
// Unlike traditional editor clients, this implementation uses CLI commands to
// interact with the Claude Code application, making it suitable for headless
// and automated deployment scenarios.
type Client struct {
	info     client.BaseClientInfo
	executor CommandExecutor
}

// New creates a new Claude Code client
func New() *Client {
	return NewWithExecutor(&DefaultCommandExecutor{})
}

// NewWithExecutor creates a new Claude Code client with a custom executor
func NewWithExecutor(executor CommandExecutor) *Client {
	return &Client{
		info: client.BaseClientInfo{
			Name:        "claude-code",
			Description: "Claude Code CLI interface",
			Platforms:   utils.DesktopPlatforms,
		},
		executor: executor,
	}
}

// Name returns the client name
func (c *Client) Name() string {
	return c.info.Name
}

// Description returns the client description
func (c *Client) Description() string {
	return c.info.Description
}

// SupportedPlatforms returns the supported platforms
func (c *Client) SupportedPlatforms() []string {
	return c.info.Platforms
}

// IsPlatformSupported checks if the current platform is supported
func (c *Client) IsPlatformSupported() bool {
	return utils.IsPlatformSupported(c.info.Platforms)
}

// IsInstalled checks if Claude Code is installed
func (c *Client) IsInstalled() bool {
	if !c.IsPlatformSupported() {
		return false
	}

	return client.ExecutableExists("claude")
}

// GetVersion returns the Claude Code version
func (c *Client) GetVersion() (string, error) {
	if !c.IsInstalled() {
		return "", fmt.Errorf("Claude Code is not installed")
	}

	version, err := client.GetExecutableVersion("claude", "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get Claude Code version: %w", err)
	}

	return strings.TrimSpace(version), nil
}

// GetSupportedScopes returns the scopes supported by Claude Code
func (c *Client) GetSupportedScopes() []pkg.ClientScope {
	return []pkg.ClientScope{pkg.LocalScope}
}

// ValidateScope validates if the scope is supported
func (c *Client) ValidateScope(scope string) error {
	if scope != "local" {
		return fmt.Errorf("claude-code only supports 'local' scope, got: %s", scope)
	}
	return nil
}

// GetCurrentConfig reads the current Claude Code configuration
func (c *Client) GetCurrentConfig(scope string) (*pkg.MCPConfig, error) {
	if err := c.ValidateScope(scope); err != nil {
		return nil, err
	}

	// Read from .mcp.json file
	configPath, err := client.ExpandPath(".mcp.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get Claude Code config path: %w", err)
	}

	config := &pkg.MCPConfig{
		Servers: make(map[string]pkg.MCPServerConfig),
	}

	// Try to read existing config
	if err := client.ReadJSONFile(configPath, config); err != nil {
		// If file doesn't exist, return empty config
		// For other errors, return the error
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("failed to read Claude Code config: %w", err)
	}

	return config, nil
}

// ValidateConfig validates the Claude Code configuration
func (c *Client) ValidateConfig(scope string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	// Try to read and parse the config file
	_, err := c.GetCurrentConfig(scope)
	return err
}

// RequiresRestart returns false as Claude Code doesn't need restart
func (c *Client) RequiresRestart() bool {
	return false
}

// TriggerReload triggers Claude Code to reload configuration
func (c *Client) TriggerReload() error {
	// Claude Code typically picks up changes automatically
	return nil
}

// RemoveServer removes a server by updating the configuration file
func (c *Client) RemoveServer(scope string, serverName string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	// Get current config
	config, err := c.GetCurrentConfig(scope)
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// Check if server exists
	if _, exists := config.Servers[serverName]; !exists {
		return fmt.Errorf("server '%s' not found", serverName)
	}

	// Remove the server
	delete(config.Servers, serverName)

	// Write the updated config back to file
	configPath, err := client.ExpandPath(".mcp.json")
	if err != nil {
		return fmt.Errorf("failed to get Claude Code config path: %w", err)
	}

	return client.WriteJSONFile(configPath, config)
}

// ListServers returns the names of configured servers from the config file
func (c *Client) ListServers(scope string) ([]string, error) {
	config, err := c.GetCurrentConfig(scope)
	if err != nil {
		return nil, err
	}

	var servers []string
	for name := range config.Servers {
		servers = append(servers, name)
	}

	return servers, nil
}

// GetLaunchCommand returns the command to launch Claude Code with the given project path
// Claude Code automatically detects and uses devcontainer configurations
func (c *Client) GetLaunchCommand(projectPath string) string {
	// Check if devcontainer config exists
	devcontainerPath := filepath.Join(projectPath, ".devcontainer", "devcontainer.json")
	if client.FileExists(devcontainerPath) {
		// Claude Code can launch directly in devcontainer context
		return fmt.Sprintf("# Launch Claude Code in devcontainer\ncd \"%s\" && claude", projectPath)
	}
	
	// Regular Claude Code launch
	return fmt.Sprintf("# Launch Claude Code\ncd \"%s\" && claude", projectPath)
}

// SupportsDevcontainers returns true if Claude Code supports devcontainers
func (c *Client) SupportsDevcontainers() bool {
	return true
}

// GenerateConfig generates Claude Code MCP configuration from manifests
func (c *Client) GenerateConfig(manifests []pkg.ServoDefinition, secretsProvider func(string) (string, error)) error {
	// Build MCP servers configuration
	servers := make(map[string]pkg.MCPServerConfig)

	for _, manifest := range manifests {
		if manifest.Server.Command != "" {
			serverConfig := pkg.MCPServerConfig{
				Command: manifest.Server.Command,
				Args:    make([]string, len(manifest.Server.Args)),
			}

			// Copy and expand args
			for i, arg := range manifest.Server.Args {
				serverConfig.Args[i] = client.ExpandSecretsInString(arg, secretsProvider)
			}

			// Build environment with secret expansion
			if len(manifest.Server.Environment) > 0 {
				serverConfig.Environment = make(map[string]string)
				for key, value := range manifest.Server.Environment {
					expandedValue := client.ExpandSecretsInString(value, secretsProvider)
					serverConfig.Environment[key] = expandedValue
				}
			}

			servers[manifest.Name] = serverConfig
		}
	}

	// Create MCP configuration
	mcpConfig := &pkg.MCPConfig{
		Servers: servers,
	}

	// Write to .mcp.json (Claude Code format)
	configPath, err := client.ExpandPath(".mcp.json")
	if err != nil {
		return fmt.Errorf("failed to get Claude Code config path: %w", err)
	}

	return client.WriteJSONFile(configPath, mcpConfig)
}

