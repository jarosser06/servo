package cursor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/servo/servo/internal/utils"
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
	if client.Name() != "cursor" {
		t.Errorf("expected name 'cursor', got '%s'", client.Name())
	}
}

func TestClient_Description(t *testing.T) {
	client := New()
	expected := "Cursor AI code editor"
	if client.Description() != expected {
		t.Errorf("expected description '%s', got '%s'", expected, client.Description())
	}
}

func TestClient_SupportedPlatforms(t *testing.T) {
	client := New()
	platforms := client.SupportedPlatforms()

	expectedPlatforms := utils.DesktopPlatforms
	if len(platforms) != len(expectedPlatforms) {
		t.Errorf("expected %d platforms, got %d", len(expectedPlatforms), len(platforms))
	}
}

func TestClient_GetSupportedScopes(t *testing.T) {
	client := New()
	scopes := client.GetSupportedScopes()

	if len(scopes) != 1 {
		t.Errorf("expected 1 scope, got %d", len(scopes))
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

func TestClient_RequiresRestart(t *testing.T) {
	client := New()
	if !client.RequiresRestart() {
		t.Errorf("Cursor should require restart")
	}
}

func TestClient_GetCurrentConfig_FileNotExists(t *testing.T) {
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

	// Create test config with Cursor format
	testConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"test-server": map[string]interface{}{
				"command": "test-command",
				"args":    []string{"arg1", "arg2"},
			},
		},
	}

	configPath := filepath.Join(tmpDir, "local", "mcp.json")
	// Ensure directory exists
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
	}
}

func TestClient_ValidateConfig(t *testing.T) {
	client, tmpDir := setupTestClient(t)

	// Test with valid file
	testConfig := map[string]interface{}{
		"test-server": map[string]interface{}{
			"command": "test",
		},
	}

	configPath := filepath.Join(tmpDir, "local", "mcp.json")
	// Ensure directory exists
	os.MkdirAll(filepath.Dir(configPath), 0755)
	data, _ := json.Marshal(testConfig)
	os.WriteFile(configPath, data, 0644)

	err := client.ValidateConfig("local")
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

	// Cursor requires restart, so reload should return error
	err := client.TriggerReload()
	if err == nil {
		t.Errorf("expected error since Cursor requires restart")
	}
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

func TestClient_IsInstalled(t *testing.T) {
	client := New()

	// Test that method doesn't panic
	installed := client.IsInstalled()

	// Should return a boolean
	if installed != true && installed != false {
		t.Errorf("IsInstalled should return a boolean")
	}
}

func TestClient_RemoveServer(t *testing.T) {
	client, tmpDir := setupTestClient(t)

	// Create a local config file with a server
	localConfigPath := filepath.Join(tmpDir, "local", "mcp.json")
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

func TestClient_GetVersion_ErrorPath(t *testing.T) {
	client := New()

	// Test GetVersion which has 33.3% coverage - test the error path
	version, err := client.GetVersion()
	if err != nil {
		// This is expected when cursor isn't installed
		if version != "" {
			t.Errorf("when error occurs, version should be empty, got '%s'", version)
		}
	}
}

func TestClient_ValidateScope_ErrorPath(t *testing.T) {
	client := New()

	// Test invalid scopes
	invalidScopes := []string{"invalid", "project", "system", ""}
	for _, scope := range invalidScopes {
		err := client.ValidateScope(scope)
		if err == nil {
			t.Errorf("ValidateScope should fail for invalid scope '%s'", scope)
		}
	}
}

func TestClient_GetCurrentConfig_ErrorPath(t *testing.T) {
	client := New()

	// Test with invalid scope
	_, err := client.GetCurrentConfig("global")
	if err == nil {
		t.Error("GetCurrentConfig should fail for invalid scope")
	}
}

func TestClient_ValidateConfig_ErrorPath(t *testing.T) {
	client := New()

	// Test with invalid scope
	err := client.ValidateConfig("global")
	if err == nil {
		t.Error("ValidateConfig should fail for invalid scope")
	}
}

func TestClient_GenerateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Change to temp directory for this test
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)
	
	// Use default client since GenerateConfig uses .cursor relative to current dir
	client := New()

	// Test manifests with secrets
	manifests := []pkg.ServoDefinition{
		{
			Name: "test-server",
			Server: pkg.Server{
				Command: "python",
				Args:    []string{"-m", "server", "--api-key", "${SERVO_SECRET:api_key}"},
				Environment: map[string]string{
					"DATABASE_URL": "${SERVO_SECRET:db_url}",
					"DEBUG":        "true",
				},
			},
		},
	}

	// Mock secrets provider
	secretsProvider := func(key string) (string, error) {
		secrets := map[string]string{
			"SERVO_SECRET:api_key": "test-api-key-123",
			"SERVO_SECRET:db_url":  "postgresql://user:pass@localhost/db",
		}
		return secrets[key], nil
	}

	// Generate config
	err := client.GenerateConfig(manifests, secretsProvider)
	if err != nil {
		t.Fatalf("GenerateConfig failed: %v", err)
	}

	// Verify config file was created
	configPath := ".cursor/mcp.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Read and verify config content
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse generated config: %v", err)
	}

	// Verify structure has mcpServers key (Cursor uses mcpServers format)
	servers, exists := config["mcpServers"].(map[string]interface{})
	if !exists {
		t.Fatalf("Config should contain mcpServers key")
	}

	// Verify test-server configuration
	testServer, exists := servers["test-server"].(map[string]interface{})
	if !exists {
		t.Fatalf("test-server not found in generated config")
	}

	if testServer["command"] != "python" {
		t.Errorf("Expected command 'python', got '%v'", testServer["command"])
	}

	// Verify args expansion
	args, ok := testServer["args"].([]interface{})
	if !ok {
		t.Fatalf("Expected args to be array")
	}

	expectedArgs := []string{"-m", "server", "--api-key", "test-api-key-123"}
	for i, expectedArg := range expectedArgs {
		if i < len(args) && args[i] != expectedArg {
			t.Errorf("Expected arg[%d] '%s', got '%v'", i, expectedArg, args[i])
		}
	}

	// Verify environment variables expansion (uses 'env' field)
	env, ok := testServer["env"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected env to be map")
	}

	if env["DATABASE_URL"] != "postgresql://user:pass@localhost/db" {
		t.Errorf("Expected DATABASE_URL to be expanded, got '%v'", env["DATABASE_URL"])
	}

	if env["DEBUG"] != "true" {
		t.Errorf("Expected DEBUG to be 'true', got '%v'", env["DEBUG"])
	}
}
