package client

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/servo/servo/internal/utils"
	"github.com/servo/servo/pkg"
)

// BaseClientInfo holds basic client metadata
type BaseClientInfo struct {
	Name        string
	Description string
	Platforms   []string
}

// IsPlatformSupportedForPlatforms checks if the current platform is supported by given platforms list
// Deprecated: Use utils.IsPlatformSupported instead
func IsPlatformSupportedForPlatforms(platforms []string) bool {
	return utils.IsPlatformSupported(platforms)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ExecutableExists checks if an executable exists in PATH
func ExecutableExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetExecutableVersion runs a command to get version information
func GetExecutableVersion(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// ExpandPath expands ~ to home directory and resolves relative paths
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return absPath, nil
}

// EnsureDirectory creates a directory if it doesn't exist
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}

// ReadJSONFile reads and parses a JSON file
func ReadJSONFile(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSONFile writes data to a JSON file
func WriteJSONFile(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := EnsureDirectory(dir); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// MergeServerConfigs merges server configurations
func MergeServerConfigs(global, local map[string]pkg.MCPServerConfig) map[string]pkg.MCPServerConfig {
	merged := make(map[string]pkg.MCPServerConfig)

	// Copy global configs
	for name, config := range global {
		merged[name] = config
	}

	// Override with local configs
	for name, config := range local {
		merged[name] = config
	}

	return merged
}

// ServerConfigsToMCP converts ServerConfig slice to MCPConfig format
// Note: This function works with ServerConfig structs which don't contain
// the actual server command/args. For proper MCP config generation from
// servo files, use the parsing logic in install command's generateMCPServersConfig
func ServerConfigsToMCP(servers []pkg.ServerConfig) map[string]pkg.MCPServerConfig {
	mcpServers := make(map[string]pkg.MCPServerConfig)

	for _, server := range servers {
		// ServerConfig doesn't contain command/args info - this is intentionally
		// a basic conversion. Real MCP configs require parsed servo definitions.
		mcpServers[server.Version] = pkg.MCPServerConfig{
			Command:     "echo",
			Args:        []string{"ServerConfig", "version", server.Version},
			Environment: make(map[string]string),
		}
	}

	return mcpServers
}

// ValidateJSONConfig validates a JSON configuration file
func ValidateJSONConfig(path string) error {
	if !FileExists(path) {
		return fmt.Errorf("config file does not exist: %s", path)
	}

	var config interface{}
	return ReadJSONFile(path, &config)
}

// BackupConfigFile creates a backup of a configuration file
func BackupConfigFile(path string) error {
	if !FileExists(path) {
		return nil // Nothing to backup
	}

	backupPath := path + ".backup"
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return os.WriteFile(backupPath, data, 0644)
}

// RestoreConfigFile restores a configuration file from backup
func RestoreConfigFile(path string) error {
	backupPath := path + ".backup"
	if !FileExists(backupPath) {
		return fmt.Errorf("backup file does not exist: %s", backupPath)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ResolveConfigPath returns the config path, using override if provided, otherwise falling back to default
func ResolveConfigPath(overridePath string, defaultPath string) (string, error) {
	if overridePath != "" {
		return overridePath, nil
	}
	return ExpandPath(defaultPath)
}

// ExpandSecretsInString expands secret placeholders in a string using the provided secrets provider
func ExpandSecretsInString(value string, secretsProvider func(string) (string, error)) string {
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
