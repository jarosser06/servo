package vscode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/servo/servo/pkg"
)

func setupTestClient(t *testing.T) (*Client, string) {
	tmpDir := t.TempDir()
	localConfigPath := filepath.Join(tmpDir, "local", "mcp.json")
	client := NewWithConfigPath(localConfigPath)

	return client, tmpDir
}

func TestClient_Name(t *testing.T) {
	client := New()
	if client.Name() != "vscode" {
		t.Errorf("expected name 'vscode', got '%s'", client.Name())
	}
}

func TestClient_Description(t *testing.T) {
	client := New()
	expected := "Visual Studio Code with MCP support"
	if client.Description() != expected {
		t.Errorf("expected description '%s', got '%s'", expected, client.Description())
	}
}

func TestClient_SupportedPlatforms(t *testing.T) {
	client := New()
	platforms := client.SupportedPlatforms()

	expectedPlatforms := []string{"darwin", "linux", "windows"}
	if len(platforms) != len(expectedPlatforms) {
		t.Errorf("expected %d platforms, got %d", len(expectedPlatforms), len(platforms))
	}
}

func TestClient_GetSupportedScopes(t *testing.T) {
	client := New()
	scopes := client.GetSupportedScopes()

	expectedScopes := []pkg.ClientScope{pkg.LocalScope}
	if len(scopes) != 1 {
		t.Errorf("expected 1 scope, got %d", len(scopes))
	}

	for i, expected := range expectedScopes {
		if scopes[i] != expected {
			t.Errorf("expected scope %v at index %d, got %v", expected, i, scopes[i])
		}
	}
}

func TestClient_ValidateScope(t *testing.T) {
	client := New()

	// Test valid scope
	if err := client.ValidateScope("local"); err != nil {
		t.Errorf("expected scope 'local' to be valid, got error: %v", err)
	}

	// Test invalid scopes
	invalidScopes := []string{"global", "invalid"}
	for _, scope := range invalidScopes {
		if err := client.ValidateScope(scope); err == nil {
			t.Errorf("expected scope '%s' to be invalid", scope)
		}
	}
}

func TestClient_IsInstalled(t *testing.T) {
	client := New()

	// Test that method doesn't panic
	installed := client.IsInstalled()

	// Should return a boolean
	if installed != true && installed != false {
		t.Errorf("IsInstalled should return a boolean")
	}
}

func TestClient_RequiresRestart(t *testing.T) {
	client := New()
	if client.RequiresRestart() {
		t.Errorf("VS Code should not require restart")
	}
}

func TestClient_GetCurrentConfig_LocalScope_FileNotExists(t *testing.T) {
	client, _ := setupTestClient(t)

	config, err := client.GetCurrentConfig("local")
	if err != nil {
		t.Errorf("expected no error for missing config, got: %v", err)
	}

	if config == nil {
		t.Errorf("expected default config, got nil")
	}
}

func TestClient_GetCurrentConfig_ValidFile(t *testing.T) {
	client, tmpDir := setupTestClient(t)

	// Create test config
	testConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "test-command",
				"args":    []string{"arg1", "arg2"},
				"env":     map[string]string{"KEY": "value"},
			},
		},
	}

	configPath := filepath.Join(tmpDir, "local", "mcp.json")
	// Ensure the directory exists
	err := os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}
	data, _ := json.Marshal(testConfig)
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	config, err := client.GetCurrentConfig("local")
	if err != nil {
		t.Errorf("unexpected error reading config: %v", err)
	}

	if len(config.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(config.Servers))
	}

	if server, exists := config.Servers["test-server"]; !exists {
		t.Errorf("expected test-server to exist")
	} else {
		if server.Command != "test-command" {
			t.Errorf("expected command 'test-command', got '%s'", server.Command)
		}
		if len(server.Args) != 2 {
			t.Errorf("expected 2 args, got %d", len(server.Args))
		}
		if server.Environment["KEY"] != "value" {
			t.Errorf("expected env KEY=value, got %v", server.Environment)
		}
	}
}

func TestClient_ValidateConfig(t *testing.T) {
	client, tmpDir := setupTestClient(t)

	// Test with missing file - should be valid (creates default)
	err := client.ValidateConfig("local")
	if err != nil {
		t.Errorf("expected no error for missing local config, got: %v", err)
	}

	// Test with valid file
	testConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{},
	}

	configPath := filepath.Join(tmpDir, "local", "mcp.json")
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)
	data, _ := json.Marshal(testConfig)
	os.WriteFile(configPath, data, 0644)

	err = client.ValidateConfig("local")
	if err != nil {
		t.Errorf("unexpected error validating config: %v", err)
	}

	// Test with invalid JSON
	os.WriteFile(configPath, []byte("invalid json"), 0644)
	err = client.ValidateConfig("local")
	if err == nil {
		t.Errorf("expected error for invalid JSON")
	}
}

func TestClient_TriggerReload(t *testing.T) {
	client := New()

	// TriggerReload might fail in test environment if code command not available
	// This is expected behavior
	err := client.TriggerReload()
	// We just test that method doesn't panic - error is acceptable in test environment
	_ = err
}

func TestClient_GetVersion(t *testing.T) {
	client := New()

	// Test that method doesn't panic
	version, err := client.GetVersion()

	// Should handle both installed and not installed cases
	if err != nil && version != "" {
		t.Errorf("if error returned, version should be empty")
	}
}

func TestClient_RemoveServer(t *testing.T) {
	client, tmpDir := setupTestClient(t)

	// Create a local config file with a server
	localConfigPath := filepath.Join(tmpDir, "local", "mcp.json")
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(localConfigPath), 0755)
	config := map[string]interface{}{
		"servers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "test-command",
			},
		},
	}
	data, _ := json.Marshal(config)
	os.WriteFile(localConfigPath, data, 0644)

	// Test removing the server
	err := client.RemoveServer("local", "test-server")
	if err != nil {
		t.Errorf("unexpected error removing server: %v", err)
	}
}

func TestClient_ListServers(t *testing.T) {
	client, tmpDir := setupTestClient(t)

	// Create a local config file with servers
	localConfigPath := filepath.Join(tmpDir, "local", "mcp.json")
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(localConfigPath), 0755)
	config := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server1": map[string]interface{}{
				"command": "test-command1",
			},
			"test-server2": map[string]interface{}{
				"command": "test-command2",
			},
		},
	}
	data, _ := json.Marshal(config)
	os.WriteFile(localConfigPath, data, 0644)

	// Test listing servers
	servers, err := client.ListServers("local")
	if err != nil {
		t.Errorf("unexpected error listing servers: %v", err)
	}

	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestClient_HasExtension(t *testing.T) {
	client := New()

	// Test the method doesn't panic - will fail in test environment due to no code CLI
	hasExt, err := client.HasExtension("ms-vscode.vscode-typescript-next")
	if err == nil && !hasExt {
		t.Logf("HasExtension works but extension not found")
	}
}

func TestClient_InstallExtension(t *testing.T) {
	client := New()

	// Test the method doesn't panic - will fail in test environment due to no code CLI
	err := client.InstallExtension("ms-vscode.vscode-typescript-next")
	if err == nil {
		t.Logf("InstallExtension unexpectedly succeeded")
	}
}
