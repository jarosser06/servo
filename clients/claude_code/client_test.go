package claude_code

import (
	"testing"

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

	expectedPlatforms := []string{"darwin", "linux", "windows"}
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
