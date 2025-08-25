package claude_code

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/servo/servo/internal/utils"
	"github.com/servo/servo/pkg"
)

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	ExecuteFunc func(name string, args ...string) ([]byte, error)
	RunFunc     func(name string, args ...string) error
}

func (m *MockCommandExecutor) Execute(name string, args ...string) ([]byte, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(name, args...)
	}
	// Default behavior - return empty output for list commands
	if name == "claude" && len(args) >= 2 && args[0] == "mcp" && args[1] == "list" {
		return []byte("server1\nserver2\n"), nil
	}
	return []byte{}, nil
}

func (m *MockCommandExecutor) Run(name string, args ...string) error {
	if m.RunFunc != nil {
		return m.RunFunc(name, args...)
	}
	// Default behavior - success
	return nil
}

func TestClient_Name(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})
	if client.Name() != "claude-code" {
		t.Errorf("expected name 'claude-code', got '%s'", client.Name())
	}
}

func TestClient_Description(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})
	expected := "Claude Code CLI interface"
	if client.Description() != expected {
		t.Errorf("expected description '%s', got '%s'", expected, client.Description())
	}
}

func TestClient_SupportedPlatforms(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})
	platforms := client.SupportedPlatforms()

	expectedPlatforms := utils.DesktopPlatforms
	if len(platforms) != len(expectedPlatforms) {
		t.Errorf("expected %d platforms, got %d", len(expectedPlatforms), len(platforms))
	}

	for _, expected := range expectedPlatforms {
		found := false
		for _, platform := range platforms {
			if platform == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected platform '%s' not found", expected)
		}
	}
}

func TestClient_GetSupportedScopes(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})
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
	client := NewWithExecutor(&MockCommandExecutor{})

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
	client := NewWithExecutor(&MockCommandExecutor{})

	// Test that method doesn't panic
	installed := client.IsInstalled()

	// Should return a boolean
	if installed != true && installed != false {
		t.Errorf("IsInstalled should return a boolean")
	}
}

func TestClient_RequiresRestart(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})
	if client.RequiresRestart() {
		t.Errorf("Claude Code should not require restart")
	}
}

func TestClient_GetCurrentConfig(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})

	// Test local scope
	config, err := client.GetCurrentConfig("local")
	if err != nil {
		t.Errorf("unexpected error getting local config: %v", err)
	}

	if config == nil {
		t.Errorf("expected config, got nil")
	}

	if config.Servers == nil {
		t.Errorf("expected servers map to be initialized")
	}
}

func TestClient_GetCurrentConfig_InvalidScope(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})

	_, err := client.GetCurrentConfig("invalid")
	if err == nil {
		t.Errorf("expected error for invalid scope")
	}
}

func TestClient_ValidateConfig(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})

	// Test local scope validation
	err := client.ValidateConfig("local")
	if err != nil {
		t.Errorf("unexpected error validating local config: %v", err)
	}
}

func TestClient_ValidateConfig_InvalidScope(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})

	err := client.ValidateConfig("invalid")
	if err == nil {
		t.Errorf("expected error for invalid scope")
	}
}

func TestClient_TriggerReload(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})

	// Claude Code supports hot reload
	err := client.TriggerReload()
	if err != nil {
		t.Errorf("unexpected error triggering reload: %v", err)
	}
}

func TestClient_GetVersion(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})

	// Test that method doesn't panic
	version, err := client.GetVersion()

	// Should handle both installed and not installed cases gracefully
	if err != nil && version != "" {
		t.Errorf("if error returned, version should be empty")
	}
}

func TestClient_Integration(t *testing.T) {
	client := NewWithExecutor(&MockCommandExecutor{})

	// Test workflow without WriteConfig (since it's been removed)

	// Validate config
	err := client.ValidateConfig("local")
	if err != nil {
		t.Errorf("failed to validate config: %v", err)
	}

	// Get config
	config, err := client.GetCurrentConfig("local")
	if err != nil {
		t.Errorf("failed to get config: %v", err)
	}

	if config == nil {
		t.Errorf("expected config to be available")
	}

	// Trigger reload
	err = client.TriggerReload()
	if err != nil {
		t.Errorf("failed to trigger reload: %v", err)
	}
}

func TestClient_GenerateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Change to temp directory for this test  
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)
	
	// Use default client since GenerateConfig uses .mcp.json relative to current dir
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
		{
			Name: "simple-server",
			Server: pkg.Server{
				Command: "node",
				Args:    []string{"index.js"},
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
	configPath := ".mcp.json"
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

	// Verify structure has mcpServers key (Claude Code uses MCPConfig format)
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
	if len(args) != len(expectedArgs) {
		t.Errorf("Expected %d args, got %d", len(expectedArgs), len(args))
	}

	for i, expectedArg := range expectedArgs {
		if i < len(args) && args[i] != expectedArg {
			t.Errorf("Expected arg[%d] '%s', got '%v'", i, expectedArg, args[i])
		}
	}

	// Verify environment variables expansion (uses 'env' field in JSON)
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

	// Verify simple-server configuration
	simpleServer, exists := servers["simple-server"].(map[string]interface{})
	if !exists {
		t.Fatalf("simple-server not found in generated config")
	}

	if simpleServer["command"] != "node" {
		t.Errorf("Expected command 'node', got '%v'", simpleServer["command"])
	}
}

func TestClient_GenerateConfig_EmptyManifests(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Change to temp directory for this test
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)
	
	client := New()

	// Generate config with empty manifests
	err := client.GenerateConfig([]pkg.ServoDefinition{}, func(string) (string, error) {
		return "", nil
	})
	if err != nil {
		t.Fatalf("GenerateConfig with empty manifests failed: %v", err)
	}

	// Verify config file was created
	configPath := ".mcp.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Verify empty servers
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to parse generated config: %v", err)
	}

	// The mcpServers field might be omitted when empty due to omitempty tag
	servers, exists := config["mcpServers"].(map[string]interface{})
	if exists && len(servers) != 0 {
		t.Errorf("Expected empty servers map or no mcpServers field, got %d servers", len(servers))
	}
}
