package cursor

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/servo/servo/internal/client"
	"github.com/servo/servo/pkg"
)

// Client implements MCP (Model Context Protocol) integration for Cursor AI editor.
//
// This client manages MCP server configurations through Cursor's local .cursor/mcp.json
// configuration file. Unlike VS Code, Cursor requires a restart to apply configuration
// changes, so hot-reloading is not supported.
//
// Cursor provides AI-enhanced code editing capabilities and supports the same devcontainer
// workflows as VS Code, making it compatible with Servo's development environment setup.
type Client struct {
	info            client.BaseClientInfo
	localConfigPath string // Custom config path for testing scenarios
}

// New creates a new Cursor client
func New() *Client {
	return NewWithConfigPath("")
}

// NewWithConfigPath creates a new Cursor client with custom config path
func NewWithConfigPath(localPath string) *Client {
	return &Client{
		info: client.BaseClientInfo{
			Name:        "cursor",
			Description: "Cursor AI code editor",
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

// IsInstalled checks if Cursor is installed
func (c *Client) IsInstalled() bool {
	if !c.IsPlatformSupported() {
		return false
	}

	// Check for Cursor executable
	if client.ExecutableExists("cursor") {
		return true
	}

	// Check for common Cursor installation paths
	commonPaths := []string{
		"/Applications/Cursor.app/Contents/Resources/app/bin/cursor",
		"/usr/bin/cursor",
		"/usr/local/bin/cursor",
	}

	for _, path := range commonPaths {
		if client.FileExists(path) {
			return true
		}
	}

	return false
}

// GetVersion returns the Cursor version
func (c *Client) GetVersion() (string, error) {
	if !c.IsInstalled() {
		return "", fmt.Errorf("Cursor is not installed")
	}

	version, err := client.GetExecutableVersion("cursor", "--version")
	if err != nil {
		return "", fmt.Errorf("failed to get Cursor version: %w", err)
	}

	return version, nil
}

// GetSupportedScopes returns the scopes supported by Cursor
func (c *Client) GetSupportedScopes() []pkg.ClientScope {
	return []pkg.ClientScope{pkg.LocalScope}
}

// ValidateScope validates if the scope is supported
func (c *Client) ValidateScope(scope string) error {
	if scope != "local" {
		return fmt.Errorf("cursor only supports 'local' scope, got: %s", scope)
	}
	return nil
}

// GetCurrentConfig reads the current Cursor configuration
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
		return nil, fmt.Errorf("failed to read Cursor config: %w", err)
	}

	return &config, nil
}

// ValidateConfig validates the Cursor configuration
func (c *Client) ValidateConfig(scope string) error {
	if err := c.ValidateScope(scope); err != nil {
		return err
	}

	configPath, err := c.getLocalConfigPath()
	if err != nil {
		return err
	}

	return client.ValidateJSONConfig(configPath)
}

// RequiresRestart returns true as Cursor typically needs restart for config changes
func (c *Client) RequiresRestart() bool {
	return true
}

// TriggerReload attempts to trigger Cursor to reload configuration
func (c *Client) TriggerReload() error {
	// Cursor doesn't have a reliable way to hot reload MCP configs
	return fmt.Errorf("Cursor requires manual restart to apply configuration changes")
}

// getLocalConfigPath returns the local Cursor MCP config path
func (c *Client) getLocalConfigPath() (string, error) {
	if c.localConfigPath != "" {
		return c.localConfigPath, nil
	}
	return client.ExpandPath(".cursor/mcp.json")
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

// GetLaunchCommand returns the command to launch Cursor with the given project path
func (c *Client) GetLaunchCommand(projectPath string) string {
	return fmt.Sprintf("cursor \"%s\"", projectPath)
}

// SupportsDevcontainers returns true if Cursor supports devcontainers
func (c *Client) SupportsDevcontainers() bool {
	return true
}

// GenerateConfig generates Cursor MCP configuration from manifests
func (c *Client) GenerateConfig(manifests []pkg.ServoDefinition, secretsProvider func(string) (string, error)) error {
	// Build MCP servers configuration in Cursor format
	cursorConfig := map[string]interface{}{
		"mcpServers": make(map[string]interface{}),
	}

	mcpServers := cursorConfig["mcpServers"].(map[string]interface{})

	for _, manifest := range manifests {
		if manifest.Server.Command != "" {
			serverData := map[string]interface{}{
				"command": manifest.Server.Command,
			}

			if len(manifest.Server.Args) > 0 {
				args := make([]interface{}, len(manifest.Server.Args))
				for i, arg := range manifest.Server.Args {
					args[i] = c.expandSecrets(arg, secretsProvider)
				}
				serverData["args"] = args
			}

			if len(manifest.Server.Environment) > 0 {
				env := make(map[string]interface{})
				for key, value := range manifest.Server.Environment {
					expandedValue := c.expandSecrets(value, secretsProvider)
					env[key] = expandedValue
				}
				serverData["env"] = env
			}

			mcpServers[manifest.Name] = serverData
		}
	}

	// Write to .cursor/mcp.json
	configPath, err := c.getLocalConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get Cursor config path: %w", err)
	}

	// Ensure .cursor directory exists
	if err := client.EnsureDirectory(filepath.Dir(configPath)); err != nil {
		return fmt.Errorf("failed to create .cursor directory: %w", err)
	}

	return client.WriteJSONFile(configPath, cursorConfig)
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
