package claude_code

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/servo/servo/internal/client"
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
			Platforms:   []string{"darwin", "linux", "windows"},
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
	return client.IsPlatformSupportedForPlatforms(c.info.Platforms)
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

	// Claude Code manages its own configuration internally
	// We use the CLI to query current servers
	servers, err := c.listMCPServers(scope)
	if err != nil {
		return nil, err
	}

	config := &pkg.MCPConfig{
		Servers: make(map[string]pkg.MCPServerConfig),
	}

	for _, server := range servers {
		config.Servers[server] = pkg.MCPServerConfig{
			Command: "managed-by-claude-code",
		}
	}

	return config, nil
}

// ValidateConfig validates the Claude Code configuration
func (c *Client) ValidateConfig(scope string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	// Check if claude mcp commands work
	args := []string{"mcp", "list", "--local"}

	_, err := c.executor.Execute("claude", args...)
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

// listMCPServers lists currently configured MCP servers
func (c *Client) listMCPServers(scope string) ([]string, error) {
	args := []string{"mcp", "list", "--local"}

	output, err := c.executor.Execute("claude", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list MCP servers: %w", err)
	}

	// Parse output to extract server names
	lines := strings.Split(string(output), "\n")
	var servers []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			// Extract server name from output
			parts := strings.Fields(line)
			if len(parts) > 0 {
				servers = append(servers, parts[0])
			}
		}
	}

	return servers, nil
}

// AddServer adds a server using Claude Code CLI
func (c *Client) AddServer(scope string, serverName string, command string, args []string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	cmdArgs := []string{"mcp", "add", serverName, command}
	cmdArgs = append(cmdArgs, args...)

	if scope == "local" {
		cmdArgs = append(cmdArgs, "--local")
	}

	return c.executor.Run("claude", cmdArgs...)
}

// RemoveServer removes a server using Claude Code CLI
func (c *Client) RemoveServer(scope string, serverName string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	args := []string{"mcp", "remove", serverName}
	if scope == "local" {
		args = append(args, "--local")
	}

	return c.executor.Run("claude", args...)
}

// ListServers returns the names of configured servers
func (c *Client) ListServers(scope string) ([]string, error) {
	return c.listMCPServers(scope)
}

// EnableServer enables a server
func (c *Client) EnableServer(scope string, serverName string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	args := []string{"mcp", "enable", serverName}
	if scope == "local" {
		args = append(args, "--local")
	}

	return c.executor.Run("claude", args...)
}

// DisableServer disables a server
func (c *Client) DisableServer(scope string, serverName string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	args := []string{"mcp", "disable", serverName}
	if scope == "local" {
		args = append(args, "--local")
	}

	return c.executor.Run("claude", args...)
}

// GetLaunchCommand returns the command to launch Claude Code with the given project path
func (c *Client) GetLaunchCommand(projectPath string) string {
	return fmt.Sprintf("claude \"%s\"", projectPath)
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
				serverConfig.Args[i] = c.expandSecrets(arg, secretsProvider)
			}

			// Build environment with secret expansion
			if len(manifest.Server.Environment) > 0 {
				serverConfig.Environment = make(map[string]string)
				for key, value := range manifest.Server.Environment {
					expandedValue := c.expandSecrets(value, secretsProvider)
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

// expandSecrets expands secret placeholders in a string
func (c *Client) expandSecrets(value string, secretsProvider func(string) (string, error)) string {
	if !strings.Contains(value, "${") {
		return value
	}

	result := value
	start := strings.Index(result, "${")
	for start != -1 {
		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end += start

		secretName := result[start+2 : end]
		secretValue, err := secretsProvider(secretName)
		if err != nil || secretValue == "" {
			// Keep placeholder if secret not found
			start = strings.Index(result[end+1:], "${")
			if start != -1 {
				start += end + 1
			}
			continue
		}

		result = result[:start] + secretValue + result[end+1:]
		start = strings.Index(result[start+len(secretValue):], "${")
		if start != -1 {
			start += len(secretValue)
		}
	}

	return result
}
