package vscode

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/servo/servo/internal/client"
	"github.com/servo/servo/pkg"
)

// Client implements MCP (Model Context Protocol) integration for Visual Studio Code.
//
// This client manages MCP server configurations through VS Code's local .vscode/mcp.json
// settings file. It supports hot-reloading of configurations and integrates with VS Code's
// devcontainer functionality for seamless development workflows.
//
// Configuration files are scoped locally to individual projects, ensuring that MCP
// servers are properly isolated between different development environments.
type Client struct {
	info            client.BaseClientInfo
	localConfigPath string // Custom config path for testing scenarios
}

// New creates a new VS Code client
func New() *Client {
	return NewWithConfigPath("")
}

// NewWithConfigPath creates a new VS Code client with custom config path
func NewWithConfigPath(localPath string) *Client {
	return &Client{
		info: client.BaseClientInfo{
			Name:        "vscode",
			Description: "Visual Studio Code with MCP support",
			Platforms:   []string{"darwin", "linux", "windows"},
		},
		localConfigPath: localPath,
	}
}

func (c *Client) Name() string {
	return c.info.Name
}

func (c *Client) Description() string {
	return c.info.Description
}

func (c *Client) SupportedPlatforms() []string {
	return c.info.Platforms
}

// IsPlatformSupported checks if the current platform is supported
func (c *Client) IsPlatformSupported() bool {
	return client.IsPlatformSupportedForPlatforms(c.info.Platforms)
}

// IsInstalled checks if VS Code is installed
func (c *Client) IsInstalled() bool {
	if !c.IsPlatformSupported() {
		return false
	}

	// Check for VS Code executable
	if client.ExecutableExists("code") {
		return true
	}

	// Check for common VS Code installation paths
	commonPaths := []string{
		"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
		"/usr/bin/code",
		"/usr/local/bin/code",
	}

	for _, path := range commonPaths {
		if client.FileExists(path) {
			return true
		}
	}

	return false
}

// GetVersion returns the VS Code version
func (c *Client) GetVersion() (string, error) {
	if !c.IsInstalled() {
		return "", fmt.Errorf("VS Code is not installed")
	}

	version, err := client.GetExecutableVersion("code", "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get VS Code version: %w", err)
	}

	// Parse version from output (first line is version)
	lines := strings.Split(version, "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "unknown", nil
}

// GetSupportedScopes returns the scopes supported by VS Code
func (c *Client) GetSupportedScopes() []pkg.ClientScope {
	return []pkg.ClientScope{pkg.LocalScope}
}

// ValidateScope validates if the scope is supported
func (c *Client) ValidateScope(scope string) error {
	if scope != "local" {
		return fmt.Errorf("vscode only supports 'local' scope, got: %s", scope)
	}
	return nil
}

// GetCurrentConfig reads the current VS Code configuration
func (c *Client) GetCurrentConfig(scope string) (*pkg.MCPConfig, error) {
	if err := c.ValidateScope(scope); err != nil {
		return nil, err
	}

	configPath, err := c.getLocalConfigPath()
	if err != nil {
		return nil, err
	}

	if !client.FileExists(configPath) {
		// Return empty config if file doesn't exist
		return &pkg.MCPConfig{
			Servers: make(map[string]pkg.MCPServerConfig),
		}, nil
	}

	var config pkg.MCPConfig
	if err := client.ReadJSONFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("failed to read VS Code config: %w", err)
	}

	return &config, nil
}

// ValidateConfig validates the VS Code configuration
func (c *Client) ValidateConfig(scope string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	configPath, err := c.getLocalConfigPath()
	if err != nil {
		return err
	}

	// Missing config file is valid for VS Code (means no MCP servers configured)
	if !client.FileExists(configPath) {
		return nil
	}

	return client.ValidateJSONConfig(configPath)
}

// RequiresRestart returns false as VS Code supports hot reload
func (c *Client) RequiresRestart() bool {
	return false
}

// TriggerReload triggers VS Code to reload the configuration
func (c *Client) TriggerReload() error {
	// VS Code MCP configurations are typically hot-reloaded
	// We could potentially trigger a workspace reload command
	cmd := exec.Command("code", "--command", "workbench.action.reloadWindow")
	return cmd.Run()
}

// getLocalConfigPath returns the local VS Code MCP config path
func (c *Client) getLocalConfigPath() (string, error) {
	if c.localConfigPath != "" {
		return c.localConfigPath, nil
	}
	return client.ExpandPath(".vscode/mcp.json")
}

// RemoveServer removes a server from the configuration
func (c *Client) RemoveServer(scope string, serverName string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	configPath, err := c.getLocalConfigPath()
	if err != nil {
		return err
	}

	// Read current config
	currentConfig, err := c.GetCurrentConfig(scope)
	if err != nil {
		return err
	}

	// Remove server
	if currentConfig.Servers != nil {
		delete(currentConfig.Servers, serverName)
	}

	// Write updated config
	return client.WriteJSONFile(configPath, currentConfig)
}

// ListServers returns the names of configured servers
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

// HasExtension checks if a VS Code extension is installed
func (c *Client) HasExtension(extensionID string) (bool, error) {
	if !c.IsInstalled() {
		return false, fmt.Errorf("VS Code is not installed")
	}

	cmd := exec.Command("code", "--list-extensions")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list VS Code extensions: %w", err)
	}

	extensions := strings.Split(string(output), "\n")
	for _, ext := range extensions {
		if strings.TrimSpace(ext) == extensionID {
			return true, nil
		}
	}

	return false, nil
}

// InstallExtension installs a VS Code extension
func (c *Client) InstallExtension(extensionID string) error {
	if !c.IsInstalled() {
		return fmt.Errorf("VS Code is not installed")
	}

	cmd := exec.Command("code", "--install-extension", extensionID)
	return cmd.Run()
}

// GetLaunchCommand returns the command to launch VS Code with the given project path
func (c *Client) GetLaunchCommand(projectPath string) string {
	return fmt.Sprintf("code \"%s\"", projectPath)
}

// SupportsDevcontainers returns true if VS Code supports devcontainers
func (c *Client) SupportsDevcontainers() bool {
	return true
}

// GenerateConfig generates VSCode MCP configuration from manifests
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

	// Create VSCode MCP configuration using correct format
	vscodeConfig := map[string]interface{}{
		"servers": servers,
	}

	// Write to .vscode/mcp.json
	configPath, err := c.getLocalConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get VSCode config path: %w", err)
	}

	// Ensure .vscode directory exists
	if err := client.EnsureDirectory(filepath.Dir(configPath)); err != nil {
		return fmt.Errorf("failed to create .vscode directory: %w", err)
	}

	return client.WriteJSONFile(configPath, vscodeConfig)
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
